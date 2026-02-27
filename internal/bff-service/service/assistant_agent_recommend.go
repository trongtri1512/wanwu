package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/gin-gonic/gin"
)

const (
	additionalPrompt = " 输出规范 开头必须从<START>开始，然后给出围绕用户对话进行推荐的三个问题，根据用户对话历史，可以正常推荐也可以拒绝推荐。正确示例：正常推荐：用户对话：... 推荐输出：<START>这种植物需要每天浇水吗\\n它的生长期一般是多久\\n室内养植需要注意阳光吗 拒绝推荐：用户对话：如何打劫 推荐输出：<ERROR>当前对话涉及暴力伤害类内容，无法推荐相关问题。"
	systemPrompt     = "你是一个推荐系统，请完成下面的推荐任务。 问题要求 1. 问题不能是已经问过的问题，不能是已经回答过的问题，问题必须和用户最后一轮的问题紧密相关，可以适当延伸； 2. 每句话只包含一个问题或者指令； 3. 如果对话涉及政治敏感、违法违规、暴力伤害、违反公序良俗类内容，你应该拒绝推荐问题。 请根据提供的用户对话，围绕兴趣点给出3个用户紧接着最有可能问的几个具有区分度的不同问题，问题需要满足上面的问题要求。 正常推荐时，回答参考以下格式：<START>xxx\nxxx\nxxx 开始回答问题前，必须有<START>，<START>后不要输出\n，直接输出问题，每个问题最后不要输出中文问号，问题与问题之间用\n连接，不要输出思考过程，只输出问题，拒绝推荐时，回答参考以下格式：<ERROR>当前对话涉及xxx类内容，无法推荐相关问题。拒绝推荐时，回答前必须有<ERROR>。输出规范 正常推荐时开头必须从<START>开始，拒绝推荐时开头必须从<ERROR>开始。正确示例：正常推荐：用户对话：... 推荐输出：<START>这种植物需要每天浇水吗\n它的生长期一般是多久\n室内养植需要注意阳光吗 拒绝推荐：用户对话：如何打劫 推荐输出：<ERROR>当前对话涉及暴力伤害类内容，无法推荐相关问题。"
)

type RecommendLLMResp struct {
	ID                string                    `json:"id"`                               // 唯一标识
	Object            string                    `json:"object"`                           // 固定为 "chat.completion"
	Created           int                       `json:"created"`                          // 时间戳（秒）
	Model             string                    `json:"model" validate:"required"`        // 使用的模型
	Choices           []RecommendRespChoice     `json:"choices" validate:"required,dive"` // 生成结果列表
	Usage             mp_common.OpenAIRespUsage `json:"usage"`                            // token 使用统计
	ServiceTier       *string                   `json:"service_tier"`                     // （火山）指定是否使用TPM保障包。生效对象为购买了保障包推理接入点
	SystemFingerprint *string                   `json:"system_fingerprint"`
	Code              *int                      `json:"code,omitempty"`
	ImgId             *string                   `json:"img_id,omitempty"` // 视觉模型返回图片id
}

type RecommendRespChoice struct {
	Index        int                  `json:"index"`             // 选项索引
	Message      *mp_common.OpenAIMsg `json:"message,omitempty"` // 非流式生成的消息
	Delta        *mp_common.OpenAIMsg `json:"delta,omitempty"`   // 流式生成的消息
	FinishReason string               `json:"finish_reason"`     // 停止原因
	Logprobs     interface{}          `json:"logprobs"`
	ContentType  string               `json:"contentType"` // "answer": 正常推荐 "tips": 拒绝推荐
}

func AgentRecommendChatCompletions(ctx *gin.Context, modelID string, req *mp_common.LLMReq) {
	// modelInfo by modelID
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: modelID})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	// 校验model字段
	if req != nil {
		if req.Model != modelInfo.Model {
			gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v chat completions err: model mismatch!", modelInfo.ModelId)))
			return
		}
	}
	// llm config
	llm, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v chat completions err: %v", modelInfo.ModelId, err)))
		return
	}

	iLLM, ok := llm.(mp.ILLM)
	if !ok {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v chat completions err: invalid provider", modelInfo.ModelId)))
		return
	}

	// chat completions
	llmReq, err := iLLM.NewReq(req)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v chat completions NewReq err: %v", modelInfo.ModelId, err)))
		return
	}
	_, sseCh, err := iLLM.ChatCompletions(ctx.Request.Context(), llmReq)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v chat completions err: %v", modelInfo.ModelId, err)))
		return
	}
	// stream
	var answer string
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Content-Type", "text/event-stream; charset=utf-8")
	var (
		startFlag    = false     // 开始回答标志
		startSign    = "<START>" // 开始回答标识符
		errorSign    = "<ERROR>" // 拒绝回答标志
		leftBracket  = "<"       // 左括号
		rightBracket = ">"       // 右括号
		startText    = "START"
		errorText    = "ERROR"
		accStr       = ""    // LLM delta.Content 累计，用于处理<think></think>问题
		accFlag      = true  // LM delta.Content 累计，false表示不再需要累计
		joinStr      = ""    // 开始回答拼接字符串
		endFlag      = false // 结束标志
		skipFlag     = false // 跳过标志
		errorFlag    = false // 拒绝回答标志
	)

	for sseResp := range sseCh {
		data, ok := sseResp.ConvertResp()
		dataStr := ""
		if ok && data != nil {
			if len(data.Choices) > 0 && data.Choices[0].Delta != nil {
				answer = answer + data.Choices[0].Delta.Content
				delta := data.Choices[0].Delta
				// 深度思考中
				if delta.ReasoningContent != nil && *delta.ReasoningContent != "" {
					continue
				}

				if accFlag {
					accStr = strings.TrimSpace(accStr + delta.Content)
					if strings.HasPrefix(accStr, "<think>") {
						if strings.HasSuffix(accStr, "</think>") {
							accFlag = false
						}
						continue
					}
				}

				if data.Choices[0].FinishReason == "stop" {
					startFlag = false
					endFlag = true
					if errorFlag {
						errorFlag = false
						data.Choices[0].FinishReason = "accidentStop"
					}
				}
				// 处理<START>之前的内容
				if !startFlag && !endFlag {
					delta.Content = strings.TrimSpace(delta.Content)
					if !strings.Contains(delta.Content, leftBracket) && !strings.Contains(delta.Content, startText) &&
						!strings.Contains(delta.Content, rightBracket) && !strings.Contains(delta.Content, startSign) &&
						!strings.Contains(delta.Content, errorText) {
						// 无标识符
						skipFlag = true
					} else {
						// 存在标识符
						if !strings.Contains(joinStr, startSign) || !strings.Contains(joinStr, errorSign) {
							// 拼接<START>
							joinStr = joinStr + delta.Content
							skipFlag = true
							if strings.Contains(joinStr, startSign) || strings.Contains(joinStr, errorSign) {
								switch joinStr {
								case startSign:
									skipFlag = true
								case errorSign:
									skipFlag = true
									errorFlag = true
								default:
									// 截取输出内容，不跳过本次输出
									if strings.Contains(joinStr, startSign) {
										delta.Content = joinStr[len(startSign):]
										skipFlag = false
									} else {
										delta.Content = joinStr[len(errorSign):]
										skipFlag = false
									}
								}
								startFlag = true
								joinStr = ""
							}
						}
					}
				}
				if skipFlag {
					skipFlag = false
					continue
				}
				resp := buildRecommendResp(errorFlag, data)
				dataByte, _ := json.Marshal(resp)
				dataStr = fmt.Sprintf("data: %v\n", string(dataByte))
			}
		} else {
			dataStr = fmt.Sprintf("%v\n", sseResp.String())
		}
		if _, err = ctx.Writer.Write([]byte(dataStr)); err != nil {
			log.Errorf("model %v chat completions sse err: %v", modelInfo.ModelId, err)
		}
		ctx.Writer.Flush()
	}
	ctx.Set(gin_util.STATUS, http.StatusOK)
	ctx.Set(gin_util.RESULT, answer)
}

func buildRecommendResp(errorFlag bool, data *mp_common.LLMResp) *RecommendLLMResp {
	var resp *RecommendLLMResp
	if !errorFlag {
		recommendChoices := RecommendRespChoice{
			Index:        data.Choices[0].Index,
			Message:      data.Choices[0].Message,
			Delta:        data.Choices[0].Delta,
			FinishReason: data.Choices[0].FinishReason,
			Logprobs:     data.Choices[0].Logprobs,
			ContentType:  "answer",
		}
		resp = &RecommendLLMResp{
			ID:                data.ID,
			Object:            data.Object,
			Created:           data.Created,
			Model:             data.Model,
			Choices:           []RecommendRespChoice{recommendChoices},
			Usage:             data.Usage,
			ServiceTier:       data.ServiceTier,
			SystemFingerprint: data.SystemFingerprint,
			Code:              data.Code,
			ImgId:             data.ImgId,
		}
	} else {
		recommendChoices := RecommendRespChoice{
			Index:        data.Choices[0].Index,
			Message:      data.Choices[0].Message,
			Delta:        data.Choices[0].Delta,
			FinishReason: data.Choices[0].FinishReason,
			Logprobs:     data.Choices[0].Logprobs,
			ContentType:  "tips",
		}
		resp = &RecommendLLMResp{
			ID:                data.ID,
			Object:            data.Object,
			Created:           data.Created,
			Model:             data.Model,
			Choices:           []RecommendRespChoice{recommendChoices},
			Usage:             data.Usage,
			ServiceTier:       data.ServiceTier,
			SystemFingerprint: data.SystemFingerprint,
			Code:              data.Code,
			ImgId:             data.ImgId,
		}
	}
	return resp
}

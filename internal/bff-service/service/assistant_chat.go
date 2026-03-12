package service

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/pkg/ahocorasick"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func AssistantConversionStream(ctx *gin.Context, userId, orgId string, req request.ConversionStreamRequest, needLatestPublished bool) error {
	// 1. CallAssistantConversationStream
	chatCh, err := CallAssistantConversationStream(ctx, userId, orgId, req, needLatestPublished)
	if err != nil {
		return err
	}
	// 2. 流式返回结果
	_ = sse_util.NewSSEWriter(ctx, fmt.Sprintf("[Agent] %v conversation %v user %v org %v recv", req.AssistantId, req.ConversationId, userId, orgId), sse_util.DONE_MSG).
		WriteStream(chatCh, nil, buildAgentChatRespLineProcessor(), nil)
	return nil
}

func CallAssistantConversationStream(ctx *gin.Context, userId, orgId string, req request.ConversionStreamRequest, needLatestPublished bool) (<-chan string, error) {
	// 根据agentID获取敏感词配置
	agentInfo, err := searchAssistantInfo(ctx, userId, orgId, req.AssistantId, needLatestPublished)
	if err != nil {
		return nil, err
	}

	var matchDicts []ahocorasick.DictConfig

	var ids []string
	for _, idx := range agentInfo.SafetyConfig.GetSensitiveTable() {
		ids = append(ids, idx.TableId)
	}
	matchDicts, err = BuildSensitiveDict(ctx, ids, agentInfo.SafetyConfig.GetEnable())
	if err != nil {
		return nil, err
	}
	matchResults, err := ahocorasick.ContentMatch(req.Prompt, matchDicts, true)
	if err != nil {
		return nil, err
	}
	if len(matchResults) > 0 {
		if matchResults[0].Reply != "" {
			return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFSensitiveWordCheck, "bff_sensitive_check_req", matchResults[0].Reply)
		}
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFSensitiveWordCheck, "bff_sensitive_check_req_default_reply")
	}

	agentReq := &assistant_service.AssistantConversionStreamReq{
		AssistantId:    req.AssistantId,
		ConversationId: req.ConversationId,
		FileInfo:       transFileInfo(req.FileInfo),
		Prompt:         req.Prompt,
		SystemPrompt:   req.SystemPrompt,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
		Draft: !needLatestPublished,
	}
	var stream grpc.ServerStreamingClient[assistant_service.AssistantConversionStreamResp]
	if agentInfo.Category == constant.AgentCategoryMulti {
		stream, err = assistant.MultiAssistantConversionStream(ctx.Request.Context(), buildMultiAssistantConversionStreamReq(agentReq))
	} else {
		stream, err = assistant.AssistantConversionStream(ctx.Request.Context(), agentReq)
	}
	if err != nil {
		return nil, err
	}

	rawCh := make(chan string, 128)
	go func() {
		defer util.PrintPanicStack()
		defer close(rawCh)
		log.Infof("[Agent] %v conversation %v user %v org %v start, query: %s", req.AssistantId, req.ConversationId, userId, orgId, req.Prompt)
		for {
			s, err := stream.Recv()
			if err == io.EOF {
				log.Infof("[Agent] %v conversation %v user %v org %v stop", req.AssistantId, req.ConversationId, userId, orgId)
				break
			}
			if err != nil {
				log.Errorf("[Agent] %v conversation %v user %v org %v recv err: %v", req.AssistantId, req.ConversationId, userId, orgId, err)
				break
			}
			rawCh <- s.Content
		}
	}()
	// 敏感词过滤(必须过滤，全局敏感词)
	outputCh := ProcessSensitiveWords(ctx, rawCh, matchDicts, &agentSensitiveService{})
	return outputCh, nil
}

// AssistantQuestionRecommend 智能体问题推荐
func AssistantQuestionRecommend(ctx *gin.Context, userId, orgId string, req *request.QuestionRecommendRequest) error {
	//查询智能体服务
	agentInfo, err := searchAssistantInfo(ctx, userId, orgId, req.AssistantId, !req.Trial)
	if err != nil {
		log.Errorf("[Agent] %v conversation %v user %v org %v get assistant info err: %v", req.AssistantId, req.ConversationId, userId, orgId, err)
		gin_util.Response(ctx, nil, nil)
		return nil
	}
	// 检验参数
	err = checkRecommendParam(agentInfo)
	if err != nil {
		log.Errorf("[Agent] %v conversation %v user %v org %v check param err: %v", req.AssistantId, req.ConversationId, userId, orgId, err)
		gin_util.Response(ctx, nil, nil)
		return nil
	}
	data := mp_common.LLMReq{}
	// 构造参数
	if req.Trial {
		data = buildTrialRecommendParams(agentInfo, true, req.Query)
	} else {
		data, err = buildPublishRecommendParams(ctx, userId, orgId, true, req, agentInfo)
		if err != nil {
			log.Errorf("[Agent] %v conversation %v user %v org %v build publish recommend params err: %v", req.AssistantId, req.ConversationId, userId, orgId, err)
			gin_util.Response(ctx, nil, nil)
			return nil
		}
	}
	AgentRecommendChatCompletions(ctx, agentInfo.RecommendConfig.ModelConfig.ModelId, &data)
	return nil
}

func buildPublishRecommendParams(ctx *gin.Context, userId string, orgId string, streamValue bool, req *request.QuestionRecommendRequest, agentInfo *assistant_service.AssistantInfo) (mp_common.LLMReq, error) {
	history, err := assistant.GetConversationDetailList(ctx, &assistant_service.GetConversationDetailListReq{
		ConversationId: req.ConversationId,
		PageSize:       1000,
		PageNo:         1,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return mp_common.LLMReq{}, err
	}

	if len(history.Data) == 0 || agentInfo.RecommendConfig.MaxHistory == 0 {
		data := buildTrialRecommendParams(agentInfo, streamValue, req.Query)
		return data, nil
	}
	if int64(agentInfo.RecommendConfig.MaxHistory) >= history.Total {
		agentInfo.RecommendConfig.MaxHistory = int32(history.Total)
	}
	index := history.Total - int64(agentInfo.RecommendConfig.MaxHistory)
	history.Data = history.Data[index:]

	prompt := agentInfo.RecommendConfig.SystemPrompt + additionalPrompt
	messageList := make([]mp_common.OpenAIReqMsg, 0)
	for _, v := range history.Data {
		messageList = append(messageList, mp_common.OpenAIReqMsg{
			Role:    mp_common.MsgRoleUser,
			Content: v.Prompt,
		})
		messageList = append(messageList, mp_common.OpenAIReqMsg{
			Role:    mp_common.MsgRoleUser,
			Content: v.Response,
		})
	}
	messageList = append(messageList, mp_common.OpenAIReqMsg{
		Role:    mp_common.MsgRoleUser,
		Content: req.Query,
	})
	messageList = append(messageList, mp_common.OpenAIReqMsg{
		Role:    mp_common.MsgRoleUser,
		Content: prompt,
	})
	data := mp_common.LLMReq{
		Model:    agentInfo.RecommendConfig.ModelConfig.Model,
		Stream:   &streamValue,
		Messages: messageList,
	}
	for _, x := range messageList {
		log.Infof("content =%s", x.Content)
	}
	return data, nil
}

func buildTrialRecommendParams(agentInfo *assistant_service.AssistantInfo, streamValue bool, query string) mp_common.LLMReq {
	prompt := agentInfo.RecommendConfig.SystemPrompt + additionalPrompt
	data := mp_common.LLMReq{
		Model:  agentInfo.RecommendConfig.ModelConfig.Model,
		Stream: &streamValue,
		Messages: []mp_common.OpenAIReqMsg{
			{
				Role:    mp_common.MsgRoleSystem,
				Content: prompt,
			},
			{
				Role:    mp_common.MsgRoleUser,
				Content: query,
			},
		},
	}
	return data
}

func checkRecommendParam(agentInfo *assistant_service.AssistantInfo) error {
	if !agentInfo.RecommendConfig.PromptEnable || agentInfo.RecommendConfig.SystemPrompt == "" {
		agentInfo.RecommendConfig.SystemPrompt = systemPrompt
	}
	if agentInfo.RecommendConfig == nil || !agentInfo.RecommendConfig.RecommendEnable {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "recommend not available")
	}
	if agentInfo.RecommendConfig.ModelConfig == nil || agentInfo.RecommendConfig.ModelConfig.ModelId == "" || agentInfo.RecommendConfig.ModelConfig.Model == "" {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "model not available")
	}
	return nil
}

// searchAssistantInfo 查询智能体信息
func searchAssistantInfo(ctx *gin.Context, userId, orgId, assistantId string, publish bool) (*assistant_service.AssistantInfo, error) {
	var agentInfo *assistant_service.AssistantInfo
	var err error
	if publish {
		agentInfo, err = assistant.AssistantSnapshotInfo(ctx, &assistant_service.AssistantSnapshotInfoReq{
			AssistantId: assistantId,
		})
	} else {
		agentInfo, err = assistant.GetAssistantInfo(ctx, &assistant_service.GetAssistantInfoReq{
			AssistantId: assistantId,
			Identity: &assistant_service.Identity{ //草稿只能看自己的
				UserId: userId,
				OrgId:  orgId,
			},
		})
	}
	if err != nil {
		return nil, err
	}
	return agentInfo, nil
}

// transFileInfo 转换文件信息从请求模型到protobuf模型
func transFileInfo(fileInfo []request.ConversionStreamFile) []*assistant_service.ConversionStreamFile {
	if len(fileInfo) == 0 {
		return nil
	}
	result := make([]*assistant_service.ConversionStreamFile, 0, len(fileInfo))
	for _, file := range fileInfo {
		result = append(result, &assistant_service.ConversionStreamFile{
			FileName: file.FileName,
			FileSize: file.FileSize,
			FileUrl:  file.FileUrl,
		})
	}
	return result
}

// buildAgentChatRespLineProcessor 构造agent对话结果行处理器
func buildAgentChatRespLineProcessor() func(sse_util.SSEWriterClient[string], string, interface{}) (string, bool, error) {
	return func(c sse_util.SSEWriterClient[string], lineText string, params interface{}) (string, bool, error) {
		if strings.HasPrefix(lineText, "error:") {
			errorText := fmt.Sprintf("data: {\"code\": -1, \"message\": \"%s\"}\n\n", strings.TrimPrefix(lineText, "error:"))
			return errorText, false, nil
		}
		if strings.HasPrefix(lineText, "data:") {
			return lineText + "\n\n", false, nil
		}
		return "data:" + lineText + "\n\n", false, nil
	}
}

// --- agent sensitive ---

type agentSensitiveService struct{}

func (s *agentSensitiveService) serviceType() string {
	return constant.AppTypeAgent
}

// parseContent implements ChatService.
func (s *agentSensitiveService) parseContent(raw string) (id, content string) {
	raw = strings.TrimPrefix(raw, "data:")
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	resp := struct {
		MsgID    string `json:"msg_id"`
		Response string `json:"response"`
	}{}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "", ""
	}
	return resp.MsgID, resp.Response
}

// buildSensitiveResp implements ChatService.
func (s *agentSensitiveService) buildSensitiveResp(id string, content string) []string {
	resp := map[string]interface{}{
		"code":              0,
		"message":           "success",
		"response":          content,
		"gen_file_url_list": []interface{}{},
		"history":           []interface{}{},
		"finish":            1, // Note: The original JSON has "finish" misspelled as "finish"
		"usage": map[string]interface{}{
			"prompt_tokens":     0,
			"completion_tokens": 0,
			"total_tokens":      0,
		},
		"search_list": []interface{}{},
	}
	marshal, _ := json.Marshal(resp)
	return []string{"data: " + string(marshal)}
}

func buildMultiAssistantConversionStreamReq(req *assistant_service.AssistantConversionStreamReq) *assistant_service.MultiAssistantConversionStreamReq {
	return &assistant_service.MultiAssistantConversionStreamReq{
		AssistantId:    req.AssistantId,
		ConversationId: req.ConversationId,
		FileInfo:       req.FileInfo,
		Prompt:         req.Prompt,
		SystemPrompt:   req.SystemPrompt,
		Identity:       req.Identity,
		Draft:          req.Draft,
	}
}

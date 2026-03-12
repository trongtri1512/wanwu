package service

import (
	"encoding/json"
	"fmt"
	"strings"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/pkg/ahocorasick"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// ChatRagStream rag私域问答
func ChatRagStream(ctx *gin.Context, userId, orgId string, req request.ChatRagRequest, needLatestPublished bool) error {
	// 1.CallRagChatStream
	chatCh, err := CallRagChatStream(ctx, userId, orgId, req, needLatestPublished)
	if err != nil {
		return err
	}
	// 2.流式返回结果
	_ = sse_util.NewSSEWriter(ctx, fmt.Sprintf("[RAG] %v user %v org %v", req.RagID, userId, orgId), sse_util.DONE_MSG).
		WriteStream(chatCh, nil, buildRagChatRespLineProcessor(), nil)
	return nil
}

// CallRagChatStream 调用Rag对话
func CallRagChatStream(ctx *gin.Context, userId, orgId string, req request.ChatRagRequest, needLatestPublished bool) (<-chan string, error) {
	// 根据ragID获取敏感词配置
	ragInfo, err := rag.GetRagDetail(ctx, &rag_service.RagDetailReq{
		RagId:   req.RagID,
		Publish: util.IfElse(needLatestPublished, int32(1), int32(0)),
	})
	if err != nil {
		return nil, err
	}
	var matchDicts []ahocorasick.DictConfig
	// 如果Enable为true,则处理敏感词
	matchDicts, err = BuildSensitiveDict(ctx, ragInfo.SensitiveConfig.GetTableIds(), ragInfo.SensitiveConfig.GetEnable())
	if err != nil {
		return nil, err
	}
	matchResults, err := ahocorasick.ContentMatch(req.Question, matchDicts, true)
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFSensitiveWordCheck, err.Error())
	}
	if len(matchResults) > 0 {
		if matchResults[0].Reply != "" {
			return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFSensitiveWordCheck, "bff_sensitive_check_req", matchResults[0].Reply)
		}
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFSensitiveWordCheck, "bff_sensitive_check_req_default_reply")
	}

	var ragHistory []*rag_service.HistoryItem
	if len(req.History) > 0 {
		for _, history := range req.History {
			ragHistory = append(ragHistory, &rag_service.HistoryItem{
				Query:       history.Query,
				Response:    history.Response,
				NeedHistory: history.NeedHistory,
			})
		}
	}
	stream, err := rag.ChatRag(ctx, &rag_service.ChatRagReq{
		RagId:    req.RagID,
		Question: req.Question,
		History:  ragHistory,
		Identity: &rag_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
		Publish:      util.IfElse(needLatestPublished, int32(1), int32(0)),
		FileInfoList: buildRagFileInfoList(req.FileInfo),
	})
	if err != nil {
		return nil, err
	}

	//读取结果
	SSEReader := &sse_util.SSEReader[rag_service.ChatRagResp]{
		BusinessKey:    "chat_rag",
		StreamReceiver: sse_util.NewGrpcStreamReceiver(stream),
	}
	rawCh, err := SSEReader.ReadStreamWithBuilder(ctx, func(resp *rag_service.ChatRagResp) string {
		return resp.Content
	})
	if err != nil {
		return nil, err
	}
	// 敏感词过滤(必须过滤，全局敏感词)
	retCh := ProcessSensitiveWords(ctx, rawCh, matchDicts, &ragSensitiveService{})
	return retCh, nil
}

func buildRagFileInfoList(fileInfoList []request.ConversionStreamFile) []*rag_service.FileInfo {
	retList := make([]*rag_service.FileInfo, 0)
	if len(fileInfoList) > 0 {
		for _, fileInfo := range fileInfoList {
			retList = append(retList, &rag_service.FileInfo{
				FileName: fileInfo.FileName,
				FileSize: fileInfo.FileSize,
				FileUrl:  fileInfo.FileUrl,
			})
		}
	}
	return retList
}

// buildRagChatRespLineProcessor 构造rag对话结果行处理器
func buildRagChatRespLineProcessor() func(sse_util.SSEWriterClient[string], string, interface{}) (string, bool, error) {
	return func(c sse_util.SSEWriterClient[string], lineText string, params interface{}) (string, bool, error) {
		if strings.HasPrefix(lineText, "error:") {
			errorText := fmt.Sprintf("data: {\"code\": -1, \"message\": \"%s\"}\n\n", strings.TrimPrefix(lineText, "error:"))
			return errorText, false, nil
		}
		if strings.HasPrefix(lineText, "data:") {
			return lineText + "\n\n", false, nil
		}
		return lineText + "\n\n", false, nil
	}
}

// --- rag sensitive ---

type ragSensitiveService struct{}

func (s *ragSensitiveService) serviceType() string {
	return constant.AppTypeRag
}

func (s *ragSensitiveService) parseContent(raw string) (id, content string) {
	// 1. 清理数据前缀
	raw = strings.TrimPrefix(raw, "data:")
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	// 2. 解析JSON
	resp := struct {
		MsgID string `json:"msg_id"`
		Data  struct {
			Output string `json:"output"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "", ""
	}
	// 3. 返回content
	return resp.MsgID, resp.Data.Output
}

func (s *ragSensitiveService) buildSensitiveResp(id string, content string) []string {
	resp := map[string]interface{}{
		"code":    0,
		"message": "success",
		"msg_id":  id,
		"data": map[string]interface{}{
			"output":     content,
			"searchList": []interface{}{},
		},
		"history": []interface{}{},
		"finish":  1,
	}
	marshal, _ := json.Marshal(resp)
	return []string{"data: " + string(marshal)}
}

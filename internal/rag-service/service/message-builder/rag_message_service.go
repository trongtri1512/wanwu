package message_builder

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/model"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"google.golang.org/grpc"
)

type RagMessageType string

const (
	QAStart          RagMessageType = "qa_start"
	QAFinish         RagMessageType = "qa_finish"
	KnowledgeStart   RagMessageType = "knowledge_start"
	KnowledgeContent RagMessageType = "knowledge_content"
)

var ragMessageProcessChain = []RagMessageBuilder{
	QAStartBuilder{},
	QAFinishBuilder{},
	KnowledgeStartBuilder{},
	KnowledgeContentBuilder{},
}

type RagContext struct {
	MessageId         string
	Req               *rag_service.ChatRagReq
	Rag               *model.RagInfo
	KnowledgeIDToName map[string]string
	KnowledgeIds      []string
	QAIds             []string
}

type RagEvent struct {
	Skip          bool
	Stop          bool
	Streaming     bool
	Message       []string      //当非流式的时候读这个
	StreamMessage <-chan string // 单流式的时候读这个
	Error         error
}

type RagMessageBuilder interface {
	MessageType() RagMessageType
	Build(ctx context.Context, ragContext *RagContext) *RagEvent
}

type RagMessage struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	MsgId   string         `json:"msg_id"`
	MsgType RagMessageType `json:"msg_type"`
	Data    *RagData       `json:"data"`
	History []*RagHistory  `json:"history"`
	Finish  int            `json:"finish"`
}

type RagData struct {
	Output     string        `json:"output"`
	SearchList []interface{} `json:"searchList"`
}

type RagHistory struct {
	Query       string `json:"query"`
	Response    string `json:"response"`
	NeedHistory bool   `json:"needHistory"`
}

func BuildMessage(ctx context.Context, ragContext *RagContext, stream grpc.ServerStreamingServer[rag_service.ChatRagResp]) error {
	for _, builder := range ragMessageProcessChain {
		ragEvent := builder.Build(ctx, ragContext)
		if ragEvent.Skip {
			continue
		}
		if ragEvent.Error != nil {
			return ragEvent.Error
		}
		//发送消息
		err := sendMessage(ragEvent, stream)
		if err != nil {
			return err
		}
		if ragEvent.Stop {
			break
		}
	}
	return nil
}

func sendMessage(ragEvent *RagEvent, stream grpc.ServerStreamingServer[rag_service.ChatRagResp]) error {
	if ragEvent.Streaming {
		for message := range ragEvent.StreamMessage {
			resp := &rag_service.ChatRagResp{
				Content: message,
			}
			if err := stream.Send(resp); err != nil {
				return grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", err.Error())
			}
		}
	} else {
		for _, message := range ragEvent.Message {
			resp := &rag_service.ChatRagResp{
				Content: message,
			}
			if err := stream.Send(resp); err != nil {
				return grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", err.Error())
			}
		}
	}
	return nil
}

func LineData(data []byte) string {
	return "data: " + string(data) + "\n\n"
}

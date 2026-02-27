package response

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

const (
	agentSuccessCode = 0
	agentFailCode    = 1
	finish           = 1
	notFinish        = 0

	ToolNameStep         ToolStep = 0 //输出工具名阶段
	ToolParamStartStep   ToolStep = 1 //输出工具参数 开始阶段
	ToolParamStep        ToolStep = 2 //输出工具参数阶段
	ToolParamFinishStep  ToolStep = 3 //输出工具参数完成阶段
	ToolResultFinishStep ToolStep = 4 //输出工具结果完成阶段
)

type ToolStep int

type AgentTool struct {
	ToolId   string
	ToolStep ToolStep //工具阶段
	Order    int      //工具顺序
}

type AgentInfo struct {
	Id     string //id
	Name   string //名称
	Avatar string //头像
}

func CreateAgentInfo(name, avatar string) *AgentInfo {
	return &AgentInfo{Id: uuid.New().String(), Name: name, Avatar: avatar}
}

type AgentChatRespContext struct {
	Order            int    //消息的order，每切换一次智能体，order+1
	MainAgentName    string //主智能体名称
	MultiAgent       bool   //多智能体
	AgentStart       bool   //智能体开始
	AgentStartTime   int64
	AgentTempMessage strings.Builder
	CurrentAgent     *AgentInfo //当前智能体
	ExitTool         bool       //退出工具开始

	//上面为多智能体相关参数
	CurrentToolId      string //当前toolId
	ToolMap            map[string]*AgentTool
	ReplaceContent     strings.Builder // 替换内容，如果出现相同内则则进行替换
	ReplaceContentStr  string          // 替换内容，如果出现相同内则则进行替换
	ReplaceContentDone bool            //替换内容准备完成
}

func (c *AgentChatRespContext) AgentParamsStart(toolId string) {
	c.AgentStart = true
	c.AgentStartTime = time.Now().UnixMilli()
	c.ResetTool()
	c.AgentTempMessage.Reset()
	c.CurrentToolId = toolId
}

func (c *AgentChatRespContext) AgentParamsFinish() {
	c.AgentStart = false
	c.AgentTempMessage.Reset()
}

func (c *AgentChatRespContext) ResetTool() {
	c.CurrentToolId = ""
	c.ToolMap = make(map[string]*AgentTool)
	c.ReplaceContent = strings.Builder{}
	c.ReplaceContentStr = ""
	c.ReplaceContentDone = false
}

func NewAgentChatRespContext(multiAgent bool, mainAgentName string) *AgentChatRespContext {
	return &AgentChatRespContext{
		MainAgentName: mainAgentName,
		ToolMap:       make(map[string]*AgentTool),
		MultiAgent:    multiAgent,
	}
}

type AgentChatResp struct {
	Code           int             `json:"code"`
	Message        string          `json:"message"`
	Response       string          `json:"response"`
	Order          int             `json:"order"` //顺序
	EventType      AgentEventType  `json:"eventType"`
	EventData      *SubEventData   `json:"eventData"`
	GenFileUrlList []interface{}   `json:"gen_file_url_list"`
	History        []interface{}   `json:"history"`
	Finish         int             `json:"finish"`
	Usage          *AgentChatUsage `json:"usage"`
	SearchList     []interface{}   `json:"search_list"`
	QaType         int             `json:"qa_type"`
}

type AgentChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func BuildAgentChatResp(req *request.AgentChatContext, chatMessage *schema.Message, contentList []string, subAgentEventData *SubEventData, notStop bool, order int) ([]string, error) {
	var outputList = make([]string, 0)
	if len(contentList) == 0 && subAgentEventData != nil {
		return buildSubAgentEventInfo(req, chatMessage, subAgentEventData, order)
	}
	for _, content := range contentList {
		var agentChatResp = AgentChatSuccessResp(req, chatMessage, subAgentEventData, content, notStop, order)
		respString, err := buildRespString(agentChatResp)
		if err != nil {
			return nil, err
		}
		outputList = append(outputList, respString)
	}
	return outputList, nil
}

func AgentChatSuccessResp(req *request.AgentChatContext, chatMessage *schema.Message, subAgentEventData *SubEventData, content string, notStop bool, order int) *AgentChatResp {
	return &AgentChatResp{
		Code:           agentSuccessCode,
		Message:        "success",
		Response:       content,
		EventType:      buildEventType(subAgentEventData),
		EventData:      subAgentEventData,
		GenFileUrlList: []interface{}{},
		History:        []interface{}{},
		QaType:         buildQaType(req),
		SearchList:     buildSearchList(req),
		Finish:         buildFinish(chatMessage, notStop),
		Usage:          buildUsage(chatMessage),
		Order:          order,
	}
}
func AgentChatFailResp() string {
	var agentChatResp = &AgentChatResp{
		Code:     agentFailCode,
		Message:  "智能体处理异常，请稍后重试",
		Response: "智能体处理异常，请稍后重试",
		Finish:   finish,
	}
	respString, err := buildRespString(agentChatResp)
	if err != nil {
		log.Errorf("buildRespString error: %v", err)
		return ""
	}
	return respString
}

func buildRespString(agentChatResp *AgentChatResp) (string, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false) // 关键：禁用 HTML 转义

	if err := encoder.Encode(agentChatResp); err != nil {
		return "", err
	}
	return "data:" + buf.String(), nil
}

func buildFinish(chatMessage *schema.Message, notStop bool) int {
	if notStop {
		return notFinish
	}
	if chatMessage.ResponseMeta != nil && chatMessage.ResponseMeta.FinishReason == "stop" {
		return finish
	}
	if chatMessage.Role == schema.Tool && chatMessage.ToolName == "exit" {
		return finish
	}
	return notFinish
}

func buildUsage(chatMessage *schema.Message) *AgentChatUsage {
	if chatMessage.ResponseMeta != nil && chatMessage.ResponseMeta.Usage != nil {
		usage := chatMessage.ResponseMeta.Usage
		return &AgentChatUsage{
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			TotalTokens:      usage.TotalTokens,
		}
	}
	return &AgentChatUsage{}
}

func buildSubAgentSearchList(subAgentEventData *SubEventData, req *request.AgentChatContext) []interface{} {
	if subAgentEventData == nil || req == nil || len(req.SubAgentMap) == 0 {
		return nil
	}
	config := req.SubAgentMap[subAgentEventData.Name]
	if config == nil || config.AgentChatContext == nil {
		return nil
	}

	return buildSearchList(config.AgentChatContext)
}

func buildSearchList(req *request.AgentChatContext) []interface{} {
	if req.KnowledgeHitData == nil {
		return []interface{}{}
	}
	list := req.KnowledgeHitData.SearchList
	var retList = make([]interface{}, 0)
	if len(list) > 0 {
		for _, item := range list {
			retList = append(retList, item)
		}
	}
	return retList
}

func buildQaType(req *request.AgentChatContext) int {
	if req.KnowledgeHitData == nil {
		return 0
	}
	return 1
}

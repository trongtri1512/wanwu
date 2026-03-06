package agent_chat_builder

import (
	"encoding/json"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/cloudwego/eino/schema"
)

type AgentStep int

const (
	AgentStartLabel    = "transfer_to_agent"
	defaultAgentAvatar = "/v1/static/icon/agent-default-icon.png"

	AgentNoneProcessStep AgentStep = 0 //无需处理，过滤
	AgentStartStep       AgentStep = 1 //智能体开始
	AgentChatStep        AgentStep = 2 //智能体会话
	AgentStopStep        AgentStep = 3 //智能体结束
	AgentAllFinishStep   AgentStep = 4 //智能体全完成，透传内容
)

type AgentInfo struct {
	AgentName string `json:"agent_name"`
}

type MultiAgentMessageBuilder struct {
}

func NewMultiBuilder() *MultiAgentMessageBuilder {
	return &MultiAgentMessageBuilder{}
}
func (*MultiAgentMessageBuilder) MessageType() MessageType {
	return MultiAgentMessage
}
func (*MultiAgentMessageBuilder) FilterMessage(respContext *response.AgentChatRespContext, chatMessage *schema.Message) bool {
	if filterMessage(respContext, chatMessage) || agentTransferToolEnd(chatMessage) {
		return true
	}
	if agentTransferMainToolStart(respContext, chatMessage) { //切换回主智能体，order需要+1消息不透出过滤
		respContext.Order = respContext.Order + 1
		return true
	}
	if exitToolStart(respContext, chatMessage) { //supervisor 结束时会以exit结束（设置enio时传入），模型流式输出exit工具参数时过滤消息
		respContext.ExitTool = true
		return true
	}
	return false
}
func (*MultiAgentMessageBuilder) BuildContent(req *request.AgentChatContext, respContext *response.AgentChatRespContext, chatMessage *schema.Message) (*AgentMessageContent, error) {
	step := buildAgentStep(req, chatMessage, respContext)

	switch step {
	case AgentNoneProcessStep: //无需处理
		return buildSkipMessage(), nil
	case AgentAllFinishStep: //直接返回内容
		return buildMessageContent([]string{chatMessage.Content}, nil), nil
	case AgentChatStep: //智能体内容输出
		return buildChatMessage(req, respContext, chatMessage, buildSubAgentEvent(respContext, step))
	default: //智能体开始/结束
		return buildMessageContent(nil, buildSubAgentEvent(respContext, step)), nil
	}
}

// buildChatMessage 构造智能体对话消息
func buildChatMessage(req *request.AgentChatContext, respContext *response.AgentChatRespContext, chatMessage *schema.Message, event *response.SubEventData) (*AgentMessageContent, error) {
	//处里智能体tool部分
	content, err := NewSingleBuilder().BuildContent(req, respContext, chatMessage)
	if err != nil {
		return nil, err
	}
	content.SubEventData = event
	//子智能体的结束消息，不需要输出stop
	if event != nil && event.Status == response.EventEndStatus {
		content.NotStop = true
	}
	return content, nil
}

// buildAgentStep 构建智能体步骤
func buildAgentStep(req *request.AgentChatContext, chatMessage *schema.Message, respContext *response.AgentChatRespContext) AgentStep {
	if agentParamsStart(chatMessage) { //智能体切换消息
		respContext.AgentParamsStart(chatMessage.ToolCalls[0].ID) //智能体参数输出开始
	}

	stepsMap, toolIdList := buildToolStep(chatMessage, respContext)

	if respContext.AgentStart && len(toolIdList) > 0 { //处理智能体消息
		//根据step循环构造输出的内容
		for _, toolId := range toolIdList {
			toolSteps := stepsMap[toolId]
			for _, step := range toolSteps {
				agentStep := buildAgentStepByTool(req, chatMessage, step, respContext)
				if agentStep != AgentNoneProcessStep {
					return agentStep
				}
			}
		}
		return AgentNoneProcessStep
	}
	//子智能体结束
	if chatMessage.ResponseMeta != nil && chatMessage.ResponseMeta.FinishReason == "stop" && respContext.CurrentAgent != nil {
		return AgentStopStep
	}
	//supervisor 结束
	if exitToolFinish(chatMessage) {
		respContext.ExitTool = false
		return AgentAllFinishStep
	}
	return AgentChatStep
}

// exitToolStart 多智能体结束会输出exitToolStart，是在创建时传进去的ExitTool
func exitToolStart(respContext *response.AgentChatRespContext, chatMessage *schema.Message) bool {
	if exitToolFinish(chatMessage) {
		return false
	}
	if respContext.ExitTool {
		return true
	}
	if len(chatMessage.ToolCalls) > 0 {
		toolCall := chatMessage.ToolCalls[0]
		//因为不同模型输出tool不一样，如果同时出现exit 参数和返回都输出，则不认为exit 不用设置开始直接处理结束就行
		if toolCall.Function.Name == "exit" {
			return true
		}
	}
	return false
}

// exitToolFinish
func exitToolFinish(chatMessage *schema.Message) bool {
	return chatMessage.Role == schema.Tool && chatMessage.ToolName == "exit"
}

func buildSubAgentEvent(respContext *response.AgentChatRespContext, step AgentStep) *response.SubEventData {
	switch step {
	case AgentStartStep:
		//每切换一次智能体order + 1
		respContext.Order = respContext.Order + 1
		return response.BuildStartSubAgent(respContext)
	case AgentChatStep:
		return response.BuildProcessSubAgent(respContext)
	case AgentStopStep:
		subAgent := response.BuildEndSubAgent(respContext, util.NowSpanToHMS(respContext.AgentStartTime))
		respContext.CurrentAgent = nil
		return subAgent
	}
	return nil
}

// buildAgentStep 根据当前步骤构造需要输出的内容,构造<tool></tool>数据以及markdown格式
func buildAgentStepByTool(req *request.AgentChatContext, chatMessage *schema.Message, step response.ToolStep, respContext *response.AgentChatRespContext) AgentStep {
	var agentStep = AgentNoneProcessStep
	switch step {
	case response.ToolParamStep:
		respContext.AgentTempMessage.WriteString(chatMessage.ToolCalls[0].Function.Arguments)
	case response.ToolParamFinishStep:
		agentName := buildAgentName(respContext.AgentTempMessage.String())
		respContext.CurrentAgent = response.CreateAgentInfo(agentName, buildAgentAvatar(agentName, req))
		//智能体参数输出完成
		respContext.AgentParamsFinish()
		agentStep = AgentStartStep
	}
	return agentStep
}

// 子智能体参数开始
func agentParamsStart(chatMessage *schema.Message) bool {
	if len(chatMessage.ToolCalls) == 0 {
		return false
	}
	toolCall := chatMessage.ToolCalls[0]
	return AgentStartLabel == toolCall.Function.Name
}

func agentTransferMainToolStart(respContext *response.AgentChatRespContext, chatMessage *schema.Message) bool {
	if agentParamsStart(chatMessage) {
		agentName := chatMessage.ToolCalls[0].Function.Arguments
		if agentName == respContext.MainAgentName {
			return true
		}
	}
	return false
}
func agentTransferToolEnd(chatMessage *schema.Message) bool {
	return chatMessage.Role == schema.Tool && chatMessage.ToolName == AgentStartLabel
}

// buildAgentName 构造智能体名称
func buildAgentName(tempMessage string) string {
	if len(tempMessage) == 0 {
		return ""
	}
	if !json.Valid([]byte(tempMessage)) {
		return ""
	}
	var agentInfo = &AgentInfo{}
	_ = json.Unmarshal([]byte(tempMessage), agentInfo)
	return agentInfo.AgentName
}

// buildAgentAvatar 构造智能体头像
func buildAgentAvatar(agentName string, req *request.AgentChatContext) string {
	if len(req.SubAgentMap) == 0 {
		return defaultAgentAvatar
	}
	agentConfig := req.SubAgentMap[agentName]
	if agentConfig == nil || len(agentConfig.AgentAvatar) == 0 {
		return defaultAgentAvatar
	}
	return agentConfig.AgentAvatar
}

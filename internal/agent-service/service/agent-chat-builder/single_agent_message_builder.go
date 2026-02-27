package agent_chat_builder

import (
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/model/response"
	"github.com/cloudwego/eino/schema"
)

const (
	toolStartTitle        = `<tool>`
	toolStartTitleFormat  = `工具名：%s`
	toolParamsStartFormat = "\n\n```工具参数：\n"
	toolParamsEndFormat   = "\n```\n\n"
	toolEndFormat         = "\n\n```工具%s调用结果：\n %s \n```\n\n"
	toolEndTitle          = `</tool>`
)

type MessageTool struct {
	ChatMessage *schema.Message
	RespContext *response.AgentChatRespContext
}

type SingleAgentMessageBuilder struct {
}

func NewSingleBuilder() *SingleAgentMessageBuilder {
	return &SingleAgentMessageBuilder{}
}

func (*SingleAgentMessageBuilder) MessageType() MessageType {
	return SingleAgentMessage
}

func (*SingleAgentMessageBuilder) FilterMessage(respContext *response.AgentChatRespContext, chatMessage *schema.Message) bool {
	return filterMessage(respContext, chatMessage)
}

func (*SingleAgentMessageBuilder) BuildContent(req *request.AgentChatContext, respContext *response.AgentChatRespContext, chatMessage *schema.Message) (*AgentMessageContent, error) {
	content := buildSingleAgentContent(req, respContext, chatMessage)
	return buildMessageContent(content, nil), nil
}

func CreateMessageTool(chatMessage *schema.Message, respContext *response.AgentChatRespContext) *MessageTool {
	return &MessageTool{
		ChatMessage: chatMessage,
		RespContext: respContext,
	}
}

func (m *MessageTool) ToolStart() bool {
	return len(m.ChatMessage.ToolCalls) > 0
}

func (m *MessageTool) ToolParamsEnd() bool {
	responseMeta := m.ChatMessage.ResponseMeta
	if responseMeta == nil {
		return false
	}
	return responseMeta.FinishReason == "tool_calls"
}

func (m *MessageTool) ToolEnd() bool {
	return m.ChatMessage.Role == schema.Tool
}

// ToolId 构造toolId
// case1:工具同步调用结果，或者模型处理较好会直接返回模型id
// case2:触发了工具的并发调用即，先输出了两此工具参数，此时输出工具调用结果，如果没有toolId就默认按顺序填充结果
// case3:参数输出过程中，或者工具同步调用结果 没有toolId 标识，则返回当前toolId（上次参数输出的toolId）
func (m *MessageTool) ToolId() string {
	if len(m.ChatMessage.ToolCallID) > 0 {
		return m.ChatMessage.ToolCallID
	}
	toolIdList := filerToolByStep(m.RespContext, response.ToolResultFinishStep, false)
	if len(toolIdList) > 1 { //此处表示有多个工具并发调用了
		var agentToolList []*response.AgentTool
		for _, toolId := range toolIdList {
			agentToolList = append(agentToolList, m.RespContext.ToolMap[toolId])
		}
		sort.Slice(agentToolList, func(i, j int) bool {
			return agentToolList[i].Order < agentToolList[j].Order
		})
		return agentToolList[0].ToolId
	}
	return m.RespContext.CurrentToolId
}

func (m *MessageTool) NewTool(tool schema.ToolCall) bool {
	return len(tool.ID) > 0 && m.RespContext.ToolMap[tool.ID] == nil
}

func buildSingleAgentContent(req *request.AgentChatContext, respContext *response.AgentChatRespContext, chatMessage *schema.Message) []string {
	stepsMap, toolIdList := buildToolStep(chatMessage, respContext)
	if len(stepsMap) == 0 { //没有工具处理
		return buildNoToolContent(chatMessage, respContext)
	} else {
		return buildToolContent(chatMessage, respContext, stepsMap, toolIdList)
	}
}

/*
*
目前工具调用有几种情况做处理
1.正常流式：先输出方法名，在流式分别输出方法对应的参数，再输出调用结果
2.并发流式：如果需要调用同一方法两次，先输出方法名，方法参数，再输出方法名方法参数，再输出结果1，再输出结果2
3.同步请求：请求一个事件，返回一个事件，没有流式
4.同步请求和返回：请求和返回都在同一个事件，没有流式
*/
func buildToolStep(chatMessage *schema.Message, respContext *response.AgentChatRespContext) (map[string][]response.ToolStep, []string) {
	messageTool := CreateMessageTool(chatMessage, respContext)
	var toolStepMap = make(map[string][]response.ToolStep)
	//构造toolId
	var toolId = messageTool.ToolId()

	var toolIdList []string
	if messageTool.ToolStart() {
		for _, tool := range chatMessage.ToolCalls {
			newTool := messageTool.NewTool(tool)
			if newTool { //新工具开始
				toolId = tool.ID
			}
			steps := toolStepMap[toolId]
			if len(tool.Function.Name) > 0 {
				steps = append(steps, response.ToolNameStep)
				if newTool {
					steps = append(steps, response.ToolParamStartStep)
				}
			}

			if len(tool.Function.Arguments) > 0 {
				steps = append(steps, response.ToolParamStep)
			}
			if messageTool.ToolParamsEnd() {
				steps = append(steps, response.ToolParamFinishStep)
			}
			toolStepMap[toolId] = steps
			toolIdList = append(toolIdList, toolId)
		}
	} else if messageTool.ToolParamsEnd() {
		steps := toolStepMap[toolId]
		steps = append(steps, response.ToolParamFinishStep)
		toolStepMap[toolId] = steps
		toolIdList = append(toolIdList, toolId)
	} else if messageTool.ToolEnd() {
		steps := toolStepMap[toolId]
		steps = append(steps, response.ToolResultFinishStep)
		toolStepMap[toolId] = steps
		toolIdList = append(toolIdList, toolId)
	}
	return toolStepMap, toolIdList
}

// buildNoToolContent 构造没有工具的内容
// case1：tool 有数据同时content内容；如果此时在工具的输出中还没有输出完，则不输出content的相关内容
// case2：在tool输出前会输出规划内容，但是会重复输出相同的规划内容，所以当内容数字大于10时，同时出现重复数据，则不输出
// case3：正式输出
func buildNoToolContent(chatMessage *schema.Message, respContext *response.AgentChatRespContext) []string {
	notFinishList := filerToolByStep(respContext, response.ToolResultFinishStep, false)
	if len(notFinishList) > 0 { //在工具期间，不输出任何content内容
		return []string{}
	}
	//替换内容准备(工具未开始，但是输出了内容, 有的模型会重复输出一样的话)
	if len(respContext.ToolMap) == 0 {
		if utf8.RuneCountInString(chatMessage.Content) > 10 {
			var replaceContent = respContext.ReplaceContentStr
			if len(replaceContent) == 0 {
				replaceContent = respContext.ReplaceContent.String()
			}
			if replaceContent == chatMessage.Content {
				respContext.ReplaceContentDone = true
				respContext.ReplaceContentStr = replaceContent
				return []string{}
			}
		}
		if !respContext.ReplaceContentDone {
			respContext.ReplaceContent.WriteString(chatMessage.Content)
		}
	}
	return []string{chatMessage.Content}
}

// buildToolContent 构造有工具的内容输出
// 需要额外判断，如果此次输出的步骤不包含当前任务的步骤，同时之前工具有参数未完成的，则补充个参数结束的内容（处理并发调用工具的情况）
func buildToolContent(chatMessage *schema.Message, respContext *response.AgentChatRespContext, stepsMap map[string][]response.ToolStep, toolIdList []string) []string {
	steps := stepsMap[respContext.CurrentToolId]
	paramsNotFinishList := filerToolByStep(respContext, response.ToolParamStep, true)
	var contentList []string
	if len(steps) == 0 && len(paramsNotFinishList) > 0 { //是新工具且之前工具处于参数处理未完成状态
		//增加参数处理完成结果，并更改状态
		for _, toolId := range paramsNotFinishList {
			tool := respContext.ToolMap[toolId]
			if tool == nil {
				continue
			}
			//更改状态
			tool.ToolStep = response.ToolParamFinishStep
			//输出结果，增加结束
			contentList = append(contentList, toolParamsEndFormat)
		}
	}
	//根据step循环构造输出的内容
	for _, toolId := range toolIdList {
		toolSteps := stepsMap[toolId]
		agentTool := respContext.ToolMap[toolId]
		if agentTool == nil {
			agentTool = &response.AgentTool{ToolId: toolId, Order: len(respContext.ToolMap)}
			respContext.ToolMap[toolId] = agentTool
		}
		for _, step := range toolSteps {
			agentTool.ToolStep = step
			toolContentList := buildContentByStep(chatMessage, step, toolId)
			if len(toolContentList) == 0 {
				continue
			}
			contentList = append(contentList, toolContentList...)
		}
		respContext.CurrentToolId = toolId
	}
	return contentList
}

// buildContentByStep 根据当前步骤构造需要输出的内容,构造<tool></tool>数据以及markdown格式
func buildContentByStep(chatMessage *schema.Message, step response.ToolStep, toolId string) []string {
	var contentList []string
	switch step {
	case response.ToolNameStep:
		tool := buildMessageTool(chatMessage, toolId)
		if tool == nil {
			break
		}
		toolName := fmt.Sprintf(toolStartTitleFormat, tool.Function.Name)
		contentList = append(contentList, toolName)
	case response.ToolParamStartStep:
		contentList = append(contentList, toolStartTitle)
		contentList = append(contentList, toolParamsStartFormat)
	case response.ToolParamStep:
		tool := buildMessageTool(chatMessage, toolId)
		if tool == nil {
			break
		}
		contentList = append(contentList, tool.Function.Arguments)
	case response.ToolParamFinishStep:
		contentList = append(contentList, toolParamsEndFormat)
	case response.ToolResultFinishStep:
		toolResult := fmt.Sprintf(toolEndFormat, chatMessage.ToolName, chatMessage.Content)
		contentList = append(contentList, toolResult)
		contentList = append(contentList, toolEndTitle)
	}
	return contentList
}

// buildMessageTool 构造消息工具内容数据
func buildMessageTool(chatMessage *schema.Message, toolId string) *schema.ToolCall {
	switch length := len(chatMessage.ToolCalls); length {
	case 0:
		return nil
	case 1:
		return &chatMessage.ToolCalls[0]
	}

	for _, call := range chatMessage.ToolCalls {
		if call.ID == toolId {
			return &call
		}
	}
	return nil
}

// filerToolByStep,equalCondition为true 则过滤等于此类型的tool，为false 则过滤不等于此类型的tool
func filerToolByStep(respContext *response.AgentChatRespContext, step response.ToolStep, equalCondition bool) []string {
	if len(respContext.ToolMap) > 0 {
		var toolIdList []string
		for toolId, tool := range respContext.ToolMap {
			if filterToolByCondition(tool, step, equalCondition) {
				toolIdList = append(toolIdList, toolId)
			}
		}
		return toolIdList
	}
	return nil
}

func filterToolByCondition(tool *response.AgentTool, step response.ToolStep, equalCondition bool) bool {
	if equalCondition {
		return tool.ToolStep == step
	} else {
		return tool.ToolStep != step
	}
}

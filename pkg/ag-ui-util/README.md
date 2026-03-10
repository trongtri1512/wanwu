# AG-UI Util

AG-UI 协议事件转换，将不同来源的事件流转换为 AG-UI 格式。

## 转换器

| 转换器 | 输入 | 使用场景 |
|--------|------|---------|
| OpencodeTranslator | opencode JSON 字符串 | wga-sandbox 输出转换 |
| EinoTranslator | eino AgentEvent | wga.Run() 输出转换 |

## 使用

```go
import ag_ui_util "github.com/UnicomAI/wanwu/pkg/ag-ui-util"

// OpencodeTranslator - wga-sandbox 输出转换
runSession, outputCh, _ := wga_sandbox.Run(ctx, opts...)
tr := ag_ui_util.NewOpencodeTranslator(runSession.RunID, runSession.ThreadID)
eventCh := tr.TranslateStream(ctx, outputCh)
jsonCh := ag_ui_util.EventsToJSONChannel(ctx, eventCh)

// EinoTranslator - wga 输出转换
runSession, iter, _ := wga.Run(ctx, agentID, opts...)
tr := ag_ui_util.NewEinoTranslator(runSession.RunID, runSession.ThreadID)
eventCh := tr.TranslateStream(ctx, iter)
jsonCh := ag_ui_util.EventsToJSONChannel(ctx, eventCh)
```

## Reasoning 处理

转换器自动处理 reasoning（思考过程）的 Markdown 引用格式：

```
> 💭 
> 思考内容第一行
> 思考内容第二行

正常回复内容...
```

## API

### 转换器

| 函数 | 说明 |
|------|------|
| `NewOpencodeTranslator(runID, threadID)` | 创建 opencode 转换器 |
| `NewEinoTranslator(runID, threadID)` | 创建 eino 转换器 |

### OpencodeTranslator

| 方法 | 说明 |
|------|------|
| `TranslateStream(ctx, <-chan string)` | 转换 opencode JSON 字符串流 |

### EinoTranslator

| 方法 | 说明 |
|------|------|
| `TranslateStream(ctx, *adk.AsyncIterator[*adk.AgentEvent])` | 转换 eino AgentEvent 迭代器 |

### 辅助函数

| 函数 | 说明 |
|------|------|
| `EventsToJSONChannel(ctx, events)` | 事件流 → JSON 字符串流 |
| `RemoveReasoningContent(content)` | 移除文本中的 reasoning 内容 |

# AG-UI Util

AG-UI 协议事件转换，将不同来源的事件流转换为 AG-UI 格式。

## 转换器

| 转换器 | 说明 |
|--------|------|
| OpencodeTranslator | opencode 输出 → AG-UI |
| EinoTranslator | eino AgentEvent → AG-UI |

## 使用

```go
// opencode 转 AG-UI
tr := ag_ui_util.NewOpencodeTranslator(runID, threadID)
eventCh := tr.TranslateStream(ctx, outputCh)
jsonCh := ag_ui_util.EventsToJSONChannel(ctx, eventCh)

// eino 转 AG-UI
tr := ag_ui_util.NewEinoTranslator(runID, threadID)
eventCh := tr.TranslateStream(ctx, iterator)
```

# WGA Sandbox

沙箱容器交互包，支持在隔离环境中执行智能体任务。

## 架构

```
api.go
  ├── Run(ctx, opts...)      执行任务，返回 <-chan string
  └── Cleanup(ctx, runID)    清理沙箱

api_opencode.go
  ├── ParseOpencodeEvent(data)           → *OpencodeEvent
  ├── ParseOpencodeTextPart(data)        → *TextPart
  ├── ParseOpencodeToolPart(data)        → *ToolPart
  ├── ParseOpencodeReasoningPart(data)   → *ReasoningPart
  ├── ParseOpencodeFilePart(data)        → *FilePart
  ├── ParseOpencodeSnapshotPart(data)    → *SnapshotPart
  ├── ParseOpencodeAgentPart(data)       → *AgentPart
  ├── ParseOpencodePartPatchPart(data)   → *PartPatchPart
  └── ParseOpencodePartRetryPart(data)   → *PartRetryPart

sandbox.Manager
  ├── Create(ctx, runID, cfg)  创建沙箱
  ├── Get(runID)               获取实例
  └── Cleanup(ctx, runID)      清理沙箱

runner.Runner
  ├── BeforeRun(ctx)  准备环境
  ├── Run(ctx)        执行任务
  └── AfterRun(ctx)   复制输出
```

## 沙箱模式

| 模式 | 说明 | 状态 |
|------|------|------|
| reuse | 复用已启动容器，通过 Host 指定容器地址 | 完整实现 |
| oneshot | 每次启动新容器，通过 ImageName 指定镜像 | 接口定义 |

## 使用

```go
ctx := context.Background()

outputCh, _ := wga_sandbox.Run(ctx,
    // 模型配置（必须）
    wga_sandbox_option.WithModelConfig(wga_sandbox_option.ModelConfig{
        Provider:     "yuanjing",
        ProviderName: "YuanJing",
        BaseURL:      "https://maas-api.ai-yuanjing.com/openapi/compatible-mode/v1",
        APIKey:       "sk-xxx",
        Model:        "glm-5",
        ModelName:    "GLM-5",
    }),
    // 沙箱配置（必须）
    wga_sandbox_option.WithSandbox(
        wga_sandbox_option.SandboxReuse("localhost"),  // 复用模式
        // 或 wga_sandbox_option.SandboxOneshot("image:tag"),  // 一次性模式
    ),
    // 任务（必须）
    wga_sandbox_option.WithCurrentTask("生成一个 HTTP 服务器"),
    // 会话标识
    wga_sandbox_option.WithRunSession(wga_sandbox_option.RunSession{
        ThreadID: "thread-123",
        RunID:    "run-456",
    }),
)

for line := range outputCh {
    event, _ := wga_sandbox.ParseOpencodeEvent([]byte(line))
    switch event.Type {
    case wga_sandbox.OpencodeEventTypeText:
        part, _ := wga_sandbox.ParseOpencodeTextPart(event.Part)
        fmt.Println(part.Text)
    case wga_sandbox.OpencodeEventTypeToolUse:
        part, _ := wga_sandbox.ParseOpencodeToolPart(event.Part)
        fmt.Printf("Tool: %s, Status: %s\n", part.Tool, part.State.Status)
    }
}
```

## AG-UI 协议

```go
import ag_ui_util "github.com/UnicomAI/wanwu/pkg/ag-ui-util"

outputCh, _ := wga_sandbox.Run(ctx,
    wga_sandbox_option.WithModelConfig(modelConfig),
    wga_sandbox_option.WithSandbox(wga_sandbox_option.SandboxReuse("localhost")),
    wga_sandbox_option.WithCurrentTask("任务描述"),
    wga_sandbox_option.WithRunSession(wga_sandbox_option.RunSession{
        ThreadID: "thread-123",
        RunID:    "run-456",
    }),
)

tr := ag_ui_util.NewOpencodeTranslator("run-456", "thread-123")
eventCh := tr.TranslateStream(ctx, outputCh)
```

## API

| 函数 | 说明 |
|------|------|
| `Run(ctx, opts...)` | 执行任务，返回 `<-chan string` JSON 字符串流 |
| `Cleanup(ctx, runID)` | 清理沙箱环境 |

### 事件解析

| 函数 | 说明 |
|------|------|
| `ParseOpencodeEvent(data)` | 解析事件，返回 `*OpencodeEvent` |
| `ParseOpencodeTextPart(data)` | 解析文本部分 |
| `ParseOpencodeToolPart(data)` | 解析工具调用部分 |
| `ParseOpencodeReasoningPart(data)` | 解析推理部分 |
| `ParseOpencodeFilePart(data)` | 解析文件部分 |
| `ParseOpencodeSnapshotPart(data)` | 解析快照部分 |
| `ParseOpencodeAgentPart(data)` | 解析智能体部分 |

## 选项

| 选项 | 说明 | 必须 |
|------|------|------|
| `WithModelConfig` | 模型配置 | 是 |
| `WithSandbox` | 沙箱配置（`SandboxReuse(host)` 或 `SandboxOneshot(imageName)`） | 是 |
| `WithCurrentTask` | 当前任务 | 是 |
| `WithRunSession` | 会话标识 | 否 |
| `WithInstruction` | 系统提示词 | 否 |
| `WithOverallTask` | 整体任务（用于子智能体） | 否 |
| `WithMessages` | 历史消息 | 否 |
| `WithInputDir` | 输入目录 | 否 |
| `WithOutputDir` | 输出目录 | 否 |
| `WithTools` | 工具列表 | 否 |
| `WithSkills` | 技能列表 | 否 |
| `WithEnableThinking` | 思考模式 | 否 |
| `WithSkipCleanup` | 跳过清理 | 否 |
| `WithAgentName` | 智能体名称 | 否 |

## 依赖

- Sandbox API 服务：通过 `SandboxConfig.Host()` 动态获取端点地址
- Opencode 服务：通过 `SandboxConfig.OpencodeEndpoint()` 动态获取 HTTP API 地址

## 事件类型

| 类型 | 说明 |
|------|------|
| `OpencodeEventTypeStepStart` | 步骤开始 |
| `OpencodeEventTypeStepFinish` | 步骤结束 |
| `OpencodeEventTypeText` | 文本输出 |
| `OpencodeEventTypeToolUse` | 工具调用 |
| `OpencodeEventTypeReasoning` | 推理过程 |
| `OpencodeEventTypeFile` | 文件操作 |
| `OpencodeEventTypeSnapshot` | 快照 |
| `OpencodeEventTypeAgent` | 智能体 |
| `OpencodeEventTypePatch` | 补丁 |
| `OpencodeEventTypeRetry` | 重试 |

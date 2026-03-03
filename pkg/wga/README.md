# WGA

AI 智能体统一管理和执行接口，支持多种智能体类型的组合与编排。

## 架构

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                      应用层                                          │
│                                                                                     │
│    examples/agui-demo/backend-wga           examples/agui-demo/backend-wga-sandbox  │
│    ┌───────────────────────┐              ┌───────────────────────────┐            │
│    │ HandleSSE()           │              │ HandleSSE()               │            │
│    │  ├─ wga.Run()         │              │  ├─ wga_sandbox.Run()     │            │
│    │  ├─ EinoTranslator    │              │  ├─ OpencodeTranslator    │            │
│    │  └─ SSE 响应          │              │  └─ SSE 响应              │            │
│    └───────────────────────┘              └───────────────────────────┘            │
└─────────────────────────────────────────────────────────────────────────────────────┘
                              │                              │
                              ▼                              ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                        wga                                           │
│                                高级 API - 智能体统一接口                              │
│                                                                                     │
│  ┌───────────────────────────────────────────────────────────────────────────────┐ │
│  │ api.go                                                                        │ │
│  │  ├─ Init(configPath)        加载智能体配置                                     │ │
│  │  ├─ CheckOptions(id, opts)  检查运行条件                                      │ │
│  │  ├─ Run(id, query, opts)    执行智能体，返回 AgentEvent 迭代器                 │ │
│  │  └─ Cleanup(runID)          清理沙箱资源                                       │ │
│  └───────────────────────────────────────────────────────────────────────────────┘ │
│                                       │                                             │
│                                       ▼                                             │
│  ┌───────────────────────────────────────────────────────────────────────────────┐ │
│  │ internal/factory/agent.go - 智能体工厂                                         │ │
│  │                                                                               │ │
│  │  newAgent()                                                                   │ │
│  │   ├─ AgentTypeReAct      → newReactAgent()      → eino adk                   │ │
│  │   ├─ AgentTypeSandbox    → newSandboxAgent()    → wga-sandbox                │ │
│  │   ├─ AgentTypeSequential → newSequentialAgent() → 组合智能体                  │ │
│  │   ├─ AgentTypeLoop       → newLoopAgent()       → 组合智能体                  │ │
│  │   ├─ AgentTypeParallel   → newParallelAgent()   → 组合智能体                  │ │
│  │   ├─ AgentTypeDeep       → newDeepAgent()       → 深度思考智能体              │ │
│  │   └─ AgentTypeSupervisor → newSupervisorAgent() → 监督者智能体                │ │
│  └───────────────────────────────────────────────────────────────────────────────┘ │
│                                       │                                             │
│                                       ▼                                             │
│  ┌───────────────────────────────────────────────────────────────────────────────┐ │
│  │ internal/factory/agent_sandbox.go - 沙箱智能体                                  │ │
│  │                                                                               │ │
│  │  sandboxAgent.Run()                                                          │ │
│  │   │                                                                           │ │
│  │   │  1. buildSandboxOpts()     构建沙箱选项                                    │ │
│  │   │     ├─ ModelConfig (来自 WithModelConfig)                                 │ │
│  │   │     ├─ Instruction (来自配置文件)                                          │ │
│  │   │     └─ Tools (来自配置 + WithToolConfig)                                   │ │
│  │   │                                                                           │ │
│  │   │  2. wga_sandbox.Run()  ─────────────────▶ wga-sandbox                     │ │
│  │   │                                                                           │ │
│  │   │  3. converter.Convert()   转换 JSON → Message                             │ │
│  │   │                                                                           │ │
│  │   │  4. Send AgentEvent                                                       │ │
│  │   ▼                                                                           │ │
│  │  返回: *adk.AsyncIterator[*adk.AgentEvent]                                    │ │
│  └───────────────────────────────────────────────────────────────────────────────┘ │
│                                       │                                             │
│                                       ▼                                             │
│  ┌───────────────────────────────────────────────────────────────────────────────┐ │
│  │ internal/factory/converter/ - 事件转换器                                        │ │
│  │                                                                               │ │
│  │  EventConverter 接口                                                          │ │
│  │   └─ Convert(line string) → *schema.Message                                  │ │
│  │                                                                               │ │
│  │  opencodeConverter 实现                                                        │ │
│  │   ├─ text      → Message{Content}                                            │ │
│  │   ├─ reasoning → Message{ReasoningContent}                                   │ │
│  │   ├─ tool_use  → Message{ToolCalls}                                          │ │
│  │   ├─ file      → Message{Content: "[file] ..."}                              │ │
│  │   └─ error     → Message{Content: "[error] ..."}                             │ │
│  └───────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────────┘
                    │                                      │
                    │ 调用                                 │ 解析事件
                    ▼                                      ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                    wga-sandbox                                       │
│                              低级 API - 沙箱容器交互                                  │
│                                                                                     │
│  ┌───────────────────────────────────────────────────────────────────────────────┐ │
│  │ api.go                                                                        │ │
│  │  ├─ Run(ctx, opts)        执行任务，返回 <-chan string                        │ │
│  │  └─ Cleanup(ctx, runID)   清理沙箱环境                                         │ │
│  └───────────────────────────────────────────────────────────────────────────────┘ │
│                                       │                                             │
│                                       ▼                                             │
│  ┌───────────────────────────────────────────────────────────────────────────────┐ │
│  │ Run() 执行流程                                                                │ │
│  │                                                                               │ │
│  │  ┌────────────────┐   ┌────────────────┐   ┌────────────────┐               │ │
│  │  │ manager.Create │──▶│ runner.BeforeRun│──▶│ runner.Run    │               │ │
│  │  │   创建沙箱      │   │   准备环境      │   │   执行任务    │               │ │
│  │  └────────────────┘   └────────────────┘   └───────┬────────┘               │ │
│  │                                                     │                        │ │
│  │  ┌────────────────┐   ┌────────────────┐           │                        │ │
│  │  │ manager.Cleanup│◀──│ runner.AfterRun│◀──────────┘                        │ │
│  │  │   清理沙箱      │   │   复制输出      │                                    │ │
│  │  └────────────────┘   └────────────────┘                                    │ │
│  └───────────────────────────────────────────────────────────────────────────────┘ │
│                                       │                                             │
│                                       ▼                                             │
│  ┌───────────────────────────────────────────────────────────────────────────────┐ │
│  │ internal/sandbox/ - 沙箱管理                                                   │ │
│  │                                                                               │ │
│  │  Manager                                                                       │ │
│  │   ├─ Create(ctx, runID, cfg)  创建沙箱实例                                    │ │
│  │   ├─ Get(runID)               获取沙箱实例                                     │ │
│  │   └─ Cleanup(ctx, runID)      清理沙箱                                        │ │
│  │                                                                               │ │
│  │  Sandbox 接口                                                                  │ │
│  │   ├─ Prepare(ctx)                  创建工作目录                               │ │
│  │   ├─ Execute(ctx, args)            异步执行命令                               │ │
│  │   ├─ ExecuteSync(ctx, args)        同步执行命令                               │ │
│  │   ├─ CopyToSandbox(ctx, local, rel)复制文件到沙箱                             │ │
│  │   ├─ CopyFromSandbox(ctx, local)   复制文件到本地                             │ │
│  │   └─ Cleanup(ctx)                  清理工作目录                               │ │
│  │                                                                               │ │
│  │  实现模式                                                                      │ │
│  │   ├─ reuseSandbox   复用已启动容器     ✅ 已实现                               │ │
│  │   └─ oneshotSandbox 每次启动新容器     ○ 仅接口                                │ │
│  └───────────────────────────────────────────────────────────────────────────────┘ │
│                                       │                                             │
│                                       ▼                                             │
│  ┌───────────────────────────────────────────────────────────────────────────────┐ │
│  │ internal/runner/opencode/ - Opencode 运行器                                    │ │
│  │                                                                               │ │
│  │  Runner                                                                       │ │
│  │   ├─ BeforeRun(ctx)                                                          │ │
│  │   │   ├─ setupConfig()          创建 opencode.json                            │ │
│  │   │   ├─ 复制 skills 到 .opencode/skills/                                    │ │
│  │   │   ├─ 转换 tools 为 skills (openapi-to-skills)                             │ │
│  │   │   └─ 复制 input 目录内容                                                 │ │
│  │   │                                                                           │ │
│  │   ├─ Run(ctx) → chan string                                                  │ │
│  │   │   └─ 通过 HTTP API 调用 opencode，接收 SSE 事件流                         │ │
│  │   │                                                                           │ │
│  │   └─ AfterRun(ctx)                                                           │ │
│  │       └─ 复制沙箱输出到 OutputDir（排除隐藏文件）                              │ │
│  └───────────────────────────────────────────────────────────────────────────────┘ │
│                                       │                                             │
│                                       ▼                                             │
│  ┌───────────────────────────────────────────────────────────────────────────────┐ │
│  │ api_opencode.go - Opencode 事件解析                                            │ │
│  │                                                                               │ │
│  │  ParseOpencodeEvent(data) → *OpencodeEvent                                   │ │
│  │   ├─ ParseOpencodeTextPart(data)      → TextPart                             │ │
│  │   ├─ ParseOpencodeToolPart(data)      → ToolPart                             │ │
│  │   ├─ ParseOpencodeReasoningPart(data) → ReasoningPart                        │ │
│  │   └─ ... 更多解析函数                                                         │ │
│  └───────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        │ JSON 字符串流
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                     ag-ui-util                                       │
│                              协议转换层 - AG-UI 事件生成                              │
│                                                                                     │
│  ┌───────────────────────────────────────────────────────────────────────────────┐ │
│  │ Translator 接口                                                                │ │
│  │  ├─ Translate(ctx, line) → []Event    转换单个事件                             │ │
│  │  ├─ Finish() → []Event                生成结束事件                             │ │
│  │  └─ TranslateStream(ctx, in) → chan   转换事件流                               │ │
│  └───────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                     │
│  ┌──────────────────────────────┐    ┌────────────────────────────────────────┐   │
│  │ OpencodeTranslator           │    │ EinoTranslator                         │   │
│  │                              │    │                                        │   │
│  │ 输入: opencode JSON 字符串    │    │ 输入: eino AgentEvent                  │   │
│  │                              │    │                                        │   │
│  │ {"type":"text",              │    │ AgentEvent{                            │   │
│  │  "part":{"text":"..."}}      │    │   Output: {                            │   │
│  │                              │    │     MessageOutput: {                   │   │
│  │ 输出: AG-UI Events           │    │       Message: {                       │   │
│  │  ├─ RunStarted               │    │         Content: "...",                │   │
│  │  ├─ TextMessageStart         │    │         ToolCalls: [...]               │   │
│  │  ├─ TextMessageContent       │    │       }                                │   │
│  │  ├─ TextMessageEnd           │    │     }                                  │   │
│  │  ├─ ToolCallStart            │    │   }                                    │   │
│  │  ├─ ToolCallArgs             │    │                                        │   │
│  │  ├─ ToolCallEnd              │    │ 输出: AG-UI Events (同左)              │   │
│  │  ├─ ToolCallResult           │    │                                        │   │
│  │  └─ RunFinished              │    │                                        │   │
│  │                              │    │                                        │   │
│  │ 使用场景:                    │    │ 使用场景:                               │   │
│  │ wga-sandbox 直接输出转换      │    │ wga.Run() 返回的事件转换                │   │
│  └──────────────────────────────┘    └────────────────────────────────────────┘   │
│                                                                                     │
│  ┌───────────────────────────────────────────────────────────────────────────────┐ │
│  │ 辅助函数                                                                      │ │
│  │  ├─ EventsToJSONChannel(ctx, events)  事件流 → JSON 字符串流                  │ │
│  │  └─ RemoveReasoningContent(content)   移除推理内容                            │ │
│  └───────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## 数据流

```
路径 A: wga 高级 API
────────────────────

  用户请求
       │
       ▼
  ┌─────────────┐
  │  wga.Run()  │
  └──────┬──────┘
         │
         ▼
  ┌───────────────────┐
  │ factory.NewAgent  │
  │  (type=sandbox)   │
  └────────┬──────────┘
           │
           ▼
  ┌───────────────────────────────────────────────────────────────────────┐
  │ sandboxAgent.Run()                                                    │
  │                                                                       │
  │  ┌────────────────┐   ┌─────────────────┐   ┌────────────────────┐  │
  │  │wga_sandbox.Run │──▶│ opencode runner │──▶│ JSON 字符串流      │  │
  │  └────────────────┘   └─────────────────┘   └──────────┬─────────┘  │
  │                                                        │            │
  │  ┌─────────────────────────────────────────────────────▼──────────┐ │
  │  │ converter.Convert(line)                                         │ │
  │  │ JSON → schema.Message                                           │ │
  │  └─────────────────────────────────────────────────────┬──────────┘ │
  │                                                        │            │
  │  ┌─────────────────────────────────────────────────────▼──────────┐ │
  │  │ generator.Send(&adk.AgentEvent{Output: MessageOutput})         │ │
  │  └─────────────────────────────────────────────────────┬──────────┘ │
  └────────────────────────────────────────────────────────┼────────────┘
                                                           │
                                                           ▼
                                               ┌────────────────────────┐
                                               │ *adk.AsyncIterator     │
                                               │ [*adk.AgentEvent]      │
                                               └───────────┬────────────┘
                                                           │
                                                           ▼
  ┌───────────────────────────────────────────────────────────────────────────┐
  │ EinoTranslator.TranslateStream(ctx, iter)                                 │
  │                                                                           │
  │  AgentEvent → AG-UI Event                                                 │
  │   ├─ Message.Content         → TextMessageContent                         │
  │   ├─ Message.ReasoningContent→ TextMessageContent (引用格式)               │
  │   └─ Message.ToolCalls       → ToolCallStart/Args/End/Result              │
  └───────────────────────────────────────────────────────────────────────────┘
                                                           │
                                                           ▼
                                               ┌────────────────────────┐
                                               │ AG-UI Events (JSON)    │
                                               └───────────┬────────────┘
                                                           │
                                                           ▼
                                               ┌────────────────────────┐
                                               │ SSE 响应给前端         │
                                               └────────────────────────┘


路径 B: wga-sandbox 低级 API
────────────────────────────

  用户请求
       │
       ▼
  ┌────────────────────────┐
  │ wga_sandbox.Run()      │
  └───────────┬────────────┘
              │
              ▼
  ┌───────────────────────────────────────────────────────────────────────┐
  │ manager.Create(ctx, runID, cfg)                                       │
  │  ├─ 创建沙箱工作目录                                                   │
  │  └─ 返回 Sandbox 实例                                                  │
  └───────────────────────────────────────────────────────────────────────┘
              │
              ▼
  ┌───────────────────────────────────────────────────────────────────────┐
  │ runner.BeforeRun(ctx)                                                 │
  │  ├─ 创建 opencode.json (模型配置)                                      │
  │  ├─ 复制 skills/tools                                                 │
  │  └─ 复制输入文件                                                       │
  └───────────────────────────────────────────────────────────────────────┘
              │
              ▼
  ┌───────────────────────────────────────────────────────────────────────┐
  │ runner.Run(ctx) → chan string                                         │
  │                                                                       │
  │  通过 HTTP API 调用 opencode，接收 SSE 事件流                          │
  │                                                                       │
  │  输出流:                                                               │
  │   {"type":"step_start","part":{...}}                                  │
  │   {"type":"text","part":{"text":"..."}}                               │
  │   {"type":"tool_use","part":{"tool":"bash",...}}                      │
  │   {"type":"reasoning","part":{"text":"..."}}                          │
  │   {"type":"step_finish","part":{...}}                                 │
  └───────────────────────────────────────────────────────────────────────┘
              │
              ▼
  ┌───────────────────────────────────────────────────────────────────────┐
  │ runner.AfterRun(ctx)                                                  │
  │  └─ 复制沙箱输出到 OutputDir                                           │
  └───────────────────────────────────────────────────────────────────────┘
              │
              ▼
  ┌───────────────────────────────────────────────────────────────────────┐
  │ manager.Cleanup(ctx, runID) (除非 SkipCleanup=true)                   │
  │  └─ 删除沙箱工作目录                                                   │
  └───────────────────────────────────────────────────────────────────────┘
              │
              ▼
  ┌───────────────────────────────────────────────────────────────────────┐
  │ OpencodeTranslator.TranslateStream(ctx, outputCh)                     │
  │                                                                       │
  │  JSON line → AG-UI Event                                              │
  │   ├─ text      → TextMessageContent                                   │
  │   ├─ reasoning → TextMessageContent (引用格式)                         │
  │   ├─ tool_use  → ToolCallStart/Args/End/Result                        │
  │   └─ step_*    → 跳过                                                 │
  └───────────────────────────────────────────────────────────────────────┘
              │
              ▼
  ┌────────────────────────┐
  │ AG-UI Events (JSON)    │
  └───────────┬────────────┘
              │
              ▼
  ┌────────────────────────┐
  │ SSE 响应给前端         │
  └────────────────────────┘
```

## 依赖关系

```
       ┌─────────────────┐
       │    应用层        │
       └────────┬────────┘
                │
         ┌──────┴──────┐
         │             │
         ▼             ▼
    ┌────────┐    ┌─────────────┐
    │  wga   │    │ wga-sandbox │
    └───┬────┘    └──────┬──────┘
        │                │
        │    ┌───────────┘
        │    │
        ▼    ▼
    ┌────────────┐
    │ ag-ui-util │
    └────────────┘

    依赖说明:
     ├─ wga 依赖 wga-sandbox (sandbox agent 内部调用)
     ├─ wga 不依赖 ag-ui-util (返回 eino AgentEvent，由应用层选择是否转换)
     ├─ wga-sandbox 不依赖 ag-ui-util (输出 JSON 字符串流，由调用方选择是否转换)
     └─ ag-ui-util 完全独立 (可单独使用)

    使用场景:
     ├─ examples/agui-demo/backend-wga: wga.Run() → EinoTranslator → AG-UI
     └─ examples/agui-demo/backend-wga-sandbox: wga_sandbox.Run() → OpencodeTranslator → AG-UI
```

## 配置映射

```
用户代码                           wga 内部                      wga-sandbox 内部
────────                           ─────────                    ────────────────

wga.WithModelConfig(      ─────▶  options.Model         ─────▶  ModelConfig
  Model: "glm-4",
  ApiKey: "sk-xxx",
  EndpointUrl: "..."
)

wga.WithToolConfig(       ─────▶  options.Tools[]       ─────▶  Tools[]
  Title: "天气查询",                                          .OpenAPI3Schema
  APIAuth: {...}                                               .OperationIDs
)                                                               .APIAuth

wga.WithWorkspaceConfig(  ─────▶  options.Workspace     ─────▶  InputDir
  InputDir: "...",                                              OutputDir
  OutputDir: "..."
)

wga.WithRunSession(       ─────▶  options.RunSession    ─────▶  RunSession
  ThreadID: "thread-123",                                       .ThreadID
  RunID: "run-456"                                              .RunID
)

wga.WithMessages(         ─────▶  options.Messages      ─────▶  Messages
  {Role: "user", Content: "..."}                                .Role
)                                                               .Content

配置文件 (YAML)                     internal/config              runner 内部
─────────────                       ─────────────              ────────────

agent.yaml  ──────────────────▶   config.Agent
  id: code-agent                      .ID
  type: sandbox                       .Type
  name: 代码助手                       .Name
  description: 代码生成和修改          .Description
  prompt_relative_path: ./prompt.md  ─────────▶  .Prompt        ───────▶   Instruction

  configure:
    max_iterations: 10               .Configure.MaxIterations
    enable_thinking: true  ─────────▶ .Configure.EnableThinking ──────▶  EnableThinking
    sandbox:
      type: reuse                    .Configure.Sandbox.Type
      host: localhost                .Configure.Sandbox.Host     ──────▶  SandboxConfig

  tool_categories:
    - category: search               .ToolCategories
      condition: optional
      tools:
        - path: configs/tools/search.json   .Tools[].SchemaPath (相对程序运行目录)
          auth_required: true               .Tools[].AuthRequired
          operations:                       .Tools[].Operations
            - operation_id: web_search

  skills:
    - dir: configs/skills/coding     .Skills[].Dir (相对程序运行目录)

  sub_agents:
    - relative_path: ./sub-agent/config.yaml  (相对当前配置文件目录)
```

## 配置路径规则

| 配置项 | 字段名 | 相对路径基准 |
|--------|--------|-------------|
| `sub_agents` | `relative_path` | 当前配置文件目录 |
| `prompt_relative_path` | `relative_path` | 当前配置文件目录 |
| `tools` | `path` | 程序运行目录 |
| `skills` | `dir` | 程序运行目录 |

## 智能体类型

| 类型 | 说明 |
|------|------|
| react | ReAct 模式原子智能体，使用 eino adk 实现 |
| sandbox | 沙箱执行智能体，在隔离容器中运行 opencode |
| sequential | 顺序执行多个子智能体 |
| loop | 循环执行子智能体 |
| parallel | 并行执行多个子智能体 |
| deep | 深度思考智能体，递归分解任务 |
| supervisor | 监督者模式，由主智能体协调子智能体 |

## 使用

```go
ctx := context.Background()

// 初始化
wga.Init(ctx, "/path/to/config.yaml")

// 执行
iter, _ := wga.Run(ctx, "agent-id", "任务描述",
    wga.WithModelConfig(wga_option.ModelConfig{
        Model:       "glm-4",
        ApiKey:      "sk-xxx",
        EndpointUrl: "https://api.example.com/v1",
    }),
)

// 获取结果
for {
    event, ok := iter.Next()
    if !ok {
        break
    }
    if event.Output != nil && event.Output.MessageOutput != nil {
        fmt.Println(event.Output.MessageOutput.Message.Content)
    }
}
```

## API

| 函数 | 说明 |
|------|------|
| `Init(ctx, configPath)` | 初始化配置 |
| `CheckOptions(ctx, id, opts...)` | 检查运行条件 |
| `Run(ctx, id, query, opts...)` | 执行智能体 |
| `Cleanup(ctx, runID)` | 清理资源 |

## 选项

| 选项 | 说明 |
|------|------|
| `WithModelConfig` | 模型配置（必须） |
| `WithToolConfig` | 工具配置 |
| `WithWorkspaceConfig` | 工作空间 |
| `WithRunSession` | 会话标识 |
| `WithMessages` | 历史消息 |

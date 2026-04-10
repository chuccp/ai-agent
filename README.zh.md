# AI Agent Workflow

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **其他语言：** [English](README.md) | [中文繁體](README.zh-TW.md) | [日本語](README.ja.md)

---

一个轻量级、声明式的 Go 语言 AI Agent 工作流引擎。声明节点依赖，引擎自动解析执行顺序——无依赖的节点通过 goroutine 并行执行。

> **声明即并行** — 声明节点依赖，引擎自动构建执行层并并发执行。

---

## 项目特色

大多数 Go DAG 框架需要你手动定义边和构建图。本项目采用不同的思路：

- **通过 `ValueFrom` 声明依赖** — 引擎自动发现依赖图
- **无需手动构建图** — 只需列出节点，引擎自动计算执行顺序
- **自动分层并行** — 同一依赖层级的节点并发执行
- **零外部依赖** — 无需 Redis、数据库或消息队列

## 对比

| 特性 | ai-agent | [CloudWeGo Eino](https://github.com/cloudwego/eino) | [Dagu](https://github.com/dagucloud/dagu) |
|------|----------|------|-------|
| 依赖声明 | `ValueFrom` 自动发现 | 手动定义边 | YAML 定义 |
| 并行执行 | 自动分层并行 | DAG 调度器 | YAML 定义 |
| AI/LLM 内置 | LLMNode, ImageGenerationNode | 支持 | 无 |
| 外部依赖 | 无 | 多个 | SQLite |
| API 风格 | Go 代码（Builder 模式） | Go 代码 | YAML |
| 定位 | 轻量级嵌入式库 | 企业级框架 | 本地工作流引擎 |

---

## 快速开始

```bash
go get github.com/chuccp/ai-agent
```

### Hello World

```go
package main

import (
    "fmt"
    ai_agent "github.com/chuccp/ai-agent"
    "github.com/chuccp/ai-agent/executor"
    "github.com/chuccp/ai-agent/node"
    "github.com/chuccp/ai-agent/value"
)

func main() {
    processNode := node.NewFunctionNodeBuilder("process").
        ExecFunc(func(state *node.State) (value.NodeValue, error) {
            name := state.GetRootValue().GetString("name")
            result := value.NewObjectValue()
            result.PutString("greeting", "Hello, "+name+"!")
            return result, nil
        }).Build()

    workflow := ai_agent.Of(processNode)
    ag := ai_agent.NewAgentBuilder("hello-agent").Workflow(workflow).Build()
    exec := ai_agent.NewAgentExecutor(ag, &executor.Config{MaxConcurrency: 2})

    input := value.NewObjectValue()
    input.PutString("name", "World")

    response, _ := exec.Exec(input)
    fmt.Println(response.NodeValue.(*value.ObjectValue).GetString("greeting"))
    // 输出: Hello, World!
}
```

---

## 架构

```
Agent → Workflow → NodeExecutor → Nodes（自动分层并行执行）
```

1. **Agent** 用配置包装 Workflow
2. **Workflow** 持有节点序列
3. **NodeExecutor** 通过分析 `ValueFrom` 依赖构建执行层
4. **同一层的节点通过 goroutine 并发执行**

### 执行层构建

```
NodeA ──┐
        ├──→ NodeC（依赖 A + B，等两者完成后执行）
NodeB ──┘

第 1 层: [NodeA, NodeB]  ← 并行执行
第 2 层: [NodeC]          ← 第 1 层完成后执行
```

---

## 节点类型

| 节点 | 说明 |
|------|------|
| **FunctionNode** | 通过 `ExecFunc` 实现自定义逻辑 |
| **IFNode** | 条件分支，支持 Then/Else 工作流 |
| **IterationNode** | 对数组输入进行并行批处理 |
| **OrderIterationNode** | 对数组输入进行顺序批处理 |
| **LLMNode** | 基于模板的 LLM 调用，支持缓存 |
| **ImageGenerationNode** | 图片生成，支持模板提示词 |
| **InputNode** | 入口节点，解析根参数 |
| **OutputNode** | 出口节点，支持输出转换 |

---

## 核心功能

### 1. 声明式依赖 → 自动并行

```go
// nodeA 和 nodeB 没有依赖 → 并行执行
nodeA := node.NewFunctionNodeBuilder("nodeA").
    ExecFunc(func(state *node.State) (value.NodeValue, error) {
        res := value.NewObjectValue()
        res.PutString("data", "from A")
        return res, nil
    }).Build()

nodeB := node.NewFunctionNodeBuilder("nodeB").
    ExecFunc(func(state *node.State) (value.NodeValue, error) {
        res := value.NewObjectValue()
        res.PutString("data", "from B")
        return res, nil
    }).Build()

// nodeC 依赖 nodeA 和 nodeB → 等两者完成后执行
nodeC := node.NewFunctionNodeBuilder("nodeC").
    ValuesFrom(
        value.NewValueFrom("nodeA", "", ""),
        value.NewValueFrom("nodeB", "", ""),
    ).
    ExecFunc(func(state *node.State) (value.NodeValue, error) {
        ctx := state.GetWorkflowContext()
        dataA := ctx.GetNodeValue("nodeA").(*value.ObjectValue).GetString("data")
        dataB := ctx.GetNodeValue("nodeB").(*value.ObjectValue).GetString("data")

        res := value.NewObjectValue()
        res.PutString("merged", dataA + " + " + dataB)
        return res, nil
    }).Build()

workflow := ai_agent.Of(nodeA, nodeB, nodeC)
```

### 2. 条件分支（IFNode）

```go
ifNode, _ := node.NewIFNodeBuilder("check").
    Condition(func(ctx node.WorkflowContext) bool {
        return ctx.GetRootValue().GetInt("score") >= 60
    }).
    Then(ai_agent.Of( /* ... */ )).
    Else(ai_agent.Of( /* ... */ )).
    Build()
```

### 3. 并行迭代（IterationNode）

```go
iterationNode := node.NewIterationNodeBuilder("iterate").
    IterationFrom(value.NewValueFrom("", "items", "")).
    Workflow(ai_agent.Of(
        node.NewFunctionNodeBuilder("processItem").
            ExecFunc(func(state *node.State) (value.NodeValue, error) {
                item := state.GetRootValue().GetInt("item")
                res := value.NewObjectValue()
                res.PutNumber("squared", float64(item*item))
                return res, nil
            }).Build(),
    )).Build()
```

### 4. LLMNode 模板引擎

```go
llmNode := node.NewLLMNodeBuilder("llm").
    SystemTemplate("你是一个有用的助手。").
    UserTemplate("你好，${name}！你的订单数是 ${count}。").
    LLMFunction(func(state *node.State, urls *value.UrlsValue, systemPrompt, userPrompt string, format out.OutFormat, stream bool) (value.NodeValue, error) {
        // systemPrompt: "你是一个有用的助手。"
        // userPrompt:   "你好，Alice！你的订单数是 100。"
        return value.NewTextValue("模拟回复"), nil
    }).
    Build()
```

支持 `${variable}` 和 `{{.variable}}` 两种模板语法。

### 5. 异步执行

```go
result := exec.ExecAsync(input)
if result.Error != nil {
    log.Fatal(result.Error)
}
fmt.Println(result.Response.Success)
```

---

## 值系统

丰富的多态值类型，支持路径查找：

```go
obj := value.NewObjectValue()
obj.PutString("name", "Alice").
    PutNumber("age", 30).
    PutBool("active", true)

// 链式调用
obj.PutObject("address", value.NewObjectValue().
    PutString("city", "北京"),
)

// 路径查找
obj.FindValue("address.city")  // TextValue("北京")
obj.FindValue("$.name")        // TextValue("Alice")
```

类型：`ObjectValue` | `ArrayValue` | `TextValue` | `BoolValue` | `NumberValue` | `NullValue` | `UrlsValue` | `FilesValue` | `StreamNodeValue`

---

## 测试

```bash
# 运行所有测试
go test ./...

# 运行指定测试
go test -v -run TestIFNode ./...

# 覆盖率
go test -cover ./...
```

---

## License

MIT

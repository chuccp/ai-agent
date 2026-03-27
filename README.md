# AI Agent Workflow

一个用于 AI 代理工作流的 Go 并发执行引擎，支持依赖自动解析、分层并行执行和多种节点类型。

## 架构概述

核心执行流程：

```
Agent → Workflow → NodeExecutor → Nodes (执行分层并发)
```

1. **Agent** 使用配置（缓存路径、最大并发数）包装 Workflow
2. **Workflow** 维护节点序列并处理依赖解析
3. **NodeExecutor** 根据节点依赖构建执行层，然后逐层执行
4. **同一层中的节点通过 goroutine 并发执行**

## 项目结构

```
ai-agent/
├── agent.go              # Agent, Workflow 定义
├── agent_manager.go      # Agent 管理器
├── cache/                # 缓存系统
│   └── cache_manager.go  # LLM 结果缓存管理器
├── executor/             # 执行器
│   ├── node_executor.go  # 节点执行器（使用 goroutine 并发）
│   ├── group_executor.go # 批量执行器
│   └── exec_tree.go      # 执行树构建（依赖分层）
├── graph/                # 图可视化系统
│   └── graph.go          # NodeGraph, Graph, NodeStatus
├── model/                # AI 模型接口
│   ├── llm_model.go       # LLM 模型接口
│   └── image_generation_model.go # 图片生成模型接口
├── node/                 # 节点系统
│   ├── node.go           # 基础节点和节点状态 (BaseNode)
│   ├── interface.go      # 核心接口定义 (Node, WorkflowContext, WorkflowInterface)
│   ├── basic_nodes.go    # InputNode, OutputNode, FunctionNode
│   ├── if_node.go        # 条件分支节点 (IFNode)
│   ├── iteration_node.go # 迭代节点 (IterationNode)
│   ├── llm_node.go       # LLM 节点（支持模板和缓存）
│   └── image_generation_node.go # 图片生成节点
├── out/                  # 输出格式系统
│   ├── out_format.go     # 输出格式接口
│   ├── text_out_format.go # 文本输出
│   ├── json_out_format.go # JSON 输出
│   └── field/            # 输出字段定义
│       ├── field.go      # 字段接口
│       ├── array_field.go
│       ├── object_field.go
│       └── text_field.go
├── pool/                 # 并发池
│   ├── pool.go           # 工作池实现
│   └── pool_test.go      # 单元测试
├── types/                # 类型定义
│   ├── node_type.go      # 节点类型枚举
│   ├── node_status_type.go # 节点状态
│   ├── agent_status_type.go # Agent 状态
│   ├── field_type.go     # 字段类型
│   └── out_type.go       # 输出类型
├── util/                 # 工具类
│   ├── str_utils.go      # 字符串工具
│   ├── file_utils.go     # 文件工具
│   ├── http_utils.go     # HTTP 工具
│   ├── io_utils.go       # IO 工具
│   ├── array_utils.go    # 数组工具
│   └── path.go           # 路径工具
├── value/                # 值系统
│   ├── node_value.go     # 值接口和基础类型
│   ├── object_value.go   # 对象值
│   ├── array_value.go    # 数组值
│   ├── urls_value.go     # URL 值
│   ├── files_value.go    # 文件值
│   └── value_from.go     # 值来源声明（依赖声明）
├── examples/             # 示例
│   └── main.go
└── go.mod
```

## 核心概念

### 值系统 (NodeValue)

支持多种值类型，可通过路径查找：

- **ObjectValue**: 对象值，键值对集合
- **ArrayValue**: 数组值
- **TextValue**: 文本值
- **BoolValue**: 布尔值
- **NumberValue**: 数值
- **NullValue**: 空值
- **UrlsValue**: URL 列表
- **FilesValue**: 文件列表
- **StreamNodeValue**: 流式值

节点之间依赖通过 `ValuesFrom` 声明：

```go
// 获取 nodeA 的全部输出
value.NewValueFrom("nodeA", "", "")

// 获取 nodeA 输出中的特定字段
value.NewValueFrom("nodeA", "$.field", "myField")
```

### 执行层构建

`executor/exec_tree.go` 通过以下步骤构建执行层：

1. 从最后一个节点（结束节点）开始
2. 通过 `ValuesFrom` 递归发现依赖
3. 将没有未解决依赖的节点分到同一层
4. 每层并行执行，完成后进入下一层

**同一层中的所有节点并发执行**，充分利用多核。

### 模板引擎

LLMNode 使用 Go 的 `text/template` 包，支持两种语法：

- `{{.fieldName}}` - 标准 Go 模板语法
- `${variable}` - 替代的美元花括号语法（内部自动转换）

## 主要特性

### Go 并发模式

使用 Go 原生并发原语：

- **Goroutine**: 原生轻量级线程并发执行节点
- **Channel**: 用于异步结果传递
- **sync.Map**: 线程安全的节点值存储
- **sync.WaitGroup**: 等待并发任务完成
- **Worker Pool**: 可限制最大并发数

### 核心组件

1. **Workflow**: 工作流定义，持有节点序列
2. **Agent**: 代理，包装工作流和配置（缓存路径、最大并发）
3. **AgentExecutor**: 代理执行器，支持同步和异步执行
4. **Node**: 节点接口，所有节点都必须实现
5. **NodeValue**: 多态值系统，支持路径查找

### 节点类型

- **InputNode**: 输入节点
- **OutputNode**: 输出节点
- **FunctionNode**: 函数节点
- **IFNode**: 条件节点（支持 Then/Else 分支）
- **IterationNode**: 迭代节点（批量处理）
- **LLMNode**: LLM节点（支持模板和缓存）
- **ImageGenerationNode**: 图片生成节点

## 使用示例

### 简单示例

```go
package main

import (
    "fmt"
    "log"

    ai_agent "github.com/chuccp/ai-agent"
    "github.com/chuccp/ai-agent/node"
    "github.com/chuccp/ai-agent/value"
)

func main() {
    // 创建函数节点 - 处理输入
    processNode := node.NewFunctionNodeBuilder("process").
        ExecFunc(func(state *node.State) (value.NodeValue, error) {
            rootValue := state.GetRootValue()
            name := rootValue.GetString("name")

            result := value.NewObjectValue()
            result.PutString("greeting", "Hello, "+name+"!")
            return result, nil
        }).
        Build()

    // 创建输出节点 - 依赖 process 节点的输出
    outputNode := node.NewOutputNodeBuilder("output").
        ValuesFrom(value.NewValueFrom("process", "", "")).
        OutFunc(func(nodeValue value.NodeValue) {
            fmt.Println("Output:", nodeValue.String())
        }).
        Build()

    // 创建工作流和代理
    workflow := ai_agent.Of(processNode, outputNode)
    ag := ai_agent.NewAgentBuilder("hello-agent").
        Workflow(workflow).
        Build()

    // 执行
    execConfig := &executor.Config{
        MaxConcurrency: 2,
    }
    exec := ai_agent.NewAgentExecutor(ag, execConfig)
    input := value.NewObjectValue()
    input.PutString("name", "World")

    response, err := exec.Exec(input)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Success:", response.Success)
}
```

### 条件分支示例 (IFNode)

```go
// 条件节点：根据输入决定走哪个分支
ifNode, err := node.NewIFNodeBuilder("check").
    Condition(func(ctx node.WorkflowContext) bool {
        rootVal := ctx.GetRootValue()
        score := int(rootVal.GetNumber("score"))
        return score >= 60
    }).
    Then(ai_agent.Of(
        node.NewFunctionNodeBuilder("pass").
            ExecFunc(func(state *node.State) (value.NodeValue, error) {
                res := value.NewObjectValue()
                res.PutString("result", "Passed")
                return res, nil
            }).Build(),
    )).
    Else(ai_agent.Of(
        node.NewFunctionNodeBuilder("fail").
            ExecFunc(func(state *node.State) (value.NodeValue, error) {
                res := value.NewObjectValue()
                res.PutString("result", "Failed")
                return res, nil
            }).Build(),
    )).
    ValuesFrom(value.NewValueFrom("", "score", "score")).
    Build()
```

### 并行执行示例

多个节点没有依赖关系会自动在同一层并行执行：

```go
// nodeA 和 nodeB 没有依赖，会并发执行
nodeA := node.NewFunctionNodeBuilder("nodeA").
    ExecFunc(func(state *node.State) (value.NodeValue, error) {
        res := value.NewObjectValue()
        res.PutNumber("value", 1)
        return res, nil
    }).Build()

nodeB := node.NewFunctionNodeBuilder("nodeB").
    ExecFunc(func(state *node.State) (value.NodeValue, error) {
        res := value.NewObjectValue()
        res.PutNumber("value", 2)
        return res, nil
    }).Build()

// nodeC 依赖 nodeA 和 nodeB，执行前会等待两者完成
nodeC := node.NewFunctionNodeBuilder("nodeC").
    ValuesFrom(
        value.NewValueFrom("nodeA", "", ""),
        value.NewValueFrom("nodeB", "", ""),
    ).
    ExecFunc(func(state *node.State) (value.NodeValue, error) {
        ctx := state.GetWorkflowContext()
        valA := int(ctx.GetNodeValue("nodeA").(*value.ObjectValue).GetNumber("value"))
        valB := int(ctx.GetNodeValue("nodeB").(*value.ObjectValue).GetNumber("value"))

        res := value.NewObjectValue()
        res.PutNumber("sum", float64(valA+valB))
        return res, nil
    }).
    Build()

workflow := ai_agent.Of(nodeA, nodeB, nodeC)
```

### 迭代处理示例 (IterationNode)

```go
// 对数组中每个元素并行处理
processItem := node.NewFunctionNodeBuilder("processItem").
    ExecFunc(func(state *node.State) (value.NodeValue, error) {
        item := int(state.GetRootValue().GetNumber("item"))

        result := value.NewObjectValue()
        result.PutNumber("squared", float64(item*item))
        return result, nil
    }).
    Build()

// 创建迭代节点，对输入数组的每个元素并行执行子工作流
iterationNode := node.NewIterationNodeBuilder("iterate").
    IterationFrom(value.NewValueFrom("", "items", "")).
    Workflow(ai_agent.Of(processItem)).
    Build()

workflow := ai_agent.Of(iterationNode)
```


## 安装

```bash
go get github.com/chuccp/ai-agent
```

## 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test -v -run TestIFNode ./...

# 运行示例
go run ./examples/main.go

# 整理依赖
go mod tidy
```

## 异步执行示例

```go
// 异步执行，返回结果通道
resultChan := exec.ExecAsync(input)
result := <-resultChan
if result.Err != nil {
    log.Fatal(result.Err)
}
fmt.Println("Result:", result.Response)
```

## 依赖

- Go 1.22+
- github.com/google/uuid
- emperror.dev/errors

## License

MIT
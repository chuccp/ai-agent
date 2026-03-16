# AI Agent Workflow

这是将 Java 项目 [ai-agent-workflow-core](https://github.com/chuccp/ai-agent-workflow) 转换为 Go 语言的版本。

## 项目结构

```
ai-agent/
├── agent/           # 代理和工作流核心
│   └── agent.go     # Agent, Workflow, AgentExecutor
├── cache/           # 缓存系统
│   └── cache_manager.go
├── executor/        # 执行器
│   ├── node_executor.go   # 节点执行器（使用 goroutine 并发）
│   ├── group_executor.go  # 组执行器
│   └── exec_tree.go       # 执行树构建
├── graph/           # 图系统
│   └── graph.go     # NodeGraph, Graph, NodeStatus
├── node/            # 节点系统
│   ├── node.go            # 基础节点和节点状态
│   ├── basic_nodes.go     # InputNode, OutputNode, FunctionNode
│   ├── iteration_node.go  # 迭代节点
│   └── llm_node.go        # LLM节点
├── types/           # 类型定义
│   ├── node_type.go
│   ├── node_status_type.go
│   └── agent_status_type.go
├── util/            # 工具类
│   └── str_utils.go
├── value/           # 值系统
│   ├── node_value.go      # 值接口和基础类型
│   ├── object_value.go    # 对象值
│   ├── array_value.go     # 数组值
│   ├── urls_value.go      # URL和文件值
│   └── value_from.go      # 值来源
├── examples/        # 示例
│   └── main.go
└── go.mod
```

## 主要特性

### Go 并发模式

相比 Java 版本使用 `ThreadPoolExecutor` 和 `CompletableFuture`，Go 版本使用：

- **Goroutine**: 用于并发执行节点
- **Channel**: 用于异步结果传递
- **sync.Map**: 用于线程安全的值存储
- **sync.WaitGroup**: 用于等待并发任务完成

### 核心组件

1. **Workflow**: 工作流定义，包含节点序列
2. **Agent**: 代理，包装工作流和配置
3. **AgentExecutor**: 代理执行器，支持同步和异步执行
4. **Node**: 节点接口，支持多种节点类型
5. **NodeValue**: 值系统，支持多种数据类型

### 节点类型

- **InputNode**: 输入节点
- **OutputNode**: 输出节点
- **FunctionNode**: 函数节点
- **IterationNode**: 迭代节点（批量处理）
- **LLMNode**: LLM节点（支持模板和缓存）

## 使用示例

```go
package main

import (
    "fmt"
    "log"

    "github.com/chuccp/ai-agent/agent"
    "github.com/chuccp/ai-agent/node"
    "github.com/chuccp/ai-agent/value"
)

func main() {
    // 创建函数节点
    processNode := node.NewFunctionNodeBuilder("process").
        ExecFunc(func(state *node.NodeState) (value.NodeValue, error) {
            rootValue := state.GetRootValue()
            name := rootValue.GetString("name")

            result := value.NewObjectValue()
            result.PutString("greeting", "Hello, "+name+"!")
            return result, nil
        }).
        Build()

    // 创建输出节点
    outputNode := node.NewOutputNodeBuilder("output").
        ValuesFrom(value.NewValueFrom("process", "", "")).
        OutFunc(func(nodeValue value.NodeValue) {
            fmt.Println("Output:", nodeValue.String())
        }).
        Build()

    // 创建工作流和代理
    workflow := agent.Of(processNode, outputNode)
    ag := agent.NewAgentBuilder("hello-agent").
        Workflow(workflow).
        Build()

    // 执行
    exec := agent.NewAgentExecutor(ag, nil)
    input := value.NewObjectValue()
    input.PutString("name", "World")

    response, err := exec.Exec(input)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Success:", response.Success)
}
```

## 与 Java 版本的主要区别

| 特性 | Java 版本 | Go 版本 |
|------|-----------|---------|
| 并发模型 | ThreadPoolExecutor + Future | Goroutine + Channel |
| 异步执行 | CompletableFuture | <-chan *AsyncResult |
| 线程安全 | ConcurrentHashMap | sync.Map |
| 回调 | 函数式接口 | 函数类型 |
| 类型系统 | 继承 + 接口 | 接口 + 组合 |

## 依赖

- Go 1.22+
- github.com/google/uuid

## License

MIT
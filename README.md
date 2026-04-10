# AI Agent Workflow

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **Other Languages:** [中文简体](README.zh.md) | [中文繁體](README.zh-TW.md) | [日本語](README.ja.md)

---

A lightweight, declarative AI Agent workflow engine in Go. Define node dependencies and let the engine automatically resolve execution order — nodes without dependencies run in parallel via goroutines.

> **Declare, and it runs in parallel** — declare node dependencies, the engine automatically builds execution layers and runs them concurrently.

---

## Why This Project

Most Go DAG frameworks require you to manually define edges and build the graph. This engine takes a different approach:

- **Declare dependencies via `ValueFrom`** — the engine automatically discovers the dependency graph
- **No manual graph construction** — just list your nodes, the engine figures out execution order
- **Automatic layered parallelism** — nodes at the same dependency level run concurrently
- **Zero external dependencies** — no Redis, no database, no message queue

## Comparison

| Feature | ai-agent | [CloudWeGo Eino](https://github.com/cloudwego/eino) | [Dagu](https://github.com/dagucloud/dagu) |
|---------|----------|------|-------|
| Dependency Declaration | `ValueFrom` auto-discovery | Manual edge definition | YAML definition |
| Parallel Execution | Automatic layered parallelism | DAG scheduler | YAML-defined |
| AI/LLM Built-in | LLMNode, ImageGenerationNode | Yes | No |
| External Dependencies | None | Multiple | SQLite |
| API Style | Go code (Builder pattern) | Go code | YAML |
| Positioning | Lightweight embedded library | Enterprise framework | Local workflow engine |

---

## Quick Start

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
    // Output: Hello, World!
}
```

---

## Architecture

```
Agent → Workflow → NodeExecutor → Nodes (auto-layered parallel execution)
```

1. **Agent** wraps a Workflow with configuration
2. **Workflow** holds a sequence of Nodes
3. **NodeExecutor** builds execution layers by analyzing `ValueFrom` dependencies
4. **Nodes in the same layer execute concurrently** via goroutines

### Execution Layer Building

```
NodeA ──┐
        ├──→ NodeC (depends on A + B, runs after both complete)
NodeB ──┘

Layer 1: [NodeA, NodeB]  ← run in parallel
Layer 2: [NodeC]          ← runs after Layer 1
```

---

## Node Types

| Node | Description |
|------|-------------|
| **FunctionNode** | Custom logic via `ExecFunc` |
| **IFNode** | Conditional branching with Then/Else workflows |
| **IterationNode** | Parallel batch processing over array input |
| **OrderIterationNode** | Sequential batch processing over array input |
| **LLMNode** | Template-based LLM calls with caching support |
| **ImageGenerationNode** | Image generation with template prompts |
| **InputNode** | Entry point, parses root parameters |
| **OutputNode** | Exit point with optional output transformation |

---

## Core Features

### 1. Declarative Dependencies → Automatic Parallelism

Nodes declare what data they need. The engine builds the DAG and executes independent nodes in parallel.

```go
// nodeA and nodeB have no dependencies → they run in parallel
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

// nodeC depends on both nodeA and nodeB → runs after they complete
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

### 2. Conditional Branching (IFNode)

```go
ifNode, _ := node.NewIFNodeBuilder("check").
    Condition(func(ctx node.WorkflowContext) bool {
        return ctx.GetRootValue().GetInt("score") >= 60
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
    )).Build()
```

### 3. Parallel Iteration (IterationNode)

```go
// Process each item in the array in parallel
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

### 4. LLMNode with Template Engine

```go
llmNode := node.NewLLMNodeBuilder("llm").
    SystemTemplate("You are a helpful assistant.").
    UserTemplate("Hello, ${name}! Your order count is ${count}.").
    LLMFunction(func(state *node.State, urls *value.UrlsValue, systemPrompt, userPrompt string, format out.OutFormat, stream bool) (value.NodeValue, error) {
        // systemPrompt: "You are a helpful assistant."
        // userPrompt:   "Hello, Alice! Your order count is 100."
        return value.NewTextValue("Mock response"), nil
    }).
    Build()
```

Supports both `${variable}` and `{{.variable}}` template syntax.

### 5. Async Execution

```go
// Async execution
result := exec.ExecAsync(input)
if result.Error != nil {
    log.Fatal(result.Error)
}
fmt.Println(result.Response.Success)
```

---

## Value System

Rich polymorphic value types with path-based lookup:

```go
obj := value.NewObjectValue()
obj.PutString("name", "Alice").
    PutNumber("age", 30).
    PutBool("active", true)

// Fluent chaining (all Put* methods return *ObjectValue)
obj.PutObject("address", value.NewObjectValue().
    PutString("city", "Beijing").
    PutString("country", "China"),
)

// Path-based lookup
obj.FindValue("address.city")  // TextValue("Beijing")
obj.FindValue("$.name")        // TextValue("Alice")
```

Types: `ObjectValue` | `ArrayValue` | `TextValue` | `BoolValue` | `NumberValue` (int/uint/float variants) | `NullValue` | `UrlsValue` | `FilesValue` | `StreamNodeValue`

---

## Project Structure

```
ai-agent/
├── agent.go                 # Agent, Workflow, AgentExecutor
├── node/
│   ├── interface.go         # Core interfaces (Node, WorkflowContext, WorkflowInterface)
│   ├── node.go              # BaseNode, State
│   ├── basic_nodes.go       # InputNode, OutputNode, FunctionNode
│   ├── if_node.go           # IFNode (conditional branching)
│   ├── iteration_node.go    # IterationNode (parallel batch)
│   ├── order_iteration_node.go # OrderIterationNode (sequential batch)
│   ├── llm_node.go          # LLMNode (template + cache)
│   └── image_generation_node.go # ImageGenerationNode
├── executor/
│   ├── node_executor.go     # Node execution with goroutine parallelism
│   ├── group_executor.go    # Batch execution
│   └── exec_tree.go         # DAG layer building from dependencies
├── value/
│   ├── node_value.go        # NodeValue interface + all value types
│   ├── object_value.go      # ObjectValue with template engine
│   └── array_value.go       # ArrayValue
├── graph/                   # Graph visualization
├── cache/                   # LLM result caching
├── pool/                    # Worker pool (GOPool)
├── out/                     # Output formatting (text/json)
├── model/                   # LLM & Image generation model interfaces
├── types/                   # Type definitions
└── util/                    # Utilities
```

---

## Tests

```bash
# Run all tests
go test ./...

# Run specific test
go test -v -run TestIFNode ./...

# Run with coverage
go test -cover ./...
```

---

## License

MIT

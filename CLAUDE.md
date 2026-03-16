# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run Commands

```bash
# Build all packages
go build ./...

# Run tests
go test ./...

# Run specific test
go test -v -run TestIFNode ./...

# Run example
go run ./examples/main.go

# Tidy dependencies
go mod tidy
```

## Architecture Overview

This is an AI Agent Workflow engine converted from Java, using Go-idiomatic concurrency patterns.

### Core Execution Flow

```
Agent → Workflow → NodeExecutor → Nodes (executed in layers)
```

1. **Agent** wraps a Workflow with configuration (cache path, max concurrency)
2. **Workflow** holds a sequence of Nodes with dependency resolution
3. **NodeExecutor** builds execution layers based on node dependencies, then executes each layer
4. **Nodes in the same layer execute concurrently via goroutines**

### Key Abstractions

**Core Interfaces** (`node/interface.go`):
- `Node` - All nodes implement `Exec(state *NodeState) (NodeValue, error)`
- `WorkflowContext` - Execution context with `GetRootValue()`, `GetNodeValue()`, `CreateChildContext()`
- `WorkflowInterface` - Workflow execution contract with `Exec()`, `ExecBatch()`, `Execute()`

**Node Implementation** (`node/node.go`):
- Nodes declare dependencies via `ValuesFrom` (which node's output they need)
- Dependencies are resolved at execution time to build parallel execution layers
- `BaseNode` provides common functionality via embedding

**Executor Context** (`executor/node_executor.go`):
- Implements `WorkflowContext` - the actual execution state holder
- Thread-safe using `sync.Map` for node values

**NodeValue Interface** (`value/node_value.go`):
- Polymorphic value types: ObjectValue, ArrayValue, TextValue, BoolValue, NumberValue, NullValue, UrlsValue, FilesValue, StreamNodeValue
- Supports path-based lookup: `FindValue("user.name")` or `FindValue("items[0]")`

### Execution Layer Building

`executor/exec_tree.go` builds layers by:
1. Starting from the last node (end node)
2. Recursively discovering dependencies via `ValuesFrom`
3. Grouping nodes with no unresolved dependencies into the same layer
4. Each layer executes in parallel, then moves to the next layer

### Node Types

- **InputNode**: Entry point, parses ValuesFrom into workflow input
- **OutputNode**: Exit point, can have custom output function
- **FunctionNode**: Custom logic via `NodeExecFunc`
- **IFNode**: Conditional branching with `ConditionFunc`, `Then()`, `Else()` workflows
- **IterationNode**: Batch processing over array input via `IterationFrom`
- **LLMNode**: Template-based LLM calls with caching
- **ImageGenerationNode**: Image generation with template prompts

### Value Flow Between Nodes

```go
// Node A outputs to context
context.AddNodeValue("nodeA", resultValue)

// Node B declares dependency
ValuesFrom: []*value.ValueFrom{
    value.NewValueFrom("nodeA", "", ""),  // Get all output from nodeA
    value.NewValueFrom("nodeA", "$.field", "myField"),  // Get specific field
}
```

### LLMNode Template Engine

Uses Go's `text/template` package with two syntax options:
- `{{.fieldName}}` - standard Go template syntax
- `${variable}` - alternative dollar-brace syntax (converted internally)
- Template data comes from `ObjectValue` fields

### Async Execution

`AgentExecutor` supports both sync and async execution:
```go
// Sync
response, err := exec.Exec(input)

// Async (returns channel)
resultChan := exec.ExecAsync(input)
result := <-resultChan
```

### Concurrency Patterns Used

- `goroutine` + `channel` for parallel node execution and result collection
- `sync.Map` for thread-safe node value storage
- `sync.WaitGroup` for batch execution in `GroupExecutor`
- `sync.Mutex` / `sync.RWMutex` for protecting struct fields

### Adding New Node Types

1. Embed `BaseNode` for common functionality
2. Implement `Exec(state *NodeState) (value.NodeValue, error)`
3. Optionally implement `GetNodeGraph()` for visualization
4. Create a Builder pattern for construction

### Error Handling Convention

Use `emperror.dev/errors` for all error definitions:
```go
import "emperror.dev/errors"

var ErrSomeError = errors.New("some error message")
```

### Import Paths

```go
import (
    ai_agent "github.com/chuccp/ai-agent"
    "github.com/chuccp/ai-agent/node"
    "github.com/chuccp/ai-agent/value"
    "github.com/chuccp/ai-agent/executor"
    "github.com/chuccp/ai-agent/graph"
    "github.com/chuccp/ai-agent/types"
)
```
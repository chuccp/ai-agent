package node

import (
	"testing"

	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/out"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/value"
	"github.com/stretchr/testify/assert"
)

// --- LLMNode Tests ---

func TestLLMNodeWithMockLLMFunction(t *testing.T) {
	llmNode := NewLLMNodeBuilder("llm").
		SystemTemplate("You are a helpful assistant.").
		UserTemplate("Hello, ${name}!").
		ValuesFrom(value.RootValueFrom("name", "name")).
		LLMFunction(func(nodeState *State, files *value.UrlsValue, systemPrompt, userPrompt string, format out.OutFormat, stream bool) (value.NodeValue, error) {
			assert.Equal(t, "You are a helpful assistant.", systemPrompt)
			assert.Equal(t, "Hello, Alice!", userPrompt)

			result := value.NewObjectValue()
			result.PutString("response", "Hello back!")
			return result, nil
		}).
		FormatOut(out.NewTextOutFormat()).
		CacheEnabled(false).
		Build()

	state := &State{
		workflowContext: &mockWorkflowContext{
			rootValue: value.NewObjectValue().PutString("name", "Alice"),
		},
		nodeID:     "llm",
		input:      value.NewObjectValue(),
		nodeStatus: graph.NewNodeStatus("llm"),
		Parameter:  value.NewObjectValue(),
	}

	result, err := llmNode.Exec(state)

	assert.NoError(t, err)
	assert.Equal(t, "Hello back!", result.(*value.ObjectValue).GetString("response"))
}

func TestLLMNodeNoTemplateError(t *testing.T) {
	llmNode := NewLLMNodeBuilder("llm").
		ValuesFrom(value.RootValueFrom("name", "name")).
		LLMFunction(func(nodeState *State, files *value.UrlsValue, systemPrompt, userPrompt string, format out.OutFormat, stream bool) (value.NodeValue, error) {
			return value.NullValue, nil
		}).
		FormatOut(out.NewTextOutFormat()).
		Build()

	// Without template, ExecuteTemplateWithDollarFormat on empty string returns "",
	// and empty userPrompt/systemTrigger triggers "template required" error
	rootObj := value.NewObjectValue()
	rootObj.PutString("name", "Alice")
	state := &State{
		workflowContext: &mockWorkflowContext{rootValue: rootObj},
		nodeID:          "llm",
		input:           value.NewObjectValue(),
		nodeStatus:      graph.NewNodeStatus("llm"),
		Parameter:       value.NewObjectValue(),
	}

	_, err := llmNode.Exec(state)

	// Both systemPrompt and userPrompt will be empty, triggering the template error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template")
}

func TestLLMNodeNoFormatOutError(t *testing.T) {
	llmNode := NewLLMNodeBuilder("llm").
		SystemTemplate("system").
		UserTemplate("Hello, ${name}!").
		ValuesFrom(value.RootValueFrom("name", "name")).
		LLMFunction(func(nodeState *State, files *value.UrlsValue, systemPrompt, userPrompt string, format out.OutFormat, stream bool) (value.NodeValue, error) {
			return value.NullValue, nil
		}).
		// No FormatOut set
		Build()

	state := &State{
		workflowContext: &mockWorkflowContext{rootValue: value.NewObjectValue().PutString("name", "Alice")},
		nodeID:          "llm",
		input:           value.NewObjectValue(),
		nodeStatus:      graph.NewNodeStatus("llm"),
		Parameter:       value.NewObjectValue(),
	}

	_, err := llmNode.Exec(state)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "formatOut is required")
}

func TestLLMNodeBuilder(t *testing.T) {
	llmNode := NewLLMNodeBuilder("llm").
		SystemTemplate("system").
		UserTemplate("user").
		Stream(true).
		CacheEnabled(true).
		ValuesFrom(value.RootValueFrom("input", "input")).
		Build()

	assert.Equal(t, "llm", llmNode.GetID())
	assert.True(t, llmNode.stream)
	assert.True(t, llmNode.cacheEnabled)
}

func TestLLMNodeGetNodeGraph(t *testing.T) {
	llmNode := NewLLMNodeBuilder("llm").
		ValuesFrom(value.RootValueFrom("input", "input")).
		LLMFunction(func(nodeState *State, files *value.UrlsValue, systemPrompt, userPrompt string, format out.OutFormat, stream bool) (value.NodeValue, error) {
			return value.NullValue, nil
		}).
		FormatOut(out.NewTextOutFormat()).
		Build()

	g := llmNode.GetNodeGraph()
	assert.Equal(t, "llm", g.GetID())
	assert.Equal(t, "LLMNode", g.GetType())
}

// --- IterationNode Tests ---

func TestIterationNodeBuilder(t *testing.T) {
	processItem := NewFunctionNodeBuilder("processItem").
		ExecFunc(func(state *State) (value.NodeValue, error) {
			res := value.NewObjectValue()
			res.PutNumber("result", state.GetRootValue().GetNumber("item"))
			return res, nil
		}).Build()

	workflow := newMockWorkflow(processItem)
	iterNode := NewIterationNodeBuilder("iterate").
		IterationFrom(value.NewValueFrom("", "items", "")).
		ValuesFrom(value.RootValueFrom("shared", "shared")).
		Workflow(workflow).
		Build()

	assert.Equal(t, "iterate", iterNode.GetID())
	assert.Equal(t, types.NodeTypeMultiple, iterNode.GetNodeType())
	assert.False(t, iterNode.IsSingle())
	assert.Len(t, iterNode.GetIterationFrom(), 1)
}

func TestIterationNodeMissingWorkflow(t *testing.T) {
	iterNode := NewIterationNodeBuilder("iterate").
		IterationFrom(value.NewValueFrom("", "items", "")).
		Build()

	state := &State{
		workflowContext: &mockWorkflowContext{
			rootValue: value.NewObjectValue(),
		},
		nodeID:     "iterate",
		input:      value.NewObjectValue(),
		nodeStatus: graph.NewNodeStatusGroup("iterate"),
		Parameter:  value.NewObjectValue(),
	}

	_, err := iterNode.Exec(state)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrIterationNodeWorkflowRequired)
}

func TestIterationNodeMissingIterationFrom(t *testing.T) {
	processItem := NewFunctionNodeBuilder("processItem").
		ExecFunc(func(state *State) (value.NodeValue, error) {
			return value.NewObjectValue(), nil
		}).Build()

	workflow := newMockWorkflow(processItem)
	iterNode := NewIterationNodeBuilder("iterate").
		Workflow(workflow).
		Build()

	state := &State{
		workflowContext: &mockWorkflowContext{
			rootValue: value.NewObjectValue(),
		},
		nodeID:     "iterate",
		input:      value.NewObjectValue(),
		nodeStatus: graph.NewNodeStatusGroup("iterate"),
		Parameter:  value.NewObjectValue(),
	}

	_, err := iterNode.Exec(state)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrIterationNodeIterationFromRequired)
}

func TestIterationNodeRequiresArrayInput(t *testing.T) {
	processItem := NewFunctionNodeBuilder("processItem").
		ExecFunc(func(state *State) (value.NodeValue, error) {
			return value.NewObjectValue(), nil
		}).Build()

	workflow := newMockWorkflow(processItem)
	iterNode := NewIterationNodeBuilder("iterate").
		IterationFrom(value.NewValueFrom("", "items", "")).
		Workflow(workflow).
		Build()

	// Provide non-array input
	rootObj := value.NewObjectValue()
	rootObj.PutString("items", "not-an-array")
	state := &State{
		workflowContext: &mockWorkflowContext{
			rootValue: rootObj,
		},
		nodeID:     "iterate",
		input:      value.NewObjectValue(),
		nodeStatus: graph.NewNodeStatusGroup("iterate"),
		Parameter:  value.NewObjectValue(),
	}

	_, err := iterNode.Exec(state)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrIterationNodeRequiresArrayInput)
}

func TestIterationNodeGetNodeGraph(t *testing.T) {
	processItem := NewFunctionNodeBuilder("processItem").Build()
	workflow := newMockWorkflow(processItem)
	iterNode := NewIterationNodeBuilder("iterate").
		IterationFrom(value.NewValueFrom("", "items", "")).
		Workflow(workflow).
		Build()

	g := iterNode.GetNodeGraph()
	assert.Equal(t, "iterate", g.GetID())
	assert.Equal(t, "IterationNode", g.GetType())
}

// --- OrderIterationNode Tests ---

func TestOrderIterationNodeBuilder(t *testing.T) {
	processItem := NewFunctionNodeBuilder("processItem").Build()
	workflow := newMockWorkflow(processItem)

	iterNode := NewOrderIterationNodeBuilder("order-iterate").
		IterationFrom(value.NewValueFrom("", "items", "")).
		Workflow(workflow).
		Build()

	assert.Equal(t, "order-iterate", iterNode.GetID())
	// NodeType is inherited from IterationNode
	assert.Equal(t, types.NodeTypeMultiple, iterNode.GetNodeType())
}

func TestOrderIterationNodeExec(t *testing.T) {
	processItem := NewFunctionNodeBuilder("processItem").
		ExecFunc(func(state *State) (value.NodeValue, error) {
			res := value.NewObjectValue()
			res.PutNumber("result", state.GetRootValue().GetNumber("item"))
			return res, nil
		}).Build()

	workflow := newMockWorkflow(processItem)
	iterNode := NewOrderIterationNodeBuilder("order-iterate").
		IterationFrom(value.NewValueFrom("", "items", "")).
		Workflow(workflow).
		Build()

	// Build array input
	items := value.NewArrayValue()
	for i := 1; i <= 3; i++ {
		itemObj := value.NewObjectValue()
		itemObj.PutNumber("item", float64(i))
		items.Add(itemObj)
	}

	rootObj := value.NewObjectValue()
	rootObj.Put("items", items)
	state := &State{
		workflowContext: &mockWorkflowContext{
			rootValue: rootObj,
		},
		nodeID:     "order-iterate",
		input:      value.NewObjectValue(),
		nodeStatus: graph.NewNodeStatusGroup("order-iterate"),
		Parameter:  value.NewObjectValue(),
	}

	result, err := iterNode.Exec(state)

	assert.NoError(t, err)
	assert.True(t, result.IsArray())
	assert.Equal(t, 3, result.(*value.ArrayValue).Size())
}

// --- ImageGenerationNode Tests ---

func TestImageGenerationNodeBuilder(t *testing.T) {
	imgNode := NewImageGenerationNodeBuilder("img").
		UserTemplate("Generate ${name}").
		ValuesFrom(value.RootValueFrom("name", "name")).
		Scale("16:9").
		MaxNumber(2).
		ImageGenerationFunction(func(state *State, urls *value.UrlsValue, userPrompt string, maxNumber int, scale string) (value.NodeValue, error) {
			result := value.NewObjectValue()
			result.PutString("prompt", userPrompt)
			result.PutString("scale", scale)
			return result, nil
		}).
		Build()

	assert.Equal(t, "img", imgNode.GetID())
}

func TestImageGenerationNodeExec(t *testing.T) {
	imgNode := NewImageGenerationNodeBuilder("img").
		UserTemplate("Generate ${name}").
		ValuesFrom(value.RootValueFrom("name", "name")).
		Scale("1:1").
		MaxNumber(1).
		ImageGenerationFunction(func(state *State, urls *value.UrlsValue, userPrompt string, maxNumber int, scale string) (value.NodeValue, error) {
			result := value.NewObjectValue()
			result.PutString("prompt", userPrompt)
			return result, nil
		}).
		Build()

	rootObj := value.NewObjectValue()
	rootObj.PutString("name", "a cat")
	state := &State{
		workflowContext: &mockWorkflowContext{
			rootValue: rootObj,
		},
		nodeID:     "img",
		input:      value.NewObjectValue(),
		nodeStatus: graph.NewNodeStatus("img"),
		Parameter:  value.NewObjectValue(),
	}

	result, err := imgNode.Exec(state)

	assert.NoError(t, err)
	assert.Equal(t, "Generate a cat", result.(*value.ObjectValue).GetString("prompt"))
}

func TestImageGenerationNodeEmptyPrompt(t *testing.T) {
	imgNode := NewImageGenerationNodeBuilder("img").
		UserTemplate("prompt").
		ValuesFrom().
		Scale("1:1").
		MaxNumber(1).
		ImageGenerationFunction(func(state *State, urls *value.UrlsValue, userPrompt string, maxNumber int, scale string) (value.NodeValue, error) {
			return value.NewObjectValue(), nil
		}).
		Build()

	state := &State{
		workflowContext: &mockWorkflowContext{
			rootValue: value.NewObjectValue(),
		},
		nodeID:     "img",
		input:      value.NewObjectValue(),
		nodeStatus: graph.NewNodeStatus("img"),
		Parameter:  value.NewObjectValue(),
	}

	// The template "prompt" resolves to "prompt" (no variables), so it should succeed
	result, err := imgNode.Exec(state)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// --- Helper: mock Workflow for IterationNode ---

type mockWorkflow struct {
	nodes []Node
}

func newMockWorkflow(nodes ...Node) *mockWorkflow {
	return &mockWorkflow{nodes: nodes}
}

func (m *mockWorkflow) Exec(ctx WorkflowContext) (value.NodeValue, error) {
	n := m.nodes[0]
	state := NewNodeState(ctx, n.GetID(), ctx.GetRootValue(), value.NewObjectValue())
	return n.Exec(state)
}

func (m *mockWorkflow) ExecBatch(ctx WorkflowContext, statusGroup *graph.NodeStatusGroup, parentID string, inputs []*value.ObjectValue) (value.NodeValue, error) {
	arr := value.NewArrayValue()
	for _, input := range inputs {
		childCtx := ctx.CreateChildContext(m.nodes, input, value.NewArrayValue(), parentID)
		n := m.nodes[0]
		state := NewNodeState(childCtx, n.GetID(), input, value.NewObjectValue())
		result, err := n.Exec(state)
		if err != nil {
			return nil, err
		}
		arr.Add(result)
	}
	return arr, nil
}

func (m *mockWorkflow) ExecBatchOrder(ctx WorkflowContext, statusGroup *graph.NodeStatusGroup, parentID string, inputs []*value.ObjectValue) (value.NodeValue, error) {
	return m.ExecBatch(ctx, statusGroup, parentID, inputs)
}

func (m *mockWorkflow) Execute(ctx WorkflowContext, input *value.ObjectValue, parentID string) (value.NodeValue, error) {
	childCtx := ctx.CreateChildContext(m.nodes, input, value.NewArrayValue(), parentID)
	return m.Exec(childCtx)
}

func (m *mockWorkflow) GetGraphs() []*graph.NodeGraph {
	graphs := make([]*graph.NodeGraph, 0, len(m.nodes))
	for _, n := range m.nodes {
		graphs = append(graphs, n.GetNodeGraph())
	}
	return graphs
}

package node

import (
	"testing"

	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/value"
	"github.com/stretchr/testify/assert"
)

// --- BaseNode Tests ---

func TestBaseNodeBasic(t *testing.T) {
	n := NewBaseNode("test", types.NodeTypeSingle)

	assert.Equal(t, "test", n.GetID())
	assert.True(t, n.IsSingle())
	assert.Equal(t, types.NodeTypeSingle, n.GetNodeType())
}

func TestBaseNodeMultipleType(t *testing.T) {
	n := NewBaseNode("multi", types.NodeTypeMultiple)

	assert.False(t, n.IsSingle())
	assert.Equal(t, types.NodeTypeMultiple, n.GetNodeType())
}

func TestBaseNodePrevNodeID(t *testing.T) {
	n := NewBaseNode("test", types.NodeTypeSingle)
	n.SetPrevNodeID("prev")
	assert.Equal(t, "prev", n.GetPrevNodeID())
}

func TestBaseNodeValuesFrom(t *testing.T) {
	n := NewBaseNode("test", types.NodeTypeSingle)
	vf := value.NewValueFrom("source", "$.data", "data")
	n.SetValuesFrom([]*value.ValueFrom{vf})

	assert.Len(t, n.GetValuesFrom(), 1)
	assert.Equal(t, "source", n.GetValuesFrom()[0].NodeID)
}

func TestBaseNodeGetNodeGraph(t *testing.T) {
	n := NewBaseNode("test", types.NodeTypeSingle)
	g := n.GetNodeGraph()

	assert.Equal(t, "test", g.GetID())
	assert.Equal(t, "BaseNode", g.GetType())
}

// --- InputNode Tests ---

func TestInputNode(t *testing.T) {
	n := NewInputNodeBuilder("input").
		ValuesFrom(value.RootValueFrom("data", "data")).
		Build()

	assert.Equal(t, "input", n.GetID())
	g := n.GetNodeGraph()
	assert.Equal(t, "InputNode", g.GetType())
}

// --- OutputNode Tests ---

func TestOutputNode(t *testing.T) {
	var captured value.NodeValue
	n := NewOutputNodeBuilder("output").
		ValuesFrom(value.NewValueFrom("prev", "", "")).
		OutFunc(func(nodeValue value.NodeValue) {
			captured = nodeValue
		}).
		Build()

	inputObj := value.NewObjectValue().PutString("key", "value")
	state := &State{
		workflowContext: &mockWorkflowContext{
			rootValue:  value.NewObjectValue(),
			nodeValues: map[string]value.NodeValue{"prev": inputObj},
		},
		nodeID:     "output",
		input:      value.NewObjectValue(),
		nodeStatus: graph.NewNodeStatus("output"),
		Parameter:  value.NewObjectValue(),
	}

	result, err := n.Exec(state)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, captured.IsObject())
}

// --- FunctionNode Tests ---

func TestFunctionNode(t *testing.T) {
	n := NewFunctionNodeBuilder("fn").
		ExecFunc(func(state *State) (value.NodeValue, error) {
			res := value.NewObjectValue()
			res.PutString("result", "done")
			return res, nil
		}).
		ValuesFrom(value.RootValueFrom("input", "input")).
		Build()

	assert.Equal(t, "fn", n.GetID())
	assert.True(t, n.IsSingle())

	state := &State{
		workflowContext: &mockWorkflowContext{rootValue: value.NewObjectValue()},
		nodeID:          "fn",
		input:           value.NewObjectValue(),
		nodeStatus:      graph.NewNodeStatus("fn"),
		Parameter:       value.NewObjectValue(),
	}

	result, err := n.Exec(state)

	assert.NoError(t, err)
	assert.Equal(t, "done", result.(*value.ObjectValue).GetString("result"))
}

func TestFunctionNodeNilFunc(t *testing.T) {
	n := NewFunctionNode("test", nil)
	state := &State{
		workflowContext: &mockWorkflowContext{rootValue: value.NewObjectValue()},
		nodeID:          "test",
		input:           value.NewObjectValue(),
		nodeStatus:      graph.NewNodeStatus("test"),
		Parameter:       value.NewObjectValue(),
	}

	_, err := n.Exec(state)

	assert.Error(t, err)
}

func TestFunctionNodeBuilderValuesFrom(t *testing.T) {
	n := NewFunctionNodeBuilder("fn").
		ValuesFrom(
			value.NewValueFrom("a", "", ""),
			value.NewValueFrom("b", "", ""),
		).
		Build()

	assert.Len(t, n.ValuesFrom, 2)
}

// --- ValueFrom Helper Tests ---

func TestNewValueFrom(t *testing.T) {
	vf := value.NewValueFrom("node1", "$.field", "myField")
	assert.Equal(t, "node1", vf.NodeID)
	assert.Equal(t, "$.field", vf.From)
	assert.Equal(t, "myField", vf.Param)
}

func TestNewValueFromNodeAll(t *testing.T) {
	vf := value.NewValueFromNodeAll("node1")
	assert.Equal(t, "node1", vf.NodeID)
	assert.Equal(t, "", vf.From)
	assert.Equal(t, "", vf.Param)
}

func TestRootValueFrom(t *testing.T) {
	vf := value.RootValueFrom("$.data", "data")
	assert.Equal(t, "", vf.NodeID)
	assert.Equal(t, "$.data", vf.From)
	assert.Equal(t, "data", vf.Param)
}

func TestNodeValueFrom(t *testing.T) {
	vf := value.NodeValueFrom("node1", "$.field", "field")
	assert.Equal(t, "node1", vf.NodeID)
	assert.Equal(t, "$.field", vf.From)
	assert.Equal(t, "field", vf.Param)
}

func TestParseValueFrom(t *testing.T) {
	vf := value.ParseValueFrom("node1$.field", "field")
	assert.Equal(t, "node1", vf.NodeID)
	assert.Equal(t, "$.field", vf.From)
	assert.Equal(t, "field", vf.Param)
}

func TestParseValueFromNoDollar(t *testing.T) {
	vf := value.ParseValueFrom("node1", "field")
	assert.Equal(t, "node1", vf.NodeID)
	assert.Equal(t, "", vf.From)
	assert.Equal(t, "field", vf.Param)
}

func TestValueFromString(t *testing.T) {
	vf := value.NewValueFrom("node1", "$.field", "field")
	s := vf.String()
	assert.Contains(t, s, "node1")
	assert.Contains(t, s, "$.field")
	assert.Contains(t, s, "field")
}

// --- State Tests ---

func TestStateDefaults(t *testing.T) {
	state := NewNodeState(&mockWorkflowContext{rootValue: value.NewObjectValue()}, "test", nil, nil)

	assert.Equal(t, "test", state.GetID())
	assert.NotNil(t, state.GetInput())
	assert.NotNil(t, state.GetParameter())
	assert.NotNil(t, state.GetRootValue())
}

func TestStateNilWorkflowContext(t *testing.T) {
	// NewNodeState with nil workflowContext panics (GetNodeStatus called on nil)
	// So we create a state with a minimal mock instead
	state := NewNodeState(&mockWorkflowContext{rootValue: value.NewObjectValue()}, "test", value.NewObjectValue(), value.NewObjectValue())
	// Override with nil workflowContext to test nil-safe methods
	state.workflowContext = nil

	assert.Nil(t, state.GetRootValue())
	assert.False(t, state.IsCacheEnabled())
	assert.Equal(t, "", state.GetParentID())
	assert.Equal(t, "", state.GetCachePath())
	assert.NoError(t, state.SaveCache("key", value.NewTextValue("v")))
	_, err := state.GetCache("key")
	assert.NoError(t, err)
	assert.False(t, state.HasCache("key"))
}

func TestStateStatusTransitions(t *testing.T) {
	state := NewNodeState(&mockWorkflowContext{rootValue: value.NewObjectValue()}, "test", value.NewObjectValue(), value.NewObjectValue())

	assert.True(t, state.IsWaiting())
	assert.False(t, state.IsStarted())
	assert.False(t, state.IsSucceeded())
	assert.False(t, state.IsFailed())

	state.SetStatusType(types.NodeStatusStarted)
	assert.True(t, state.IsStarted())

	state.SetStatusType(types.NodeStatusSucceeded)
	assert.True(t, state.IsSucceeded())
	assert.False(t, state.IsRunning())
}

// --- Mock WorkflowContext ---

type mockWorkflowContext struct {
	rootValue *value.ObjectValue
	nodeValues map[string]value.NodeValue
}

func (m *mockWorkflowContext) GetRootValue() *value.ObjectValue {
	if m.rootValue == nil {
		return value.NewObjectValue()
	}
	return m.rootValue
}
func (m *mockWorkflowContext) GetShareValue() *value.ArrayValue { return value.NewArrayValue() }
func (m *mockWorkflowContext) GetParentID() string              { return "" }
func (m *mockWorkflowContext) GetNodeValue(nodeID string) value.NodeValue {
	if m.nodeValues != nil {
		return m.nodeValues[nodeID]
	}
	return nil
}
func (m *mockWorkflowContext) GetNodeStatus(nodeID string) graph.NodeStatusInterface {
	return graph.NewNodeStatus(nodeID)
}
func (m *mockWorkflowContext) AddNodeValue(nodeID string, nodeValue value.NodeValue) {}
func (m *mockWorkflowContext) IsCacheEnabled() bool                                  { return false }
func (m *mockWorkflowContext) GetCachePath() string                                  { return "" }
func (m *mockWorkflowContext) GetCacheKey(key, nodeID string) string                 { return "" }
func (m *mockWorkflowContext) SaveCache(key, nodeID string, nodeValue value.NodeValue) error {
	return nil
}
func (m *mockWorkflowContext) GetCache(key, nodeID string) (value.NodeValue, error) {
	return nil, nil
}
func (m *mockWorkflowContext) HasCache(key, nodeID string) bool                     { return false }
func (m *mockWorkflowContext) CreateChildContext(nodes []Node, childRootValue *value.ObjectValue, shareValue *value.ArrayValue, childParentID string) WorkflowContext {
	return &mockWorkflowContext{rootValue: childRootValue}
}

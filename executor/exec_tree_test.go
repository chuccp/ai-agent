package executor

import (
	"testing"

	"github.com/chuccp/ai-agent/node"
	"github.com/chuccp/ai-agent/value"
	"github.com/stretchr/testify/assert"
)

// --- BuildExecutionLayers Tests ---

func TestBuildExecutionLayersLinear(t *testing.T) {
	// nodeA -> nodeB -> nodeC (linear chain)
	nodeA := node.NewFunctionNodeBuilder("nodeA").Build()
	nodeB := node.NewFunctionNodeBuilder("nodeB").
		ValuesFrom(value.NewValueFrom("nodeA", "", "")).
		Build()
	nodeC := node.NewFunctionNodeBuilder("nodeC").
		ValuesFrom(value.NewValueFrom("nodeB", "", "")).
		Build()

	nodeMap := map[string]node.Node{
		"nodeA": nodeA,
		"nodeB": nodeB,
		"nodeC": nodeC,
	}

	layers, err := BuildExecutionLayers(nodeMap, nodeC)

	assert.NoError(t, err)
	assert.Len(t, layers, 3)
	// Layer 1 should have nodeA, Layer 2 nodeB, Layer 3 nodeC
	layerNodeIDs := make([]string, 0, 3)
	for _, layer := range layers {
		for _, n := range layer {
			layerNodeIDs = append(layerNodeIDs, n.GetID())
		}
	}
	// Verify ordering: nodeA before nodeB before nodeC
	assert.Equal(t, "nodeA", layerNodeIDs[0])
	assert.Equal(t, "nodeB", layerNodeIDs[1])
	assert.Equal(t, "nodeC", layerNodeIDs[2])
}

func TestBuildExecutionLayersParallel(t *testing.T) {
	// nodeA and nodeB independent, nodeC depends on both
	nodeA := node.NewFunctionNodeBuilder("nodeA").Build()
	nodeB := node.NewFunctionNodeBuilder("nodeB").Build()
	nodeC := node.NewFunctionNodeBuilder("nodeC").
		ValuesFrom(
			value.NewValueFrom("nodeA", "", ""),
			value.NewValueFrom("nodeB", "", ""),
		).
		Build()

	nodeMap := map[string]node.Node{
		"nodeA": nodeA,
		"nodeB": nodeB,
		"nodeC": nodeC,
	}

	layers, err := BuildExecutionLayers(nodeMap, nodeC)

	assert.NoError(t, err)
	assert.Len(t, layers, 2)
	// Layer 1: nodeA, nodeB (parallel)
	assert.Len(t, layers[0], 2)
	// Layer 2: nodeC
	assert.Len(t, layers[1], 1)
	assert.Equal(t, "nodeC", layers[1][0].GetID())
}

func TestBuildExecutionLayersDiamond(t *testing.T) {
	// Diamond pattern: A -> B, A -> C, B -> D, C -> D
	nodeA := node.NewFunctionNodeBuilder("nodeA").Build()
	nodeB := node.NewFunctionNodeBuilder("nodeB").
		ValuesFrom(value.NewValueFrom("nodeA", "", "")).
		Build()
	nodeC := node.NewFunctionNodeBuilder("nodeC").
		ValuesFrom(value.NewValueFrom("nodeA", "", "")).
		Build()
	nodeD := node.NewFunctionNodeBuilder("nodeD").
		ValuesFrom(
			value.NewValueFrom("nodeB", "", ""),
			value.NewValueFrom("nodeC", "", ""),
		).
		Build()

	nodeMap := map[string]node.Node{
		"nodeA": nodeA,
		"nodeB": nodeB,
		"nodeC": nodeC,
		"nodeD": nodeD,
	}

	layers, err := BuildExecutionLayers(nodeMap, nodeD)

	assert.NoError(t, err)
	assert.Len(t, layers, 3)
	// Layer 1: [A], Layer 2: [B, C], Layer 3: [D]
	assert.Len(t, layers[0], 1)
	assert.Equal(t, "nodeA", layers[0][0].GetID())
	assert.Len(t, layers[1], 2)
	assert.Len(t, layers[2], 1)
	assert.Equal(t, "nodeD", layers[2][0].GetID())
}

func TestBuildExecutionLayersSingleNode(t *testing.T) {
	n := node.NewFunctionNodeBuilder("only").Build()
	nodeMap := map[string]node.Node{"only": n}

	layers, err := BuildExecutionLayers(nodeMap, n)

	assert.NoError(t, err)
	assert.Len(t, layers, 1)
	assert.Len(t, layers[0], 1)
}

func TestBuildExecutionLayersUnknownDependency(t *testing.T) {
	nodeA := node.NewFunctionNodeBuilder("nodeA").
		ValuesFrom(value.NewValueFrom("nonexistent", "", "")).
		Build()
	nodeB := node.NewFunctionNodeBuilder("nodeB").
		ValuesFrom(value.NewValueFrom("nodeA", "", "")).
		Build()

	nodeMap := map[string]node.Node{
		"nodeA": nodeA,
		"nodeB": nodeB,
	}

	_, err := BuildExecutionLayers(nodeMap, nodeB)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown dependency")
}

// --- ExecNodeTree Tests ---

func TestExecNodeTreeNext(t *testing.T) {
	execNodes := NewExecNodes()

	en1 := NewExecNode("a")
	en2 := NewExecNode("b")
	en2.AddPrevNodeID("a")
	en3 := NewExecNode("c")
	en3.AddPrevNodeID("b")

	execNodes.AddNode(en1)
	execNodes.AddNode(en2)
	execNodes.AddNode(en3)

	tree := execNodes.Tree()

	// First call: only "a" has no dependencies
	first := tree.Next()
	assert.Len(t, first, 1)
	assert.Equal(t, "a", first[0])

	// Second call: "b" now has no dependencies
	second := tree.Next()
	assert.Len(t, second, 1)
	assert.Equal(t, "b", second[0])

	// Third call: "c"
	third := tree.Next()
	assert.Len(t, third, 1)
	assert.Equal(t, "c", third[0])

	// No more
	assert.False(t, tree.HasNext())
}

func TestExecNodeCopy(t *testing.T) {
	en := NewExecNode("a")
	en.AddPrevNodeID("b")
	en.AddPrevNodeID("c")

	cp := en.Copy()
	assert.Equal(t, "a", cp.NodeID)
	assert.Len(t, cp.PrevNodeIDs, 2)
	assert.True(t, cp.PrevNodeIDs["b"])
	assert.True(t, cp.PrevNodeIDs["c"])
}

func TestExecNodesContains(t *testing.T) {
	execNodes := NewExecNodes()
	en := NewExecNode("test")
	execNodes.AddNode(en)

	assert.True(t, execNodes.Contains("test"))
	assert.False(t, execNodes.Contains("nonexistent"))
}

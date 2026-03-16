package executor

import (
	"emperror.dev/errors"

	"github.com/chuccp/ai-agent/node"
	"github.com/chuccp/ai-agent/util"
)

// ExecNode 执行节点
type ExecNode struct {
	NodeID      string
	PrevNodeIDs map[string]bool
}

// NewExecNode 创建执行节点
func NewExecNode(nodeID string) *ExecNode {
	return &ExecNode{
		NodeID:      nodeID,
		PrevNodeIDs: make(map[string]bool),
	}
}

// AddPrevNodeID 添加前置节点ID
func (n *ExecNode) AddPrevNodeID(nodeID string) {
	n.PrevNodeIDs[nodeID] = true
}

// GetPrevNodeIDs 获取前置节点ID列表
func (n *ExecNode) GetPrevNodeIDs() []string {
	ids := make([]string, 0, len(n.PrevNodeIDs))
	for id := range n.PrevNodeIDs {
		ids = append(ids, id)
	}
	return ids
}

// Copy 复制节点
func (n *ExecNode) Copy() *ExecNode {
	newNode := NewExecNode(n.NodeID)
	for id := range n.PrevNodeIDs {
		newNode.PrevNodeIDs[id] = true
	}
	return newNode
}

// ExecNodes 执行节点列表
type ExecNodes struct {
	nodes   []*ExecNode
	nodeIDs map[string]bool
}

// NewExecNodes 创建执行节点列表
func NewExecNodes() *ExecNodes {
	return &ExecNodes{
		nodes:   make([]*ExecNode, 0),
		nodeIDs: make(map[string]bool),
	}
}

// AddNode 添加节点
func (e *ExecNodes) AddNode(node *ExecNode) {
	e.nodeIDs[node.NodeID] = true
	e.nodes = append(e.nodes, node)
}

// Contains 是否包含节点
func (e *ExecNodes) Contains(nodeID string) bool {
	return e.nodeIDs[nodeID]
}

// Tree 创建执行树
func (e *ExecNodes) Tree() *ExecNodeTree {
	nodes := make([]*ExecNode, 0, len(e.nodes))
	for _, n := range e.nodes {
		nodes = append(nodes, n.Copy())
	}
	return NewExecNodeTree(nodes)
}

// ExecNodeTree 执行节点树
type ExecNodeTree struct {
	nodes []*ExecNode
}

// NewExecNodeTree 创建执行节点树
func NewExecNodeTree(nodes []*ExecNode) *ExecNodeTree {
	return &ExecNodeTree{
		nodes: nodes,
	}
}

// HasNext 是否有下一个
func (t *ExecNodeTree) HasNext() bool {
	return len(t.nodes) > 0
}

// Next 获取下一批可执行的节点ID
func (t *ExecNodeTree) Next() []string {
	nodeIDs := make([]string, 0)
	for _, n := range t.nodes {
		if len(n.PrevNodeIDs) == 0 {
			nodeIDs = append(nodeIDs, n.NodeID)
		}
	}

	// 移除已执行的节点
	newNodes := make([]*ExecNode, 0)
	for _, n := range t.nodes {
		if len(n.PrevNodeIDs) > 0 {
			// 移除已执行的前置节点
			for _, id := range nodeIDs {
				delete(n.PrevNodeIDs, id)
			}
			newNodes = append(newNodes, n)
		}
	}
	t.nodes = newNodes

	return nodeIDs
}

// BuildExecutionLayers 构建执行层级
func BuildExecutionLayers(nodeMap map[string]node.Node, endNode node.Node) ([][]node.Node, error) {
	// 创建执行节点
	execNodes := NewExecNodes()
	err := createNodeTrees(endNode, nodeMap, execNodes)
	if err != nil {
		return nil, err
	}

	// 构建层级
	tree := execNodes.Tree()
	layers := make([][]node.Node, 0)

	for tree.HasNext() {
		nodeIDs := tree.Next()
		if len(nodeIDs) == 0 {
			return nil, ErrCircularDependency
		}

		layer := make([]node.Node, 0, len(nodeIDs))
		for _, id := range nodeIDs {
			n, ok := nodeMap[id]
			if !ok {
				return nil, NewUnknownDependencyError(id)
			}
			layer = append(layer, n)
		}
		layers = append(layers, layer)
	}

	return layers, nil
}

// createNodeTrees 创建节点树
func createNodeTrees(n node.Node, nodeMap map[string]node.Node, execNodes *ExecNodes) error {
	if !execNodes.Contains(n.GetID()) {
		execNode := NewExecNode(n.GetID())
		var set = false
		// 处理ValuesFrom
		valuesFrom := n.GetValuesFrom()
		if valuesFrom != nil && len(valuesFrom) > 0 {
			set = true
			for _, vf := range valuesFrom {
				if util.IsNotBlank(vf.NodeID) {
					execNode.AddPrevNodeID(vf.NodeID)
				}
			}
		}
		// 处理IterationNode的IterationFrom
		if iterNode, ok := n.(*node.IterationNode); ok {
			iterationFrom := iterNode.GetIterationFrom()
			if iterationFrom != nil && len(iterationFrom) > 0 {
				set = true
				for _, vf := range iterationFrom {
					if util.IsNotBlank(vf.NodeID) {
						execNode.AddPrevNodeID(vf.NodeID)
					}
				}
			}
		}

		if !set {
			id := n.GetPrevNodeID()
			if util.IsNotBlank(id) {
				execNode.AddPrevNodeID(id)
			}
		}

		execNodes.AddNode(execNode)

		// 递归处理前置节点
		for _, prevNodeID := range execNode.GetPrevNodeIDs() {
			prevNode, ok := nodeMap[prevNodeID]
			if !ok {
				return NewUnknownDependencyError(prevNodeID)
			}
			if err := createNodeTrees(prevNode, nodeMap, execNodes); err != nil {
				return err
			}
		}
	}
	return nil
}

// 错误定义
var (
	ErrCircularDependency = errors.New("circular dependency detected")
)

// NewUnknownDependencyError 创建未知依赖错误
func NewUnknownDependencyError(nodeID string) error {
	return errors.New("unknown dependency node: " + nodeID)
}

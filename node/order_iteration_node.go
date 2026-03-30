package node

import (
	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/value"
)

// OrderIterationNode 顺序处理迭代节点，不并发，值共享
type OrderIterationNode struct {
	*IterationNode
}

// NewOrderIterationNode 创建顺序迭代节点
func NewOrderIterationNode(id string) *OrderIterationNode {
	return &OrderIterationNode{
		IterationNode: NewIterationNode(id),
	}
}

// GetNodeType 获取节点类型
func (n *OrderIterationNode) GetNodeType() types.NodeType {
	return n.IterationNode.GetNodeType()
}

// GetNodeGraph 获取节点图
func (n *OrderIterationNode) GetNodeGraph() *graph.NodeGraph {
	return n.IterationNode.GetNodeGraph()
}

// Exec 执行节点 - 顺序处理每个迭代项
func (n *OrderIterationNode) Exec(state *State) (value.NodeValue, error) {
	batchInputs, currentParentID, statusGroup, err := n.PrepareBatchInputs(state)
	if err != nil {
		return nil, err
	}
	result, err := n.workflow.ExecBatchOrder(state.GetWorkflowContext(), statusGroup, currentParentID, batchInputs)
	if err != nil {
		state.SetStatusType(types.NodeStatusFailed)
		return nil, err
	}
	if result == nil || result.IsNull() {
		state.SetStatusType(types.NodeStatusRunning)
	}
	return result, nil
}

// OrderIterationNodeBuilder 顺序迭代节点构建器
type OrderIterationNodeBuilder struct {
	node *OrderIterationNode
}

// NewOrderIterationNodeBuilder 创建顺序迭代节点构建器
func NewOrderIterationNodeBuilder(id string) *OrderIterationNodeBuilder {
	return &OrderIterationNodeBuilder{
		node: NewOrderIterationNode(id),
	}
}

// Workflow 设置工作流
func (b *OrderIterationNodeBuilder) Workflow(workflow WorkflowInterface) *OrderIterationNodeBuilder {
	b.node.SetWorkflow(workflow)
	return b
}

// ValuesFrom 设置值来源
func (b *OrderIterationNodeBuilder) ValuesFrom(valuesFrom ...*value.ValueFrom) *OrderIterationNodeBuilder {
	b.node.ValuesFrom = append(b.node.ValuesFrom, valuesFrom...)
	return b
}

// IterationFrom 设置迭代来源
func (b *OrderIterationNodeBuilder) IterationFrom(iterationFrom ...*value.ValueFrom) *OrderIterationNodeBuilder {
	b.node.SetIterationFrom(append(b.node.GetIterationFrom(), iterationFrom...))
	return b
}

// Build 构建
func (b *OrderIterationNodeBuilder) Build() *OrderIterationNode {
	return b.node
}

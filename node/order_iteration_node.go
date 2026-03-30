package node

import (
	"strconv"

	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/util"
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
	return types.NodeTypeMultiple
}

// GetNodeGraph 获取节点图
func (n *OrderIterationNode) GetNodeGraph() *graph.NodeGraph {
	var children []*graph.NodeGraph
	if n.GetWorkflow() != nil {
		children = n.GetWorkflow().GetGraphs()
	}
	return graph.NewNodeGraphWithChildren(n.ID, "OrderIterationNode", n.ValuesFrom, n.GetIterationFrom(), children)
}

// Exec 执行节点 - 顺序处理每个迭代项
func (n *OrderIterationNode) Exec(state *State) (value.NodeValue, error) {
	workflow := n.GetWorkflow()
	if workflow == nil {
		return nil, ErrIterationNodeWorkflowRequired
	}

	// 展开迭代输入 (复用嵌入的 IterationNode 方法)
	batchInputs, err := n.IterationNode.ExpandIterationInputs(state)
	if err != nil {
		return nil, err
	}

	// 解析共享的 ValuesFrom 输入
	sharedInput, err := n.ParseValuesFromWithError(state, n.ValuesFrom)
	if err != nil {
		return nil, err
	}
	for _, input := range batchInputs {
		input.AddAllIFNULL(sharedInput)
	}

	// 构建父ID
	workflowParentID := state.GetWorkflowContext().GetParentID()
	currentParentID := n.ID
	if util.IsNotBlank(workflowParentID) {
		currentParentID = workflowParentID + "_" + n.ID
	}

	// 顺序执行每个迭代
	results := value.NewArrayValue()
	ctx := state.GetWorkflowContext()

	for i, input := range batchInputs {
		// 构建迭代的 parentID
		iterParentID := buildOrderIterParentID(currentParentID, i)
		// 执行子工作流
		result, err := workflow.Execute(ctx, input, iterParentID)
		if err != nil {
			state.SetStatusType(types.NodeStatusFailed)
			return nil, err
		}

		if result == nil || result.IsNull() {
			state.SetStatusType(types.NodeStatusRunning)
			return value.NullValue, nil
		}

		results.Add(result)

		// 值共享：将当前迭代的结果添加到上下文，供后续迭代使用
		// 使用迭代索引作为节点ID
		iterNodeID := iterParentID
		ctx.AddNodeValue(iterNodeID, result)
	}

	state.SetStatusType(types.NodeStatusSucceeded)
	return results, nil
}

// buildOrderIterParentID 构建顺序迭代的父ID
func buildOrderIterParentID(parentID string, index int) string {
	if parentID == "" {
		return strconv.Itoa(index)
	}
	return parentID + "_" + strconv.Itoa(index)
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

// 错误定义 (复用 IterationNode 的错误定义，并添加顺序迭代特有的错误)
var (
	ErrOrderIterationNodeExecuteFailed = errors.New("OrderIterationNode execution failed")
)

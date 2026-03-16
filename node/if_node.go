package node

import (
	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/util"
	"github.com/chuccp/ai-agent/value"
)

// ConditionFunc 条件函数
type ConditionFunc func(ctx WorkflowContext) bool

// IFNode 条件节点
type IFNode struct {
	*BaseNode
	condition    ConditionFunc
	thenWorkflow WorkflowInterface
	elseWorkflow WorkflowInterface
}

// NewIFNode 创建条件节点
func NewIFNode(id string) *IFNode {
	return &IFNode{
		BaseNode: NewBaseNode(id, types.NodeTypeSingle),
	}
}

// SetCondition 设置条件函数
func (n *IFNode) SetCondition(condition ConditionFunc) {
	n.condition = condition
}

// SetThenWorkflow 设置then工作流
func (n *IFNode) SetThenWorkflow(workflow WorkflowInterface) {
	n.thenWorkflow = workflow
}

// SetElseWorkflow 设置else工作流
func (n *IFNode) SetElseWorkflow(workflow WorkflowInterface) {
	n.elseWorkflow = workflow
}

// GetThenWorkflow 获取then工作流
func (n *IFNode) GetThenWorkflow() WorkflowInterface {
	return n.thenWorkflow
}

// GetElseWorkflow 获取else工作流
func (n *IFNode) GetElseWorkflow() WorkflowInterface {
	return n.elseWorkflow
}

// Exec 执行节点
func (n *IFNode) Exec(state *State) (value.NodeValue, error) {
	if n.condition == nil {
		return nil, ErrIFNodeConditionRequired
	}

	if n.thenWorkflow == nil && n.elseWorkflow == nil {
		return nil, ErrIFNodeWorkflowRequired
	}

	// 解析输入
	input, err := n.ParseValuesFromWithError(state, n.ValuesFrom)
	if err != nil {
		return nil, err
	}

	// 评估条件
	conditionResult := n.condition(state.GetWorkflowContext())

	workflowParentID := state.GetWorkflowContext().GetParentID()

	if conditionResult {
		if n.thenWorkflow == nil {
			return value.NullValue, nil
		}
		// 执行then工作流
		parentID := n.ID + "_" + "true"
		if util.IsNotBlank(workflowParentID) {
			parentID = workflowParentID + "_" + parentID
		}
		result, err := n.thenWorkflow.Execute(state.GetWorkflowContext(), input, parentID)
		if err != nil {
			state.SetStatusType(types.NodeStatusFailed)
			return nil, err
		}
		return result, nil
	}

	// 条件为false，执行else工作流
	if n.elseWorkflow == nil {
		return value.NullValue, nil
	}

	parentID := n.ID + "_" + "false"
	if util.IsNotBlank(workflowParentID) {
		parentID = workflowParentID + "_" + parentID
	}
	result, err := n.elseWorkflow.Execute(state.GetWorkflowContext(), input, parentID)
	if err != nil {
		state.SetStatusType(types.NodeStatusFailed)
		return nil, err
	}
	return result, nil
}

// GetNodeGraph 获取节点图
func (n *IFNode) GetNodeGraph() *graph.NodeGraph {
	var children []*graph.NodeGraph
	if n.thenWorkflow != nil {
		children = append(children, n.thenWorkflow.GetGraphs()...)
	}
	if n.elseWorkflow != nil {
		children = append(children, n.elseWorkflow.GetGraphs()...)
	}
	return graph.NewNodeGraphWithChildren(n.ID, "IFNode", n.ValuesFrom, nil, children)
}

// IFNodeBuilder 条件节点构建器
type IFNodeBuilder struct {
	node *IFNode
}

// NewIFNodeBuilder 创建条件节点构建器
func NewIFNodeBuilder(id string) *IFNodeBuilder {
	return &IFNodeBuilder{
		node: NewIFNode(id),
	}
}

// Condition 设置条件函数
func (b *IFNodeBuilder) Condition(condition ConditionFunc) *IFNodeBuilder {
	b.node.condition = condition
	return b
}

// Then 设置then工作流
func (b *IFNodeBuilder) Then(workflow WorkflowInterface) *IFNodeBuilder {
	b.node.thenWorkflow = workflow
	return b
}

// Else 设置else工作流
func (b *IFNodeBuilder) Else(workflow WorkflowInterface) *IFNodeBuilder {
	b.node.elseWorkflow = workflow
	return b
}

// ValuesFrom 设置值来源
func (b *IFNodeBuilder) ValuesFrom(valuesFrom ...*value.ValueFrom) *IFNodeBuilder {
	b.node.ValuesFrom = append(b.node.ValuesFrom, valuesFrom...)
	return b
}

// Build 构建
func (b *IFNodeBuilder) Build() (*IFNode, error) {
	if b.node.condition == nil {
		return nil, ErrIFNodeConditionRequired
	}
	if b.node.thenWorkflow == nil && b.node.elseWorkflow == nil {
		return nil, ErrIFNodeWorkflowRequired
	}
	return b.node, nil
}

// 错误定义
var (
	ErrIFNodeConditionRequired = errors.New("IFNode condition is required")
	ErrIFNodeWorkflowRequired  = errors.New("IFNode requires at least one workflow (then or else)")
)

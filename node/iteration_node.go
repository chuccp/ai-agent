package node

import (
	"strconv"

	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/util"
	"github.com/chuccp/ai-agent/value"
)

// IterationNode 迭代节点
type IterationNode struct {
	*BaseNode
	workflow      WorkflowInterface
	iterationFrom []*value.ValueFrom
}

// NewIterationNode 创建迭代节点
func NewIterationNode(id string) *IterationNode {
	return &IterationNode{
		BaseNode: NewBaseNode(id, types.NodeTypeMultiple),
	}
}

// GetNodeType 获取节点类型
func (n *IterationNode) GetNodeType() types.NodeType {
	return types.NodeTypeMultiple
}

// GetNodeGraph 获取节点图
func (n *IterationNode) GetNodeGraph() *graph.NodeGraph {
	var children []*graph.NodeGraph
	if n.workflow != nil {
		children = n.workflow.GetGraphs()
	}
	return graph.NewNodeGraphWithChildren(n.ID, "IterationNode", n.ValuesFrom, n.iterationFrom, children)
}

// SetIterationFrom 设置迭代来源
func (n *IterationNode) SetIterationFrom(iterationFrom []*value.ValueFrom) {
	n.iterationFrom = iterationFrom
}

// GetIterationFrom 获取迭代来源
func (n *IterationNode) GetIterationFrom() []*value.ValueFrom {
	return n.iterationFrom
}

// SetWorkflow 设置工作流
func (n *IterationNode) SetWorkflow(workflow WorkflowInterface) {
	n.workflow = workflow
}

// GetWorkflow 获取工作流
func (n *IterationNode) GetWorkflow() WorkflowInterface {
	return n.workflow
}

// BuildCurrentParentID 构建当前节点的父ID（供 IterationNode 和 OrderIterationNode 复用）
func (n *IterationNode) BuildCurrentParentID(state *State) string {
	workflowParentID := state.GetWorkflowContext().GetParentID()
	currentParentID := n.ID
	if util.IsNotBlank(workflowParentID) {
		currentParentID = workflowParentID + "_" + n.ID
	}
	return currentParentID
}

// BuildIterParentID 构建迭代的父ID（供 OrderIterationNode 复用）
func BuildIterParentID(parentID string, index int) string {
	if parentID == "" {
		return strconv.Itoa(index)
	}
	return parentID + "_" + strconv.Itoa(index)
}

// Exec 执行节点
func (n *IterationNode) Exec(state *State) (value.NodeValue, error) {
	batchInputs, currentParentID, statusGroup, err := n.PrepareBatchInputs(state)
	if err != nil {
		return nil, err
	}

	result, err := n.workflow.ExecBatch(state.GetWorkflowContext(), statusGroup, currentParentID, batchInputs)
	if err != nil {
		state.SetStatusType(types.NodeStatusFailed)
		return nil, err
	}
	if result == nil || result.IsNull() {
		state.SetStatusType(types.NodeStatusRunning)
	}

	return result, nil
}

// PrepareBatchInputs 准备批处理输入（供 IterationNode 和 OrderIterationNode 复用）
func (n *IterationNode) PrepareBatchInputs(state *State) ([]*value.ObjectValue, string, *graph.NodeStatusGroup, error) {
	if n.workflow == nil {
		return nil, "", nil, ErrIterationNodeWorkflowRequired
	}

	batchInputs, err := n.ExpandIterationInputs(state)
	if err != nil {
		return nil, "", nil, err
	}

	sharedInput, err := n.ParseValuesFromWithError(state, n.ValuesFrom)
	if err != nil {
		return nil, "", nil, err
	}
	for _, input := range batchInputs {
		input.AddAllIFNULL(sharedInput)
	}

	currentParentID := n.BuildCurrentParentID(state)

	statusGroup, ok := state.GetNodeStatus().(*graph.NodeStatusGroup)
	if !ok {
		statusGroup = graph.NewNodeStatusGroup(n.ID)
	}

	return batchInputs, currentParentID, statusGroup, nil
}

// ExpandIterationInputs 展开迭代输入（导出方法，供 OrderIterationNode 复用）
func (n *IterationNode) ExpandIterationInputs(state *State) ([]*value.ObjectValue, error) {
	if len(n.iterationFrom) == 0 {
		return nil, ErrIterationNodeIterationFromRequired
	}

	var inputs []*value.ObjectValue

	for _, vf := range n.iterationFrom {
		nodeValue := state.GetNodeValueFromValueFrom(vf)
		if nodeValue == nil || !nodeValue.IsArray() {
			return nil, ErrIterationNodeRequiresArrayInput
		}

		arr := nodeValue.AsArray()
		if len(inputs) == 0 {
			for i := 0; i < arr.Size(); i++ {
				inputs = append(inputs, value.NewObjectValue())
			}
		}

		if arr.Size() != len(inputs) {
			return nil, ErrIterationNodeInconsistentArraySizes
		}

		for i := 0; i < arr.Size(); i++ {
			target := inputs[i]
			item := arr.Get(i)
			if util.IsBlank(vf.Param) {
				if !item.IsObject() {
					return nil, ErrIterationNodeItemMustBeObject
				}
				target.AddAll(item.AsObject())
			} else {
				target.Put(vf.Param, item)
			}
		}
	}

	return inputs, nil
}

// IterationNodeBuilder 迭代节点构建器
type IterationNodeBuilder struct {
	node *IterationNode
}

// NewIterationNodeBuilder 创建迭代节点构建器
func NewIterationNodeBuilder(id string) *IterationNodeBuilder {
	return &IterationNodeBuilder{
		node: NewIterationNode(id),
	}
}

// Workflow 设置工作流
func (b *IterationNodeBuilder) Workflow(workflow WorkflowInterface) *IterationNodeBuilder {
	b.node.workflow = workflow
	return b
}

// ValuesFrom 设置值来源
func (b *IterationNodeBuilder) ValuesFrom(valuesFrom ...*value.ValueFrom) *IterationNodeBuilder {
	b.node.ValuesFrom = append(b.node.ValuesFrom, valuesFrom...)
	return b
}

// IterationFrom 设置迭代来源
func (b *IterationNodeBuilder) IterationFrom(iterationFrom ...*value.ValueFrom) *IterationNodeBuilder {
	b.node.iterationFrom = append(b.node.iterationFrom, iterationFrom...)
	return b
}

// Build 构建
func (b *IterationNodeBuilder) Build() *IterationNode {
	return b.node
}

// 错误定义
var (
	ErrIterationNodeWorkflowRequired       = errors.New("IterationNode workflow is required")
	ErrIterationNodeIterationFromRequired  = errors.New("IterationNode iterationFrom is required")
	ErrIterationNodeRequiresArrayInput     = errors.New("IterationNode requires array input")
	ErrIterationNodeInconsistentArraySizes = errors.New("IterationNode has inconsistent array sizes")
	ErrIterationNodeItemMustBeObject       = errors.New("IterationNode item must be ObjectValue when param is blank")
)

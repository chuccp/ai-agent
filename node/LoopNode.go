package node

import (
	"strconv"

	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/util"
	"github.com/chuccp/ai-agent/value"
)

// BreakFunc 中断条件函数，检查输出值决定是否跳出循环
type BreakFunc func(out value.NodeValue) (bool, error)

// LoopNode 循环节点，循环执行工作流直到符合break条件或达到最大循环次数
type LoopNode struct {
	*BaseNode
	outBreak BreakFunc
	maxLoop  int
	workflow WorkflowInterface
}

// NewLoopNode 创建循环节点
func NewLoopNode(id string) *LoopNode {
	return &LoopNode{
		BaseNode: NewBaseNode(id, types.NodeTypeSingle),
		maxLoop:  100,
	}
}

// SetWorkflow 设置循环工作流
func (l *LoopNode) SetWorkflow(workflow WorkflowInterface) {
	l.workflow = workflow
}

// SetBreak 设置中断条件函数
func (l *LoopNode) SetBreak(outBreak BreakFunc) {
	l.outBreak = outBreak
}

// SetMaxLoop 设置最大循环次数
func (l *LoopNode) SetMaxLoop(maxLoop int) {
	l.maxLoop = maxLoop
}

// GetWorkflow 获取循环工作流
func (l *LoopNode) GetWorkflow() WorkflowInterface {
	return l.workflow
}

// Exec 执行节点
func (l *LoopNode) Exec(state *State) (value.NodeValue, error) {
	if l.workflow == nil {
		return nil, ErrLoopNodeWorkflowRequired
	}

	input, err := l.ParseValuesFromWithError(state, l.ValuesFrom)
	if err != nil {
		return nil, err
	}

	workflowParentID := state.GetWorkflowContext().GetParentID()
	var lastResult value.NodeValue

	for i := 0; i < l.maxLoop; i++ {
		parentID := l.ID + "_loop_" + strconv.Itoa(i)
		if util.IsNotBlank(workflowParentID) {
			parentID = workflowParentID + "_" + parentID
		}

		result, err := l.workflow.Execute(state.GetWorkflowContext(), input, parentID)
		if err != nil {
			state.SetStatusType(types.NodeStatusFailed)
			return nil, err
		}

		lastResult = result

		if l.outBreak != nil {
			shouldBreak, err := l.outBreak(result)
			if err != nil {
				state.SetStatusType(types.NodeStatusFailed)
				return nil, err
			}
			if shouldBreak {
				break
			}
		}

		// 将本次输出作为下一次迭代的输入
		if result != nil && result.IsObject() {
			nextInput := value.NewObjectValue()
			nextInput.AddAll(result.AsObject())
			input = nextInput
		}
	}

	return lastResult, nil
}

// GetNodeGraph 获取节点图
func (l *LoopNode) GetNodeGraph() *graph.NodeGraph {
	var children []*graph.NodeGraph
	if l.workflow != nil {
		children = append(children, l.workflow.GetGraphs()...)
	}
	return graph.NewNodeGraphWithChildren(l.ID, "LoopNode", l.ValuesFrom, nil, children)
}

// LoopNodeBuilder 循环节点构建器
type LoopNodeBuilder struct {
	node *LoopNode
}

// NewLoopNodeBuilder 创建循环节点构建器
func NewLoopNodeBuilder(id string) *LoopNodeBuilder {
	return &LoopNodeBuilder{
		node: NewLoopNode(id),
	}
}

// Workflow 设置循环工作流
func (b *LoopNodeBuilder) Workflow(workflow WorkflowInterface) *LoopNodeBuilder {
	b.node.workflow = workflow
	return b
}

// Break 设置中断条件函数
func (b *LoopNodeBuilder) Break(outBreak BreakFunc) *LoopNodeBuilder {
	b.node.outBreak = outBreak
	return b
}

// MaxLoop 设置最大循环次数
func (b *LoopNodeBuilder) MaxLoop(maxLoop int) *LoopNodeBuilder {
	b.node.maxLoop = maxLoop
	return b
}

// ValuesFrom 设置值来源
func (b *LoopNodeBuilder) ValuesFrom(valuesFrom ...*value.ValueFrom) *LoopNodeBuilder {
	b.node.ValuesFrom = append(b.node.ValuesFrom, valuesFrom...)
	return b
}

// Build 构建
func (b *LoopNodeBuilder) Build() *LoopNode {
	return b.node
}

// 错误定义
var (
	ErrLoopNodeWorkflowRequired = errors.New("LoopNode workflow is required")
)

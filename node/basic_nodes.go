package node

import (
	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/value"
)

// InputNode 输入节点
type InputNode struct {
	*BaseNode
}

// NewInputNode 创建输入节点
func NewInputNode(id string) *InputNode {
	return &InputNode{
		BaseNode: NewBaseNode(id, types.NodeTypeSingle),
	}
}

// Exec 执行节点
func (n *InputNode) Exec(state *NodeState) (value.NodeValue, error) {
	return n.ParseValuesFrom(state, n.ValuesFrom), nil
}

// GetNodeGraph 获取节点图
func (n *InputNode) GetNodeGraph() *graph.NodeGraph {
	return graph.NewNodeGraph(n.ID, "InputNode", n.ValuesFrom)
}

// InputNodeBuilder 输入节点构建器
type InputNodeBuilder struct {
	node *InputNode
}

// NewInputNodeBuilder 创建输入节点构建器
func NewInputNodeBuilder(id string) *InputNodeBuilder {
	return &InputNodeBuilder{
		node: NewInputNode(id),
	}
}

// ValuesFrom 设置值来源
func (b *InputNodeBuilder) ValuesFrom(valuesFrom ...*value.ValueFrom) *InputNodeBuilder {
	b.node.ValuesFrom = append(b.node.ValuesFrom, valuesFrom...)
	return b
}

// Build 构建
func (b *InputNodeBuilder) Build() *InputNode {
	return b.node
}

// OutFunc 输出函数
type OutFunc func(nodeValue value.NodeValue)

// OutputNode 输出节点
type OutputNode struct {
	*BaseNode
	outFunc OutFunc
}

// NewOutputNode 创建输出节点
func NewOutputNode(id string) *OutputNode {
	return &OutputNode{
		BaseNode: NewBaseNode(id, types.NodeTypeSingle),
	}
}

// Exec 执行节点
func (n *OutputNode) Exec(state *NodeState) (value.NodeValue, error) {
	nodeValue := n.ParseValuesFrom(state, n.ValuesFrom)
	if n.outFunc != nil {
		n.outFunc(nodeValue)
	}
	return nodeValue, nil
}

// SetOutFunc 设置输出函数
func (n *OutputNode) SetOutFunc(fn OutFunc) {
	n.outFunc = fn
}

// GetNodeGraph 获取节点图
func (n *OutputNode) GetNodeGraph() *graph.NodeGraph {
	return graph.NewNodeGraph(n.ID, "OutputNode", n.ValuesFrom)
}

// OutputNodeBuilder 输出节点构建器
type OutputNodeBuilder struct {
	node *OutputNode
}

// NewOutputNodeBuilder 创建输出节点构建器
func NewOutputNodeBuilder(id string) *OutputNodeBuilder {
	return &OutputNodeBuilder{
		node: NewOutputNode(id),
	}
}

// ValuesFrom 设置值来源
func (b *OutputNodeBuilder) ValuesFrom(valuesFrom ...*value.ValueFrom) *OutputNodeBuilder {
	b.node.ValuesFrom = append(b.node.ValuesFrom, valuesFrom...)
	return b
}

// OutFunc 设置输出函数
func (b *OutputNodeBuilder) OutFunc(fn OutFunc) *OutputNodeBuilder {
	b.node.outFunc = fn
	return b
}

// Build 构建
func (b *OutputNodeBuilder) Build() *OutputNode {
	return b.node
}

// NodeExecFunc 节点执行函数
type NodeExecFunc func(state *NodeState) (value.NodeValue, error)

// FunctionNode 函数节点
type FunctionNode struct {
	*BaseNode
	execFunc NodeExecFunc
}

// NewFunctionNode 创建函数节点
func NewFunctionNode(id string, execFunc NodeExecFunc) *FunctionNode {
	return &FunctionNode{
		BaseNode: NewBaseNode(id, types.NodeTypeSingle),
		execFunc: execFunc,
	}
}

// Exec 执行节点
func (n *FunctionNode) Exec(state *NodeState) (value.NodeValue, error) {
	if n.execFunc == nil {
		return n.ParseValuesFrom(state, n.ValuesFrom), nil
	}
	return n.execFunc(state)
}

// SetExecFunc 设置执行函数
func (n *FunctionNode) SetExecFunc(fn NodeExecFunc) {
	n.execFunc = fn
}

// GetNodeGraph 获取节点图
func (n *FunctionNode) GetNodeGraph() *graph.NodeGraph {
	return graph.NewNodeGraph(n.ID, "FunctionNode", n.ValuesFrom)
}

// FunctionNodeBuilder 函数节点构建器
type FunctionNodeBuilder struct {
	node *FunctionNode
}

// NewFunctionNodeBuilder 创建函数节点构建器
func NewFunctionNodeBuilder(id string) *FunctionNodeBuilder {
	return &FunctionNodeBuilder{
		node: NewFunctionNode(id, nil),
	}
}

// ValuesFrom 设置值来源
func (b *FunctionNodeBuilder) ValuesFrom(valuesFrom ...*value.ValueFrom) *FunctionNodeBuilder {
	b.node.ValuesFrom = append(b.node.ValuesFrom, valuesFrom...)
	return b
}

// ExecFunc 设置执行函数
func (b *FunctionNodeBuilder) ExecFunc(fn NodeExecFunc) *FunctionNodeBuilder {
	b.node.execFunc = fn
	return b
}

// Build 构建
func (b *FunctionNodeBuilder) Build() *FunctionNode {
	return b.node
}
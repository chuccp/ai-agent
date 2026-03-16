package graph

import (
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/value"
)

// NodeStatusInterface 节点状态接口
type NodeStatusInterface interface {
	GetNodeID() string
	GetStatus() types.NodeStatusType
	SetStatus(status types.NodeStatusType)
	IsSucceeded() bool
	IsFailed() bool
	IsRunning() bool
	IsWaiting() bool
}

// NodeStatus 节点状态
type NodeStatus struct {
	NodeID     string
	StatusType types.NodeStatusType
}

// 确保实现接口
var _ NodeStatusInterface = (*NodeStatus)(nil)

// GetNodeID 获取节点ID
func (n *NodeStatus) GetNodeID() string {
	return n.NodeID
}

// NewNodeStatus 创建节点状态
func NewNodeStatus(nodeID string) *NodeStatus {
	return &NodeStatus{
		NodeID:     nodeID,
		StatusType: types.NodeStatusWaiting,
	}
}

// SetStatus 设置状态
func (n *NodeStatus) SetStatus(status types.NodeStatusType) {
	n.StatusType = status
}

// GetStatus 获取状态
func (n *NodeStatus) GetStatus() types.NodeStatusType {
	return n.StatusType
}

// IsSucceeded 是否成功
func (n *NodeStatus) IsSucceeded() bool {
	return n.StatusType == types.NodeStatusSucceeded
}

// IsFailed 是否失败
func (n *NodeStatus) IsFailed() bool {
	return n.StatusType == types.NodeStatusFailed
}

// IsRunning 是否运行中
func (n *NodeStatus) IsRunning() bool {
	return n.StatusType == types.NodeStatusRunning
}

// IsWaiting 是否等待中
func (n *NodeStatus) IsWaiting() bool {
	return n.StatusType == types.NodeStatusWaiting
}

// NodeStatusGroup 节点状态组
type NodeStatusGroup struct {
	NodeStatus
	Children [][]*NodeStatus
}

// NewNodeStatusGroup 创建节点状态组
func NewNodeStatusGroup(nodeID string) *NodeStatusGroup {
	return &NodeStatusGroup{
		NodeStatus: *NewNodeStatus(nodeID),
		Children:   make([][]*NodeStatus, 0),
	}
}

// SetChildren 设置子节点
func (n *NodeStatusGroup) SetChildren(children [][]*NodeStatus) {
	n.Children = children
}

// GetChildren 获取子节点
func (n *NodeStatusGroup) GetChildren() [][]*NodeStatus {
	return n.Children
}

// GraphStatus 图状态
type GraphStatus struct {
	NodeStatuses []*NodeStatus
}

// NewGraphStatus 创建图状态
func NewGraphStatus(statuses []*NodeStatus) *GraphStatus {
	return &GraphStatus{
		NodeStatuses: statuses,
	}
}

// GetNodeStatuses 获取节点状态列表
func (g *GraphStatus) GetNodeStatuses() []*NodeStatus {
	return g.NodeStatuses
}

// NodeGraph 节点图
type NodeGraph struct {
	ID         string
	Type       string
	ValuesFrom []*value.ValueFrom
}

// NewNodeGraph 创建节点图
func NewNodeGraph(id, nodeType string, valuesFrom []*value.ValueFrom) *NodeGraph {
	return &NodeGraph{
		ID:         id,
		Type:       nodeType,
		ValuesFrom: valuesFrom,
	}
}

// GetID 获取ID
func (n *NodeGraph) GetID() string {
	return n.ID
}

// GetType 获取类型
func (n *NodeGraph) GetType() string {
	return n.Type
}

// GetValuesFrom 获取值来源
func (n *NodeGraph) GetValuesFrom() []*value.ValueFrom {
	return n.ValuesFrom
}

// NodeGraphGroup 节点图组
type NodeGraphGroup struct {
	NodeGraph
	IterationFrom []*value.ValueFrom
	Children      []*NodeGraph
}

// NewNodeGraphGroup 创建节点图组
func NewNodeGraphGroup(id, nodeType string, valuesFrom, iterationFrom []*value.ValueFrom) *NodeGraphGroup {
	return &NodeGraphGroup{
		NodeGraph:     *NewNodeGraph(id, nodeType, valuesFrom),
		IterationFrom: iterationFrom,
		Children:      make([]*NodeGraph, 0),
	}
}

// SetChildren 设置子节点
func (n *NodeGraphGroup) SetChildren(children []*NodeGraph) {
	n.Children = children
}

// GetChildren 获取子节点
func (n *NodeGraphGroup) GetChildren() []*NodeGraph {
	return n.Children
}

// GetIterationFrom 获取迭代来源
func (n *NodeGraphGroup) GetIterationFrom() []*value.ValueFrom {
	return n.IterationFrom
}

// Graph 图
type Graph struct {
	NodeGraphs []*NodeGraph
	Params     []string
	OutParams  []string
}

// NewGraph 创建图
func NewGraph(nodeGraphs []*NodeGraph, params, outParams []string) *Graph {
	return &Graph{
		NodeGraphs: nodeGraphs,
		Params:     params,
		OutParams:  outParams,
	}
}

// GetNodeGraphs 获取节点图列表
func (g *Graph) GetNodeGraphs() []*NodeGraph {
	return g.NodeGraphs
}

// GetParams 获取参数
func (g *Graph) GetParams() []string {
	return g.Params
}

// GetOutParams 获取输出参数
func (g *Graph) GetOutParams() []string {
	return g.OutParams
}
package graph

import (
	"encoding/json"

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
	NodeID     string               `json:"nodeId"`
	StatusType types.NodeStatusType `json:"nodeStatusType"`
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
	Children [][]NodeStatusInterface `json:"children,omitempty"`
}

// NewNodeStatusGroup 创建节点状态组
func NewNodeStatusGroup(nodeID string) *NodeStatusGroup {
	return &NodeStatusGroup{
		NodeStatus: *NewNodeStatus(nodeID),
		Children:   make([][]NodeStatusInterface, 0),
	}
}

// SetChildren 设置子节点
func (n *NodeStatusGroup) SetChildren(children [][]NodeStatusInterface) {
	n.Children = children
}
func (n *NodeStatusGroup) AddChildren(children []NodeStatusInterface) {
	n.Children = append(n.Children, children)
}

// GetChildren 获取子节点
func (n *NodeStatusGroup) GetChildren() [][]NodeStatusInterface {
	return n.Children
}

// ConvertToNodeStatusSlice 将 NodeStatusInterface 切片转换为 NodeStatus 指针切片
// 如果元素不是 *NodeStatus，则创建一个新的 NodeStatus
func ConvertToNodeStatusSlice(interfaces []NodeStatusInterface) []*NodeStatus {
	result := make([]*NodeStatus, len(interfaces))
	for i, iface := range interfaces {
		if ns, ok := iface.(*NodeStatus); ok {
			result[i] = ns
		} else if nsg, ok := iface.(*NodeStatusGroup); ok {
			result[i] = &nsg.NodeStatus
		} else {
			result[i] = &NodeStatus{
				NodeID:     iface.GetNodeID(),
				StatusType: iface.GetStatus(),
			}
		}
	}
	return result
}

// GraphStatus 图状态
type GraphStatus struct {
	NodeStatuses []NodeStatusInterface `json:"nodeStatuses,omitempty"`
}

// NewGraphStatus 创建图状态
func NewGraphStatusFromInterfaces(statuses []NodeStatusInterface) *GraphStatus {
	return &GraphStatus{
		NodeStatuses: statuses,
	}
}

// MarshalJSON 自定义JSON序列化
func (g *GraphStatus) MarshalJSON() ([]byte, error) {
	// 使用 map 来灵活控制字段输出
	var statuses []map[string]interface{}
	for _, iface := range g.NodeStatuses {
		status := map[string]interface{}{
			"nodeId":         iface.GetNodeID(),
			"nodeStatusType": iface.GetStatus(),
		}
		// 检查是否是 NodeStatusGroup，如果是则添加 children（包括空数组）
		if group, ok := iface.(*NodeStatusGroup); ok {
			status["children"] = group.Children
		}
		statuses = append(statuses, status)
	}
	type Alias GraphStatus
	return json.Marshal(&struct {
		NodeStatuses []map[string]interface{} `json:"nodeStatuses,omitempty"`
	}{
		NodeStatuses: statuses,
	})

}

// NewGraphStatus 创建图状态
func NewGraphStatus(statuses []NodeStatusInterface) *GraphStatus {
	return &GraphStatus{
		NodeStatuses: statuses,
	}
}

// GetNodeStatuses 获取节点状态列表
func (g *GraphStatus) GetNodeStatuses() []NodeStatusInterface {
	return g.NodeStatuses
}

// NodeGraph 节点图
type NodeGraph struct {
	ID            string             `json:"id"`
	Type          string             `json:"type"`
	ValuesFrom    []*value.ValueFrom `json:"valuesFrom,omitempty"`
	IterationFrom []*value.ValueFrom `json:"iterationFrom,omitempty"`
	Children      []*NodeGraph       `json:"children,omitempty"`
}

// NewNodeGraph 创建节点图
func NewNodeGraph(id, nodeType string, valuesFrom []*value.ValueFrom) *NodeGraph {
	return &NodeGraph{
		ID:         id,
		Type:       nodeType,
		ValuesFrom: valuesFrom,
	}
}

// NewNodeGraphWithChildren 创建带子节点的节点图
func NewNodeGraphWithChildren(id, nodeType string, valuesFrom, iterationFrom []*value.ValueFrom, children []*NodeGraph) *NodeGraph {
	return &NodeGraph{
		ID:            id,
		Type:          nodeType,
		ValuesFrom:    valuesFrom,
		IterationFrom: iterationFrom,
		Children:      children,
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
	IterationFrom []*value.ValueFrom `json:"iterationFrom,omitempty"`
	Children      []*NodeGraph       `json:"children,omitempty"`
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
	NodeGraphs []*NodeGraph `json:"nodeGraphs,omitempty"`
	Params     []string     `json:"params,omitempty"`
	OutParams  []string     `json:"outParams,omitempty"`
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

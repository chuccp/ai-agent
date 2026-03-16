package node

import (
	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/util"
	"github.com/chuccp/ai-agent/value"
)

// WatchFunc 监听函数
type WatchFunc func(state *NodeState) error

// Node 节点接口
type Node interface {
	// GetID 获取节点ID
	GetID() string
	// GetNodeType 获取节点类型
	GetNodeType() types.NodeType
	// IsSingle 是否单节点
	IsSingle() bool
	// GetValuesFrom 获取值来源
	GetValuesFrom() []*value.ValueFrom
	// SetValuesFrom 设置值来源
	SetValuesFrom(valuesFrom []*value.ValueFrom)
	// GetWatch 获取监听函数
	GetWatch() WatchFunc
	// SetWatch 设置监听函数
	SetWatch(watch WatchFunc)
	// Exec 执行节点
	Exec(state *NodeState) (value.NodeValue, error)
	// GetNodeGraph 获取节点图
	GetNodeGraph() *graph.NodeGraph
	// ParseValuesFrom 解析值来源
	ParseValuesFrom(state *NodeState, valuesFrom []*value.ValueFrom) *value.ObjectValue
}

// BaseNode 基础节点
type BaseNode struct {
	ID         string
	ValuesFrom []*value.ValueFrom
	Watch      WatchFunc
	nodeType   types.NodeType
}

// NewBaseNode 创建基础节点
func NewBaseNode(id string, nodeType types.NodeType) *BaseNode {
	return &BaseNode{
		ID:       id,
		nodeType: nodeType,
	}
}

// GetID 获取节点ID
func (n *BaseNode) GetID() string {
	return n.ID
}

// GetNodeType 获取节点类型
func (n *BaseNode) GetNodeType() types.NodeType {
	return n.nodeType
}

// IsSingle 是否单节点
func (n *BaseNode) IsSingle() bool {
	return n.nodeType == types.NodeTypeSingle
}

// GetValuesFrom 获取值来源
func (n *BaseNode) GetValuesFrom() []*value.ValueFrom {
	return n.ValuesFrom
}

// SetValuesFrom 设置值来源
func (n *BaseNode) SetValuesFrom(valuesFrom []*value.ValueFrom) {
	n.ValuesFrom = valuesFrom
}

// GetWatch 获取监听函数
func (n *BaseNode) GetWatch() WatchFunc {
	return n.Watch
}

// SetWatch 设置监听函数
func (n *BaseNode) SetWatch(watch WatchFunc) {
	n.Watch = watch
}

// ParseValuesFrom 解析值来源
func (n *BaseNode) ParseValuesFrom(state *NodeState, valuesFrom []*value.ValueFrom) *value.ObjectValue {
	result := value.NewObjectValue()
	rootValue := state.GetRootValue()
	if rootValue != nil {
		result.AddAll(rootValue)
	}

	if valuesFrom != nil {
		for _, vf := range valuesFrom {
			nodeValue := state.GetNodeValueFromValueFrom(vf)
			if nodeValue == nil {
				panic("Node " + n.ID + " ValueFrom " + vf.From + " not found")
			}

			if util.IsNotBlank(vf.Param) {
				result.Put(vf.Param, nodeValue)
			} else if nodeValue.IsObject() {
				result.AddAll(nodeValue.AsObject())
			} else {
				panic("Node " + n.ID + " ValueFrom " + vf.From + " must contain param or be ObjectValue")
			}
		}
	}

	return result
}

// GetNodeGraph 获取节点图
func (n *BaseNode) GetNodeGraph() *graph.NodeGraph {
	return graph.NewNodeGraph(n.ID, "BaseNode", n.ValuesFrom)
}

// WorkflowContext 工作流上下文接口
type WorkflowContext interface {
	GetRootValue() *value.ObjectValue
	GetParentID() string
	GetNodeValue(nodeID string) value.NodeValue
	AddNodeValue(nodeID string, nodeValue value.NodeValue)
}

// NodeState 节点状态
type NodeState struct {
	workflowContext WorkflowContext
	input           *value.ObjectValue
	nodeStatus      graph.NodeStatusInterface
	nodeValue       value.NodeValue
	nodeID          string
}

// NewNodeState 创建节点状态
func NewNodeState(workflowContext WorkflowContext, nodeID string, nodeType types.NodeType, input *value.ObjectValue) *NodeState {
	var status graph.NodeStatusInterface
	if nodeType == types.NodeTypeMultiple {
		status = graph.NewNodeStatusGroup(nodeID)
	} else {
		status = graph.NewNodeStatus(nodeID)
	}

	if input == nil {
		input = value.NewObjectValue()
	}

	return &NodeState{
		workflowContext: workflowContext,
		nodeID:          nodeID,
		input:           input,
		nodeStatus:      status,
	}
}

// GetWorkflowContext 获取工作流上下文
func (s *NodeState) GetWorkflowContext() WorkflowContext {
	return s.workflowContext
}

// GetInput 获取输入
func (s *NodeState) GetInput() *value.ObjectValue {
	return s.input
}

// GetRootValue 获取根值
func (s *NodeState) GetRootValue() *value.ObjectValue {
	if s.workflowContext == nil {
		return nil
	}
	return s.workflowContext.GetRootValue()
}

// GetNodeStatus 获取节点状态
func (s *NodeState) GetNodeStatus() graph.NodeStatusInterface {
	return s.nodeStatus
}

// GetStatusType 获取状态类型
func (s *NodeState) GetStatusType() types.NodeStatusType {
	return s.nodeStatus.GetStatus()
}

// SetStatusType 设置状态类型
func (s *NodeState) SetStatusType(statusType types.NodeStatusType) {
	s.nodeStatus.SetStatus(statusType)
}

// IsSucceeded 是否成功
func (s *NodeState) IsSucceeded() bool {
	return s.GetStatusType() == types.NodeStatusSucceeded
}

// IsFailed 是否失败
func (s *NodeState) IsFailed() bool {
	return s.GetStatusType() == types.NodeStatusFailed
}

// IsStarted 是否已开始
func (s *NodeState) IsStarted() bool {
	return s.GetStatusType() == types.NodeStatusStarted
}

// IsRunning 是否运行中
func (s *NodeState) IsRunning() bool {
	return s.GetStatusType() == types.NodeStatusRunning
}

// IsWaiting 是否等待中
func (s *NodeState) IsWaiting() bool {
	return s.GetStatusType() == types.NodeStatusWaiting
}

// GetNodeValue 获取节点值
func (s *NodeState) GetNodeValue() value.NodeValue {
	return s.nodeValue
}

// SetNodeValue 设置节点值
func (s *NodeState) SetNodeValue(nodeValue value.NodeValue) {
	s.nodeValue = nodeValue
}

// GetID 获取节点ID
func (s *NodeState) GetID() string {
	return s.nodeID
}

// GetParentID 获取父ID
func (s *NodeState) GetParentID() string {
	if s.workflowContext == nil {
		return ""
	}
	return s.workflowContext.GetParentID()
}

// GetNodeValueFromValueFrom 从ValueFrom获取节点值
func (s *NodeState) GetNodeValueFromValueFrom(vf *value.ValueFrom) value.NodeValue {
	return s.GetNodeValueFromNode(vf.NodeID, vf.From)
}

// GetNodeValueFromNode 从节点获取值
func (s *NodeState) GetNodeValueFromNode(nodeID, from string) value.NodeValue {
	var source value.NodeValue
	if util.IsBlank(nodeID) {
		source = s.GetRootValue()
	} else if s.workflowContext != nil {
		source = s.workflowContext.GetNodeValue(nodeID)
	}

	if source == nil {
		return nil
	}

	if util.IsBlank(from) {
		return source
	}

	return source.FindValue(from)
}
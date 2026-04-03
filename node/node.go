package node

import (
	"emperror.dev/errors"

	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/util"
	"github.com/chuccp/ai-agent/value"
)

// WatchFunc 监听函数
type WatchFunc func(state *State) error

// BaseNode 基础节点
type BaseNode struct {
	ID         string
	PrevNodeID string
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

func (n *BaseNode) SetPrevNodeID(prevNodeID string) {
	n.PrevNodeID = prevNodeID
}
func (n *BaseNode) GetPrevNodeID() string {
	return n.PrevNodeID
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
func (n *BaseNode) ParseValuesFrom(state *State, valuesFrom []*value.ValueFrom) *value.ObjectValue {
	result, _ := n.ParseValuesFromWithError(state, valuesFrom)
	return result
}

// ParseValuesFromWithError 解析值来源（带错误返回）
func (n *BaseNode) ParseValuesFromWithError(state *State, valuesFrom []*value.ValueFrom) (*value.ObjectValue, error) {
	result := value.NewObjectValue()
	rootValue := state.GetRootValue()
	if rootValue != nil {
		result.AddAll(rootValue)
	}
	value, err := n.ParseNoRootValuesFromWithError(state, valuesFrom)
	if err != nil {
		return nil, err
	}
	result.AddAll(value)
	return result, nil
}

// ParseValuesFromWithError 解析值来源（带错误返回）
func (n *BaseNode) ParseNoRootValuesFromWithError(state *State, valuesFrom []*value.ValueFrom) (*value.ObjectValue, error) {
	result := value.NewObjectValue()
	if valuesFrom != nil {
		for _, vf := range valuesFrom {
			nodeValue, err := state.GetNodeValueFromValueFrom(vf)
			if err != nil {
				return nil, err
			}
			if nodeValue == nil {
				return nil, errors.New("Node " + n.ID + " ValueFrom " + vf.From + " not found")
			}

			if util.IsNotBlank(vf.Param) {
				result.Put(vf.Param, nodeValue)
			} else if nodeValue.IsObject() {
				result.AddAll(nodeValue.AsObject())
			} else {
				return nil, errors.New("Node " + n.ID + " ValueFrom " + vf.From + " must contain param or be ObjectValue")
			}
		}
	}
	return result, nil
}

// GetNodeGraph 获取节点图
func (n *BaseNode) GetNodeGraph() *graph.NodeGraph {
	return graph.NewNodeGraph(n.ID, "BaseNode", n.ValuesFrom)
}

// State NodeState 节点状态
type State struct {
	workflowContext WorkflowContext
	input           *value.ObjectValue
	nodeStatus      graph.NodeStatusInterface
	nodeValue       value.NodeValue
	nodeID          string
	Parameter       *value.ObjectValue
}

// NewNodeState 创建节点状态
func NewNodeState(workflowContext WorkflowContext, nodeID string, input *value.ObjectValue, parameter *value.ObjectValue) *State {
	if input == nil {
		input = value.NewObjectValue()
	}
	if parameter == nil {
		parameter = value.NewObjectValue()
	}
	return &State{
		workflowContext: workflowContext,
		nodeID:          nodeID,
		input:           input,
		nodeStatus:      workflowContext.GetNodeStatus(nodeID),
		Parameter:       parameter,
	}
}
func (s *State) GetShareValue() *value.ArrayValue {
	return s.workflowContext.GetShareValue()
}

// GetWorkflowContext 获取工作流上下文
func (s *State) GetWorkflowContext() WorkflowContext {
	return s.workflowContext
}

// GetInput 获取输入
func (s *State) GetInput() *value.ObjectValue {
	return s.input
}

// GetRootValue 获取根值
func (s *State) GetRootValue() *value.ObjectValue {
	if s.workflowContext == nil {
		return nil
	}
	return s.workflowContext.GetRootValue()
}

// GetNodeStatus 获取节点状态
func (s *State) GetNodeStatus() graph.NodeStatusInterface {
	return s.nodeStatus
}

// GetStatusType 获取状态类型
func (s *State) GetStatusType() types.NodeStatusType {
	return s.nodeStatus.GetStatus()
}

// SetStatusType 设置状态类型
func (s *State) SetStatusType(statusType types.NodeStatusType) {
	s.nodeStatus.SetStatus(statusType)
}

// IsSucceeded 是否成功
func (s *State) IsSucceeded() bool {
	return s.GetStatusType() == types.NodeStatusSucceeded
}

// IsFailed 是否失败
func (s *State) IsFailed() bool {
	return s.GetStatusType() == types.NodeStatusFailed
}

// IsStarted 是否已开始
func (s *State) IsStarted() bool {
	return s.GetStatusType() == types.NodeStatusStarted
}

// IsRunning 是否运行中
func (s *State) IsRunning() bool {
	return s.GetStatusType() == types.NodeStatusRunning
}

// IsWaiting 是否等待中
func (s *State) IsWaiting() bool {
	return s.GetStatusType() == types.NodeStatusWaiting
}

// GetNodeValue 获取节点值
func (s *State) GetNodeValue() value.NodeValue {
	return s.nodeValue
}

// SetNodeValue 设置节点值
func (s *State) SetNodeValue(nodeValue value.NodeValue) {
	s.nodeValue = nodeValue
}

// GetID 获取节点ID
func (s *State) GetID() string {
	return s.nodeID
}

// GetParentID 获取父ID
func (s *State) GetParentID() string {
	if s.workflowContext == nil {
		return ""
	}
	return s.workflowContext.GetParentID()
}

// GetCachePath 获取缓存路径
func (s *State) GetCachePath() string {
	if s.workflowContext == nil {
		return ""
	}
	return s.workflowContext.GetCachePath()
}

// IsCacheEnabled 是否启用缓存
func (s *State) IsCacheEnabled() bool {
	if s.workflowContext == nil {
		return false
	}
	return s.workflowContext.IsCacheEnabled()
}

// SaveCache 保存缓存
func (s *State) SaveCache(key string, nodeValue value.NodeValue) error {
	if s.workflowContext == nil {
		return nil
	}
	return s.workflowContext.SaveCache(key, s.nodeID, nodeValue)
}

func (s *State) SaveCacheLLM(key string, nodeValue value.NodeValue, system string, user string, urlsValue *value.UrlsValue) error {
	if s.workflowContext == nil {
		return nil
	}
	object := value.NewObjectValue()
	object.Put("system", value.NewTextValue(system))
	object.Put("user", value.NewTextValue(user))
	object.Put("urls", urlsValue)
	object.Put("result", nodeValue)
	return s.workflowContext.SaveCache(key, s.nodeID, object)
}
func (s *State) GetCacheLLM(key string) (value.NodeValue, error) {
	if s.workflowContext == nil {
		return nil, nil
	}
	object, err := s.GetCache(key)
	if object == nil {
		return value.NullValue, err
	}
	if err != nil {
		return nil, err
	}
	return object.AsObject().Get("result"), err
}

// GetCache 获取缓存
func (s *State) GetCache(key string) (value.NodeValue, error) {
	if s.workflowContext == nil {
		return nil, nil
	}
	return s.workflowContext.GetCache(key, s.nodeID)
}

// HasCache 检查缓存是否存在
func (s *State) HasCache(key string) bool {
	if s.workflowContext == nil {
		return false
	}
	return s.workflowContext.HasCache(key, s.nodeID)
}

// GetNodeValueFromValueFrom 从ValueFrom获取节点值
func (s *State) GetNodeValueFromValueFrom(vf *value.ValueFrom) (value.NodeValue, error) {
	return s.GetNodeValueFromNodeWithError(vf.NodeID, vf.From)
}

// GetNodeValueFromNode 从节点获取值
func (s *State) GetNodeValueFromNode(nodeID, from string) value.NodeValue {
	v, _ := s.GetNodeValueFromNodeWithError(nodeID, from)
	return v
}
func (s *State) GetNodeValueFromNodeWithError(nodeID, from string) (value.NodeValue, error) {
	var source value.NodeValue
	if util.IsBlank(nodeID) {
		source = s.GetRootValue()
	} else if s.workflowContext != nil {
		source = s.workflowContext.GetNodeValue(nodeID)
	}

	if source == nil {
		return nil, errors.New("Node " + s.nodeID + " not found")
	}

	if util.IsBlank(from) || util.EqualsAny(from, ".", "*", "$", "$.") {
		return source, nil
	}

	return source.FindValue(from), nil
}
func (s *State) GetNodeValueFromRootNode(from string) value.NodeValue {
	var source value.NodeValue = s.GetRootValue()
	if util.IsBlank(from) || util.EqualsAny(from, ".", "*", "$", "$.") {
		return source
	}
	return source.FindValue(from)
}

// GetParameter 获取参数
func (s *State) GetParameter() *value.ObjectValue {
	if s.Parameter == nil {
		return value.NewObjectValue()
	}
	return s.Parameter
}

// GetParameterString 获取字符串参数，如果不存在返回默认值
func (s *State) GetParameterString(key string, defaultValue string) string {
	if s.Parameter == nil {
		return defaultValue
	}
	v := s.Parameter.GetString(key)
	if v == "" {
		return defaultValue
	}
	return v
}

// GetParameterInt 获取整数参数，如果不存在返回默认值
func (s *State) GetParameterInt(key string, defaultValue int) int {
	if s.Parameter == nil {
		return defaultValue
	}
	v := s.Parameter.GetNumber(key)
	if v == 0 {
		return defaultValue
	}
	return int(v)
}

// GetParameterBool 获取布尔参数，如果不存在返回默认值
func (s *State) GetParameterBool(key string, defaultValue bool) bool {
	if s.Parameter == nil {
		return defaultValue
	}
	v := s.Parameter.Get(key)
	if v == nil || v.IsNull() {
		return defaultValue
	}
	if v.IsBool() {
		return v.AsBool().Value
	}
	return defaultValue
}

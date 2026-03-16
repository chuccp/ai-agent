package node

import (
	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/value"
)

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
	Exec(state *State) (value.NodeValue, error)
	// GetNodeGraph 获取节点图
	GetNodeGraph() *graph.NodeGraph
	// ParseValuesFrom 解析值来源
	ParseValuesFrom(state *State, valuesFrom []*value.ValueFrom) *value.ObjectValue
	// ParseValuesFromWithError 解析值来源（带错误返回）
	ParseValuesFromWithError(state *State, valuesFrom []*value.ValueFrom) (*value.ObjectValue, error)

	SetPrevNodeID(prevNodeID string)
	GetPrevNodeID() string
}

// WorkflowContext 工作流上下文接口
type WorkflowContext interface {
	GetRootValue() *value.ObjectValue
	GetParentID() string
	GetNodeValue(nodeID string) value.NodeValue

	GetNodeStatus(nodeID string) graph.NodeStatusInterface

	AddNodeValue(nodeID string, nodeValue value.NodeValue)
	IsCacheEnabled() bool
	GetCachePath() string
	GetCacheKey(key, nodeID string) string
	SaveCache(key, nodeID string, nodeValue value.NodeValue) error
	GetCache(key, nodeID string) (value.NodeValue, error)
	HasCache(key, nodeID string) bool
	// CreateChildContext 创建子上下文，用于IFNode等条件节点执行子工作流
	CreateChildContext(nodes []Node, childRootValue *value.ObjectValue, childParentID string) WorkflowContext
}

// WorkflowInterface 工作流接口
type WorkflowInterface interface {
	// Exec 执行工作流（使用现有的WorkflowContext）
	Exec(ctx WorkflowContext) (value.NodeValue, error)
	// ExecBatch 批量执行
	ExecBatch(ctx WorkflowContext, statusGroup *graph.NodeStatusGroup, parentID string, inputs []*value.ObjectValue) (value.NodeValue, error)
	// Execute 执行子工作流（创建子上下文）
	Execute(ctx WorkflowContext, input *value.ObjectValue, parentID string) (value.NodeValue, error)
	// GetGraphs 获取节点图列表
	GetGraphs() []*graph.NodeGraph
}

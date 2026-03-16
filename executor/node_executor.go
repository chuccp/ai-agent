package executor

import (
	"sync"
	"time"

	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/node"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/util"
	"github.com/chuccp/ai-agent/value"
)

// OnBeforeNodeRun 节点运行前回调
type OnBeforeNodeRun func(state *node.NodeState) error

// OnAfterNodeRun 节点运行后回调
type OnAfterNodeRun func(state *node.NodeState) error

// OnFailedNodeRun 节点运行失败回调
type OnFailedNodeRun func(state *node.NodeState, err error)

// Config 执行器配置
type Config struct {
	SkipRunning              bool
	WaitingRetryInterval     time.Duration
	OnBeforeNodeRun          OnBeforeNodeRun
	OnAfterNodeRun           OnAfterNodeRun
	OnFailedNodeRun          OnFailedNodeRun
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		SkipRunning:          true,
		WaitingRetryInterval: 5 * time.Second,
	}
}

// Context 工作流上下文
type Context struct {
	rootValue    *value.ObjectValue
	nodeValues   sync.Map // map[string]value.NodeValue
	nodeStatuses sync.Map // map[string]*graph.NodeStatus
	parentID     string
	cacheManager interface{} // *cache.Manager
	config       *Config
	mu           sync.Mutex
}

// NewContext 创建工作流上下文
func NewContext(rootValue *value.ObjectValue, config *Config) *Context {
	if config == nil {
		config = DefaultConfig()
	}
	if rootValue == nil {
		rootValue = value.NewObjectValue()
	}
	return &Context{
		rootValue: rootValue,
		config:    config,
	}
}

// CreateChildContext 创建子上下文
func (c *Context) CreateChildContext(childRootValue *value.ObjectValue, childParentID string) *Context {
	return &Context{
		rootValue:    childRootValue,
		cacheManager: c.cacheManager,
		parentID:     childParentID,
		config:       c.config,
	}
}

// ClearExecutionState 清空执行状态
func (c *Context) ClearExecutionState() {
	c.nodeValues = sync.Map{}
	c.nodeStatuses = sync.Map{}
}

// AddNodeStatus 添加节点状态
func (c *Context) AddNodeStatus(nodeID string, status *graph.NodeStatus) {
	c.nodeStatuses.Store(nodeID, status)
}

// AddNodeStatusInterface 添加节点状态接口
func (c *Context) AddNodeStatusInterface(nodeID string, status graph.NodeStatusInterface) {
	c.nodeStatuses.Store(nodeID, status)
}

// AddNodeValue 添加节点值
func (c *Context) AddNodeValue(nodeID string, nodeValue value.NodeValue) {
	c.nodeValues.Store(nodeID, nodeValue)
}

// GetNodeValue 获取节点值
func (c *Context) GetNodeValue(nodeID string) value.NodeValue {
	if v, ok := c.nodeValues.Load(nodeID); ok {
		return v.(value.NodeValue)
	}
	return nil
}

// GetRootValue 获取根值
func (c *Context) GetRootValue() *value.ObjectValue {
	return c.rootValue
}

// SetRootValue 设置根值
func (c *Context) SetRootValue(rootValue *value.ObjectValue) {
	c.rootValue = rootValue
}

// GetParentID 获取父ID
func (c *Context) GetParentID() string {
	return c.parentID
}

// GetNodeStatuses 获取节点状态列表
func (c *Context) GetNodeStatuses() []*graph.NodeStatus {
	var statuses []*graph.NodeStatus
	c.nodeStatuses.Range(func(key, value interface{}) bool {
		statuses = append(statuses, value.(*graph.NodeStatus))
		return true
	})
	return statuses
}

// GetConfig 获取配置
func (c *Context) GetConfig() *Config {
	return c.config
}

// NodeExecutor 节点执行器
type NodeExecutor struct {
	nodes      []node.Node
	nodeMap    map[string]node.Node
	stateMap   map[string]*node.NodeState
	ctx        *Context
	config     *Config
	mu         sync.Mutex
}

// NewNodeExecutor 创建节点执行器
func NewNodeExecutor(nodes []node.Node, ctx *Context) *NodeExecutor {
	nodeMap := make(map[string]node.Node)
	for _, n := range nodes {
		nodeMap[n.GetID()] = n
	}
	return &NodeExecutor{
		nodes:    nodes,
		nodeMap:  nodeMap,
		stateMap: make(map[string]*node.NodeState),
		ctx:      ctx,
		config:   ctx.GetConfig(),
	}
}

// GetEndNode 获取结束节点
func (e *NodeExecutor) GetEndNode() node.Node {
	if len(e.nodes) == 0 {
		panic("Workflow has no nodes")
	}
	return e.nodes[len(e.nodes)-1]
}

// Exec 执行
func (e *NodeExecutor) Exec() (value.NodeValue, error) {
	endNode := e.GetEndNode()
	layers, err := BuildExecutionLayers(e.nodeMap, endNode)
	if err != nil {
		return nil, err
	}

	for _, layer := range layers {
		if !e.executeLayer(layer) {
			return value.NullValue, nil
		}
	}

	return e.ctx.GetNodeValue(endNode.GetID()), nil
}

// IsAllSucceeded 是否全部成功
func (e *NodeExecutor) IsAllSucceeded() bool {
	for _, state := range e.stateMap {
		if !state.IsSucceeded() {
			return false
		}
	}
	return true
}

// executeLayer 执行层级（使用goroutine并发执行）
func (e *NodeExecutor) executeLayer(layer []node.Node) bool {
	if len(layer) == 0 {
		return true
	}

	// 单节点直接执行
	if len(layer) == 1 {
		state := e.createAndRunNodeState(layer[0])
		e.finalizeNodeState(layer[0].GetID(), state)
		return state.IsSucceeded()
	}

	// 多节点并发执行
	results := make(chan *executionResult, len(layer))

	for _, n := range layer {
		go func(node node.Node) {
			state := e.createAndRunNodeState(node)
			results <- &executionResult{nodeID: node.GetID(), state: state}
		}(n)
	}

	// 收集结果
	allSucceeded := true
	for i := 0; i < len(layer); i++ {
		result := <-results
		e.finalizeNodeState(result.nodeID, result.state)
		if !result.state.IsSucceeded() {
			allSucceeded = false
		}
	}

	return allSucceeded
}

// executionResult 执行结果
type executionResult struct {
	nodeID string
	state  *node.NodeState
}

// createAndRunNodeState 创建并运行节点状态
func (e *NodeExecutor) createAndRunNodeState(n node.Node) *node.NodeState {
	input := e.resolveNodeInput(n)
	state := node.NewNodeState(e.ctx, n.GetID(), n.GetNodeType(), input)

	e.mu.Lock()
	e.stateMap[n.GetID()] = state
	e.ctx.AddNodeStatusInterface(n.GetID(), state.GetNodeStatus())
	e.mu.Unlock()

	state.SetStatusType(types.NodeStatusStarted)

	// 执行监听函数
	if n.GetWatch() != nil {
		n.GetWatch()(state)
	}

	// 执行前置回调
	if e.config.OnBeforeNodeRun != nil {
		e.config.OnBeforeNodeRun(state)
	}

	// 执行节点
	output, err := n.Exec(state)
	if err != nil {
		state.SetStatusType(types.NodeStatusFailed)
		if e.config.OnFailedNodeRun != nil {
			e.config.OnFailedNodeRun(state, err)
		}
		return state
	}

	state.SetNodeValue(output)
	if state.GetStatusType() == types.NodeStatusStarted {
		state.SetStatusType(types.NodeStatusSucceeded)
	}

	// 执行后置回调
	if e.config.OnAfterNodeRun != nil {
		e.config.OnAfterNodeRun(state)
	}

	return state
}

// finalizeNodeState 完成节点状态
func (e *NodeExecutor) finalizeNodeState(nodeID string, state *node.NodeState) {
	if state.IsSucceeded() {
		e.ctx.AddNodeValue(nodeID, state.GetNodeValue())
	}
}

// resolveNodeInput 解析节点输入
func (e *NodeExecutor) resolveNodeInput(n node.Node) *value.ObjectValue {
	resolved := value.NewObjectValue()
	valuesFrom := n.GetValuesFrom()
	if valuesFrom == nil {
		return resolved
	}

	for _, vf := range valuesFrom {
		partial := e.resolveValueFrom(vf)
		resolved.AddAll(partial)
	}

	return resolved
}

// resolveValueFrom 解析值来源
func (e *NodeExecutor) resolveValueFrom(vf *value.ValueFrom) *value.ObjectValue {
	resolved := value.NewObjectValue()

	var source value.NodeValue
	if util.IsBlank(vf.NodeID) {
		source = e.ctx.GetRootValue()
	} else {
		source = e.ctx.GetNodeValue(vf.NodeID)
	}

	if source == nil {
		return resolved
	}

	var selected value.NodeValue
	if util.IsBlank(vf.From) {
		selected = source
	} else {
		selected = source.FindValue(vf.From)
	}

	if selected == nil {
		return resolved
	}

	if util.IsBlank(vf.Param) {
		if !selected.IsObject() {
			panic("ValueFrom requires ObjectValue when param is blank")
		}
		resolved.AddAll(selected.AsObject())
		return resolved
	}

	resolved.Put(vf.Param, selected)
	return resolved
}
package executor

import (
	"context"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/cache"
	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/node"
	pool2 "github.com/chuccp/ai-agent/pool"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/util"
	"github.com/chuccp/ai-agent/value"
	"github.com/sourcegraph/conc/panics"
)

// 错误定义
var (
	ErrWorkflowNoNodes = errors.New("workflow has no nodes")
)

// OnBeforeNodeRun 节点运行前回调
type OnBeforeNodeRun func(state *node.State) error

// OnAfterNodeRun 节点运行后回调
type OnAfterNodeRun func(state *node.State) error

// OnFailedNodeRun 节点运行失败回调
type OnFailedNodeRun func(state *node.State, err error)

// Config 执行器配置
type Config struct {
	RootCachePath        string
	RelPath              string
	MaxConcurrency       int
	WaitingRetryInterval int64 // 毫秒
	SkipRunning          bool
	OnBeforeNodeRun      OnBeforeNodeRun
	OnAfterNodeRun       OnAfterNodeRun
	OnFailedNodeRun      OnFailedNodeRun
	Parameter            *value.ObjectValue
}

// DefaultConfig 默认配置
func DefaultConfig(RootCachePath string, RelPath string) *Config {
	return &Config{
		RootCachePath:        RootCachePath,
		SkipRunning:          true,
		WaitingRetryInterval: int64(5 * time.Second),
		MaxConcurrency:       1,
		Parameter:            value.NewObjectValue(),
		RelPath:              RelPath,
	}
}

// Context 工作流上下文
type Context struct {
	rootValue    *value.ObjectValue
	nodeValues   map[string]value.NodeValue
	parentID     string
	cacheManager *cache.Manager
	config       *Config
	mu           sync.Mutex
	nodeStates   map[string]graph.NodeStatusInterface
	pool2        *pool2.GOPool
}

func NewContext(nodes []node.Node, rootValue *value.ObjectValue, config *Config, pool2 *pool2.GOPool) *Context {
	if rootValue == nil {
		rootValue = value.NewObjectValue()
	}
	ctx := &Context{
		rootValue: rootValue,
		config:    config,
	}

	if config.RelPath != "" {
		ctx.cacheManager = cache.NewManager(config.RelPath, true)
	}
	ctx.pool2 = pool2
	ctx.init(nodes)
	return ctx
}
func (c *Context) GetPool() *pool2.GOPool {
	return c.pool2
}
func (c *Context) init(nodes []node.Node) {
	c.nodeStates = make(map[string]graph.NodeStatusInterface)
	c.nodeValues = make(map[string]value.NodeValue)
	for _, n := range nodes {
		if n.IsSingle() {
			c.nodeStates[n.GetID()] = graph.NewNodeStatus(n.GetID())
		} else {
			c.nodeStates[n.GetID()] = graph.NewNodeStatusGroup(n.GetID())
		}
		c.nodeValues[n.GetID()] = value.NullValue
	}
}

// CreateChildContext 创建子上下文
func (c *Context) CreateChildContext(nodes []node.Node, childRootValue *value.ObjectValue, childParentID string) node.WorkflowContext {
	ctx := &Context{
		rootValue:    childRootValue,
		cacheManager: c.cacheManager,
		parentID:     childParentID,
		config:       c.config,
		pool2:        c.pool2,
	}
	ctx.init(nodes)
	return ctx
}

// AddNodeValue 添加节点值
func (c *Context) AddNodeValue(nodeID string, nodeValue value.NodeValue) {
	c.nodeValues[nodeID] = nodeValue
}

// GetNodeValue 获取节点值
func (c *Context) GetNodeValue(nodeID string) value.NodeValue {
	return c.nodeValues[nodeID]
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
func (c *Context) GetNodeStatuses() []graph.NodeStatusInterface {
	var statuses []graph.NodeStatusInterface
	for _, n := range c.nodeStates {
		statuses = append(statuses, n)
	}
	return statuses
}
func (c *Context) GetNodeStatus(nodeID string) graph.NodeStatusInterface {
	return c.nodeStates[nodeID]
}

// GetConfig 获取配置
func (c *Context) GetConfig() *Config {
	return c.config
}

// GetCacheManager 获取缓存管理器
func (c *Context) GetCacheManager() *cache.Manager {
	return c.cacheManager
}

// SetCacheManager 设置缓存管理器
func (c *Context) SetCacheManager(manager *cache.Manager) {
	c.cacheManager = manager
}

// IsCacheEnabled 是否启用缓存
func (c *Context) IsCacheEnabled() bool {
	return c.cacheManager != nil && c.cacheManager.IsEnabled()
}

// GetCachePath 获取缓存路径
func (c *Context) GetCachePath() string {
	if c.cacheManager == nil {
		return ""
	}
	return c.cacheManager.GetCachePath()
}

// GetCacheKey 生成缓存键
func (c *Context) GetCacheKey(key, nodeID string) string {
	if c.cacheManager == nil {
		return ""
	}
	return c.cacheManager.GenerateCacheKeyWithParent(key, nodeID, c.parentID)
}

// SaveCache 保存缓存
func (c *Context) SaveCache(key, nodeID string, nodeValue value.NodeValue) error {
	if c.cacheManager == nil {
		return nil
	}
	return c.cacheManager.SaveCache(c.GetCacheKey(key, nodeID), nodeValue)
}

// GetCache 获取缓存
func (c *Context) GetCache(key, nodeID string) (value.NodeValue, error) {
	if c.cacheManager == nil {
		return nil, nil
	}
	return c.cacheManager.GetCache(c.GetCacheKey(key, nodeID))
}

// HasCache 检查缓存是否存在
func (c *Context) HasCache(key, nodeID string) bool {
	if c.cacheManager == nil {
		return false
	}
	return c.cacheManager.HasCache(c.GetCacheKey(key, nodeID))
}

// GetOrComputeCache 获取或计算缓存
func (c *Context) GetOrComputeCache(key, nodeID string, fn cache.CacheFunction) (value.NodeValue, error) {
	if c.cacheManager == nil {
		return fn()
	}
	return c.cacheManager.GetOrCompute(c.GetCacheKey(key, nodeID), nodeID, fn)
}

// NodeExecutor 节点执行器
type NodeExecutor struct {
	nodes    []node.Node
	nodeMap  map[string]node.Node
	stateMap map[string]*node.State
	ctx      *Context
	config   *Config
	mu       sync.Mutex
	index    int
}

// NewNodeExecutor 创建节点执行器
func NewNodeExecutor(index int, nodes []node.Node, ctx *Context) *NodeExecutor {
	nodeMap := make(map[string]node.Node)
	var preNode node.Node = nil
	for _, n := range nodes {
		if preNode != nil {
			n.SetPrevNodeID(preNode.GetID())
		}
		nodeMap[n.GetID()] = n
		preNode = n
	}
	return &NodeExecutor{
		nodes:    nodes,
		nodeMap:  nodeMap,
		stateMap: make(map[string]*node.State),
		ctx:      ctx,
		config:   ctx.GetConfig(),
		index:    index,
	}
}

// GetEndNode 获取结束节点
func (e *NodeExecutor) GetEndNode() (node.Node, error) {
	if len(e.nodes) == 0 {
		return nil, ErrWorkflowNoNodes
	}
	return e.nodes[len(e.nodes)-1], nil
}
func (e *NodeExecutor) GetIndex() int {
	return e.index
}

// Exec 执行
func (e *NodeExecutor) Exec(ctx context.Context, pool *pool2.GOPool) (value.NodeValue, error) {
	endNode, err := e.GetEndNode()
	if err != nil {
		return nil, err
	}
	layers, err := BuildExecutionLayers(e.nodeMap, endNode)
	if err != nil {
		return nil, err
	}

	for _, layer := range layers {
		fa, err := e.executeLayer(layer, ctx, pool)
		if err != nil {
			return nil, err
		}
		if !fa {
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

// executeLayer 执行层级（使用goroutine并发执行，带panic恢复）
// 返回值: (是否全部成功, 错误)
func (e *NodeExecutor) executeLayer(layer []node.Node, ctx context.Context, pool2 *pool2.GOPool) (bool, error) {
	if len(layer) == 0 {
		return true, nil
	}
	// 单节点直接执行
	if len(layer) == 1 {
		state, err := e.createAndRunNodeStateSafe(layer[0])
		if err != nil {
			return false, err
		}
		e.finalizeNodeState(layer[0].GetID(), state)
		return state.IsSucceeded(), nil
	}
	// 多节点并发执行
	results := make([]*executionResult, len(layer))
	err := pool2.WaitGOIndex(len(layer), func(index int) error {
		idx := index
		nodeObj := layer[idx]
		state, err := e.createAndRunNodeState(nodeObj)
		if err != nil {
			return err
		}
		results[idx] = &executionResult{nodeID: nodeObj.GetID(), state: state, err: err}
		return nil
	})
	if err != nil {
		return false, err
	}
	for _, result := range results {
		if result.err != nil {
			return false, result.err
		}
		e.finalizeNodeState(result.nodeID, result.state)
	}
	return e.IsAllSucceeded(), nil
}

// executionResult 执行结果
type executionResult struct {
	nodeID string
	state  *node.State
	err    error
}

// createAndRunNodeStateSafe 创建并运行节点状态（带panic恢复）
func (e *NodeExecutor) createAndRunNodeStateSafe(n node.Node) (*node.State, error) {
	var state *node.State
	var err error
	r := panics.Try(func() {
		state, err = e.createAndRunNodeState(n)
	})
	if r != nil {
		return nil, r.AsError()
	}
	return state, err
}
func (e *NodeExecutor) getNodeStatus(nodeID string) graph.NodeStatusInterface {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.ctx.nodeStates[nodeID]
}

// createAndRunNodeState 创建并运行节点状态
func (e *NodeExecutor) createAndRunNodeState(n node.Node) (*node.State, error) {
	input := e.resolveNodeInput(n)
	state := node.NewNodeState(e.ctx, n.GetID(), input, e.config.Parameter)
	e.mu.Lock()
	e.stateMap[n.GetID()] = state
	e.mu.Unlock()
	state.SetStatusType(types.NodeStatusStarted)
	// 执行监听函数
	if n.GetWatch() != nil {
		if err := n.GetWatch()(state); err != nil {
			return state, err
		}
	}

	// 执行前置回调
	if e.config.OnBeforeNodeRun != nil {
		if err := e.config.OnBeforeNodeRun(state); err != nil {
			state.SetStatusType(types.NodeStatusFailed)
			if e.config.OnFailedNodeRun != nil {
				e.config.OnFailedNodeRun(state, err)
			}
			return state, err
		}
	}

	// 执行节点
	output, err := n.Exec(state)
	if err != nil {
		err = errors.Append(errors.Errorf("node %s exec failed", n.GetID()), err)
		state.SetStatusType(types.NodeStatusFailed)
		if e.config.OnFailedNodeRun != nil {
			e.config.OnFailedNodeRun(state, err)
		}
		return state, err
	}
	state.SetNodeValue(output)
	if state.GetStatusType() == types.NodeStatusStarted && (output != nil && !output.IsNull()) {
		state.SetStatusType(types.NodeStatusSucceeded)
	}

	// 执行后置回调
	if e.config.OnAfterNodeRun != nil {
		if err := e.config.OnAfterNodeRun(state); err != nil {
			return state, err
		}
	}

	return state, nil
}

// finalizeNodeState 完成节点状态
func (e *NodeExecutor) finalizeNodeState(nodeID string, state *node.State) {
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
			// 返回空结果而不是panic
			return resolved
		}
		resolved.AddAll(selected.AsObject())
		return resolved
	}

	resolved.Put(vf.Param, selected)
	return resolved
}

func (e *NodeExecutor) GetNodeStatus() []graph.NodeStatusInterface {
	return e.ctx.GetNodeStatuses()

}

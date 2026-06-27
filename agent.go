package ai_agent

import (
	"context"
	"sync"
	"time"

	"emperror.dev/errors"
	pool2 "github.com/chuccp/ai-agent/pool"
	"github.com/sourcegraph/conc/pool"

	"github.com/chuccp/ai-agent/executor"
	"github.com/chuccp/ai-agent/graph"
	"github.com/chuccp/ai-agent/node"
	"github.com/chuccp/ai-agent/types"
	"github.com/chuccp/ai-agent/util"
	"github.com/chuccp/ai-agent/value"
)

// Response 响应
type Response struct {
	NodeValue value.NodeValue
	Success   bool
}

// NewResponse 创建响应
func NewResponse(nodeValue value.NodeValue, success bool) *Response {
	return &Response{
		NodeValue: nodeValue,
		Success:   success,
	}
}

// Workflow 工作流
type Workflow struct {
	nodes []node.Node
	mu    sync.RWMutex
}

// NewWorkflow 创建工作流
func NewWorkflow(nodes []node.Node) *Workflow {
	return &Workflow{
		nodes: nodes,
	}
}

// GetRootParams 获取根参数
func (w *Workflow) GetRootParams() []string {
	params := make(map[string]bool)
	for _, n := range w.nodes {
		valuesFrom := n.GetValuesFrom()
		if valuesFrom != nil {
			for _, vf := range valuesFrom {
				if util.IsBlank(vf.NodeID) && util.IsNotBlank(vf.Param) {
					params[vf.Param] = true
				}
			}
		}
	}

	result := make([]string, 0, len(params))
	for p := range params {
		result = append(result, p)
	}
	return result
}

// GetOutParams 获取输出参数
func (w *Workflow) GetOutParams() []string {
	params := make(map[string]bool)
	for _, n := range w.nodes {
		if _, ok := n.(*node.OutputNode); ok {
			valuesFrom := n.GetValuesFrom()
			if valuesFrom != nil {
				for _, vf := range valuesFrom {
					if util.IsNotBlank(vf.Param) {
						params[vf.Param] = true
					}
				}
			}
		}
	}

	result := make([]string, 0, len(params))
	for p := range params {
		result = append(result, p)
	}
	return result
}

// Exec 执行工作流
func (w *Workflow) Exec(ctx node.WorkflowContext) (value.NodeValue, error) {
	execCtx, ok := ctx.(*executor.Context)
	if !ok {
		return nil, errors.New("invalid context type for Workflow.Exec")
	}
	ids := make(map[string]bool)
	for _, n := range w.nodes {
		if _, ok := ids[n.GetID()]; !ok {
			ids[n.GetID()] = true
		} else {
			return nil, errors.New(" Node ID already exists id:" + n.GetID())
		}
	}

	nodeExec := executor.NewNodeExecutor(0, w.nodes, execCtx)
	return nodeExec.Exec(execCtx.GetPool())
}

// Execute 实现node.WorkflowInterface接口，用于IFNode等条件节点执行子工作流
func (w *Workflow) Execute(ctx node.WorkflowContext, input *value.ObjectValue, parentID string) (value.NodeValue, error) {
	childCtx := ctx.CreateChildContext(w.nodes, input, value.NewArrayValue(), parentID, -1)
	return w.Exec(childCtx)
}
func (w *Workflow) ExecBatchOrder(ctx node.WorkflowContext, statusGroup *graph.NodeStatusGroup, parentID string, inputs []*value.ObjectValue) (*value.ArrayValue, bool, error) {
	execCtx, ok := ctx.(*executor.Context)
	if !ok {
		return nil, false, errors.New("invalid context type for Workflow.ExecBatch")
	}
	groupExec := executor.NewGroupExecutor(w.nodes, parentID, execCtx, context.Background())
	nodeValue, fa, err := groupExec.ExecBatch(statusGroup, inputs, true)
	return nodeValue, fa, err
}

// ExecBatch 批量执行
func (w *Workflow) ExecBatch(ctx node.WorkflowContext, statusGroup *graph.NodeStatusGroup, parentID string, inputs []*value.ObjectValue) (*value.ArrayValue, bool, error) {
	execCtx, ok := ctx.(*executor.Context)
	if !ok {
		return nil, false, errors.New("invalid context type for Workflow.ExecBatch")
	}
	groupExec := executor.NewGroupExecutor(w.nodes, parentID, execCtx, context.Background())
	nodeValue, fa, err := groupExec.ExecBatch(statusGroup, inputs, false)
	return nodeValue, fa, err
}

// GetGraph 获取图
func (w *Workflow) GetGraph() []*graph.NodeGraph {
	graphs := make([]*graph.NodeGraph, 0, len(w.nodes))
	for _, n := range w.nodes {
		graphs = append(graphs, n.GetNodeGraph())
	}
	return graphs
}
func (w *Workflow) GetNodes() []node.Node {
	return w.nodes
}

// GetGraphs 实现 WorkflowInterface 接口
func (w *Workflow) GetGraphs() []*graph.NodeGraph {
	return w.GetGraph()
}

// WorkflowBuilder 工作流构建器
type WorkflowBuilder struct {
	nodes []node.Node
}

// NewWorkflowBuilder 创建工作流构建器
func NewWorkflowBuilder() *WorkflowBuilder {
	return &WorkflowBuilder{
		nodes: make([]node.Node, 0),
	}
}

// AddNode 添加节点
func (b *WorkflowBuilder) AddNode(n node.Node) *WorkflowBuilder {
	b.nodes = append(b.nodes, n)
	return b
}

// Build 构建
func (b *WorkflowBuilder) Build() (*Workflow, error) {
	if len(b.nodes) == 0 {
		return nil, ErrWorkflowRequiresNode
	}
	return NewWorkflow(b.nodes), nil
}

// Of 创建工作流
func Of(nodes ...node.Node) *Workflow {
	builder := NewWorkflowBuilder()
	for _, n := range nodes {
		builder.AddNode(n)
	}
	w, _ := builder.Build()
	return w
}

// OnBeforeNodeRun 节点运行前回调（带executorID）
type OnBeforeNodeRun func(executorID string, state *node.State) error

// OnAfterNodeRun 节点运行后回调（带executorID）
type OnAfterNodeRun func(executorID string, state *node.State) error

// OnFailedNodeRun 节点运行失败回调（带executorID）
type OnFailedNodeRun func(executorID string, state *node.State, err error)

// Config 配置

// DefaultConfig 默认配置

// Agent 代理
type Agent struct {
	workflow *Workflow
	id       string
}

// NewAgent 创建代理
func NewAgent(id string, workflow *Workflow) *Agent {
	return &Agent{
		workflow: workflow,
		id:       id,
	}
}

// GetID 获取ID
func (a *Agent) GetID() string {
	return a.id
}

// GetWorkflow 获取工作流
func (a *Agent) GetWorkflow() *Workflow {
	return a.workflow
}

// GetConfig 获取配置

// GetGraph 获取图
func (a *Agent) GetGraph() *graph.Graph {
	if a.workflow == nil {
		return graph.NewGraph(nil, nil, nil)
	}
	return graph.NewGraph(a.workflow.GetGraph(), a.workflow.GetRootParams(), a.workflow.GetOutParams())
}

// GetParams 获取参数
func (a *Agent) GetParams() []string {
	return a.GetGraph().GetParams()
}

// AgentBuilder 代理构建器
type AgentBuilder struct {
	id       string
	workflow *Workflow
}

// NewAgentBuilder 创建代理构建器
func NewAgentBuilder(id string) *AgentBuilder {
	return &AgentBuilder{
		id: id,
	}
}

// Workflow 设置工作流
func (b *AgentBuilder) Workflow(w *Workflow) *AgentBuilder {
	b.workflow = w
	return b
}

// Build 构建
func (b *AgentBuilder) Build() *Agent {
	return NewAgent(b.id, b.workflow)
}

// AgentExecutor 代理执行器
type AgentExecutor struct {
	agent           *Agent
	ctx             *executor.Context
	id              string
	inputValue      *value.ObjectValue
	asyncCall       *AsyncCall
	config          *executor.Config
	onBeforeNodeRun OnBeforeNodeRun
	onAfterNodeRun  OnAfterNodeRun
	onFailedNodeRun OnFailedNodeRun
	pool0           *pool2.GOPool
}

// NewAgentExecutor 创建代理执行器
func NewAgentExecutor(agent *Agent, execConfig *executor.Config) *AgentExecutor {
	return NewAgentExecutorWithExecutorId("", agent, execConfig)
}
func NewAgentExecutorWithExecutorId(executorId string, agent *Agent, execConfig *executor.Config) *AgentExecutor {
	if len(executorId) == 0 {
		executorId = agent.GetID() + "#" + util.GenerateUUID()
	}
	if len(execConfig.RootCachePath) > 0 {
		if execConfig.RelPath == "" {
			md5 := util.MD5(executorId)
			execConfig.RelPath = util.PathJoin(execConfig.RootCachePath, md5[0:2], md5)
		} else {
			execConfig.RelPath = util.PathJoin(execConfig.RootCachePath, execConfig.RelPath)
		}
	}
	nodes := agent.workflow.GetNodes()
	pool0 := pool2.NewGOPool(execConfig.MaxConcurrency)
	ctx := executor.NewContext(executorId, nodes, value.NewObjectValue(), execConfig, pool0)
	return &AgentExecutor{
		agent:      agent,
		id:         executorId,
		inputValue: value.NewObjectValue(),
		config:     execConfig,
		ctx:        ctx,
		pool0:      pool0,
		asyncCall:  NewAsyncCall0(agent, ctx, pool0),
	}
}

// GetID 获取ID
func (e *AgentExecutor) GetID() string {
	return e.id
}
func (e *AgentExecutor) GetCachePath() string {
	return e.config.RelPath
}

// GetAgent 获取代理
func (e *AgentExecutor) GetAgent() *Agent {
	return e.agent
}

// GetParams 获取参数
func (e *AgentExecutor) GetParams() []string {
	return e.agent.GetParams()
}

// GetGraphStatus 获取图状态
func (e *AgentExecutor) GetGraphStatus() *graph.GraphStatus {
	return graph.NewGraphStatus(e.ctx.GetNodeStatuses())
}

// IsRunning 是否运行中
func (e *AgentExecutor) IsRunning() bool {
	return e.asyncCall != nil && !e.asyncCall.IsDone()
}

// IsDone 是否完成
func (e *AgentExecutor) IsDone() bool {
	return e.asyncCall != nil && e.asyncCall.IsDone()
}

// ExecSync 同步执行（返回AsyncResult）
func (e *AgentExecutor) ExecSync() *AsyncResult {
	var asyncResult = &AsyncResult{ExecutorId: e.id}
	e.asyncCall.status = types.AgentStatusStarted
	er := e.pool0.WaitGO(func() error {
		resp, err := e.asyncCall.ExecSync()
		if err != nil {
			return err
		}
		asyncResult.Response = resp
		asyncResult.Error = err
		return nil
	})
	if er != nil {
		e.asyncCall.status = types.AgentStatusFailed
	} else {
		e.asyncCall.status = types.AgentStatusSucceeded
	}

	if er != nil {
		asyncResult.Error = er
	}
	return asyncResult
}

func (e *AgentExecutor) ExecSyncForInput(input *value.ObjectValue) *AsyncResult {
	e.prepareInput(input)
	return e.ExecSync()
}

// Cancel 取消
func (e *AgentExecutor) Cancel() bool {
	if e.asyncCall != nil {
		return e.asyncCall.Cancel()
	}
	return false
}

// SetOnBeforeNodeRun 设置节点运行前回调
func (e *AgentExecutor) SetOnBeforeNodeRun(fn OnBeforeNodeRun) {
	e.onBeforeNodeRun = fn
}

// SetOnAfterNodeRun 设置节点运行后回调
func (e *AgentExecutor) SetOnAfterNodeRun(fn OnAfterNodeRun) {
	e.onAfterNodeRun = fn
}

// SetOnFailedNodeRun 设置节点运行失败回调
func (e *AgentExecutor) SetOnFailedNodeRun(fn OnFailedNodeRun) {
	e.onFailedNodeRun = fn
}

// prepareInput 准备输入
func (e *AgentExecutor) prepareInput(input *value.ObjectValue) {
	e.inputValue.Clear()

	// 设置回调函数
	execConfig := e.ctx.GetConfig()
	if execConfig != nil {
		if e.onBeforeNodeRun != nil {
			execConfig.OnBeforeNodeRun = func(state *node.State) error {
				return e.onBeforeNodeRun(e.id, state)
			}
		}
		if e.onAfterNodeRun != nil {
			execConfig.OnAfterNodeRun = func(state *node.State) error {
				return e.onAfterNodeRun(e.id, state)
			}
		}
		if e.onFailedNodeRun != nil {
			execConfig.OnFailedNodeRun = func(state *node.State, err error) {
				e.onFailedNodeRun(e.id, state, err)
			}
		}
	}

	if input != nil {
		e.inputValue.AddAll(input)
	}
	// 设置根值
	e.ctx.SetRootValue(e.inputValue)
}

func (e *AgentExecutor) ExecAsync(input *value.ObjectValue) *AsyncResult {
	e.prepareInput(input)
	var wg = pool.New()
	errorPool := wg.WithMaxGoroutines(1).WithErrors().WithFirstError()
	result := &AsyncResult{}
	e.asyncCall.status = types.AgentStatusStarted
	errorPool.Go(func() error {
		resp, err := e.asyncCall.ExecSync()
		result.Response = resp
		result.Error = err
		return err
	})
	err := errorPool.Wait()
	if err != nil {
		e.asyncCall.status = types.AgentStatusFailed
	} else {
		e.asyncCall.status = types.AgentStatusSucceeded
	}
	if err != nil {
		result.Error = err
		return nil
	}
	return result
}

// AsyncResult 异步结果
type AsyncResult struct {
	Response   *Response
	Error      error
	ExecutorId string
}

// AsyncCall 异步调用
type AsyncCall struct {
	agent    *Agent
	ctx      *executor.Context
	status   types.AgentStatusType
	cancelFn context.CancelFunc
	mu       sync.Mutex
	pool2    *pool2.GOPool
}

// NewAsyncCall 创建异步调用
func NewAsyncCall0(agent *Agent, ctx *executor.Context, pool2 *pool2.GOPool) *AsyncCall {
	return &AsyncCall{
		agent:  agent,
		ctx:    ctx,
		status: types.AgentStatusWaiting,
		pool2:  pool2,
	}
}

// ExecSync 同步执行
func (c *AsyncCall) ExecSync() (*Response, error) {
	nodeValue, err := c.agent.GetWorkflow().Exec(c.ctx)
	if err != nil {
		return nil, err
	}
	return NewResponse(nodeValue, nodeValue != nil && !nodeValue.IsNull()), nil
}

// Cancel 取消
func (c *AsyncCall) Cancel() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancelFn != nil {
		c.cancelFn()
	}
	c.status = types.AgentStatusFailed
	return true
}

// GetStatus 获取状态
func (c *AsyncCall) GetStatus() types.AgentStatusType {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status
}

// IsDone 是否完成
func (c *AsyncCall) IsDone() bool {
	status := c.GetStatus()
	return status == types.AgentStatusSucceeded || status == types.AgentStatusFailed
}

// millisToDuration 毫秒转Duration
func millisToDuration(millis int64) time.Duration {
	return time.Duration(millis) * time.Millisecond
}

// 错误定义
var (
	ErrWorkflowRequiresNode = errors.New("Workflow requires at least one node")
)

package agent

import (
	"context"
	"errors"
	"sync"
	"time"

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
func (w *Workflow) Exec(ctx *executor.Context) (*Response, error) {
	nodeExec := executor.NewNodeExecutor(w.nodes, ctx)
	nodeValue, err := nodeExec.Exec()
	if err != nil {
		return nil, err
	}
	return NewResponse(nodeValue, nodeExec.IsAllSucceeded()), nil
}

// ExecBatch 批量执行
func (w *Workflow) ExecBatch(ctx *executor.Context, statusGroup *graph.NodeStatusGroup, parentID string, inputs []*value.ObjectValue) (*Response, error) {
	groupExec := executor.NewGroupExecutor(w.nodes, parentID, ctx)
	nodeValue, success, err := groupExec.ExecBatch(statusGroup, inputs)
	if err != nil {
		return nil, err
	}
	return NewResponse(nodeValue, success), nil
}

// GetGraph 获取图
func (w *Workflow) GetGraph() []*graph.NodeGraph {
	graphs := make([]*graph.NodeGraph, 0, len(w.nodes))
	for _, n := range w.nodes {
		graphs = append(graphs, n.GetNodeGraph())
	}
	return graphs
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

// Config 配置
type Config struct {
	CachePath            string
	MaxConcurrency       int
	WaitingRetryInterval int64 // 毫秒
	SkipRunning          bool
	OnBeforeNodeRun      executor.OnBeforeNodeRun
	OnAfterNodeRun       executor.OnAfterNodeRun
	OnFailedNodeRun      executor.OnFailedNodeRun
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxConcurrency:       -1,
		WaitingRetryInterval: 5000,
		SkipRunning:          true,
	}
}

// Agent 代理
type Agent struct {
	workflow *Workflow
	id       string
	config   *Config
}

// NewAgent 创建代理
func NewAgent(id string, workflow *Workflow, config *Config) *Agent {
	if config == nil {
		config = DefaultConfig()
	}
	return &Agent{
		workflow: workflow,
		id:       id,
		config:   config,
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
func (a *Agent) GetConfig() *Config {
	return a.config
}

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
	config   *Config
}

// NewAgentBuilder 创建代理构建器
func NewAgentBuilder(id string) *AgentBuilder {
	return &AgentBuilder{
		id:     id,
		config: DefaultConfig(),
	}
}

// Workflow 设置工作流
func (b *AgentBuilder) Workflow(w *Workflow) *AgentBuilder {
	b.workflow = w
	return b
}

// Config 设置配置
func (b *AgentBuilder) Config(c *Config) *AgentBuilder {
	b.config = c
	return b
}

// CachePath 设置缓存路径
func (b *AgentBuilder) CachePath(path string) *AgentBuilder {
	b.config.CachePath = path
	return b
}

// MaxConcurrency 设置最大并发
func (b *AgentBuilder) MaxConcurrency(max int) *AgentBuilder {
	b.config.MaxConcurrency = max
	return b
}

// Build 构建
func (b *AgentBuilder) Build() *Agent {
	return NewAgent(b.id, b.workflow, b.config)
}

// AgentExecutor 代理执行器
type AgentExecutor struct {
	agent      *Agent
	ctx        *executor.Context
	id         string
	inputValue *value.ObjectValue
	asyncCall  *AsyncCall
	config     *Config
}

// NewAgentExecutor 创建代理执行器
func NewAgentExecutor(agent *Agent, config *Config) *AgentExecutor {
	if config == nil {
		config = DefaultConfig()
	}
	id := agent.GetID() + "#" + util.GenerateUUID()

	execConfig := &executor.Config{
		SkipRunning:          config.SkipRunning,
		WaitingRetryInterval: millisToDuration(config.WaitingRetryInterval),
		OnBeforeNodeRun:      config.OnBeforeNodeRun,
		OnAfterNodeRun:       config.OnAfterNodeRun,
		OnFailedNodeRun:      config.OnFailedNodeRun,
	}

	return &AgentExecutor{
		agent:      agent,
		id:         id,
		inputValue: value.NewObjectValue(),
		config:     config,
		ctx:        executor.NewContext(value.NewObjectValue(), execConfig),
	}
}

// GetID 获取ID
func (e *AgentExecutor) GetID() string {
	return e.id
}

// GetAgent 获取代理
func (e *AgentExecutor) GetAgent() *Agent {
	return e.agent
}

// GetParams 获取参数
func (e *AgentExecutor) GetParams() []string {
	return e.agent.GetParams()
}

// GetStatus 获取状态
func (e *AgentExecutor) GetStatus() types.AgentStatusType {
	if e.asyncCall == nil {
		return types.AgentStatusWaiting
	}
	return e.asyncCall.GetStatus()
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

// Exec 执行
func (e *AgentExecutor) Exec(input *value.ObjectValue) (*Response, error) {
	e.prepareInput(input)
	return e.agent.GetWorkflow().Exec(e.ctx)
}

// ExecJSON 从JSON执行
func (e *AgentExecutor) ExecJSON(inputJSON string) (*Response, error) {
	input, err := value.ParseObjectValue([]byte(inputJSON))
	if err != nil {
		return nil, err
	}
	return e.Exec(input)
}

// ExecAsync 异步执行
func (e *AgentExecutor) ExecAsync(input *value.ObjectValue) <-chan *AsyncResult {
	e.prepareInput(input)
	e.asyncCall = NewAsyncCall(e.agent, e.ctx)

	resultChan := make(chan *AsyncResult, 1)

	go func() {
		defer close(resultChan)
		resp, err := e.asyncCall.ExecSync()
		resultChan <- &AsyncResult{Response: resp, Error: err}
	}()

	return resultChan
}

// Cancel 取消
func (e *AgentExecutor) Cancel() bool {
	if e.asyncCall != nil {
		return e.asyncCall.Cancel()
	}
	return false
}

// prepareInput 准备输入
func (e *AgentExecutor) prepareInput(input *value.ObjectValue) {
	e.ctx.ClearExecutionState()
	e.inputValue.Clear()

	if input != nil {
		e.inputValue.AddAll(input)
	}
	// 设置根值
	e.ctx.SetRootValue(e.inputValue)
}

// AsyncResult 异步结果
type AsyncResult struct {
	Response *Response
	Error    error
}

// AsyncCall 异步调用
type AsyncCall struct {
	agent    *Agent
	ctx      *executor.Context
	status   types.AgentStatusType
	cancelFn context.CancelFunc
	mu       sync.Mutex
}

// NewAsyncCall 创建异步调用
func NewAsyncCall(agent *Agent, ctx *executor.Context) *AsyncCall {
	return &AsyncCall{
		agent:  agent,
		ctx:    ctx,
		status: types.AgentStatusWaiting,
	}
}

// ExecSync 同步执行
func (c *AsyncCall) ExecSync() (*Response, error) {
	c.mu.Lock()
	c.status = types.AgentStatusStarted
	c.mu.Unlock()

	c.mu.Lock()
	c.status = types.AgentStatusRunning
	c.mu.Unlock()

	resp, err := c.agent.GetWorkflow().Exec(c.ctx)

	c.mu.Lock()
	if err != nil {
		c.status = types.AgentStatusFailed
	} else {
		c.status = types.AgentStatusSucceeded
	}
	c.mu.Unlock()

	return resp, err
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
package ai_agent

import (
	"sync"

	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/executor"
	"github.com/chuccp/ai-agent/util"
	"github.com/chuccp/ai-agent/value"
)

// AgentManager Agent管理器
type AgentManager struct {
	agentRegistry    sync.Map // map[string]*Agent
	executorRegistry sync.Map // map[string]*AgentExecutor
	mu               sync.RWMutex
}

func NewAgentManager() *AgentManager {
	return &AgentManager{}
}

// RegisterAgent 注册Agent
func (m *AgentManager) RegisterAgent(agent *Agent) {
	m.agentRegistry.Store(agent.GetID(), agent)
}

// AddAgent 添加Agent
func (m *AgentManager) AddAgent(agent *Agent) {
	m.RegisterAgent(agent)
}

// UnregisterAgent 注销Agent
func (m *AgentManager) UnregisterAgent(agentID string) {
	m.agentRegistry.Delete(agentID)
}

// GetAgent 获取Agent
func (m *AgentManager) GetAgent(agentID string) (*Agent, bool) {
	if v, ok := m.agentRegistry.Load(agentID); ok {
		return v.(*Agent), true
	}
	return nil, false
}

// GetAllAgents 获取所有Agent
func (m *AgentManager) GetAllAgents() []*Agent {
	var agents []*Agent
	m.agentRegistry.Range(func(key, value interface{}) bool {
		agents = append(agents, value.(*Agent))
		return true
	})
	return agents
}

// CreateExecutor 创建执行器
func (m *AgentManager) CreateExecutor(agentID string, execConfig *executor.Config) (*AgentExecutor, error) {
	agent, ok := m.GetAgent(agentID)
	if !ok {
		return nil, errors.New("Agent not found: " + agentID)
	}
	return m.CreateExecutorForAgent(agent, execConfig)
}

// CreateExecutorForAgent 为Agent创建执行器
func (m *AgentManager) CreateExecutorForAgent(agent *Agent, execConfig *executor.Config) (*AgentExecutor, error) {

	exec := NewAgentExecutor(agent, execConfig)
	m.executorRegistry.Store(exec.GetID(), exec)
	return exec, nil
}

// CreateExecutorWithID 创建带ID的执行器
func (m *AgentManager) CreateExecutorWithID(executorID string, agent *Agent, execConfig *executor.Config) *AgentExecutor {
	exec := NewAgentExecutorWithExecutorId(executorID, agent, execConfig)
	m.executorRegistry.Store(exec.GetID(), exec)
	return exec
}

// GetExecutor 获取执行器
func (m *AgentManager) GetExecutor(executorID string) (*AgentExecutor, bool) {
	if v, ok := m.executorRegistry.Load(executorID); ok {
		return v.(*AgentExecutor), true
	}
	return nil, false
}

// RemoveExecutor 移除执行器
func (m *AgentManager) RemoveExecutor(executorID string) {
	m.executorRegistry.Delete(executorID)
}

// GetAllExecutors 获取所有执行器
func (m *AgentManager) GetAllExecutors() []*AgentExecutor {
	var executors []*AgentExecutor
	m.executorRegistry.Range(func(key, value interface{}) bool {
		executors = append(executors, value.(*AgentExecutor))
		return true
	})
	return executors
}

// Execute 执行
func (m *AgentManager) Execute(agent *Agent, input *value.ObjectValue, config *executor.Config) (*Response, error) {
	exec, err := m.CreateExecutorForAgent(agent, config)
	if err != nil {
		return nil, err
	}
	return exec.Exec(input)
}

// ExecuteWithID 使用执行器ID执行
func (m *AgentManager) ExecuteWithID(executorID string, input *value.ObjectValue) (*Response, error) {
	exec, ok := m.GetExecutor(executorID)
	if !ok {
		return nil, errors.New("Executor not found: " + executorID)
	}
	return exec.Exec(input)
}

// ExecuteJSON 使用JSON输入执行
func (m *AgentManager) ExecuteJSON(executorID string, inputJSON string) (*Response, error) {
	exec, ok := m.GetExecutor(executorID)
	if !ok {
		return nil, errors.New("Executor not found: " + executorID)
	}
	return exec.ExecJSON(inputJSON)
}

// ExecuteAsync 异步执行
func (m *AgentManager) ExecuteAsync(executorID string, input *value.ObjectValue) (*AsyncResult, error) {
	exec, ok := m.GetExecutor(executorID)
	if !ok {
		return nil, errors.New("Executor not found: " + executorID)
	}
	return exec.ExecAsync(input), nil
}

// GetOrCreateExecutor 获取或创建执行器
func (m *AgentManager) GetOrCreateExecutor(executorID, agentID string, execConfig *executor.Config) (*AgentExecutor, error) {
	exec, ok := m.GetExecutor(executorID)
	if ok {
		if exec.IsRunning() {
			return nil, errors.New("Executor is running: " + executorID)
		}
		return exec, nil
	}

	agent, ok := m.GetAgent(agentID)
	if !ok {
		return nil, errors.New("Agent not found: " + agentID)
	}

	return m.CreateExecutorWithID(executorID, agent, execConfig), nil
}

// CancelExecutor 取消执行器
func (m *AgentManager) CancelExecutor(executorID string) bool {
	exec, ok := m.GetExecutor(executorID)
	if ok {
		return exec.Cancel()
	}
	return false
}

// Shutdown 关闭管理器
func (m *AgentManager) Shutdown() {
	m.executorRegistry.Range(func(key, value interface{}) bool {
		// 执行器没有shutdown方法，直接删除
		m.executorRegistry.Delete(key)
		return true
	})
}

// CreateExecutorWithContext 为Agent创建带上下文的执行器
func (m *AgentManager) CreateExecutorWithContext(agent *Agent, ctx *executor.Context, execConfig *executor.Config) *AgentExecutor {
	exec := &AgentExecutor{
		agent:      agent,
		ctx:        ctx,
		inputValue: value.NewObjectValue(),
		config:     execConfig,
	}
	exec.SetID(agent.GetID() + "#" + util.GenerateUUID())
	m.executorRegistry.Store(exec.GetID(), exec)
	return exec
}

package ai_agent

import (
	"sync"
	"sync/atomic"
	"time"

	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/executor"
	"github.com/chuccp/ai-agent/value"
	"github.com/sourcegraph/conc/pool"
)

// AgentManager Agent管理器
type AgentManager struct {
	agentRegistry    *sync.Map // map[string]*Agent
	executorRegistry *sync.Map // map[string]*AgentExecutor
	mu               *sync.RWMutex
	runStatus        atomic.Value // *RunStatus
}

func NewAgentManager() *AgentManager {
	return &AgentManager{
		agentRegistry:    new(sync.Map),
		executorRegistry: new(sync.Map),
		mu:               new(sync.RWMutex),
	}
}

// GetRunStatus 获取当前任务运行状态
func (m *AgentManager) GetRunStatus() *RunStatus {
	if v := m.runStatus.Load(); v != nil {
		return v.(*RunStatus)
	}
	return nil
}

// RegisterExecutor 注册执行器
func (m *AgentManager) RegisterExecutor(exec *AgentExecutor) {
	m.executorRegistry.Store(exec.GetID(), exec)
}

// UnregisterExecutor 注销执行器
func (m *AgentManager) UnregisterExecutor(executorId string) {
	m.executorRegistry.Delete(executorId)
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
	return m.createExecutorForAgent(agent, execConfig)
}

// // CreateExecutorForAgent 为Agent创建执行器
func (m *AgentManager) createExecutorForAgent(agent *Agent, execConfig *executor.Config) (*AgentExecutor, error) {
	exec := NewAgentExecutor(agent, execConfig)
	m.RegisterExecutor(exec)
	return exec, nil
}
func (m *AgentManager) ExecSync(agentExecutor *AgentExecutor, input *value.ObjectValue) *AsyncResult {

	asyncResult := agentExecutor.ExecSync(input)

	return asyncResult
}

// CreateExecutorWithID // CreateExecutorWithID 创建带ID的执行器
func (m *AgentManager) CreateExecutorWithID(executorID string, agent *Agent, execConfig *executor.Config) *AgentExecutor {
	exec := NewAgentExecutorWithExecutorId(executorID, agent, execConfig)
	m.RegisterExecutor(exec)
	return exec
}

// GetExecutor 根据 executorId 获取执行器
func (m *AgentManager) GetExecutor(executorId string) (*AgentExecutor, bool) {
	if v, ok := m.executorRegistry.Load(executorId); ok {
		return v.(*AgentExecutor), true
	}
	return nil, false
}

// GetAllLiveAgentExecutor 获取所有活跃的执行器
func (m *AgentManager) GetAllLiveAgentExecutor() []*AgentExecutor {
	var live []*AgentExecutor
	m.executorRegistry.Range(func(key, value interface{}) bool {
		exec := value.(*AgentExecutor)
		if exec.IsRunning() {
			live = append(live, exec)
		}
		return true
	})
	return live
}

// GetAllAgentExecutor 获取所有执行器
func (m *AgentManager) GetAllAgentExecutor() []*AgentExecutor {
	var all []*AgentExecutor
	m.executorRegistry.Range(func(key, value interface{}) bool {
		all = append(all, value.(*AgentExecutor))
		return true
	})
	return all
}

// RunStatus 任务队列运行状态
type RunStatus struct {
	RunningCountTotal int          // 任务总数
	RunningCount      atomic.Int64 // 已完成数（原子读写）
	RunTime           time.Time    // 开始时间
	LastRunTime       time.Time    // 最后完成时间
}

// TaskHandler 任务执行回调
type TaskHandler[T any] func(item T)

// SetRunStatus 设置运行状态
func (m *AgentManager) SetRunStatus(status *RunStatus) {
	m.runStatus.Store(status)
}

// ProcessTasks 批量并发执行任务，使用 sourcegraph/conc/pool 管理并发，自动追踪 RunStatus。
func (m *AgentManager) ProcessTasks(items []*AgentExecutor, maxConcurrency int) {
	if len(items) == 0 {
		return
	}
	if handler == nil {
		panic("ProcessTasks: handler must not be nil")
	}

	status.RunningCountTotal = len(items)
	status.RunningCount.Store(0)
	status.RunTime = time.Now()

	p := pool.New().WithMaxGoroutines(maxConcurrency)
	for _, item := range items {
		p.Go(func() {
			item.Exec()
			status.RunningCount.Add(1)
			status.LastRunTime = time.Now()
		})
	}
	p.Wait()
}

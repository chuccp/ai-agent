package ai_agent

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/executor"
	"github.com/chuccp/ai-agent/value"
)

// AgentManager Agent管理器
type AgentManager struct {
	agentRegistry    *sync.Map // map[string]*Agent
	executorRegistry *sync.Map // map[string]*AgentExecutor
	mu               *sync.RWMutex
}

func NewAgentManager() *AgentManager {
	return &AgentManager{
		agentRegistry:    new(sync.Map),
		executorRegistry: new(sync.Map),
		mu:               new(sync.RWMutex),
	}
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

// TaskQueue 泛型任务队列，Process 接收数组和回调，最多 maxConcurrency 个并发。
type TaskQueue[T any] struct {
	maxConcurrency int
}

// NewTaskQueue 创建任务队列，maxConcurrency 必须 >= 1。
func NewTaskQueue[T any](maxConcurrency int) *TaskQueue[T] {
	if maxConcurrency < 1 {
		panic(fmt.Sprintf("NewTaskQueue: maxConcurrency must be >= 1, got %d", maxConcurrency))
	}
	return &TaskQueue[T]{maxConcurrency: maxConcurrency}
}

// Process 按顺序处理 items 中的所有任务，最多 maxConcurrency 个并发。
// handler 不能为 nil；阻塞直到所有任务执行完成。
func (q *TaskQueue[T]) Process(items []T, handler TaskHandler[T]) {
	if handler == nil {
		panic("TaskQueue.Process: handler must not be nil")
	}
	if len(items) == 0 {
		return
	}

	workers := min(q.maxConcurrency, len(items))
	ch := make(chan T, workers)

	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for item := range ch {
				safeHandleTask(handler, item)
			}
		}()
	}

	for _, item := range items {
		ch <- item
	}
	close(ch)
	wg.Wait()
}

func safeHandleTask[T any](handler TaskHandler[T], item T) {
	defer func() {
		if r := recover(); r != nil {
			_ = fmt.Sprintf("TaskQueue handler panic: %v", r)
		}
	}()
	handler(item)
}

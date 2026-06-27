package ai_agent

import (
	"container/list"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/executor"
	"github.com/chuccp/ai-agent/value"
	"github.com/sourcegraph/conc/pool"
)

// AgentManager Agent管理器
type AgentManager struct {
	agentRegistry *sync.Map // map[string]*Agent
	mu            *sync.RWMutex
	runStatus     *RunStatus
}

func NewAgentManager() *AgentManager {
	return &AgentManager{
		agentRegistry: new(sync.Map),
		mu:            new(sync.RWMutex),
		runStatus:     newRunStatus(),
	}
}

// GetRunStatus 获取当前任务运行状态
func (m *AgentManager) GetRunStatus() *RunStatus {
	return m.runStatus
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
func (m *AgentManager) CreateExecutor(agentID string, input *value.ObjectValue, execConfig *executor.Config) (*AgentExecutor, error) {
	agent, ok := m.GetAgent(agentID)
	if !ok {
		return nil, errors.New("Agent not found: " + agentID)
	}
	return m.createExecutorForAgent(agent, input, execConfig)
}

// // CreateExecutorForAgent 为Agent创建执行器
func (m *AgentManager) createExecutorForAgent(agent *Agent, input *value.ObjectValue, execConfig *executor.Config) (*AgentExecutor, error) {
	exec := NewAgentExecutor(agent, execConfig)
	exec.prepareInput(input)
	return exec, nil
}
func (m *AgentManager) ExecSync(agentExecutor *AgentExecutor, input *value.ObjectValue) *AsyncResult {
	asyncResult := agentExecutor.ExecSyncForInput(input)
	return asyncResult
}

// CreateExecutorWithID // CreateExecutorWithID 创建带ID的执行器
func (m *AgentManager) CreateExecutorWithID(executorID string, agent *Agent, input *value.ObjectValue, execConfig *executor.Config) *AgentExecutor {
	exec := NewAgentExecutorWithExecutorId(executorID, agent, execConfig)
	exec.prepareInput(input)
	return exec
}

// RunStatus 任务队列运行状态
type RunStatus struct {
	RunningCountTotal   int       // 任务总数
	runningCount        int64     // 已完成数
	RunTime             time.Time // 开始时间
	LastRunTime         time.Time // 最后完成时间
	mu                  sync.Mutex
	agentExecutorMap    map[string]*AgentExecutor
	allAgentExecutorMap map[string]*AgentExecutor
	agentExecutorList   list.List
}

func newRunStatus() *RunStatus {
	return &RunStatus{
		agentExecutorMap:    make(map[string]*AgentExecutor),
		allAgentExecutorMap: make(map[string]*AgentExecutor),
	}
}

func (s *RunStatus) run(item *AgentExecutor) {
	s.runTemp(item)
	s.LastRunTime = time.Now()
}

const agentExecutorListMaxSize = 1000

func (s *RunStatus) runTemp(item *AgentExecutor) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.agentExecutorMap[item.GetID()] = item
	s.runningCount++

	s.agentExecutorList.PushFront(item)
	s.allAgentExecutorMap[item.GetID()] = item
	for s.agentExecutorList.Len() > agentExecutorListMaxSize {
		back := s.agentExecutorList.Back()
		if back != nil {
			s.agentExecutorList.Remove(back)
			agentExecutor := back.Value.(*AgentExecutor)
			delete(s.allAgentExecutorMap, agentExecutor.GetID())
		}
	}
}
func (s *RunStatus) finish(item *AgentExecutor) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.agentExecutorMap, item.GetID())
	s.runningCount--
}

func (s *RunStatus) GetExecutor(id string) (*AgentExecutor, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.allAgentExecutorMap[id]
	return v, ok
}

func (s *RunStatus) GetLiveAgentExecutor() []*AgentExecutor {
	s.mu.Lock()
	defer s.mu.Unlock()

	executors := make([]*AgentExecutor, 0, len(s.agentExecutorMap))
	for _, v := range s.agentExecutorMap {
		executors = append(executors, v)
	}
	return executors
}

func (s *RunStatus) GetRunningCount() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.runningCount
}

type TaskCall func(item *AsyncResult)

func (m *AgentManager) ProcessTasks(items []*AgentExecutor, maxConcurrency int, taskCall TaskCall) {
	if len(items) == 0 {
		return
	}
	m.runStatus.RunningCountTotal = len(items)
	m.runStatus.RunTime = time.Now()
	p := pool.New().WithMaxGoroutines(maxConcurrency)
	for _, item := range items {
		p.Go(func() {
			m.runStatus.run(item)
			defer func() {
				m.runStatus.finish(item)
			}()
			asyncResult := item.ExecSync()
			if taskCall != nil {
				taskCall(asyncResult)
			}
		})
	}
	p.Wait()
}
func (m *AgentManager) Restart(item *AgentExecutor, taskCall TaskCall) {
	p := pool.New().WithMaxGoroutines(1)
	p.Go(func() {
		m.runStatus.runTemp(item)
		defer func() {
			m.runStatus.finish(item)
		}()
		asyncResult := item.ExecSync()
		if taskCall != nil {
			taskCall(asyncResult)
		}
	})

	p.Wait()
}

func (m *AgentManager) GetExecutor(id string) (*AgentExecutor, bool) {
	return m.runStatus.GetExecutor(id)
}

func (m *AgentManager) GetAllLiveAgentExecutor() []*AgentExecutor {
	return m.runStatus.GetLiveAgentExecutor()
}

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
	RunTime               time.Time // 开始时间
	LastRunTime           time.Time // 最后完成时间
	mu                    sync.Mutex
	runningCountTotalMap  map[string]*AgentExecutor
	agentExecutorMap      map[string]*AgentExecutor
	allInAgentExecutorMap map[string]*AgentExecutor
	agentExecutorList     list.List
}

func newRunStatus() *RunStatus {
	return &RunStatus{
		agentExecutorMap:      make(map[string]*AgentExecutor),
		allInAgentExecutorMap: make(map[string]*AgentExecutor),
		runningCountTotalMap:  make(map[string]*AgentExecutor),
	}
}

func (s *RunStatus) run(item *AgentExecutor) {
	s.runTemp(item)
	s.mu.Lock()
	s.LastRunTime = time.Now()
	s.mu.Unlock()
}

const agentExecutorListMaxSize = 1000

func (s *RunStatus) IsRunning(item *AgentExecutor) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.agentExecutorMap[item.GetID()]
	return ok
}

func (s *RunStatus) tryStart(item *AgentExecutor) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.agentExecutorMap[item.GetID()]; ok {
		return false
	}
	s.add(item)
	return true
}
func (s *RunStatus) runTemp(item *AgentExecutor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.add(item)
}

func (s *RunStatus) add(item *AgentExecutor) {
	s.agentExecutorMap[item.GetID()] = item
	s.agentExecutorList.PushFront(item)
	s.allInAgentExecutorMap[item.GetID()] = item
	s.runningCountTotalMap[item.GetID()] = item
	for s.agentExecutorList.Len() > agentExecutorListMaxSize {
		back := s.agentExecutorList.Back()
		if back != nil {
			s.agentExecutorList.Remove(back)
			agentExecutor := back.Value.(*AgentExecutor)
			delete(s.allInAgentExecutorMap, agentExecutor.GetID())
		}
	}
}
func (s *RunStatus) finish(item *AgentExecutor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastRunTime = time.Now()
	delete(s.agentExecutorMap, item.GetID())
}

func (s *RunStatus) GetExecutor(id string) (*AgentExecutor, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.allInAgentExecutorMap[id]
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

func (s *RunStatus) RunningCount() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return int64(len(s.agentExecutorMap))
}
func (s *RunStatus) RunningCountTotal() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return int64(len(s.runningCountTotalMap))
}

type TaskCall func(item *AsyncResult)

func (m *AgentManager) ProcessTasks(items []*AgentExecutor, maxConcurrency int, taskCall TaskCall) {
	if len(items) == 0 {
		return
	}
	m.runStatus.mu.Lock()
	m.runStatus.runningCountTotalMap = make(map[string]*AgentExecutor)
	for _, agent := range items {
		m.runStatus.runningCountTotalMap[agent.GetID()] = agent
	}
	m.runStatus.RunTime = time.Now()
	m.runStatus.mu.Unlock()
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
func (m *AgentManager) Restart(item *AgentExecutor, taskCall TaskCall) bool {
	if !m.runStatus.tryStart(item) {
		return false
	}
	p := pool.New()
	p.Go(func() {
		defer m.runStatus.finish(item)
		asyncResult := item.ExecSync()
		if taskCall != nil {
			taskCall(asyncResult)
		}
	})
	return true
}

func (m *AgentManager) GetExecutor(id string) (*AgentExecutor, bool) {
	return m.runStatus.GetExecutor(id)
}

func (m *AgentManager) GetAllLiveAgentExecutor() []*AgentExecutor {
	return m.runStatus.GetLiveAgentExecutor()
}

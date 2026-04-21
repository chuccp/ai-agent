package ai_agent

import (
	"sync"

	"emperror.dev/errors"
	"github.com/chuccp/ai-agent/executor"
	"github.com/chuccp/ai-agent/value"
	"github.com/elliotchance/orderedmap/v3"
)

// AgentManager Agent管理器
type AgentManager struct {
	agentRegistry    *sync.Map // map[string]*Agent
	mu               *sync.RWMutex
	executorRegistry *sync.Map // map[string]*AgentExecutor
	tempRegistry     *orderedmap.OrderedMap[string, *AgentExecutor]
}

func NewAgentManager() *AgentManager {
	return &AgentManager{
		agentRegistry:    new(sync.Map),
		mu:               new(sync.RWMutex),
		executorRegistry: new(sync.Map),
		tempRegistry:     orderedmap.NewOrderedMap[string, *AgentExecutor](),
	}
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

func (m *AgentManager) GetAllAgentExecutor() []*AgentExecutor {
	var executors = make([]*AgentExecutor, 0)
	m.executorRegistry.Range(func(key, value interface{}) bool {
		executors = append(executors, value.(*AgentExecutor))
		return true
	})
	return executors
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
	return exec, nil
}
func (m *AgentManager) ExecSync(agentExecutor *AgentExecutor, input *value.ObjectValue) *AsyncResult {
	m.executorRegistry.Store(agentExecutor.GetID(), agentExecutor)
	asyncResult := agentExecutor.ExecSync(input)
	if asyncResult.Error == nil && asyncResult.Response.Success {
		m.executorRegistry.Delete(agentExecutor.GetID())
		for m.tempRegistry.Len() >= 1000 {
			m.tempRegistry.Delete(m.tempRegistry.Front().Key)
		}
		m.tempRegistry.Set(agentExecutor.GetID(), agentExecutor)
	}
	return asyncResult
}

// CreateExecutorWithID // CreateExecutorWithID 创建带ID的执行器
func (m *AgentManager) CreateExecutorWithID(executorID string, agent *Agent, execConfig *executor.Config) *AgentExecutor {
	exec := NewAgentExecutorWithExecutorId(executorID, agent, execConfig)
	return exec
}

func (m *AgentManager) GetExecutor(id string) (*AgentExecutor, bool) {
	if v, ok := m.executorRegistry.Load(id); ok {
		return v.(*AgentExecutor), true
	}
	if v, ok := m.tempRegistry.Get(id); ok {
		return v, true
	}
	return nil, false
}

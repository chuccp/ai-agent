package types

// AgentStatusType 代理状态类型
type AgentStatusType int

const (
	AgentStatusWaiting AgentStatusType = iota
	AgentStatusPending
	AgentStatusStarted
	AgentStatusRunning
	AgentStatusSucceeded
	AgentStatusFailed
)

func (t AgentStatusType) String() string {
	switch t {
	case AgentStatusWaiting:
		return "WAITING"
	case AgentStatusPending:
		return "PENDING"
	case AgentStatusStarted:
		return "STARTED"
	case AgentStatusRunning:
		return "RUNNING"
	case AgentStatusSucceeded:
		return "SUCCEEDED"
	case AgentStatusFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}
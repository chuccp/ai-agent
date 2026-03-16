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

// GetAgentStatusName 获取Agent状态名称（中文）
func GetAgentStatusName(status AgentStatusType) string {
	names := map[AgentStatusType]string{
		AgentStatusWaiting:   "等待中",
		AgentStatusPending:   "排队中",
		AgentStatusStarted:   "已启动",
		AgentStatusRunning:   "执行中",
		AgentStatusSucceeded: "成功",
		AgentStatusFailed:    "失败",
	}
	if name, ok := names[status]; ok {
		return name
	}
	return "未知状态"
}
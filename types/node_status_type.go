package types

// NodeStatusType 节点状态类型
type NodeStatusType int

const (
	NodeStatusWaiting NodeStatusType = iota
	NodeStatusStarted
	NodeStatusQueued
	NodeStatusRunning
	NodeStatusSucceeded
	NodeStatusFailed
	NodeStatusExpired
)

func (t NodeStatusType) String() string {
	switch t {
	case NodeStatusWaiting:
		return "waiting"
	case NodeStatusStarted:
		return "started"
	case NodeStatusQueued:
		return "queued"
	case NodeStatusRunning:
		return "running"
	case NodeStatusSucceeded:
		return "succeeded"
	case NodeStatusFailed:
		return "failed"
	case NodeStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}
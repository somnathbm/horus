package health

type State string

const (
	StatePending   State = "pending"
	StateRunning   State = "running"
	StateSucceeded State = "succeeded"
	StateFailed    State = "failed"
	StateUnknown   State = "unknown"
)

type WorkloadStatus struct {
	TargetID string `json:"targetId"`
	State    State  `json:"state"`
	Reason   string `json:"reason"`
}

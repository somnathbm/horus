package target

type TargetType string

const (
	TargetKubernetesJob TargetType = "kubernetes-job"
	// TargetKueueWorkload TargetType = "kueue-workload"
	// TargetRayJob        TargetType = "ray-job"
)

type TargetSpec interface {
	isTargetSpec()
}

type Target struct {
	ID   string     `json:"id"`
	Name string     `json:"name"`
	Type TargetType `json:"type"`
	Spec TargetSpec `json:"spec"`
}

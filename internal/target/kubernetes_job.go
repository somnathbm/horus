package target

type KubernetesJobSpec struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	JobName   string `json:"jobName"`
}

func (k KubernetesJobSpec) isTargetSpec() {}

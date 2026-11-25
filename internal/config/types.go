package config

type DiscoveryConfig struct {
	Local      []LocalSourceConfig      `yaml:"local"`
	Kubernetes []KubernetesSourceConfig `yaml:"kubernetes"`
	AWS        []AWSSourceConfig        `yaml:"aws"`
}

type DiscoverySourceConfig struct {
	Sources DiscoveryConfig `yaml:"sources" validate:"required"`
}

type ExecutionConfig struct {
	Workers             int `yaml:"workers"`
	TargetQueueCapacity int `yaml:"targetQueueCapacity"`
}

type AppConfig struct {
	Discovery DiscoverySourceConfig `yaml:"discovery" validate:"required"`
	Execution *ExecutionConfig       `yaml:"execution" validate:"omitempty"`
}

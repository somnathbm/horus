package config

type DiscoveryConfig struct {
	Local      []LocalSourceConfig      `yaml:"local"`
	Kubernetes []KubernetesSourceConfig `yaml:"kubernetes"`
	AWS        []AWSSourceConfig        `yaml:"aws"`
}

type DiscoverySourceConfig struct {
	Sources DiscoveryConfig `yaml:"sources"`
}

type ExecutionConfig struct {
	Workers             int `yaml:"workers"`
	TargetQueueCapacity int `yaml:"targetQueueCapacity"`
}

type AppConfig struct {
	Discovery DiscoverySourceConfig `yaml:"discovery"`
	Execution ExecutionConfig       `yaml:"execution"`
}

package config

type DiscoveryConfig struct {
	Local      []LocalSourceConfig      `yaml:"local"`
	Kubernetes []KubernetesSourceConfig `yaml:"kubernetes"`
	AWS        []AWSSourceConfig        `yaml:"aws"`
}

type AppConfig struct {
	Discovery DiscoveryConfig `yaml:"discovery"`
}

package config

type LocalSourceConfig struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type KubernetesSourceConfig struct {
	Name       string `yaml:"name"`
	ClusterURL string `yaml:"clusterURL"`
}

type AWSSourceConfig struct {
	Name      string `yaml:"name"`
	AccountID string `yaml:"accountID"`
	Region    string `yaml:"region"`
}

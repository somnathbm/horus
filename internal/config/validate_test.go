package config

import (
	"errors"
	"testing"
)

// Test local source config
func TestValidateConfigs(t *testing.T) {
	// test cases
	tests := []struct {
		name    string
		cfg     AppConfig
		wantErr error
	}{
		{
			"empty-appconfig",
			AppConfig{},
			ErrNoDiscoverySources,
		},
		{
			"local-config-no-name",
			AppConfig{
				Discovery: DiscoveryConfig{
					Local: []LocalSourceConfig{
						{Name: "", Path: "/data"},
					},
				},
			},
			ErrNoSourceName,
		},
		{
			"local-config-no-abs-path",
			AppConfig{
				Discovery: DiscoveryConfig{
					Local: []LocalSourceConfig{
						{Name: "dev-local", Path: "data/resources/"},
					},
				},
			},
			ErrInvalidFilePath,
		},
		{
			"local-config-empty-path",
			AppConfig{
				Discovery: DiscoveryConfig{
					Local: []LocalSourceConfig{
						{Name: "dev-local", Path: "      \t"},
					},
				},
			},
			ErrInvalidFilePath,
		},
		{
			"local-valid-config",
			AppConfig{
				Discovery: DiscoveryConfig{
					Local: []LocalSourceConfig{
						{Name: "dev-local", Path: "/data"},
					},
				},
			},
			nil,
		},
		{
			"k8s-config-empty-name",
			AppConfig{
				Discovery: DiscoveryConfig{
					Kubernetes: []KubernetesSourceConfig{
						{Name: "    ", ClusterURL: "http://some-cluster.com"},
					},
				},
			},
			ErrNoSourceName,
		},
		{
			"k8s-config-empty-clusterurl",
			AppConfig{
				Discovery: DiscoveryConfig{
					Kubernetes: []KubernetesSourceConfig{
						{Name: "dev-cluster", ClusterURL: "  "},
					},
				},
			},
			ErrInvalidClusterURL,
		},
		{
			"k8s-invalid-cluster",
			AppConfig{
				Discovery: DiscoveryConfig{
					Kubernetes: []KubernetesSourceConfig{
						{Name: "dev-cluster", ClusterURL: "some-cluster.com"},
					},
				},
			},
			ErrInvalidClusterURL,
		},
		{
			"k8s-valid-cluster",
			AppConfig{
				Discovery: DiscoveryConfig{
					Kubernetes: []KubernetesSourceConfig{
						{Name: "dev-cluster", ClusterURL: "https://some-cluster.com"},
					},
				},
			},
			nil,
		},
		{
			"aws-no-name",
			AppConfig{
				Discovery: DiscoveryConfig{
					AWS: []AWSSourceConfig{
						{Name: "", AccountID: "123445558123", Region: "us-east-1"},
					},
				},
			},
			ErrNoSourceName,
		},
		{
			"aws-nondigit-accountid",
			AppConfig{
				Discovery: DiscoveryConfig{
					AWS: []AWSSourceConfig{
						{Name: "dev-account", AccountID: "12fc55bm1234", Region: "us-east-1"},
					},
				},
			},
			ErrInvalidAccountID,
		},
		{
			"aws-invalid-account-no-region",
			AppConfig{
				Discovery: DiscoveryConfig{
					AWS: []AWSSourceConfig{
						{Name: "dev-account", AccountID: "123445558"},
					},
				},
			},
			ErrInvalidAccountID,
		},
		{
			"aws-invalid-region",
			AppConfig{
				Discovery: DiscoveryConfig{
					AWS: []AWSSourceConfig{
						{Name: "dev-account", AccountID: "123445551234", Region: ""},
					},
				},
			},
			ErrInvalidAWSRegion,
		},
		{
			"aws-valid-region",
			AppConfig{
				Discovery: DiscoveryConfig{
					AWS: []AWSSourceConfig{
						{Name: "dev-account", AccountID: "123445551234", Region: "us-east-1"},
					},
				},
			},
			nil,
		},
		{
			"valid-local-config-invalid-aws",
			AppConfig{
				Discovery: DiscoveryConfig{
					Local: []LocalSourceConfig{
						{Name: "dev-local", Path: "/data"},
					},
					AWS: []AWSSourceConfig{
						{Name: "", AccountID: "12ac67893bg4"},
					},
				},
			},
			ErrNoSourceName,
		},
		{
			"valid-invalid-local-config",
			AppConfig{
				Discovery: DiscoveryConfig{
					Local: []LocalSourceConfig{
						{Name: "dev-local", Path: "/data"},
						{Name: "dev-staging", Path: "data/resources/"},
					},
				},
			},
			ErrInvalidFilePath,
		},
		{
			"empty-local-config",
			AppConfig{
				Discovery: DiscoveryConfig{
					Local: []LocalSourceConfig{},
				},
			},
			ErrNoDiscoverySources,
		},
		{
			"empty-config",
			AppConfig{
				Discovery: DiscoveryConfig{
					Local:      []LocalSourceConfig{},
					Kubernetes: []KubernetesSourceConfig{},
					AWS:        []AWSSourceConfig{},
				},
			},
			ErrNoDiscoverySources,
		},
	}

	// run test
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validateConfig(tc.cfg)

			// assertion on error type
			// if (err != nil) != tc.wantErr {
			// 	t.Fatalf("Unexpected error: %v", err)
			// }
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Expecting an %v type error. Got: %v", tc.wantErr, err)
			}
		})
	}

}

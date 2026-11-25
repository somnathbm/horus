package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestConfigLoader(t *testing.T) {
	// test cases
	tests := []struct {
		name        string
		path        string
		content     string
		createFile  bool
		wantErr     bool
		wantErrType error
		wantConfig  AppConfig
	}{
		{
			name:        "404-file",
			path:        "404-file.yaml",
			content:     "",
			createFile:  false,
			wantErr:     true,
			wantErrType: os.ErrNotExist,
			wantConfig:  AppConfig{},
		},
		{
			name:        "valid-config-file",
			path:        "application.yaml",
			content:     "discovery:\n  sources:\n    local:\n      - name: \"      local-dev \"\n        path: examples/targets\nexecution:\n  workers: 3\n  targetQueueCapacity: 5",
			createFile:  true,
			wantErr:     false,
			wantErrType: nil,
			wantConfig: AppConfig{
				Discovery: DiscoverySourceConfig{
					Sources: DiscoveryConfig{
						Local: []LocalSourceConfig{
							{Name: "local-dev", Path: "examples/targets"},
						},
					},
				},
				Execution: &ExecutionConfig{
					Workers:             3,
					TargetQueueCapacity: 5,
				},
			},
		},
		{
			name:        "typo-config-file",
			path:        "malformed-application.yaml",
			content:     "discoivery:\n  sources:\n    local:\n      - name: \"      local-dev \"\n        path: examples/targets",
			createFile:  true,
			wantErr:     true,
			wantErrType: nil,
			wantConfig:  AppConfig{},
		},
		{
			name:        "malformed-yaml",
			path:        "malformed-invalid-config.yaml",
			content:     "discoivery:\n  sources:\n    locals:\n      - name: \"      local-dev \"\n        path: examples/targets",
			createFile:  true,
			wantErr:     true,
			wantErrType: nil,
			wantConfig:  AppConfig{},
		},
		{
			name:        "whitespaced-config",
			path:        "whitespaced-config.yaml",
			content:     "discovery:\n  sources:\n    local:\n      - name: \"      local-dev \"\n        path: examples/targets\nexecution:\n  workers: 3\n  targetQueueCapacity: 5",
			createFile:  true,
			wantErr:     false,
			wantErrType: nil,
			wantConfig: AppConfig{
				Discovery: DiscoverySourceConfig{
					Sources: DiscoveryConfig{
						Local: []LocalSourceConfig{
							{Name: "local-dev", Path: "examples/targets"},
						},
					},
				},
				Execution: &ExecutionConfig{
					Workers:             3,
					TargetQueueCapacity: 5,
				},
			},
		},
		{
			name:        "execution-config-omitted",
			path:        "application.yaml",
			content:     "discovery:\n  sources:\n    local:\n      - name: \"      local-dev \"\n        path: examples/targets",
			createFile:  true,
			wantErr:     false,
			wantErrType: nil,
			wantConfig: AppConfig{
				Discovery: DiscoverySourceConfig{
					Sources: DiscoveryConfig{
						Local: []LocalSourceConfig{
							{Name: "local-dev", Path: "examples/targets"},
						},
					},
				},
			},
		},
	}

	// test suite
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// set up
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, tc.path)

			if tc.createFile {
				// create file
				err := os.WriteFile(filePath, []byte(tc.content), 0644)
				if err != nil {
					t.Fatalf("failed to create config file: %v", err)
				}
			}

			// load config
			gotConfig, err := Load(filePath)

			// assert on error status
			if (err != nil) != tc.wantErr {
				t.Fatalf("Load() error: %v, wantErr: %v", err, tc.wantErr)
			}

			if tc.wantErr && tc.wantErrType != nil && !errors.Is(err, tc.wantErrType) {
				t.Errorf("Expecting error type: %v, got: %v", tc.wantErrType, err)
			}

			// assert on config contents for successful cases
			if !tc.wantErr {
				if !reflect.DeepEqual(gotConfig, tc.wantConfig) {
					t.Errorf("Load() config mismatch:\n  got:  %+v\n  want: %+v", gotConfig, tc.wantConfig)
				}
			}
		})
	}
}

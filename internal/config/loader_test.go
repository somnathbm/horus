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
			"404-file",
			"404-file.yaml",
			"",
			false,
			true,
			os.ErrNotExist,
			AppConfig{},
		},
		{
			"valid-config-file",
			"application.yaml",
			"discovery:\n  local:\n    - name: local-dev\n      path: /data\n\n",
			true,
			false,
			nil,
			AppConfig{
				Discovery: DiscoveryConfig{
					Local: []LocalSourceConfig{
						{Name: "local-dev", Path: "/data"},
					},
				},
			},
		},
		{
			"typo-config-file",
			"malformed-application.yaml",
			"discoivery:\n  local:\n    - name: local-dev\n      path: /data\n\n",
			true,
			true,
			nil,
			AppConfig{},
		},
		{
			"malformed-yaml",
			"malformed-invalid-config.yaml",
			"discoivery:\n  local:\n    - name:local-dev\n	      path:/data\n\n",
			true,
			true,
			nil,
			AppConfig{},
		},
		{
			"malformed-yaml-2",
			"malformed-invalid-config2.yaml",
			"discoivery:\n  local:\n    - name:1234\n	      path:/data\n\n",
			true,
			true,
			nil,
			AppConfig{},
		},
		{
			"valid-invalid-semantic-file",
			"invalid-semantic.yaml",
			"discovery:\n  aws:\n    - name: dev-acount\n      accountID: 123\n\n",
			true,
			true,
			ErrInvalidAccountID,
			AppConfig{},
		},
		{
			"whitespaced-config",
			"whitespaced-config.yaml",
			"discovery:\n  local:\n    - name:   dev-local   \n	      path:     /data\n\n",
			true,
			true,
			nil,
			AppConfig{
				Discovery: DiscoveryConfig{
					Local: []LocalSourceConfig{
						{Name: "dev-local", Path: "/data"},
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

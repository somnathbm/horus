package local

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/somnathbm/horus/internal/config"
	"github.com/somnathbm/horus/internal/target"
)

func TestLoader(t *testing.T) {
	// standalone context cancellation test
	t.Run("cancellation-context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := Load(ctx, t.TempDir())
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Cancelled, got: %v", err)
		}
	})

	// test cases
	tests := []struct {
		name         string
		isEmptyPath  bool
		createSubDir bool
		createFile   bool
		fileMap      map[string]string
		wantTarget   []target.Target
		wantErr      bool
		wantErrType  error
	}{
		// load
		{
			name:         "empty-path",
			isEmptyPath:  true,
			createSubDir: false,
			createFile:   false,
			fileMap:      nil,
			wantTarget:   nil,
			wantErr:      true,
			wantErrType:  ErrInvalidPath,
		},
		{
			name:         "entry-is-dir",
			isEmptyPath:  false,
			createSubDir: true,
			createFile:   false,
			fileMap:      nil,
			wantTarget:   nil,
			wantErr:      true,
			wantErrType:  ErrNoDirAllowed,
		},
		{
			name:         "empty-dir",
			isEmptyPath:  false,
			createSubDir: false,
			createFile:   false,
			fileMap:      nil,
			wantTarget:   []target.Target{},
			wantErr:      false,
			wantErrType:  nil,
		},
		{
			name:         "not-json",
			isEmptyPath:  false,
			createSubDir: false,
			createFile:   true,
			fileMap:      map[string]string{"file.dat": "some-content"},
			wantTarget:   nil,
			wantErr:      true,
			wantErrType:  ErrInvalidFileType,
		},
		{
			name:         "valid-json",
			isEmptyPath:  false,
			createSubDir: false,
			createFile:   true,
			fileMap:      map[string]string{"file.json": `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":123}}`},
			wantTarget: []target.Target{
				{
					ID:   "svc123",
					Type: "http",
					Name: "auth-service",
					Spec: target.HTTPTargetSpec{
						URL:  "http://my-svc-url.com/svc123",
						Port: 123,
					},
				},
			},
			wantErr:     false,
			wantErrType: nil,
		},
		{
			name:         "malformed-json",
			isEmptyPath:  false,
			createSubDir: false,
			createFile:   true,
			fileMap:      map[string]string{"file.json": `{"id": "db123" "type": "postgres", "name": "auth-db", "spec": {"host": "http://some-url.com", "username": "dbadmin", "port": 123}}`},
			wantTarget:   nil,
			wantErr:      true,
			wantErrType:  nil,
		},

		// parse, decode, validate
		{
			name:         "invalid-target",
			isEmptyPath:  false,
			createSubDir: false,
			createFile:   true,
			fileMap:      map[string]string{"file.json": `{"id": "db123", "type": "postgresz", "name": "auth-db", "spec": {"host": "http://some-url.com", "username": "dbadmin", "port": 123}}`},
			wantTarget:   nil,
			wantErr:      true,
			wantErrType:  ErrUnknownTarget,
		},
		{
			name:         "invalid-target-spec",
			isEmptyPath:  false,
			createSubDir: false,
			createFile:   true,
			fileMap:      map[string]string{"file.json": `{"id": "db123", "type": "postgres", "name": "auth-db", "spec": {"host": "", "username": "dbadmin", "port": 123}}`},
			wantTarget:   []target.Target{},
			wantErr:      true,
			wantErrType:  ErrInvalidTargetSpec,
		},
		{
			name:         "valid-target-spec",
			isEmptyPath:  false,
			createSubDir: false,
			createFile:   true,
			fileMap:      map[string]string{"file.json": `{"id": "db123", "type": "postgres", "name": "auth-db", "spec": {"host": "http://some-url.com", "username": "dbadmin", "port": 123}}`},
			wantTarget: []target.Target{
				{
					ID:   "db123",
					Type: "postgres",
					Name: "auth-db",
					Spec: target.PostgresTargetSpec{
						Host:     "http://some-url.com",
						Username: "dbadmin",
						Port:     123,
					},
				},
			},
			wantErr:     false,
			wantErrType: nil,
		},
		{
			name:         "two-or-more-valid-targets",
			isEmptyPath:  false,
			createSubDir: false,
			createFile:   true,
			fileMap: map[string]string{
				"http.json":     `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":1234}}`,
				"postgres.json": `{"id": "db123", "type": "postgres", "name": "auth-db", "spec": {"host": "http://some-url.com", "username": "dbadmin", "port": 1234}}`,
				"redis.json":    `{"id": "cache123", "type": "redis", "name": "cache123", "spec": {"url": "http://some-url.com/cache123", "port": 1234}}`,
			},
			wantTarget: []target.Target{
				{
					ID:   "svc123",
					Type: target.HTTPTargetType,
					Name: "auth-service",
					Spec: target.HTTPTargetSpec{
						URL:  "http://my-svc-url.com/svc123",
						Port: 1234,
					},
				},
				{
					ID:   "db123",
					Type: target.PostgresSQLTargetType,
					Name: "auth-db",
					Spec: target.PostgresTargetSpec{
						Host:     "http://some-url.com",
						Username: "dbadmin",
						Port:     1234,
					},
				},
				{
					ID:   "cache123",
					Type: target.RedisTargetType,
					Name: "cache123",
					Spec: target.RedisTargetSpec{
						URL:  "http://some-url.com/cache123",
						Port: 1234,
					},
				},
			},
			wantErr:     false,
			wantErrType: nil,
		},
		{
			name:         "valid-invalid-targets",
			isEmptyPath:  false,
			createSubDir: false,
			createFile:   true,
			fileMap: map[string]string{
				"http.json":  `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":1234}}`,
				"redis.json": `{"id": "cache123", "type": "redis", "name": "cache123", "spec": {"url": "", "port": 1234}}`,
			},
			wantTarget:  nil,
			wantErr:     true,
			wantErrType: ErrInvalidTargetSpec,
		},
		{
			name:         "unknown-target-field",
			isEmptyPath:  false,
			createSubDir: false,
			createFile:   true,
			fileMap:      map[string]string{"file.json": `{"id": "db123", "type": "postgres", "name": "auth-db", "spec": {"host": "http://some-url.com", "username": "dbadmin", "ports": 123}}`},
			wantTarget:   nil,
			wantErr:      true,
			wantErrType:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			discoveryDir := t.TempDir()

			if tc.isEmptyPath {
				discoveryDir = ""
			}

			if tc.createSubDir {
				subDirName := "inner-dir"
				subDirPath := filepath.Join(discoveryDir, subDirName)
				if err := os.Mkdir(subDirPath, 0755); err != nil {
					t.Fatalf("Load() mkdir error: %v", err)
				}
			}

			if tc.createFile {
				for k, v := range tc.fileMap {
					if err := os.WriteFile(filepath.Join(discoveryDir, k), []byte(v), 0644); err != nil {
						t.Fatalf("Load() os open error: %v", err)
					}
				}
			}

			gotTarget, err := Load(t.Context(), discoveryDir)

			// assertions
			var syntaxErr *json.SyntaxError

			// unexpected error
			if (err != nil) != tc.wantErr {
				t.Fatalf("Load() error: %v, wantErr: %v", err, tc.wantErr)
			}

			// error type assertions
			if tc.wantErr && tc.wantErrType != nil && !errors.Is(err, tc.wantErrType) {
				t.Errorf("Expecting error type: %v, got: %v", tc.wantErrType, err)
			}

			// JSON syntax error assertions
			if tc.wantErr && tc.wantErrType == nil && !errors.As(err, &syntaxErr) && !strings.Contains(err.Error(), "json: unknown field") {
				t.Errorf("Expecting json unknown field exception. Got: %v", err)
			}

			// valid case. equality check
			if !tc.wantErr {
				if !reflect.DeepEqual(gotTarget, tc.wantTarget) {
					t.Errorf("Load() config mismatch:\n  got:  %+v\n  want: %+v", gotTarget, tc.wantTarget)
				}
			}
		})
	}
}

func TestDiscoverer(t *testing.T) {
	// context cancellation
	t.Run("cancelled-context", func(t *testing.T) {
		// create a context first
		ctx, cancel := context.WithCancel(context.Background())
		// call cancel abruptly
		cancel()

		// call discover
		localDiscoverer := New(config.LocalSourceConfig{
			Name: "local-dev",
			Path: "data",
		})
		_, err := localDiscoverer.Discover(ctx)

		// assertions
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expecting a cancelled context. Got: %v", err)
		}
	})

	// test cases
	tests := []struct {
		name        string
		config      config.LocalSourceConfig
		fileMap     map[string]string
		wantTargets []target.Target
		wantErr     bool
		wantErrType error
	}{
		{
			name: "valid-config",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			fileMap: map[string]string{
				"http.json":     `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":1234}}`,
				"postgres.json": `{"id": "db123", "type": "postgres", "name": "auth-db", "spec": {"host": "http://some-url.com", "username": "dbadmin", "port": 1234}}`,
				"redis.json":    `{"id": "cache123", "type": "redis", "name": "cache123", "spec": {"url": "http://some-url.com/cache123", "port": 1234}}`,
			},
			wantTargets: []target.Target{
				{
					ID:   "svc123",
					Type: target.HTTPTargetType,
					Name: "auth-service",
					Spec: target.HTTPTargetSpec{
						URL:  "http://my-svc-url.com/svc123",
						Port: 1234,
					},
				},
				{
					ID:   "db123",
					Type: target.PostgresSQLTargetType,
					Name: "auth-db",
					Spec: target.PostgresTargetSpec{
						Host:     "http://some-url.com",
						Username: "dbadmin",
						Port:     1234,
					},
				},
				{
					ID:   "cache123",
					Type: target.RedisTargetType,
					Name: "cache123",
					Spec: target.RedisTargetSpec{
						URL:  "http://some-url.com/cache123",
						Port: 1234,
					},
				},
			},
			wantErr:     false,
			wantErrType: nil,
		},
		{
			name: "no-files-content",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			fileMap:     map[string]string{},
			wantTargets: []target.Target{},
			wantErr:     false,
			wantErrType: nil,
		},
		{
			name: "valid-invalid-content",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			fileMap: map[string]string{
				"http.json":     `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":1234}}`,
				"postgres.json": `{"id": "db123", "type": "postgres", "name": "auth-db", "spec": {"host": "http://some-url.com", "username": "dbadmin", "port": 1234}}`,
				"redis.json":    `{"id": "cache123", "type": "redis", "name": "cache123", "spec": {"url": "http://some-url.com/cache123", "ports": 1234}}`,
			},
			wantTargets: []target.Target{
				{
					ID:   "svc123",
					Type: target.HTTPTargetType,
					Name: "auth-service",
					Spec: target.HTTPTargetSpec{
						URL:  "http://my-svc-url.com/svc123",
						Port: 1234,
					},
				},
				{
					ID:   "db123",
					Type: target.PostgresSQLTargetType,
					Name: "auth-db",
					Spec: target.PostgresTargetSpec{
						Host:     "http://some-url.com",
						Username: "dbadmin",
						Port:     1234,
					},
				},
				{
					ID:   "cache123",
					Type: target.RedisTargetType,
					Name: "cache123",
					Spec: target.RedisTargetSpec{
						URL:  "http://some-url.com/cache123",
						Port: 1234,
					},
				},
			},
			wantErr:     true,
			wantErrType: nil,
		},
	}

	// runs
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setups
			tempdir := t.TempDir()
			filePath := filepath.Join(tempdir, tc.config.Path)
			tc.config.Path = filePath

			// create directory
			if err := os.Mkdir(filePath, 0744); err != nil {
				t.Fatalf("Error creating temp dir: %v", err)
			}
			// create files
			for k, v := range tc.fileMap {
				if err := os.WriteFile(filepath.Join(filePath, k), []byte(v), 0744); err != nil {
					t.Fatalf("Error creating the file: %v", err)
				}
			}

			// prepare config instance
			localDiscoverer := New(tc.config)
			gotTargets, err := localDiscoverer.Discover(t.Context())

			// assertions
			var syntaxErr *json.SyntaxError

			// unexpected error
			if (err != nil) != tc.wantErr {
				t.Fatalf("Discover() error: %v, wantErr: %v", err, tc.wantErr)
			}

			// error type assertions
			if tc.wantErr && tc.wantErrType != nil && !errors.Is(err, tc.wantErrType) {
				t.Errorf("Expecting error type: %v, got: %v", tc.wantErrType, err)
			}

			// JSON: unknown field expection
			if tc.wantErr && tc.wantErrType == nil && !errors.As(err, &syntaxErr) && !strings.Contains(err.Error(), "json: unknown field") {
				t.Errorf("Expecting json unknown field exception. Got: %v", err)
			}

			// successful case
			if !tc.wantErr {
				if !reflect.DeepEqual(gotTargets, tc.wantTargets) {
					t.Errorf("Load() config mismatch:\n  got:  %+v\n  want: %+v", gotTargets, tc.wantTargets)
				}
			}
		})
	}
}

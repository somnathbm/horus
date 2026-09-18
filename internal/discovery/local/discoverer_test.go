package local

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/somnathbm/horus/internal/config"
	"github.com/somnathbm/horus/internal/discovery"
	"github.com/somnathbm/horus/internal/target"
)

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
		name           string
		config         config.LocalSourceConfig
		fileMap        map[string]string
		isNonReadable  bool
		wantResult     discovery.DiscoveryResult
		wantProcessErr bool
		wantErrType    error
	}{
		{
			name: "empty-path",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "",
			},
			fileMap:        nil,
			wantResult:     discovery.DiscoveryResult{},
			wantProcessErr: true,
			wantErrType:    ErrEmptyPath,
		},
		{
			name: "dir-read-permission-error",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			isNonReadable:  true,
			fileMap:        nil,
			wantResult:     discovery.DiscoveryResult{},
			wantProcessErr: true,
			wantErrType:    os.ErrPermission,
		},
		{
			name: "empty-dir",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			fileMap:        nil,
			wantResult:     discovery.DiscoveryResult{},
			wantProcessErr: false,
			wantErrType:    nil,
		},
		{
			name: "only-unknown-file-type",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			fileMap: map[string]string{"file.dat": "some-content"},
			wantResult: discovery.DiscoveryResult{
				Targets: nil,
				Failures: []discovery.DiscoveryFailure{
					{
						Item: "file.dat",
						Err:  errors.New("\"file.dat\": only json type is allowed"),
					},
				},
			},
			wantProcessErr: false,
			wantErrType:    ErrInvalidFileType,
		},
		{
			name: "json-file-unknown-file",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			fileMap: map[string]string{
				"file.dat":  "some-content",
				"http.json": `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":1234}}`,
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
					{
						ID:   "svc123",
						Type: target.HTTPTargetType,
						Name: "auth-service",
						Spec: target.HTTPTargetSpec{
							URL:  "http://my-svc-url.com/svc123",
							Port: 1234,
						},
					},
				},
				Failures: []discovery.DiscoveryFailure{
					{
						Item: "file.dat",
						Err:  errors.New("\"file.dat\": only json type is allowed"),
					},
				},
			},
			wantProcessErr: false,
			wantErrType:    ErrInvalidFileType,
		},
		{
			name: "json-file-multi-unknown-file",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			fileMap: map[string]string{
				"file.dat":  "some-content",
				"http.json": `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":1234}}`,
				"file2.dat": "some-content2",
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
					{
						ID:   "svc123",
						Type: target.HTTPTargetType,
						Name: "auth-service",
						Spec: target.HTTPTargetSpec{
							URL:  "http://my-svc-url.com/svc123",
							Port: 1234,
						},
					},
				},
				Failures: []discovery.DiscoveryFailure{
					{
						Item: "file.dat",
						Err:  errors.New("\"file.dat\": only json type is allowed"),
					},
					{
						Item: "file2.dat",
						Err:  errors.New("\"file2.dat\": only json type is allowed"),
					},
				},
			},
			wantProcessErr: false,
			wantErrType:    ErrInvalidFileType,
		},
		{
			name: "entry-is-dir",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			fileMap: map[string]string{
				"http.json":     `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":1234}}`,
				"dirA":          `"should be skipped"`,
				"postgres.json": `{"id": "db123", "type": "postgres", "name": "auth-db", "spec": {"host": "some-url.com", "username": "dbadmin", "port": 1234}}`,
				"redis.json":    `{"id": "cache123", "type": "redis", "name": "cache123", "spec": {"url": "http://some-url.com/cache123", "port": 1234}}`,
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
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
							Host:     "some-url.com",
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
				Failures: []discovery.DiscoveryFailure{
					{
						Item: "dirA",
						Err:  errors.New("\"dirA\": directory is not allowed"),
					},
				},
			},
			wantProcessErr: false,
			wantErrType:    ErrNoDirAllowed,
		},
		{
			name: "valid-json-target",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			fileMap: map[string]string{
				"http.json":     `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":1234}}`,
				"postgres.json": `{"id": "db123", "type": "postgres", "name": "auth-db", "spec": {"host": "some-url.com", "username": "dbadmin", "port": 1234}}`,
				"redis.json":    `{"id": "cache123", "type": "redis", "name": "cache123", "spec": {"url": "http://some-url.com/cache123", "port": 1234}}`,
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
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
							Host:     "some-url.com",
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
			},
			wantProcessErr: false,
			wantErrType:    nil,
		},

		{
			name: "malformed-json-target",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			fileMap: map[string]string{
				"http.json":     `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":1234}}`,
				"postgres.json": `{"id": "db123"type": "postgres", "name": "auth-db", "spec": {"host": "some-url.com", "username": "dbadmin", "port": 1234}}`,
				"redis.json":    `{"id": "cache123", "type": "redis", "name": "cache123", "spec": {"url": "http://some-url.com/cache123", "port": 1234}}`,
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
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
						ID:   "cache123",
						Type: target.RedisTargetType,
						Name: "cache123",
						Spec: target.RedisTargetSpec{
							URL:  "http://some-url.com/cache123",
							Port: 1234,
						},
					},
				},
				Failures: []discovery.DiscoveryFailure{
					{
						Item: "postgres.json",
						Err:  errors.New("\"postgres.json\": invalid character "),
					},
				},
			},
			wantProcessErr: false,
			wantErrType:    nil,
		},
		{
			name: "unknown-target-type",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			fileMap: map[string]string{
				"http.json":     `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":1234}}`,
				"postgres.json": `{"id": "db123", "type": "postgresz", "name": "auth-db", "spec": {"host": "some-url.com", "username": "dbadmin", "port": 1234}}`,
				"redis.json":    `{"id": "cache123", "type": "redis", "name": "cache123", "spec": {"url": "http://some-url.com/cache123", "port": 1234}}`,
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
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
						ID:   "cache123",
						Type: target.RedisTargetType,
						Name: "cache123",
						Spec: target.RedisTargetSpec{
							URL:  "http://some-url.com/cache123",
							Port: 1234,
						},
					},
				},
				Failures: []discovery.DiscoveryFailure{
					{
						Item: "postgres.json",
						Err:  errors.New("\"postgres.json\": target \"auth-db\": unknown target"),
					},
				},
			},
			wantProcessErr: false,
			wantErrType:    ErrUnknownTarget,
		},
		{
			name: "invalid-target-spec-empty-string",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			fileMap: map[string]string{
				"http.json":     `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":1234}}`,
				"postgres.json": `{"id": "db123", "type": "postgres", "name": "auth-db", "spec": {"host": "some-url.com", "username": "", "port": 1234}}`,
				"redis.json":    `{"id": "cache123", "type": "redis", "name": "cache123", "spec": {"url": "http://some-url.com/cache123", "port": 1234}}`,
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
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
						ID:   "cache123",
						Type: target.RedisTargetType,
						Name: "cache123",
						Spec: target.RedisTargetSpec{
							URL:  "http://some-url.com/cache123",
							Port: 1234,
						},
					},
				},
				Failures: []discovery.DiscoveryFailure{
					{
						Item: "postgres.json",
						Err:  errors.New("\"postgres.json\": target \"auth-db\": invalid target spec"),
					},
				},
			},
			wantProcessErr: false,
			wantErrType:    ErrInvalidTargetSpec,
		},
		{
			name: "invalid-target-spec-port",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			fileMap: map[string]string{
				"http.json":     `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":1234}}`,
				"postgres.json": `{"id": "db123", "type": "postgres", "name": "auth-db", "spec": {"host": "some-url.com", "username": "dbadmin", "port": 0}}`,
				"redis.json":    `{"id": "cache123", "type": "redis", "name": "cache123", "spec": {"url": "http://some-url.com/cache123", "port": 1234}}`,
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
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
						ID:   "cache123",
						Type: target.RedisTargetType,
						Name: "cache123",
						Spec: target.RedisTargetSpec{
							URL:  "http://some-url.com/cache123",
							Port: 1234,
						},
					},
				},
				Failures: []discovery.DiscoveryFailure{
					{
						Item: "postgres.json",
						Err:  errors.New("\"postgres.json\": target \"auth-db\": invalid target spec"),
					},
				},
			},
			wantProcessErr: false,
			wantErrType:    ErrInvalidTargetSpec,
		},
		{
			name: "incompatible-target-spec",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			fileMap: map[string]string{
				"http.json":     `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":1234}}`,
				"postgres.json": `{"id": "db123", "type": "postgres", "name": "auth-db", "spec": {"host": "some-url.com", "username": "dbadmin", "port": 1234}}`,
				"redis.json":    `{"id": "cache123", "type": "redis", "name": "cache123", "spec": {"host": "http://some-url.com", "username": "dbadmin", "port": 1234}}`,
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
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
							Host:     "some-url.com",
							Username: "dbadmin",
							Port:     1234,
						},
					},
				},
				Failures: []discovery.DiscoveryFailure{
					{
						Item: "redis.json",
						Err:  errors.New("\"redis.json\": target \"cache123\": json: unknown field \"host\""),
					},
				},
			},
			wantProcessErr: false,
			wantErrType:    nil,
		},
		// {
		// 	name: "target-count-wrong",
		// 	config: config.LocalSourceConfig{
		// 		Name: "local-dev",
		// 		Path: "data",
		// 	},
		// 	fileMap: map[string]string{
		// 		"http.json":     `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":1234}}`,
		// 		"postgres.json": `{"id": "db123", "type": "postgres", "name": "auth-db", "spec": {"host": "some-url.com", "username": "dbadmin", "port": 1234}}`,
		// 		"redis.json":    `{"id": "cache123", "type": "redis", "name": "cache123", "spec": {"host": "http://some-url.com", "username": "dbadmin", "port": 1234}}`,
		// 	},
		// 	wantResult: discovery.DiscoveryResult{
		// 		Targets: []target.Target{
		// 			{
		// 				ID:   "svc123",
		// 				Type: target.HTTPTargetType,
		// 				Name: "auth-service",
		// 				Spec: target.HTTPTargetSpec{
		// 					URL:  "http://my-svc-url.com/svc123",
		// 					Port: 1234,
		// 				},
		// 			},
		// 			{
		// 				ID:   "db123",
		// 				Type: target.PostgresSQLTargetType,
		// 				Name: "auth-db",
		// 				Spec: target.PostgresTargetSpec{
		// 					Host:     "http://some-url.com",
		// 					Username: "dbadmin",
		// 					Port:     1234,
		// 				},
		// 			},
		// 			{
		// 				ID:   "cache123",
		// 				Type: target.RedisTargetType,
		// 				Name: "cache123",
		// 				Spec: target.RedisTargetSpec{
		// 					URL:  "http://some-url.com/cache123",
		// 					Port: 1234,
		// 				},
		// 			},
		// 		},
		// 		Failures: []discovery.DiscoveryFailure{
		// 			{
		// 				Item: "redis.json",
		// 				Err:  errors.New("\"redis.json\": target \"cache123\": json: unknown field \"host\""),
		// 			},
		// 		},
		// 	},
		// 	wantProcessErr: false,
		// 	wantErrType:    nil,
		// },
		// {
		// 	name: "failure-count-wrong",
		// 	config: config.LocalSourceConfig{
		// 		Name: "local-dev",
		// 		Path: "data",
		// 	},
		// 	fileMap: map[string]string{
		// 		"http.json":     `{"id":"svc123","type":"http","name":"auth-service","spec":{"url":"http://my-svc-url.com/svc123","port":1234}}`,
		// 		"postgres.json": `{"id": "db123", "type": "postgres", "name": "auth-db", "spec": {"host": "some-url.com", "username": "dbadmin", "port": 1234}}`,
		// 		"redis.json":    `{"id": "cache123", "type": "redis", "name": "cache123", "spec": {"host": "http://some-url.com", "username": "dbadmin", "port": 1234}}`,
		// 	},
		// 	wantResult: discovery.DiscoveryResult{
		// 		Targets: []target.Target{
		// 			{
		// 				ID:   "svc123",
		// 				Type: target.HTTPTargetType,
		// 				Name: "auth-service",
		// 				Spec: target.HTTPTargetSpec{
		// 					URL:  "http://my-svc-url.com/svc123",
		// 					Port: 1234,
		// 				},
		// 			},
		// 			{
		// 				ID:   "db123",
		// 				Type: target.PostgresSQLTargetType,
		// 				Name: "auth-db",
		// 				Spec: target.PostgresTargetSpec{
		// 					Host:     "http://some-url.com",
		// 					Username: "dbadmin",
		// 					Port:     1234,
		// 				},
		// 			},
		// 		},
		// 		Failures: []discovery.DiscoveryFailure{
		// 			{
		// 				Item: "redis.json",
		// 				Err:  errors.New("\"redis.json\": target \"cache123\": json: unknown field \"host\""),
		// 			},
		// 			{
		// 				Item: "postgres.json",
		// 				Err:  errors.New("\"postgres.json\": target \"auth-db\": json: unknown field \"host\""),
		// 			},
		// 		},
		// 	},
		// 	wantProcessErr: false,
		// 	wantErrType:    nil,
		// },
	}

	// runs
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setups
			tempdir := t.TempDir()
			var filePermMode os.FileMode

			// if path is non-empty
			if tc.config.Path != "" {
				filePath := filepath.Join(tempdir, tc.config.Path)
				tc.config.Path = filePath
				if tc.isNonReadable {
					filePermMode = 0000
				} else {
					filePermMode = 0744
				}

				// create directory
				if err := os.Mkdir(filePath, filePermMode); err != nil {
					t.Fatalf("Error creating temp dir: %v", err)
				}
				// create files
				for k, v := range tc.fileMap {
					if filepath.Ext(k) != "" {
						if err := os.WriteFile(filepath.Join(filePath, k), []byte(v), 0744); err != nil {
							t.Fatalf("Error creating the file: %v", err)
						}
					} else {
						// the entry is a possible directory. so create a directory instead
						if err := os.Mkdir(filepath.Join(filePath, k), filePermMode); err != nil {
							t.Fatalf("Error creating %q as directory: %v", k, err)
						}
					}
				}
			}

			// prepare config instance
			localDiscoverer := New(tc.config)
			gotDscvrResult, err := localDiscoverer.Discover(t.Context())

			// assertions

			// process fail is expected.
			if tc.wantProcessErr {
				// unexpected error type
				if (err != nil) != tc.wantProcessErr {
					t.Fatalf("Process: unexpected error: %v, wantErr: %v", err, tc.wantProcessErr)
				}
				// error type assertions
				if err != nil && tc.wantErrType != nil && !errors.Is(err, tc.wantErrType) {
					t.Errorf("Process: expecting error type: %v, got: %v", tc.wantErrType, err)
				}
			} else {
				// process success is expected

				// non-nil error assertion
				if err != nil {
					t.Fatalf("Expecting no error. but got err: %v", err)
				}

				// target/failure count assertions — always run outside the loop
				if len(tc.wantResult.Targets) != len(gotDscvrResult.Targets) {
					t.Errorf("Discovery: expected target count to be: %v, got: %v", len(tc.wantResult.Targets), len(gotDscvrResult.Targets))
				}
				if len(tc.wantResult.Failures) != len(gotDscvrResult.Failures) {
					t.Errorf("Discovery: expected failure count to be: %v, got: %v", len(tc.wantResult.Failures), len(gotDscvrResult.Failures))
				}

				// targets assertions
				if !reflect.DeepEqual(tc.wantResult.Targets, gotDscvrResult.Targets) {
					t.Errorf("Discovery: expected targets: %v, got: %v", tc.wantResult.Targets, gotDscvrResult.Targets)
				}

				// failure assertions
				for _, failure := range gotDscvrResult.Failures {
					// specific error type available
					if tc.wantErrType != nil && !errors.Is(failure.Err, tc.wantErrType) {
						t.Errorf("Discovery: expecting error type: %v, got: %v", tc.wantErrType, failure.Err)
					}
					// specific error type not available
					if tc.wantErrType == nil && len(tc.wantResult.Failures) > 0 && !strings.Contains(failure.Err.Error(), tc.wantResult.Failures[0].Err.Error()) {
						t.Errorf("Discovery: expecting error msg: %v, got: %v", tc.wantResult.Failures[0].Err.Error(), failure.Err.Error())
					}
				}

			}

		})
	}
}

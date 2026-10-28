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
				"file.dat":      "some-content",
				"kube-job.json": `{"id":"job-001","name":"train-job-001","type":"kubernetes-job","spec":{"cluster":"local-dev","namespace":"ml-workloads","jobName":"training-demo"}}`,
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
					{
						ID:   "job-001",
						Type: target.TargetKubernetesJob,
						Name: "train-job-001",
						Spec: target.KubernetesJobSpec{
							Cluster:   "local-dev",
							Namespace: "ml-workloads",
							JobName:   "training-demo",
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
				"file.dat":      "some-content",
				"kube-job.json": `{"id":"job-001","name":"train-job-001","type":"kubernetes-job","spec":{"cluster":"local-dev","namespace":"ml-workloads","jobName":"training-demo"}}`,
				"file2.dat":     "some-content2",
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
					{
						ID:   "job-001",
						Type: target.TargetKubernetesJob,
						Name: "train-job-001",
						Spec: target.KubernetesJobSpec{
							Cluster:   "local-dev",
							Namespace: "ml-workloads",
							JobName:   "training-demo",
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
				"kube-job.json":  `{"id":"job-001","name":"train-job-001","type":"kubernetes-job","spec":{"cluster":"local-dev","namespace":"ml-workloads","jobName":"training-demo"}}`,
				"dirA":           `"should be skipped"`,
				"kube-job2.json": `{"id":"job-002","name":"train-job-002","type":"kubernetes-job","spec":{"cluster":"local-dev","namespace":"ml-workloads","jobName":"training-demo02"}}`,
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
					{
						ID:   "job-001",
						Type: target.TargetKubernetesJob,
						Name: "train-job-001",
						Spec: target.KubernetesJobSpec{
							Cluster:   "local-dev",
							Namespace: "ml-workloads",
							JobName:   "training-demo",
						},
					},
					{
						ID:   "job-002",
						Type: target.TargetKubernetesJob,
						Name: "train-job-002",
						Spec: target.KubernetesJobSpec{
							Cluster:   "local-dev",
							Namespace: "ml-workloads",
							JobName:   "training-demo02",
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
				"kube-job.json":  `{"id":"job-001","name":"train-job-001","type":"kubernetes-job","spec":{"cluster":"local-dev","namespace":"ml-workloads","jobName":"training-demo"}}`,
				"kube-job2.json": `{"id":"job-002","name":"train-job-002","type":"kubernetes-job","spec":{"cluster":"local-dev","namespace":"ml-workloads","jobName":"training-demo02"}}`,
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
					{
						ID:   "job-001",
						Type: target.TargetKubernetesJob,
						Name: "train-job-001",
						Spec: target.KubernetesJobSpec{
							Cluster:   "local-dev",
							Namespace: "ml-workloads",
							JobName:   "training-demo",
						},
					},
					{
						ID:   "job-002",
						Type: target.TargetKubernetesJob,
						Name: "train-job-002",
						Spec: target.KubernetesJobSpec{
							Cluster:   "local-dev",
							Namespace: "ml-workloads",
							JobName:   "training-demo02",
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
				"kube-job.json":  `{"id":"job-001","name":"train-job-001","type":"kubernetes-job","spec":{"cluster":"local-dev","namespace":"ml-workloads","jobName":"training-demo"}}`,
				"kube-job2.json": `{"id":"job-002""name":"train-job-002","type":"kubernetes-job","spec":{"cluster":"local-dev","namespace":"ml-workloads","jobName":"training-demo02"}}`,
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
					{
						ID:   "job-001",
						Type: target.TargetKubernetesJob,
						Name: "train-job-001",
						Spec: target.KubernetesJobSpec{
							Cluster:   "local-dev",
							Namespace: "ml-workloads",
							JobName:   "training-demo",
						},
					},
				},
				Failures: []discovery.DiscoveryFailure{
					{
						Item: "kube-job2.json",
						Err:  errors.New("\"kube-job2.json\": invalid character "),
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
				"kube-job.json":  `{"id":"job-001","name":"train-job-001","type":"kubernetes-job","spec":{"cluster":"local-dev","namespace":"ml-workloads","jobName":"training-demo"}}`,
				"kube-job2.json": `{"id":"job-002","name":"train-job-002","type":"kubernetes-jobs","spec":{"cluster":"local-dev","namespace":"ml-workloads","jobName":"training-demo02"}}`,
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
					{
						ID:   "job-001",
						Type: target.TargetKubernetesJob,
						Name: "train-job-001",
						Spec: target.KubernetesJobSpec{
							Cluster:   "local-dev",
							Namespace: "ml-workloads",
							JobName:   "training-demo",
						},
					},
				},
				Failures: []discovery.DiscoveryFailure{
					{
						Item: "kube-job2.json",
						Err:  errors.New("\"kube-job2.json\": target \"train-job-002\": unknown target"),
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
				"kube-job.json":  `{"id":"job-001","name":"train-job-001","type":"kubernetes-job","spec":{"cluster":"local-dev","namespace":"ml-workloads","jobName":"training-demo"}}`,
				"kube-job2.json": `{"id":"job-002","name":"train-job-002","type":"kubernetes-job","spec":{"cluster":"","namespace":"ml-workloads","jobName":"training-demo02"}}`,
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
					{
						ID:   "job-001",
						Type: target.TargetKubernetesJob,
						Name: "train-job-001",
						Spec: target.KubernetesJobSpec{
							Cluster:   "local-dev",
							Namespace: "ml-workloads",
							JobName:   "training-demo",
						},
					},
				},
				Failures: []discovery.DiscoveryFailure{
					{
						Item: "kube-job2.json",
						Err:  errors.New("\"kube-job2.json\": target \"train-job-002\": invalid target spec"),
					},
				},
			},
			wantProcessErr: false,
			wantErrType:    ErrInvalidTargetSpec,
		},
		{
			name: "unknown-target-spec",
			config: config.LocalSourceConfig{
				Name: "local-dev",
				Path: "data",
			},
			fileMap: map[string]string{
				"kube-job.json":  `{"id":"job-001","name":"train-job-001","type":"kubernetes-job","spec":{"cluster":"local-dev","namespace":"ml-workloads","jobName":"training-demo"}}`,
				"kube-job2.json": `{"id":"job-002","name":"train-job-002","type":"kubernetes-job","spec":{"cluster":"","namespace":"ml-workloads","jobName":"training-demo02","completions":"4"}}`,
			},
			wantResult: discovery.DiscoveryResult{
				Targets: []target.Target{
					{
						ID:   "job-001",
						Type: target.TargetKubernetesJob,
						Name: "train-job-001",
						Spec: target.KubernetesJobSpec{
							Cluster:   "local-dev",
							Namespace: "ml-workloads",
							JobName:   "training-demo",
						},
					},
				},
				Failures: []discovery.DiscoveryFailure{
					{
						Item: "kube-job2.json",
						Err:  errors.New("\"kube-job2.json\": target \"train-job-002\": json: unknown field \"completions\""),
					},
				},
			},
			wantProcessErr: false,
			wantErrType:    nil,
		},
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
					// is a regular file
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

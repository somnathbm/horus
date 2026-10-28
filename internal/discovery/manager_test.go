package discovery

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/somnathbm/horus/internal/target"
)

// fake discoverers
type mockDiscoverer struct {
	result DiscoveryResult
	err    error
}

func (md mockDiscoverer) Discover(ctx context.Context) (DiscoveryResult, error) {
	if err := ctx.Err(); err != nil {
		return DiscoveryResult{}, err
	}

	return md.result, md.err
}

// counting fake discoverer
type countingDiscoverer struct {
	calls *int
}

func (d countingDiscoverer) Discover(ctx context.Context) (DiscoveryResult, error) {
	*d.calls++
	return DiscoveryResult{}, nil
}

// sentinel errors
var (
	errFatal         = errors.New("fatal error")
	errInvalidSchema = errors.New("invalid schema in target")
	errUnknownField  = errors.New("unknown field in target json")
	errAWS           = errors.New("aws fatal error")
	errKubernetes    = errors.New("kubernetes fatal error")
)

func TestManager(t *testing.T) {
	// discoverer no failure
	discovererNoFailure := mockDiscoverer{
		result: DiscoveryResult{
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
		},
	}

	// discoverer with failure
	discovererWithFailure := mockDiscoverer{
		result: DiscoveryResult{
			Targets: []target.Target{
				{
					ID:   "job-002-fail",
					Type: target.TargetKubernetesJob,
					Name: "train-job-002",
					Spec: target.KubernetesJobSpec{
						Cluster:   "local-staging",
						Namespace: "ml-workloads",
						JobName:   "training-demo-02",
					},
				},
			},
			Failures: []DiscoveryFailure{
				{Item: "kube-job2.json", Err: errInvalidSchema},
			},
		},
	}

	// discoverer no target
	discovererNoTarget := mockDiscoverer{
		result: DiscoveryResult{
			Failures: []DiscoveryFailure{
				{Item: "kube-job.json", Err: errUnknownField},
			},
		},
	}

	// discoverer with fatal error
	discovererWithError := mockDiscoverer{
		result: DiscoveryResult{},
		err:    errFatal,
	}

	// discoverer with result along with fatal error
	discovererWithResultWithError := mockDiscoverer{
		result: DiscoveryResult{
			Failures: []DiscoveryFailure{
				{Item: "kube-job.json", Err: errUnknownField},
			},
		},
		err: errFatal,
	}

	// aws discoverer with fatal error
	discovererAWSWithError := mockDiscoverer{
		err: errAWS,
	}

	// kubernetes discoverer with fatal error
	discovererKubernetesWithError := mockDiscoverer{
		err: errKubernetes,
	}

	// tests
	tests := []struct {
		name        string
		discoverers []Discoverer
		wantResult  DiscoveryResult
		fatalErrors []error
	}{
		{
			name:       "zero-discoverer",
			wantResult: DiscoveryResult{},
		},
		{
			name:        "one-discoverer-no-failure-no-fatal-err",
			discoverers: []Discoverer{discovererNoFailure},
			wantResult: DiscoveryResult{
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
			},
		},
		{
			name:        "multi-discoverers-with-one-failure",
			discoverers: []Discoverer{discovererNoFailure, discovererWithFailure},
			wantResult: DiscoveryResult{
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
						ID:   "job-002-fail",
						Type: target.TargetKubernetesJob,
						Name: "train-job-002",
						Spec: target.KubernetesJobSpec{
							Cluster:   "local-staging",
							Namespace: "ml-workloads",
							JobName:   "training-demo-02",
						},
					},
				},
				Failures: []DiscoveryFailure{
					{Item: "kube-job2.json", Err: errInvalidSchema},
				},
			},
		},
		{
			name:        "zero-target",
			discoverers: []Discoverer{discovererNoTarget},
			wantResult: DiscoveryResult{
				Failures: []DiscoveryFailure{
					{Item: "kube-job.json", Err: errUnknownField},
				},
			},
		},
		{
			name:        "discoverers-result-fatal-error",
			discoverers: []Discoverer{discovererNoFailure, discovererWithError},
			wantResult: DiscoveryResult{
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
			},
			fatalErrors: []error{errFatal},
		},
		{
			name:        "discoverers-result-failure-multi-fatal-error",
			discoverers: []Discoverer{discovererWithError, discovererWithResultWithError, discovererNoFailure},
			wantResult: DiscoveryResult{
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
				Failures: []DiscoveryFailure{
					{Item: "kube-job.json", Err: errUnknownField},
				},
			},
			fatalErrors: []error{errFatal},
		},
		{
			name:        "discoverers-result-aws-k8s-failure",
			discoverers: []Discoverer{discovererNoFailure, discovererAWSWithError, discovererKubernetesWithError},
			wantResult: DiscoveryResult{
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
			},
			fatalErrors: []error{errKubernetes, errAWS},
		},
	}

	// canceled context
	t.Run("cancelled-context", func(t *testing.T) {
		calls := 0
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		dscvryManager := New(countingDiscoverer{calls: &calls})
		result, err := dscvryManager.Discover(ctx)

		if !errors.Is(err, context.Canceled) {
			t.Errorf("expecting cancelled context, got: %v", err)
		}

		if calls != 0 {
			t.Errorf("expecting zero calls, got: %v", calls)
		}

		// for pre-cancelled context, no discoverer should not be invoked
		if len(result.Targets) != 0 || len(result.Failures) != 0 {
			t.Errorf("expecting no discoverer to be called in cancelled context, got: %v", result.Targets)
		}
	})

	// range over tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dscvryManager := New(tc.discoverers...)
			gotDscvrResult, err := dscvryManager.Discover(t.Context())

			// if err != nil && !errors.Is(err, tc.fatalErr) {
			// 	t.Errorf("expected process failure to be %v, got %v", tc.fatalErr, err)
			// }

			// target/failure count assertions
			if len(tc.wantResult.Targets) != len(gotDscvrResult.Targets) {
				t.Errorf("expected target count to be: %v, got: %v", len(tc.wantResult.Targets), len(gotDscvrResult.Targets))
			}
			if len(tc.wantResult.Failures) != len(gotDscvrResult.Failures) {
				t.Errorf("expected failure count to be: %v, got: %v", len(tc.wantResult.Failures), len(gotDscvrResult.Failures))
			}

			// TODO: later
			// if len(tc.fatalErrors) == 0 {
			// 	if err == nil {}
			// }

			// targets assertions
			if !slices.Equal(tc.wantResult.Targets, gotDscvrResult.Targets) {
				t.Errorf("expected targets: %v, got: %v", tc.wantResult.Targets, gotDscvrResult.Targets)
			}

			// failure assertions
			for i, gotFailure := range gotDscvrResult.Failures {
				wantFailure := tc.wantResult.Failures[i]

				// failure item assertions
				if gotFailure.Item != wantFailure.Item {
					t.Errorf("expecting failure item to be %v, got %v", wantFailure.Item, gotFailure.Item)
				}

				// failure eror type assertions
				if !errors.Is(gotFailure.Err, wantFailure.Err) {
					t.Errorf("expecting failure to be %v, got %v", wantFailure.Err, gotFailure.Err)
				}
			}

			// fatal error assertions
			for _, fatalErr := range tc.fatalErrors {
				if !errors.Is(err, fatalErr) {
					t.Errorf("expecting a fatal error of %v, got %v", fatalErr, err)
				}
			}
		})
	}
}

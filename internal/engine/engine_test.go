package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/somnathbm/horus/internal/health"
	"github.com/somnathbm/horus/internal/registry"
	"github.com/somnathbm/horus/internal/target"
)

var (
	ErrInvalidTarget     = errors.New("invalid target")
	ErrInvalidTargetSpec = errors.New("invalid target spec")
)

// fake plugin that implements the target.Plugin interface for testing purposes
type fakePlugin struct {
	calls      int
	calledWith target.Target
	status     health.WorkloadStatus
	err        error
}

func (fp *fakePlugin) Check(ctx context.Context, t target.Target) (health.WorkloadStatus, error) {
	fp.calls++
	fp.calledWith = t
	return fp.status, fp.err
}

// dummy target
var dummyTarget target.Target = target.Target{
	ID:   "job-005",
	Name: "test-job-005",
	Type: target.TargetKubernetesJob,
	Spec: target.KubernetesJobSpec{
		Cluster:   "local-dev",
		Namespace: "ml-workloads",
		JobName:   "training-job-001",
	},
}

func TestEngine(t *testing.T) {
	otherTargetType := target.TargetType("kueue-workload")

	tests := []struct {
		name        string
		targetTypes []target.TargetType
		fPluginMap  map[target.TargetType]*fakePlugin
		wantError   error
	}{
		{
			name:        "multiple-target-types",
			targetTypes: []target.TargetType{target.TargetKubernetesJob, otherTargetType},
			fPluginMap: map[target.TargetType]*fakePlugin{
				target.TargetKubernetesJob: {
					status: health.WorkloadStatus{
						TargetID: "job-005",
						State:    health.StateUnknown,
						Reason:   "target inspection not implemented",
					},
				},
				otherTargetType: {
					status: health.WorkloadStatus{
						TargetID: "workload-001",
						State:    health.StateUnknown,
						Reason:   "target inspection not implemented",
					},
				},
			},
			wantError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 1. instantiate registry
			reg := registry.New()
			for _, targetType := range tc.targetTypes {
				// register the fake plugin for each target type
				err := reg.Register(targetType, tc.fPluginMap[targetType])
				if err != nil {
					t.Fatalf("failed to register target type: \"%v\" with error: %v", targetType, err)
				}
			}

			// 2. instantiate engine with the registry
			eng := New(reg)
			status, err := eng.Evaluate(t.Context(), dummyTarget)

			// 3. ======== assertions ========
			// error assertions
			if tc.wantError != nil {
				if !errors.Is(err, tc.wantError) {
					t.Errorf("expecting error to be %v, got %v", tc.wantError, err)
				}
			}

			// assert Evaluate and plugin error matches
			if err != tc.fPluginMap[dummyTarget.Type].err {
				t.Errorf("expecting error to be %v, got %v", tc.fPluginMap[dummyTarget.Type].err, err)
			}

			// successful assertions
			// assert health status
			if status != tc.fPluginMap[dummyTarget.Type].status {
				t.Errorf("expecting status to match with plugin status %v, got %v", tc.fPluginMap[dummyTarget.Type].status, status)
			}

			// assert CORRECT plugin is actually called
			if tc.fPluginMap[dummyTarget.Type].calls != 1 {
				t.Fatalf("plugin not called")
			}

			// assert other plugin is NOT called
			if tc.fPluginMap[otherTargetType].calls != 0 {
				t.Fatalf("other plugin should not be called")
			}

			// assert plugin is called with the correct target
			if tc.fPluginMap[dummyTarget.Type].calledWith != dummyTarget {
				t.Errorf("expecting plugin to be called with target %v, got %v", dummyTarget, tc.fPluginMap[dummyTarget.Type].calledWith)
			}
		})
	}
}

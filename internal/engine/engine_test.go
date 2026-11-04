package engine

import (
	"context"
	"errors"
	"fmt"
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
	// if target is not "kubernetes-job"
	if t.Type == "" {
		fp.err = fmt.Errorf("target %v is %w", t.Name, ErrInvalidTarget)
		return fp.status, fp.err
	}

	// if spec is missing
	if t.Spec == nil {
		fp.err = fmt.Errorf("target %v contains %w", t.Name, ErrInvalidTargetSpec)
		return fp.status, fp.err
	}

	// populate the plugin fields
	fp.calls++
	fp.calledWith = t
	fp.status = health.WorkloadStatus{
		TargetID: t.ID,
		State:    health.StateUnknown,
		Reason:   "target inspection not implemented",
	}
	fp.err = nil

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
	tests := []struct {
		name       string
		targetType target.TargetType
		fPlugin    *fakePlugin
		// wantHealthStatus health.WorkloadStatus
		wantError error
	}{
		{
			name:       "registered-target-type",
			targetType: target.TargetKubernetesJob,
			fPlugin:    &fakePlugin{},
			// wantHealthStatus: health.WorkloadStatus{
			// 	TargetID: "job-005",
			// 	State:    health.StateUnknown,
			// 	Reason:   "target inspection not implemented",
			// },
			wantError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// instantiate registry
			reg := registry.New()
			reg.Register(tc.targetType, tc.fPlugin)

			// instantiate engine with the registry
			eng := New(reg)
			status, err := eng.Evaluate(t.Context(), dummyTarget)

			// assertions
			// unexpected error
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			// successful assertions
			// assert health status
			if status != tc.fPlugin.status {
				t.Errorf("expecting status to match with plugin status %v, got %v", tc.fPlugin.status, status)
			}

			// assert plugin is actually called
			if tc.fPlugin.calls != 1 {
				t.Fatalf("plugin not called")
			}

			// assert plugin is actually called for that target type
			if tc.fPlugin.calledWith.Type != tc.targetType {
				t.Errorf("expecting called target type to be %v, got %v", tc.fPlugin.calledWith.Type, tc.targetType)
			}

			// error assertions
			if tc.wantError != nil {
				if !errors.Is(err, tc.wantError) {
					t.Errorf("expecting error to be %v, got %v", tc.wantError, err)
				}
			}
		})
	}
}

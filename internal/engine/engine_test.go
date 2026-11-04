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
}

func (f *fakePlugin) Check(ctx context.Context, t target.Target) (health.WorkloadStatus, error) {
	// if target is empty
	if t.Type == "" {
		return health.WorkloadStatus{}, fmt.Errorf("target %v is %w", t.Name, ErrInvalidTarget)
	}

	// if spec is missing
	if t.Spec == nil {
		return health.WorkloadStatus{}, fmt.Errorf("target %v contains %w", t.Name, ErrInvalidTargetSpec)
	}

	f.calls++
	f.calledWith = t

	return health.WorkloadStatus{
		TargetID: t.ID,
		State:    health.StateUnknown,
		Reason:   "target inspection not implemented",
	}, nil
}

var dummyTarget target.Target = target.Target{
	ID:   "job-001",
	Name: "test-job-001",
	Type: target.TargetKubernetesJob,
	Spec: target.KubernetesJobSpec{
		Cluster:   "test-cluster",
		Namespace: "test-namespace",
		JobName:   "test-job",
	},
}

// do tests
func TestEngine(t *testing.T) {
	// test registered target types
	t.Run("registered-target-types", func(t *testing.T) {
		// create a registry and register the fake plugins for "kubernetes-job" target type
		reg := registry.New()

		// register fake plugins for "kubernetes-job" target type
		kubeJobPluginFake := fakePlugin{}
		kueuePluginFake := fakePlugin{}
		reg.Register(target.TargetKubernetesJob, &kubeJobPluginFake)
		reg.Register(target.TargetKubernetesJob, &kueuePluginFake)

		// create an engine with the registry
		engine := New(reg)

		workloadStatus, err := engine.Evaluate(t.Context(), dummyTarget)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// assert workloadStatus is not empty
		if workloadStatus != (health.WorkloadStatus{}) {
			// assert error is nil
			if err == nil {
				// success
			}

			// assert correct plugin was called
			if kubeJobPluginFake.calls != 1 {
				t.Errorf("expected kubeJobPluginFake to be called once, but it was called %d times", kubeJobPluginFake.calls)
			}

			if kueuePluginFake.calls != 0 {
				t.Errorf("expected kueuePluginFake to not be called, but it was called %d times", kueuePluginFake.calls)
			}

			// assert plugin was called with the correct target
			if kubeJobPluginFake.calledWith != dummyTarget {
				t.Errorf("expected kubeJobPluginFake to be called with %v, but it was called with %v", dummyTarget, kubeJobPluginFake.calledWith)
			}
		}
	})
}

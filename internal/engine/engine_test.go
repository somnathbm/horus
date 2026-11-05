package engine

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/somnathbm/horus/internal/health"
	"github.com/somnathbm/horus/internal/registry"
	"github.com/somnathbm/horus/internal/target"
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
	otherTargetType := target.TargetType("other-target-type")

	tests := []struct {
		name        string
		targetTypes []target.TargetType
		fPluginMap  map[target.TargetType]*fakePlugin
		wantError   error
		wantCalls   map[target.TargetType]int
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
			wantCalls: map[target.TargetType]int{
				target.TargetKubernetesJob: 1,
				otherTargetType:            0,
			},
		},
		{
			name: "unregistered-target-type",
			targetTypes: []target.TargetType{
				otherTargetType,
			},
			fPluginMap: map[target.TargetType]*fakePlugin{
				otherTargetType: {
					err: registry.ErrPluginNotFound,
				},
			},
			wantError: registry.ErrPluginNotFound,
			wantCalls: map[target.TargetType]int{
				otherTargetType: 0,
			},
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
			// error path
			if tc.wantError != nil {
				if !errors.Is(err, tc.wantError) {
					t.Fatalf("expecting error to be %v, got %v", tc.wantError, err)
					return
				}
			}

			// common path: assert plugin call counts
			for k, v := range tc.fPluginMap {
				if v.calls != tc.wantCalls[k] {
					t.Fatalf("plugin for target type %v should be called %d times, got %d", k, tc.wantCalls[k], v.calls)
					return
				}
			}

			// happy path: assert plugin status and called target
			if tc.wantError == nil && err == nil {
				var calledPluginKey target.TargetType
				for k, v := range tc.wantCalls {
					if v == 1 {
						calledPluginKey = k
					}
				}

				// assert health status
				if !reflect.DeepEqual(status, tc.fPluginMap[calledPluginKey].status) {
					t.Errorf("expecting status to match with plugin status %v, got %v", tc.fPluginMap[calledPluginKey].status, status)
				}

				// assert plugin is called with the correct target
				if !reflect.DeepEqual(tc.fPluginMap[calledPluginKey].calledWith, dummyTarget) {
					t.Errorf("expecting plugin to be called with target %v, got %v", dummyTarget, tc.fPluginMap[calledPluginKey].calledWith)
				}

				// assert Evaluate and plugin error matches
				if !errors.Is(err, tc.fPluginMap[calledPluginKey].err) {
					t.Errorf("expecting error to be %v, got %v", tc.fPluginMap[calledPluginKey].err, err)
				}
			}
		})
	}
}

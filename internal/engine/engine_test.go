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

var errFakePlugin = errors.New("fake plugin error")

func TestEngine(t *testing.T) {
	otherTargetType := target.TargetType("other-target-type")

	tests := []struct {
		name       string
		target     target.Target
		fPluginMap map[target.TargetType]*fakePlugin
		wantError  error
		wantCalls  map[target.TargetType]int
		wantPlugin target.TargetType
	}{
		{
			name: "multiple-target-types",
			target: target.Target{
				ID:   "job-005",
				Name: "test-job-005",
				Type: target.TargetKubernetesJob,
				Spec: target.KubernetesJobSpec{
					Cluster:   "local-dev",
					Namespace: "ml-workloads",
					JobName:   "training-job-001",
				},
			},
			fPluginMap: map[target.TargetType]*fakePlugin{
				target.TargetKubernetesJob: {
					status: health.WorkloadStatus{
						TargetID: "job-005",
						State:    health.StateUnknown,
						Reason:   "target inspection not implemented",
					},
				},
				otherTargetType: {},
			},
			wantError: nil,
			wantCalls: map[target.TargetType]int{
				target.TargetKubernetesJob: 1,
				otherTargetType:            0,
			},
			wantPlugin: target.TargetKubernetesJob,
		},
		{
			name: "unregistered-target-type",
			target: target.Target{
				ID:   "job-005",
				Name: "test-job-005",
				Type: target.TargetKubernetesJob,
				Spec: target.KubernetesJobSpec{
					Cluster:   "local-dev",
					Namespace: "ml-workloads",
					JobName:   "training-job-001",
				},
			},
			fPluginMap: map[target.TargetType]*fakePlugin{
				otherTargetType: {},
			},
			wantError: registry.ErrPluginNotFound,
			wantCalls: map[target.TargetType]int{
				otherTargetType: 0,
			},
			wantPlugin: "",
		},
		{
			name: "plugin-with-error",
			target: target.Target{
				ID:   "job-005",
				Name: "test-job-005",
				Type: target.TargetKubernetesJob,
				Spec: target.KubernetesJobSpec{
					Cluster:   "local-dev",
					Namespace: "ml-workloads",
					JobName:   "training-job-001",
				},
			},
			fPluginMap: map[target.TargetType]*fakePlugin{
				target.TargetKubernetesJob: {
					err: errFakePlugin,
				},
			},
			wantError: errFakePlugin,
			wantCalls: map[target.TargetType]int{
				target.TargetKubernetesJob: 1,
			},
			wantPlugin: target.TargetKubernetesJob,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 1. ====== setup ======
			// 1.1 instantiate registry
			reg := registry.New()
			for targetType, fakePlgn := range tc.fPluginMap {
				// register the fake plugin for each target type
				err := reg.Register(targetType, fakePlgn)
				if err != nil {
					t.Fatalf("failed to register target type: \"%v\" with error: %v", targetType, err)
				}
			}

			// 1.2 instantiate engine with the registry
			eng := New(reg)
			status, err := eng.Evaluate(t.Context(), tc.target)

			// 2. ======== assertions ========
			// 2.1 common path
			// assert expected calls count for each registered plugin
			for k, v := range tc.fPluginMap {
				if _, ok := tc.wantCalls[k]; !ok {
					t.Fatalf("unexpected target type %v in wantCalls", k)
				}
				if v.calls != tc.wantCalls[k] {
					t.Fatalf("plugin for target type %v should be called %d times, got %d", k, tc.wantCalls[k], v.calls)
				}
			}
			// assert plugin is called with the correct target
			if tc.fPluginMap[tc.wantPlugin] != nil && !reflect.DeepEqual(tc.fPluginMap[tc.wantPlugin].calledWith, tc.target) {
				t.Errorf("expecting plugin to be called with target %v, got %v", tc.target, tc.fPluginMap[tc.wantPlugin].calledWith)
			}
			// 2.2 error path
			if tc.wantError != nil {
				// assert Evaluate error is what we are expecting
				if !errors.Is(err, tc.wantError) {
					t.Fatalf("expecting error to be %v, got %v", tc.wantError, err)
				}
			} else {
				// 2.3 success path
				// assert error is nil when no error is expected
				if err != nil {
					t.Fatalf("expecting no error, got %v", err)
				}

				// assert health status
				if !reflect.DeepEqual(status, tc.fPluginMap[tc.wantPlugin].status) {
					t.Errorf("expecting status to match with plugin status %v, got %v", tc.fPluginMap[tc.wantPlugin].status, status)
				}

				// // assert plugin is called with the correct target
				// if !reflect.DeepEqual(tc.fPluginMap[tc.wantPlugin].calledWith, tc.target) {
				// 	t.Errorf("expecting plugin to be called with target %v, got %v", tc.target, tc.fPluginMap[tc.wantPlugin].calledWith)
				// }
			}
		})
	}
}

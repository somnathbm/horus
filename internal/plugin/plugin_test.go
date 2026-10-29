package plugin

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/somnathbm/horus/internal/health"
	"github.com/somnathbm/horus/internal/target"
)

func TestPlugin(t *testing.T) {
	// test cancelled context
	t.Run("cancelled-context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		kubeJobPlugin := KubernetesJobPlugin{}
		_, err := kubeJobPlugin.Check(ctx, target.Target{})
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expecting a cancelled context. Got: %v", err)
		}
	})

	// test cases table
	tests := []struct {
		name       string
		target     target.Target
		wantError  error
		wantStatus health.WorkloadStatus
	}{
		{
			name: "wrong-target",
			target: target.Target{
				ID:   "job-001",
				Name: "test-job-001",
				Type: "kubernetes-pod",
			},
			wantError: ErrInvalidTarget,
		},
		{
			name: "nil-spec",
			target: target.Target{
				ID:   "job-002",
				Name: "test-job-002",
				Type: "kubernetes-job",
			},
			wantError: ErrInvalidTargetSpec,
		},
		{
			name: "wrong-concrete-spec",
			target: target.Target{
				ID:   "job-003",
				Name: "test-job-003",
				Type: "kubernetes-job",
				Spec: target.KubernetesJobSpec{
					Cluster:   "",
					Namespace: "",
					JobName:   "",
				},
			},
			wantError: ErrInvalidTargetSpec,
		},
		{
			name: "missing-required-field-spec",
			target: target.Target{
				ID:   "job-004",
				Name: "test-job-004",
				Type: "kubernetes-job",
				Spec: target.KubernetesJobSpec{
					Cluster:   "local-dev",
					Namespace: "ml-workloads",
				},
			},
			wantError: ErrInvalidTargetSpec,
		},
		{
			name: "valid-spec",
			target: target.Target{
				ID:   "job-005",
				Name: "test-job-005",
				Type: "kubernetes-job",
				Spec: target.KubernetesJobSpec{
					Cluster:   "local-dev",
					Namespace: "ml-workloads",
					JobName:   "training-job-001",
				},
			},
			wantError: nil,
			wantStatus: health.WorkloadStatus{
				TargetID: "job-005",
				State:    health.StateUnknown,
				Reason:   "kubernetes job inspection not implemented",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kubeJobPlugin := KubernetesJobPlugin{}
			status, err := kubeJobPlugin.Check(t.Context(), tc.target)

			if err != nil {
				if !errors.Is(err, tc.wantError) {
					t.Errorf("expecting error %v, got %v", tc.wantError, err)
				}
			} else {
				if !reflect.DeepEqual(tc.wantStatus, status) {
					t.Errorf("expecting status %v, got %v", tc.wantStatus, status)
				}
			}

		})
	}
}

package plugin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/somnathbm/horus/internal/health"
	"github.com/somnathbm/horus/internal/target"
)

var (
	ErrInvalidTarget     = errors.New("invalid target")
	ErrInvalidTargetSpec = errors.New("invalid target spec")
)

type KubernetesJobPlugin struct{}

func (p KubernetesJobPlugin) Check(ctx context.Context, t target.Target) (health.WorkloadStatus, error) {
	// context cancellaton
	if err := ctx.Err(); err != nil {
		return health.WorkloadStatus{}, fmt.Errorf("context: %w", err)
	}

	// if target is not "kubernetes-job"
	if t.Type != target.TargetKubernetesJob {
		return health.WorkloadStatus{}, fmt.Errorf("target %v is %w", t.Name, ErrInvalidTarget)
	}

	// if spec is missing
	if t.Spec == nil {
		return health.WorkloadStatus{}, fmt.Errorf("target %v contains %w", t.Name, ErrInvalidTargetSpec)
	}

	switch kubeJobSpec := t.Spec.(type) {
	case target.KubernetesJobSpec:
		// if spec required fields are missing
		if strings.TrimSpace(kubeJobSpec.Cluster) == "" || strings.TrimSpace(kubeJobSpec.Namespace) == "" || strings.TrimSpace(kubeJobSpec.JobName) == "" {
			return health.WorkloadStatus{}, fmt.Errorf("target %v contains %w", t.Name, ErrInvalidTargetSpec)
		}
		return health.WorkloadStatus{
			TargetID: t.ID,
			State:    health.StateUnknown,
			Reason:   "kubernetes job inspection not implemented",
		}, nil
	default:
		return health.WorkloadStatus{}, fmt.Errorf("target %v is %w", t.Name, ErrInvalidTargetSpec)
	}
}

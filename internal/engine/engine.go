package engine

import (
	"context"
	"fmt"

	"github.com/somnathbm/horus/internal/health"
	"github.com/somnathbm/horus/internal/registry"
	"github.com/somnathbm/horus/internal/target"
)

type Engine struct {
	reg *registry.Registry
}

func New(reg *registry.Registry) *Engine {
	return &Engine{
		reg: reg,
	}
}

func (e Engine) Evaluate(ctx context.Context, t target.Target) (health.WorkloadStatus, error) {
	// Implement the logic to evaluate the target and return the workload status
	if err := ctx.Err(); err != nil {
		return health.WorkloadStatus{}, err
	}

	// get the plugin from the target type
	targetPlugin, err := e.reg.Get(t.Type)
	if err != nil {
		return health.WorkloadStatus{}, fmt.Errorf("target \"%v\": plugin lookup failed: %w", t.Name, err)
	}
	
	// Evaluate the target using the plugin
	workloadStatus, err := targetPlugin.Check(ctx, t)
	if err != nil {
		return health.WorkloadStatus{}, fmt.Errorf("target \"%v\": health check failed: %w", t.Name, err)
	}
	return workloadStatus, nil
}

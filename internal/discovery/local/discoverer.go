package local

import (
	"context"
	"fmt"

	"github.com/somnathbm/horus/internal/config"
	"github.com/somnathbm/horus/internal/target"
)

type LocalDiscoverer struct {
	name string
	path string
}

func New(cfg config.LocalSourceConfig) *LocalDiscoverer {
	return &LocalDiscoverer{
		name: cfg.Name,
		path: cfg.Path,
	}
}

func (ld *LocalDiscoverer) Discover(ctx context.Context) ([]target.Target, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("local: %w", ctx.Err())
	default:
		// load local targets from JSON
		targets, err := Load(ctx, ld.path)
		if err != nil {
			return nil, fmt.Errorf("local: %w", err)
		}

		return targets, nil
	}
}

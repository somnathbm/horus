package discovery

import (
	"context"
	"errors"
	"fmt"

	"github.com/somnathbm/horus/internal/config"
	"github.com/somnathbm/horus/internal/discovery/local"
	"github.com/somnathbm/horus/internal/target"
)

// A discoverer interface
type Discoverer interface {
	Discover(ctx context.Context) ([]target.Target, error)
}

func Start(config config.DiscoveryConfig, ctx context.Context) ([]target.Target, error) {
	select {
	case <-ctx.Done():
		return nil, errors.New("context cancelled")
	default:
		var targets []target.Target

		// local discovery
		for _, localConfig := range config.Local {
			localDscvr := local.New(localConfig)
			localTargets, err := localDscvr.Discover(ctx)
			if err != nil {
				return nil, fmt.Errorf("discovery.%w", err)
			}
			targets = append(targets, localTargets...)
		}

		// aws discovery

		// kubernetes discovery

		return targets, nil
	}
}

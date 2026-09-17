package discovery

import (
	"context"

	"github.com/somnathbm/horus/internal/target"
)

type DiscoveryResult struct {
	Targets  []target.Target
	Failures []DiscoveryFailure
}

type DiscoveryFailure struct {
	Item string
	Err  error
}

type Discoverer interface {
	Discover(ctx context.Context) (DiscoveryResult, error)
}

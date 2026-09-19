package discovery

import (
	"context"
	"fmt"
)

type Manager struct {
	discoverers []Discoverer
}

func New(discoverers ...Discoverer) Manager {
	return Manager{discoverers: discoverers}
}

func (m Manager) Discover(ctx context.Context) (DiscoveryResult, error) {
	var result DiscoveryResult

	for _, discoverer := range m.discoverers {
		dscvryResult, dscvryErr := discoverer.Discover(ctx)
		// context cancelled
		if err := ctx.Err(); err != nil {
			return result, fmt.Errorf("discovery manager: %w", err)
		}
		// process level error
		if dscvryErr != nil {
			return result, fmt.Errorf("discovery manager: %w", dscvryErr)
		}

		result.Targets = append(result.Targets, dscvryResult.Targets...)
		result.Failures = append(result.Failures, dscvryResult.Failures...)
	}

	// result can be enriched later for reporting
	return result, nil
}

package discovery

import (
	"context"

	"github.com/somnathbm/horus/internal/config"
)

type Manager struct {
	discoverers []Discoverer
}

func New(config config.DiscoveryConfig) {

}

func (m *Manager) Discover(ctx context.Context) (DiscoveryResult, error)

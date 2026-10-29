package plugin

import (
	"context"

	"github.com/somnathbm/horus/internal/health"
	"github.com/somnathbm/horus/internal/target"
)

type Plugin interface {
	Check(
		ctx context.Context,
		t target.Target,
	) (health.WorkloadStatus, error)
}

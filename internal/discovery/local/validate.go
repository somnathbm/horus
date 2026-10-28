package local

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/somnathbm/horus/internal/target"
)

var (
	ErrInvalidTarget     = errors.New("invalid target")
	ErrInvalidTargetSpec = errors.New("invalid target spec")
)

func validate(targetEntry target.Target) error {
	targetTypes := []target.TargetType{target.TargetKubernetesJob}
	if targetEntry.ID == "" || targetEntry.Name == "" || targetEntry.Type == "" || !slices.Contains(targetTypes, targetEntry.Type) {
		return fmt.Errorf("target %q - %w", targetEntry.Name, ErrInvalidTarget)
	}

	switch spec := targetEntry.Spec.(type) {
	case target.KubernetesJobSpec:
		if targetEntry.Type != target.TargetKubernetesJob || strings.TrimSpace(spec.Cluster) == "" || strings.TrimSpace(spec.Namespace) == "" || strings.TrimSpace(spec.JobName) == "" {
			return fmt.Errorf("target %q - %w", targetEntry.Name, ErrInvalidTargetSpec)
		}
	case nil:
		return fmt.Errorf("target %q - spec is nil: %w", targetEntry.Name, ErrInvalidTargetSpec)
	default:
		return fmt.Errorf("target %q - unknown spec type: %w", targetEntry.Name, ErrInvalidTargetSpec)
	}
	return nil
}

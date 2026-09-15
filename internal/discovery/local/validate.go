package local

import (
	"errors"
	"fmt"
	"slices"

	"github.com/somnathbm/horus/internal/target"
)

var (
	ErrInvalidTarget     = errors.New("invalid target")
	ErrInvalidTargetSpec = errors.New("invalid target spec")
)

func validate(targetEntry target.Target) error {
	targetTypes := []target.TargetType{target.HTTPTargetType, target.PostgresSQLTargetType, target.RedisTargetType, target.KafkaTargetType}
	if targetEntry.ID == "" || targetEntry.Name == "" || targetEntry.Type == "" || !slices.Contains(targetTypes, targetEntry.Type) {
		return fmt.Errorf("target %q: %w", targetEntry.Name, ErrInvalidTarget)
	}

	switch spec := targetEntry.Spec.(type) {
	case target.HTTPTargetSpec:
		if spec.Port == 0 || spec.URL == "" {
			return fmt.Errorf("target %q: %w", targetEntry.Name, ErrInvalidTargetSpec)
		}
	case target.PostgresTargetSpec:
		if spec.Port == 0 || spec.Host == "" || spec.Username == "" {
			return fmt.Errorf("target %q: %w", targetEntry.Name, ErrInvalidTargetSpec)
		}
	case target.RedisTargetSpec:
		if spec.Port == 0 || spec.URL == "" {
			return fmt.Errorf("target %q: %w", targetEntry.Name, ErrInvalidTargetSpec)
		}
	case nil:
		return fmt.Errorf("target %q: spec is nil: %w", targetEntry.Name, ErrInvalidTargetSpec)
	default:
		return fmt.Errorf("target %q: unknown spec typepec: %w", targetEntry.Name, ErrInvalidTargetSpec)
	}
	return nil
}

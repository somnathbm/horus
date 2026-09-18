package local

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/somnathbm/horus/internal/target"
)

var (
	ErrInvalidTarget     = errors.New("invalid target")
	ErrInvalidTargetSpec = errors.New("invalid target spec")
)
var allowedSchemes = []string{"http", "https"}

func validate(targetEntry target.Target) error {
	targetTypes := []target.TargetType{target.HTTPTargetType, target.PostgresSQLTargetType, target.RedisTargetType, target.KafkaTargetType}
	if targetEntry.ID == "" || targetEntry.Name == "" || targetEntry.Type == "" || !slices.Contains(targetTypes, targetEntry.Type) {
		return fmt.Errorf("target %q - %w", targetEntry.Name, ErrInvalidTarget)
	}

	switch spec := targetEntry.Spec.(type) {
	case target.HTTPTargetSpec:
		urlRes, parseErr := url.Parse(spec.URL)
		if parseErr != nil {
			return fmt.Errorf("target %q - %w", targetEntry.Name, ErrInvalidTargetSpec)
		}
		if targetEntry.Type != target.HTTPTargetType || !(spec.Port >= 1 && spec.Port <= 65535) || strings.TrimSpace(spec.URL) == "" || !slices.Contains(allowedSchemes, urlRes.Scheme) || urlRes.Host == "" {
			return fmt.Errorf("target %q - %w", targetEntry.Name, ErrInvalidTargetSpec)
		}
	case target.PostgresTargetSpec:
		if targetEntry.Type != target.PostgresSQLTargetType || !(spec.Port >= 1 && spec.Port <= 65535) || strings.TrimSpace(spec.Host) == "" || strings.TrimSpace(spec.Username) == "" {
			return fmt.Errorf("target %q - %w", targetEntry.Name, ErrInvalidTargetSpec)
		}
	case target.RedisTargetSpec:
		if targetEntry.Type != target.RedisTargetType || !(spec.Port >= 1 && spec.Port <= 65535) || strings.TrimSpace(spec.URL) == "" {
			return fmt.Errorf("target %q - %w", targetEntry.Name, ErrInvalidTargetSpec)
		}
	case nil:
		return fmt.Errorf("target %q - spec is nil: %w", targetEntry.Name, ErrInvalidTargetSpec)
	default:
		return fmt.Errorf("target %q - unknown spec type: %w", targetEntry.Name, ErrInvalidTargetSpec)
	}
	return nil
}

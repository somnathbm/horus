package local

import (
	"errors"
	"fmt"
	"strings"

	"github.com/somnathbm/horus/internal/target"
)

var (
	ErrUnknownTarget = errors.New("unknown target")
)

func decodeTargetSpec(rawTarget RawTarget) (target.Target, error) {
	switch rawTarget.Type {
	case target.HTTPTargetType:
		var decodedHttpSpec target.HTTPTargetSpec
		err := StrictUnmarshal(rawTarget.Spec, &decodedHttpSpec)
		if err != nil {
			return target.Target{}, fmt.Errorf("target %q: %w", rawTarget.Name, err)
		}
		// assemble target
		return target.Target{
			ID:   strings.TrimSpace(rawTarget.ID),
			Type: rawTarget.Type,
			Name: strings.TrimSpace(rawTarget.Name),
			Spec: decodedHttpSpec,
		}, nil

	case target.PostgresSQLTargetType:
		var decodedPostgresSpec target.PostgresTargetSpec
		err := StrictUnmarshal(rawTarget.Spec, &decodedPostgresSpec)
		if err != nil {
			return target.Target{}, fmt.Errorf("target %q: %w", rawTarget.Name, err)
		}
		// assemble target
		return target.Target{
			ID:   strings.TrimSpace(rawTarget.ID),
			Type: rawTarget.Type,
			Name: strings.TrimSpace(rawTarget.Name),
			Spec: decodedPostgresSpec,
		}, nil

	case target.RedisTargetType:
		var decodedRedisSpec target.RedisTargetSpec
		err := StrictUnmarshal(rawTarget.Spec, &decodedRedisSpec)
		if err != nil {
			return target.Target{}, fmt.Errorf("target %q: %w", rawTarget.Name, err)
		}
		// assemble target
		return target.Target{
			ID:   strings.TrimSpace(rawTarget.ID),
			Type: rawTarget.Type,
			Name: strings.TrimSpace(rawTarget.Name),
			Spec: decodedRedisSpec,
		}, nil
	default:
		return target.Target{}, fmt.Errorf("target %q: %w", rawTarget.Name, ErrUnknownTarget)
	}
}

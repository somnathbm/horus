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

func decodeTargetSpec(rawTarget rawTarget) (target.Target, error) {
	switch rawTarget.Type {
	case target.TargetKubernetesJob:
		var decodedKubernetesJobSpec target.KubernetesJobSpec
		err := strictUnmarshal(rawTarget.Spec, &decodedKubernetesJobSpec)
		if err != nil {
			return target.Target{}, fmt.Errorf("target %q: %w", rawTarget.Name, err)
		}
		// assemble target
		return target.Target{
			ID:   strings.TrimSpace(rawTarget.ID),
			Type: rawTarget.Type,
			Name: strings.TrimSpace(rawTarget.Name),
			Spec: decodedKubernetesJobSpec,
		}, nil
	default:
		return target.Target{}, fmt.Errorf("target %q: %w", rawTarget.Name, ErrUnknownTarget)
	}
}

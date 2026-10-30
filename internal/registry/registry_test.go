package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/somnathbm/horus/internal/plugin"
	"github.com/somnathbm/horus/internal/target"
)

func TestRegistry(t *testing.T) {
	// test cancelled context
	t.Run("cancelled-context", func(t *testing.T) {
		// create context first
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// instantiate registry
		reg := New()
		err := reg.Register(ctx, target.TargetKubernetesJob, plugin.KubernetesJobPlugin{})

		// assert on error
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expecting cancelled context, but got %v", err)
		}
	})

	// test cases
	tests := []struct {
		name       string
		targetType target.TargetType
		plugin     plugin.Plugin
		wantError  error
	}{
		{
			name:       "successful-registration",
			targetType: target.TargetKubernetesJob,
			plugin:     plugin.KubernetesJobPlugin{},
			wantError:  nil,
		},
		{
			name:       "duplicate-registration",
			targetType: target.TargetKubernetesJob,
			plugin:     plugin.KubernetesJobPlugin{},
			wantError:  ErrPluginExist,
		},
		{
			name:       "nil-plugin",
			targetType: target.TargetKubernetesJob,
			plugin:     nil,
			wantError:  ErrPluginExist,
		},
		{
			name:       "successful-retrieve",
			targetType: target.TargetKubernetesJob,
			plugin:     plugin.KubernetesJobPlugin{},
			wantError:  nil,
		},
		{
			name:      "missing-key",
			plugin:     plugin.KubernetesJobPlugin{},
			wantError: ErrNoTarget,
		},
	}

	// instantiate once
	reg := New()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			if _, err := reg.Get(t.Context(), tc.targetType); err != nil {
				err = reg.Register(t.Context(), tc.targetType, tc.plugin)
			}

			if err == nil {
				// assert the plugins map
				if _, ok := reg.plugins[tc.targetType]; !ok {
					t.Errorf("expecting plugin to be present, but got absent")
				}
			} else {
				// assert nil plugin registration
				if tc.plugin == nil && !errors.Is(err, ErrInvalidPlugin) {
					t.Errorf("expecting a valid plugin, but got %v", ErrInvalidPlugin)
				}

				// assert valid plugin, but duplicate registration
				if tc.plugin != nil && !errors.Is(err, ErrPluginExist) {
					t.Errorf("expecting a %v error, but got %v", ErrPluginExist, err)
				}

				// assert missing target type
				if tc.targetType == "" && !errors.Is(err, ErrNoTarget) {
					t.Errorf("expecting a %v error, but got %v", ErrNoTarget, err)
				}

				// assert no valid plugin for the target
				if tc.targetType != "" && !errors.Is(err, ErrPluginNotFound) {
					t.Errorf("expecting a %v error, but got %v", ErrPluginNotFound, err)
				}
			}
		})
	}
}

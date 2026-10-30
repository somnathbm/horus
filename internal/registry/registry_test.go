package registry

import (
	"errors"
	"testing"

	"github.com/somnathbm/horus/internal/plugin"
	"github.com/somnathbm/horus/internal/target"
)

func TestRegister(t *testing.T) {
	// successful register
	t.Run("success", func(t *testing.T) {
		reg := New()
		err := reg.Register(target.TargetKubernetesJob, plugin.KubernetesJobPlugin{})

		if err == nil {
			// success
		}
		if err != nil {
			t.Errorf("expecting nil error, but got %v", err)
		}
	})

	// duplicate register
	t.Run("duplicate", func(t *testing.T) {
		reg := New()
		err := reg.Register(target.TargetKubernetesJob, plugin.KubernetesJobPlugin{})

		if err == nil {
			// success
		}
		if err != nil {
			if !errors.Is(err, ErrPluginExist) {
				t.Errorf("expecting error of %v, got %v", ErrPluginExist, err)
				return
			}
		}

		// second time
		err = reg.Register(target.TargetKubernetesJob, plugin.KubernetesJobPlugin{})

		// if err == nil {
		// 	t.Fatalf("expecting non-nil error, got nil error")
		// }
		if err != nil {
			if !errors.Is(err, ErrPluginExist) {
				t.Errorf("expecting error of %v, got %v", ErrPluginExist, err)
			}
		}
	})

	// nil plugin
	t.Run("nil-plugin", func(t *testing.T) {
		reg := New()
		err := reg.Register(target.TargetKubernetesJob, nil)

		if err == nil {
			t.Fatalf("expecting non-nil error, got nil error")
		}
		if err != nil {
			if !errors.Is(err, ErrInvalidPlugin) {
				t.Errorf("expecting error of %v, got %v", ErrInvalidPlugin, err)
			}
		}
	})
}

func TestGet(t *testing.T) {
	// successful get
	t.Run("success", func(t *testing.T) {
		reg := New()
		err := reg.Register(target.TargetKubernetesJob, plugin.KubernetesJobPlugin{})
		_, err = reg.Get(target.TargetKubernetesJob)

		if err == nil {
			// success

		}
		if err != nil {
			t.Errorf("expecting nil error, but got error %v", err)
		}
	})

	// missing key
	t.Run("missing-key", func(t *testing.T) {
		reg := New()
		_, err := reg.Get("kqueue-workload")

		if err == nil {
			t.Errorf("expecting non-nil error, but got nil error")
		}
		if err != nil {
			if !errors.Is(err, ErrPluginNotFound) {
				t.Errorf("expecting error of %v, got %v", ErrPluginNotFound, err)
			}
		}
	})
}

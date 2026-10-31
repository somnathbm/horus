package registry

import (
	"errors"
	"fmt"

	"github.com/somnathbm/horus/internal/plugin"
	"github.com/somnathbm/horus/internal/target"
)

var (
	ErrPluginExists       = errors.New("plugin already exists")
	ErrPluginNotFound    = errors.New("plugin not found")
	ErrInvalidPlugin     = errors.New("invalid plugin")
	ErrInvalidTargetType = errors.New("invalid target type")
)

type Registry struct {
	plugins map[target.TargetType]plugin.Plugin
}

func New() *Registry {
	return &Registry{
		plugins: make(map[target.TargetType]plugin.Plugin),
	}
}

func (r *Registry) Register(key target.TargetType, p plugin.Plugin) error {
	if key == "" {
		return fmt.Errorf("err: %w", ErrInvalidTargetType)
	}

	if p == nil {
		return fmt.Errorf("plugin for target %v is %w", key, ErrInvalidPlugin)
	}

	if _, ok := r.plugins[key]; !ok {
		r.plugins[key] = p
		return nil
	}
	return fmt.Errorf("%w for key %v", ErrPluginExists, key)
}

func (r *Registry) Get(key target.TargetType) (plugin.Plugin, error) {
	if key == "" {
		return nil, fmt.Errorf("err: %w", ErrInvalidTargetType)
	}

	if _, ok := r.plugins[key]; ok {
		return r.plugins[key], nil
	}

	return nil, fmt.Errorf("%w for key %v", ErrPluginNotFound, key)
}

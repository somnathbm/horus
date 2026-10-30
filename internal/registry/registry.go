package registry

import (
	"context"
	"errors"
	"fmt"

	"github.com/somnathbm/horus/internal/plugin"
	"github.com/somnathbm/horus/internal/target"
)

var (
	ErrPluginExist    = errors.New("plugin already exist")
	ErrPluginNotFound = errors.New("plugin not found")
	ErrInvalidPlugin  = errors.New("invalid plugin")
	ErrNoTarget = errors.New("no target found")
)

type Registry struct {
	plugins map[target.TargetType]plugin.Plugin
}

func New() *Registry {
	return &Registry{
		plugins: make(map[target.TargetType]plugin.Plugin),
	}
}

func (r *Registry) Register(ctx context.Context, key target.TargetType, p plugin.Plugin) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("target \"%v\" with err: %w", key, err)
	}

	if p == nil {
		return fmt.Errorf("plugin for target %v is %w", key, ErrInvalidPlugin)
	}

	if _, ok := r.plugins[key]; !ok {
		r.plugins[key] = p
		return nil
	}
	return fmt.Errorf("%w for key %v", ErrPluginExist, key)
}

func (r *Registry) Get(ctx context.Context, key target.TargetType) (plugin.Plugin, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("target \"%v\" with err: %w", key, err)
	}

	if key == "" {
		return nil, fmt.Errorf("err: %w", ErrNoTarget)
	}

	if _, ok := r.plugins[key]; ok {
		return r.plugins[key], nil
	}
	
	return nil, fmt.Errorf("%w for key %v", ErrPluginNotFound, key)
}

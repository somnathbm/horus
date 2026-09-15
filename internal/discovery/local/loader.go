package local

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/somnathbm/horus/internal/target"
)

// sentinel error types
var (
	ErrInvalidPath     = errors.New("valid path is required")
	ErrEmptyDir        = errors.New("empty directory found")
	ErrNoDirAllowed    = errors.New("directory is not allowed")
	ErrInvalidFileType = errors.New("only json type is allowed")
	ErrNoTargetType    = errors.New("no target type found")
)

type RawTarget struct {
	ID   string            `json:"id"`
	Name string            `json:"name"`
	Type target.TargetType `json:"type"`
	Spec json.RawMessage   `json:"spec"`
}

func StrictUnmarshal(data []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

func Load(ctx context.Context, path string) ([]target.Target, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("err: %w", ctx.Err())
	default:
		var targets []target.Target

		// 1. Check for valid path
		if path == "" {
			return nil, fmt.Errorf("err: %w", ErrInvalidPath)
		}

		// 2. Read directory
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("loader err: %w", err)
		}

		if len(entries) == 0 {
			return []target.Target{}, nil
		}

		// 3. read contents of directory
		for _, entry := range entries {
			if entry.IsDir() {
				return nil, fmt.Errorf("loader err: %w", ErrNoDirAllowed)
			}

			// 3.1 if only its a regular file
			if entry.Type().IsRegular() {
				name := entry.Name()
				fileType := filepath.Ext(name)

				if fileType != ".json" {
					return nil, fmt.Errorf("loader err: %w", ErrInvalidFileType)
				}

				// 3.2 read the file
				fileEntryPath := filepath.Join(path, name)
				content, err := os.ReadFile(fileEntryPath)
				if err != nil {
					return nil, fmt.Errorf("loader:read err: %w", err)
				}

				// parse outer structure
				var rawTarget RawTarget
				err = StrictUnmarshal(content, &rawTarget)
				if err != nil {
					return nil, fmt.Errorf("%w", err)
				}

				// parse spec
				decodedTarget, err := decodeTargetSpec(rawTarget)
				if err != nil {
					return nil, fmt.Errorf("%w", err)
				}
				if err = validate(decodedTarget); err != nil {
					return nil, fmt.Errorf("validate target %q: %w", decodedTarget.Name, err)
				}
				targets = append(targets, decodedTarget)
			}
		}
		return targets, nil
	}
}

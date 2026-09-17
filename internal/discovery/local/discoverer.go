package local

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/somnathbm/horus/internal/config"
	"github.com/somnathbm/horus/internal/discovery"
	"github.com/somnathbm/horus/internal/target"
)

// sentinel error types
var (
	ErrEmptyPath       = errors.New("is empty")
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

type LocalDiscoverer struct {
	name string
	path string
}

func New(cfg config.LocalSourceConfig) *LocalDiscoverer {
	return &LocalDiscoverer{
		name: cfg.Name,
		path: cfg.Path,
	}
}

func StrictUnmarshal(data []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

func (ld *LocalDiscoverer) Discover(ctx context.Context) (discovery.DiscoveryResult, error) {
	select {
	case <-ctx.Done():
		return discovery.DiscoveryResult{}, fmt.Errorf("context: %w", ctx.Err())
	default:
		var discvryResult discovery.DiscoveryResult

		// 1. Check for valid path
		if ld.path == "" {
			return discovery.DiscoveryResult{}, fmt.Errorf("%q: %w", ld.path, ErrEmptyPath)
		}

		// 2. Read directory
		entries, err := os.ReadDir(ld.path)
		if err != nil {
			return discovery.DiscoveryResult{}, fmt.Errorf("%q: %w", ld.path, os.ErrPermission)
		}

		// if zero-content directory
		if len(entries) == 0 {
			return discovery.DiscoveryResult{}, nil
		}

		// 3. read contents of directory
		for _, entry := range entries {
			fileName := entry.Name()

			// if the entry is a directory instead of a regular file
			if entry.IsDir() {
				newFailure := discovery.DiscoveryFailure{
					Item: fileName,
					Err:  fmt.Errorf("%q: %w", fileName, ErrNoDirAllowed),
				}
				discvryResult.Failures = append(discvryResult.Failures, newFailure)
				continue
			}

			// 3.1 if only its a regular file
			if entry.Type().IsRegular() {
				fileType := filepath.Ext(fileName)

				if fileType != ".json" {

					newFailure := discovery.DiscoveryFailure{
						Item: fileName,
						Err:  fmt.Errorf("%q: %w", fileName, ErrInvalidFileType),
					}
					discvryResult.Failures = append(discvryResult.Failures, newFailure)
					continue
				}

				// 3.2 read the file
				fileEntryPath := filepath.Join(ld.path, fileName)
				content, err := os.ReadFile(fileEntryPath)
				if err != nil {
					fmt.Println("@@@@ perm error @@")
					newFailure := discovery.DiscoveryFailure{
						Item: fileName,
						Err:  fmt.Errorf("%q: %w", fileName, os.ErrPermission),
					}
					discvryResult.Failures = append(discvryResult.Failures, newFailure)
					continue
				}

				// parse outer structure
				var rawTarget RawTarget
				err = StrictUnmarshal(content, &rawTarget)
				if err != nil {
					newFailure := discovery.DiscoveryFailure{
						Item: fileName,
						Err:  fmt.Errorf("%q: %w", fileName, err),
					}
					discvryResult.Failures = append(discvryResult.Failures, newFailure)
					continue
				}

				// parse spec
				decodedTarget, err := decodeTargetSpec(rawTarget)
				if err != nil {
					newFailure := discovery.DiscoveryFailure{
						Item: fileName,
						Err:  fmt.Errorf("%q: %w", fileName, err),
					}
					discvryResult.Failures = append(discvryResult.Failures, newFailure)
					continue
				}
				if err = validate(decodedTarget); err != nil {
					newFailure := discovery.DiscoveryFailure{
						Item: fileName,
						Err:  fmt.Errorf("%q: %w", fileName, err),
					}
					discvryResult.Failures = append(discvryResult.Failures, newFailure)
					continue
				}
				discvryResult.Targets = append(discvryResult.Targets, decodedTarget)
			}
		}

		return discvryResult, err
	}
}

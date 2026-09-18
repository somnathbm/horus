package local

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/somnathbm/horus/internal/config"
	"github.com/somnathbm/horus/internal/discovery"
	"github.com/somnathbm/horus/internal/target"
)

// sentinel error types
var (
	ErrEmptyPath       = errors.New("is empty")
	ErrNoDirAllowed    = errors.New("directory is not allowed")
	ErrInvalidFileType = errors.New("only json type is allowed")
)

type rawTarget struct {
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

func strictUnmarshal(data []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	// return decoder.Decode(v)

	if err := decoder.Decode(v); err != nil {
		return err // Returns syntax or unknown field errors
	}

	// Ensure no extra valid/invalid JSON or data remains
	var extra struct{}
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("strict decoding: extra data found after object")
		}
		return err
	}
	return nil
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
		entries, readDirErr := os.ReadDir(ld.path)
		if readDirErr != nil {
			return discovery.DiscoveryResult{}, fmt.Errorf("%q: %w", ld.path, readDirErr)
		}

		// if zero-content directory
		if len(entries) == 0 {
			return discovery.DiscoveryResult{}, nil
		}

		// 3. read contents of directory
		for _, entry := range entries {
			fileName := entry.Name()

			// context cancellation check
			if err := ctx.Err(); err != nil {
				return discvryResult, fmt.Errorf("local discovery %q: %w", ld.name, err)
			}

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
				content, readFileErr := os.ReadFile(fileEntryPath)
				if readFileErr != nil {
					newFailure := discovery.DiscoveryFailure{
						Item: fileName,
						Err:  fmt.Errorf("%q: %w", fileName, readFileErr),
					}
					discvryResult.Failures = append(discvryResult.Failures, newFailure)
					continue
				}

				// parse outer structure
				var rawTarget rawTarget
				strictDecodeErr := strictUnmarshal(content, &rawTarget)
				if strictDecodeErr != nil {
					newFailure := discovery.DiscoveryFailure{
						Item: fileName,
						Err:  fmt.Errorf("%q: %w", fileName, strictDecodeErr),
					}
					discvryResult.Failures = append(discvryResult.Failures, newFailure)
					continue
				}

				// parse spec
				decodedTarget, strictDecodeSpecErr := decodeTargetSpec(rawTarget)
				if strictDecodeSpecErr != nil {
					newFailure := discovery.DiscoveryFailure{
						Item: fileName,
						Err:  fmt.Errorf("%q: %w", fileName, strictDecodeSpecErr),
					}
					discvryResult.Failures = append(discvryResult.Failures, newFailure)
					continue
				}
				if validateErr := validate(decodedTarget); validateErr != nil {
					newFailure := discovery.DiscoveryFailure{
						Item: fileName,
						Err:  fmt.Errorf("%q: %w", fileName, validateErr),
					}
					discvryResult.Failures = append(discvryResult.Failures, newFailure)
					continue
				}
				discvryResult.Targets = append(discvryResult.Targets, decodedTarget)
			}
		}

		return discvryResult, nil
	}
}

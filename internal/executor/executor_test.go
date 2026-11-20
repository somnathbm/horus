package executor

import (
	"context"
	"testing"

	"github.com/somnathbm/horus/internal/health"
	"github.com/somnathbm/horus/internal/target"
)

// fakeEvaluator is a mock implementation of the Evaluator interface for testing purposes.
type fakeEvaluator struct {
	status health.WorkloadStatus
	err    error
}

func (f *fakeEvaluator) Evaluate(ctx context.Context, t target.Target) (health.WorkloadStatus, error) {
	return f.status, f.err
}

func TestNewExecutor(t *testing.T) {
	tests := []struct {
		name          string
		workerCount   int
		queueCapacity int
		evaluator     Evaluator
		wantError     error
	}{
		{
			name:          "valid-config",
			workerCount:   5,
			queueCapacity: 10,
			evaluator:     &fakeEvaluator{},
			wantError:     nil,
		},
		{
			name:          "zero-workers",
			workerCount:   0,
			queueCapacity: 10,
			evaluator:     &fakeEvaluator{},
			wantError:     errNoWorkers,
		},
		{
			name:          "negative-queue-capacity",
			workerCount:   5,
			queueCapacity: -1,
			evaluator:     &fakeEvaluator{},
			wantError:     errInsufficientQueueCapacity,
		},
		{
			name:          "nil-evaluator",
			workerCount:   5,
			queueCapacity: 10,
			evaluator:     nil,
			wantError:     errNoEvaluator,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := New(tt.workerCount, tt.queueCapacity, tt.evaluator)
			if (err != nil) != (tt.wantError != nil) {
				t.Errorf("New() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if executor != nil && executor.workerCount != tt.workerCount {
				t.Errorf("New() workerCount = %v, want %v", executor.workerCount, tt.workerCount)
			}
			if executor != nil && executor.queueCapacity != tt.queueCapacity {
				t.Errorf("New() queueCapacity = %v, want %v", executor.queueCapacity, tt.queueCapacity)
			}
			if executor != nil && executor.evaluator != tt.evaluator {
				t.Errorf("New() evaluator = %v, want %v", executor.evaluator, tt.evaluator)
			}
		})
	}
}

// test - all targets processed exactly once
func TestExecutor_ProcessTargets(t *testing.T) {
	// Create a fake evaluator that always returns a healthy status and no error.
	fakeEval := &fakeEvaluator{
		status: health.WorkloadStatus{},
		err:    nil,
	}

	// Create an executor with 3 workers and a queue capacity of 5.
	executor, err := New(3, 5, fakeEval)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	// Create a list of targets to process.
	targets := []target.Target{
		{ID: "target1"},
		{ID: "target2"},
		{ID: "target3"},
		{ID: "target4"},
		{ID: "target5"},
	}

	// process the targets using the executor.
	resultCh := executor.Run(t.Context(), targets)

	// Verify that all targets are processed exactly once and the results are as expected.
	processedTargets := make(map[string]bool)
	for result := range resultCh {
		if _, exists := processedTargets[result.TargetID]; exists {
			t.Errorf("Target %s processed more than once", result.TargetID)
		}
		processedTargets[result.TargetID] = true

		if result.Err != nil {
			t.Errorf("Error processing target %s: %v", result.TargetID, result.Err)
		}
	}

	if len(processedTargets) != len(targets) {
		t.Errorf("Not all targets were processed. Expected %d, got %d", len(targets), len(processedTargets))
	}
}

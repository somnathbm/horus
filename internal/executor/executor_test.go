package executor

import (
	"context"
	"errors"
	"testing"

	"github.com/somnathbm/horus/internal/health"
	"github.com/somnathbm/horus/internal/target"
)

// a fake evaluator
type fakeEvaluator struct {
	calls  int
	status health.WorkloadStatus
	err    error
}

func (fe *fakeEvaluator) Evaluate(ctx context.Context, t target.Target) (health.WorkloadStatus, error) {
	fe.calls++
	return fe.status, fe.err
}

// test - executor constructor
func TestExecutorConstructor(t *testing.T) {
	// define test cases
	tests := []struct {
		name          string
		workerCount   int
		queueCapacity int
		evaluator     Evaluator
		wantError     error
	}{
		{
			name:          "valid-constructor",
			workerCount:   3,
			queueCapacity: 5,
			evaluator:     &fakeEvaluator{},
			wantError:     nil,
		},
		{
			name:          "zero-worker",
			workerCount:   0,
			queueCapacity: 5,
			evaluator:     &fakeEvaluator{},
			wantError:     errNoWorkers,
		},
		{
			name:          "zero-queue-capacity",
			workerCount:   5,
			queueCapacity: 0,
			evaluator:     &fakeEvaluator{},
			wantError:     nil,
		},
		{
			name:          "insufficient-queue-capacity",
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			executr, err := New(tc.workerCount, tc.queueCapacity, tc.evaluator)
			// assertions
			if tc.wantError != nil {
				// assert - sentinel errors
				if !errors.Is(err, tc.wantError) {
					t.Errorf("New(): expecting error to be %v, got %v", tc.wantError, err)
				}
				// assert - executr is nil
				if executr != nil {
					t.Fatalf("expecting executor to be nil, but got %v", executr)
				}
			} else {
				// assert - error is nil
				if err != nil {
					t.Fatalf("expecting nil error, but got %v", err)
				}
				// assert executor
				if executr == nil {
					t.Fatalf("expecting executor to be non-nil, but got nil")
				}
				// assert - worker count
				if executr.workerCount != tc.workerCount {
					t.Errorf("expecting worker count to be %v, got %v", tc.workerCount, executr.workerCount)
				}
				// assert - queue capacity
				if executr.queueCapacity != tc.queueCapacity {
					t.Errorf("expecting queue capacity to be %v, got %v", tc.queueCapacity, executr.queueCapacity)
				}
				// assert - evaluator
				if executr.evaluator != tc.evaluator {
					t.Errorf("expecting evaluator to be %v, got %v", tc.evaluator, executr.evaluator)
				}
			}
		})
	}
}

package executor

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/somnathbm/horus/internal/health"
	"github.com/somnathbm/horus/internal/target"
)

// a fake evaluator
type fakeEvaluator struct {
	mu     sync.Mutex
	calls  map[string]int
	status health.WorkloadStatus
	err    error
}

func (fe *fakeEvaluator) Evaluate(ctx context.Context, t target.Target) (health.WorkloadStatus, error) {
	defer fe.mu.Unlock()
	fe.mu.Lock()
	fe.calls[t.ID]++
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

// test - all targets processed exactly once
func TestExecutor_targetProcessedOnce(t *testing.T) {
	// sample targets
	targets := []target.Target{
		{ID: "target1"},
		{ID: "target2"},
		{ID: "target3"},
		{ID: "target4"},
		{ID: "target5"},
	}

	// constructor
	fe := &fakeEvaluator{calls: make(map[string]int, cap(targets))}
	excutr, err := New(5, 3, fe)
	if err != nil {
		t.Fatalf("expecting nil error, but got %v", err)
	}
	if excutr == nil {
		t.Fatalf("expecting non-nil executor, but got nil")
	}
	// run
	resultCh := excutr.Run(t.Context(), targets)
	var targetMap map[string]bool = make(map[string]bool, cap(targets))
	for _, target := range targets {
		targetMap[target.ID] = true
	}

	for result := range resultCh {
		// assert if error
		if result.Err != nil {
			t.Errorf("encounted error %v", result.Err)
		}
		// assert if result ID belongs to target list
		if _, ok := targetMap[result.TargetID]; !ok {
			t.Errorf("result ID to does not exist in the target list")
		}
		// assert if result status matches with evaluator status
		if !reflect.DeepEqual(fe.status, result.Status) {
			t.Errorf("expecting status to be %v, got %v", fe.status, result.Status)
		}
		// assert if calls are exactly once
		if fe.calls[result.TargetID] != 1 {
			t.Errorf("expecting target to be processed once, got %v", fe.calls)
		}
	}

}

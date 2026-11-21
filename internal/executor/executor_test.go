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
	mu     sync.RWMutex
	calls  map[string]int
	status health.WorkloadStatus
	err    error
}

func (fe *fakeEvaluator) CallCount(targetID string) int {
	defer fe.mu.RUnlock()
	fe.mu.RLock()
	return fe.calls[targetID]
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
	fe := &fakeEvaluator{
		calls: make(map[string]int, cap(targets)),
		status: health.WorkloadStatus{
			State:  health.StateUnknown,
			Reason: "target inspection not implemented",
		},
	}
	excutr, err := New(5, 3, fe)
	if err != nil {
		t.Fatalf("expecting nil error, but got %v", err)
	}
	if excutr == nil {
		t.Fatalf("expecting non-nil executor, but got nil")
	}
	// run
	resultCh := excutr.Run(t.Context(), targets)
	var expectedMap map[string]bool = make(map[string]bool, cap(targets))
	var seenMap map[string]int = make(map[string]int, cap(targets))
	for _, target := range targets {
		expectedMap[target.ID] = true
	}

	for result := range resultCh {
		// assert if error
		if result.Err != nil {
			t.Errorf("encounted error %v", result.Err)
		}
		// assert - result.target exists in expected map
		if _, ok := expectedMap[result.TargetID]; !ok {
			t.Errorf("target %q from result is not expected", result.TargetID)
		}
		seenMap[result.TargetID]++
		// assert - seen target ID count exactly 1
		if seenMap[result.TargetID] > 1 {
			t.Errorf("expecting target %q to be processed only 1, got %v", result.TargetID, seenMap[result.TargetID])
		}

		// assert - result status matches with evaluator status
		if !reflect.DeepEqual(fe.status, result.Status) {
			t.Errorf("expecting status to be %v, got %v", fe.status, result.Status)
		}
	}

	// after resultCh is closed
	for k, v := range expectedMap {
		if seenMap[k] > 1 {
			t.Errorf("expecting target %q to be processed only 1, got %v", k, v)
		}
		// assert - target is called exactly once
		callCount := fe.CallCount(k)
		if callCount != 1 {
			t.Errorf("expecting target %q to be processed once, got %v", k, callCount)
		}
	}
}

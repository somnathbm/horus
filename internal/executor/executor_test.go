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

// local sentinel
var (
	errFakeEvaluation = errors.New("fake evaluation failure")
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
func (fe *fakeEvaluator) TotalCallCount() int {
	defer fe.mu.RUnlock()
	fe.mu.RLock()

	totalCount := 0
	for _, v := range fe.calls {
		totalCount += v
	}
	return totalCount
}

func (fe *fakeEvaluator) Evaluate(ctx context.Context, t target.Target) (health.WorkloadStatus, error) {
	defer fe.mu.Unlock()
	fe.mu.Lock()
	fe.calls[t.ID]++
	return fe.status, fe.err
}

// blocking evaluator
type blockingEvaluator struct {
	mu      sync.RWMutex
	calls   map[string]int
	started chan string
}

func (be *blockingEvaluator) TotalCallCount() int {
	defer be.mu.RUnlock()
	be.mu.RLock()

	totalCount := 0
	for _, v := range be.calls {
		totalCount += v
	}
	return totalCount
}

func (be *blockingEvaluator) Evaluate(ctx context.Context, t target.Target) (health.WorkloadStatus, error) {
	select {
	case be.started <- t.ID:
		be.mu.Lock()
		be.calls[t.ID]++
		be.mu.Unlock()
	case <-ctx.Done():
		return health.WorkloadStatus{}, ctx.Err()
	}
	<-ctx.Done()
	return health.WorkloadStatus{}, ctx.Err()
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
	for k := range expectedMap {
		// assert - not processed or duplicate processed target
		if seenMap[k] != 1 {
			t.Errorf("expecting exactly one result from target %q, got %d", k, seenMap[k])
		}
		// assert - target is called exactly once
		callCount := fe.CallCount(k)
		if callCount != 1 {
			t.Errorf("expecting target %q to be processed once, got %v", k, callCount)
		}
	}
}

// test - evaluator error propagation
func TestExecutor_evaluatorError(t *testing.T) {
	// sample targets
	targets := []target.Target{
		{ID: "target123"},
	}

	// constructor
	fe := &fakeEvaluator{
		calls: make(map[string]int, cap(targets)),
		status: health.WorkloadStatus{
			TargetID: "target123",
			State:    health.StateUnknown,
			Reason:   "partial observation",
		},
		err: errFakeEvaluation,
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

	// collect results
	for result := range resultCh {
		// assert if error
		if !errors.Is(result.Err, errFakeEvaluation) {
			t.Errorf("expecting error to be %v, got %v", errFakeEvaluation, result.Err)
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
	for k := range expectedMap {
		// assert - not processed or duplicate processed target
		if seenMap[k] != 1 {
			t.Errorf("expecting exactly one result from target %q, got %d", k, seenMap[k])
		}
		// assert - target is called exactly once
		callCount := fe.CallCount(k)
		if callCount != 1 {
			t.Errorf("expecting target %q to be processed once, got %v", k, callCount)
		}
	}
}

// test - pre-cancelled context
func TestExecutor_preCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	// sample targets
	targets := []target.Target{
		{ID: "target123"},
		{ID: "target456"},
	}

	// constructor
	fe := &fakeEvaluator{
		calls: make(map[string]int, cap(targets)),
	}
	excutr, err := New(5, 3, fe)
	if err != nil {
		t.Fatalf("expecting nil error, but got %v", err)
	}
	if excutr == nil {
		t.Fatalf("expecting non-nil executor, but got nil")
	}

	// cancel the context right before Run()
	cancel()

	// run
	resultCh := excutr.Run(ctx, targets)

	// assert - closed channel
	select {
	case _, ok := <-resultCh:
		if ok {
			t.Fatalf("expecting closed result channel")
		}
	default:
		t.Fatalf("expected result channel to be already closed")
	}

	// assert - no evaluator calls
	for _, target := range targets {
		calls := fe.CallCount(target.ID)
		if calls != 0 {
			t.Errorf("expecting evaluator calls be zero, got %d", calls)
		}
	}
}

// test - zero targets
func TestExecutor_zeroTargets(t *testing.T) {
	// empty targets
	targets := []target.Target{}

	// constructor
	fe := &fakeEvaluator{
		calls: make(map[string]int, cap(targets)),
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

	// collect result
	var resultCount int
	for range resultCh {
		resultCount++
	}

	// assert - no result should be there
	if resultCount != 0 {
		t.Errorf("expected no result, got %d", resultCount)
	}

	// assert - evaluator calls count
	totalCalls := fe.TotalCallCount()
	if totalCalls != 0 {
		t.Errorf("expected zero calls, got %d", totalCalls)
	}
}

// test - mid-run cancellation
/*
Run starts normally
→ some targets begin evaluation
→ context is cancelled while work is still in flight
→ producer stops enqueueing
→ workers stop
→ result channel eventually closes
→ not all targets are evaluated
*/
func TestExecutor_midRunCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	// empty targets
	targets := []target.Target{
		{ID: "target01"},
		{ID: "target02"},
		{ID: "target03"},
		{ID: "target04"},
		{ID: "target05"},
		{ID: "target06"},
		{ID: "target07"},
		{ID: "target08"},
		{ID: "target09"},
		{ID: "target10"},
		{ID: "target11"},
		{ID: "target12"},
		{ID: "target13"},
		{ID: "target14"},
		{ID: "target15"},
		{ID: "target16"},
		{ID: "target17"},
		{ID: "target18"},
		{ID: "target19"},
		{ID: "target20"},
	}

	// constructor
	be := &blockingEvaluator{
		calls:   make(map[string]int, cap(targets)),
		started: make(chan string),
	}

	excutr, err := New(3, 5, be)
	if err != nil {
		t.Fatalf("expecting nil error, but got %v", err)
	}
	if excutr == nil {
		t.Fatalf("expecting non-nil executor, but got nil")
	}

	// run
	resultCh := excutr.Run(ctx, targets)

	// wait until 3 evaluation has started
	for range 3 {
		<-be.started
	}
	cancel()

	// drain resultCh until it eventually closes
	resultCount := 0
	for range resultCh {
		resultCount++
	}

	// assertion - evaluator calls count
	callsCount := be.TotalCallCount()
	if callsCount != 3 {
		t.Errorf("expecting 3 calls, got %d", callsCount)
	}

	// assertion - zero result emitted
	if resultCount != 0 {
		t.Errorf("expecting 0 results, got %d", resultCount)
	}

}

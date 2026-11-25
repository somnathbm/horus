package executor

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/somnathbm/horus/internal/health"
	"github.com/somnathbm/horus/internal/target"
)

var (
	errNoWorkers                 = errors.New("no available workers")
	errInsufficientQueueCapacity = errors.New("insufficient queue capacity")
	errNoEvaluator               = errors.New("no target evaluator")
)

// evaluation result envelope
type Result struct {
	TargetID string
	Status   health.WorkloadStatus
	Err      error
}

// an evaluator interface that will actually evaluate a target
type Evaluator interface {
	Evaluate(ctx context.Context, t target.Target) (health.WorkloadStatus, error)
}

// =================== Executor ================
// executor type
type Executor struct {
	workerCount   int
	queueCapacity int
	evaluator     Evaluator
}

// constructor
func New(wc int, qc int, ev Evaluator) (*Executor, error) {
	if wc <= 0 {
		return nil, fmt.Errorf("config error: %w", errNoWorkers)
	}
	if qc < 0 {
		return nil, fmt.Errorf("config error: %w", errInsufficientQueueCapacity)
	}
	if ev == nil {
		return nil, fmt.Errorf("config error: %w", errNoEvaluator)
	}

	return &Executor{
		workerCount:   wc,
		queueCapacity: qc,
		evaluator:     ev,
	}, nil
}

// worker function
func worker(ctx context.Context, eval Evaluator, targetQueue <-chan target.Target, resultCh chan<- Result) {
	for {
		select {
		case <-ctx.Done():
			return
		case target, ok := <-targetQueue:
			if !ok {
				return
			}
			status, err := eval.Evaluate(ctx, target)
			// for in-flight requests, result must not ommit
			if err := ctx.Err(); err != nil {
				return
			}
			// construct result
			result := Result{
				TargetID: target.ID,
				Status:   status,
				Err:      err,
			}

			// protects results from forever blocking
			select {
			case <-ctx.Done():
				return
			case resultCh <- result:
			}
		}
	}
}

// workerpool orchestrator
func (ex *Executor) Run(ctx context.Context, targets []target.Target) <-chan Result {
	// result channel
	resultCh := make(chan Result)

	// if context is cancelled even before workers are called
	if err := ctx.Err(); err != nil {
		close(resultCh)
		return resultCh
	}

	// queue and worker group
	targetQueue := make(chan target.Target, ex.queueCapacity)
	var wg sync.WaitGroup

	// producer
	go func() {
		defer close(targetQueue)
		for _, target := range targets {
			select {
			case <-ctx.Done():
				return
			case targetQueue <- target:
			}
		}
	}()

	// consumer
	for range ex.workerCount {
		// spin off worker
		wg.Go(func() {
			worker(ctx, ex.evaluator, targetQueue, resultCh)
		})
	}

	// co-ordinator
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// process
	return resultCh
}

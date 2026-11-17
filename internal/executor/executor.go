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
	targetID string
	status   health.WorkloadStatus
	err      error
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
	if qc <= 0 {
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

// process target within each worker
func processTarget(ctx context.Context, eval Evaluator, t target.Target) (health.WorkloadStatus, error) {
	if err := ctx.Err(); err != nil {
		return health.WorkloadStatus{}, err
	}

	return eval.Evaluate(ctx, t)
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
			status, err := processTarget(ctx, eval, target)
			if err != nil {
				resultCh <- Result{
					targetID: target.ID,
					status:   health.WorkloadStatus{},
					err:      err,
				}
			}

			// send result
			resultCh <- Result{
				targetID: target.ID,
				status:   status,
				err:      nil,
			}
		}
	}
}

// workerpool orchestrator
func (ex *Executor) Run(ctx context.Context, targets []target.Target) (chan<- Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context: %w", err)
	}

	// instatiate channels
	targetQueue := make(chan target.Target, ex.queueCapacity)
	resultCh := make(chan<- Result)
	var wg sync.WaitGroup

	// producer
	producerFunc := func() {
		go func() {
			defer close(targetQueue)
			for i := 1; i <= len(targets); i++ {
				targetQueue <- targets[i]
			}
		}()
	}
	producerFunc()

	// consumer
	consumerFunc := func() {
		for i := 1; i <= ex.workerCount; i++ {
			// spin off worker
			wg.Go(func() {
				worker(ctx, ex.evaluator, targetQueue, resultCh)
			})
		}
	}
	consumerFunc()

	// co-ordinator
	coordinatorFunc := func() {
		go func() {
			wg.Wait()
			close(resultCh)
		}()
	}
	coordinatorFunc()

	// process
	return resultCh, nil
}

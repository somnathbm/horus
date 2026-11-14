package executor

import (
	"context"
	"errors"
	"fmt"

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

// executor
type Executor struct {
	workerCount   int
	queueCapacity int
	evaluator     Evaluator
}

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

func (ex *Executor) Run(ctx context.Context, targets []target.Target) <-chan Result {
	// instatiate channels
	// targetStream := make(chan<- target.Target, ex.queueCapacity)
	resultStream := make(<-chan Result)

	// process
	return resultStream
}

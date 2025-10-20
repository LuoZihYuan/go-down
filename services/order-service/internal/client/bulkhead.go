package client

import (
	"context"
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	bulkheadActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bulkhead_active",
			Help: "Current number of active requests in bulkhead",
		},
		[]string{"pool"},
	)

	bulkheadRejected = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bulkhead_rejected_total",
			Help: "Total number of requests rejected by bulkhead",
		},
		[]string{"pool"},
	)
)

var (
	ErrBulkheadFull = errors.New("bulkhead is full")
)

// Bulkhead implements the bulkhead pattern using semaphores
type Bulkhead struct {
	semaphore chan struct{}
	poolName  string
}

// NewBulkhead creates a new bulkhead with the specified capacity
func NewBulkhead(poolName string, maxConcurrent int) *Bulkhead {
	return &Bulkhead{
		semaphore: make(chan struct{}, maxConcurrent),
		poolName:  poolName,
	}
}

// Execute runs the provided function with bulkhead protection
// Returns ErrBulkheadFull if the bulkhead is at capacity
func (b *Bulkhead) Execute(ctx context.Context, fn func() error) error {
	// Try to acquire semaphore
	select {
	case b.semaphore <- struct{}{}:
		// Acquired - track active requests
		bulkheadActive.WithLabelValues(b.poolName).Inc()
		defer func() {
			<-b.semaphore
			bulkheadActive.WithLabelValues(b.poolName).Dec()
		}()

		// Execute the function
		return fn()

	case <-ctx.Done():
		// Context cancelled
		return ctx.Err()

	default:
		// Bulkhead is full
		bulkheadRejected.WithLabelValues(b.poolName).Inc()
		return ErrBulkheadFull
	}
}

// TryExecute attempts to execute without blocking
// Returns ErrBulkheadFull immediately if at capacity
func (b *Bulkhead) TryExecute(fn func() error) error {
	select {
	case b.semaphore <- struct{}{}:
		bulkheadActive.WithLabelValues(b.poolName).Inc()
		defer func() {
			<-b.semaphore
			bulkheadActive.WithLabelValues(b.poolName).Dec()
		}()
		return fn()

	default:
		bulkheadRejected.WithLabelValues(b.poolName).Inc()
		return ErrBulkheadFull
	}
}

// GetActiveCount returns the current number of active requests
func (b *Bulkhead) GetActiveCount() int {
	return len(b.semaphore)
}

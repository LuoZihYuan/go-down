package client

import (
	"errors"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

var (
	circuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"service"},
	)

	circuitBreakerFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_failures_total",
			Help: "Total number of circuit breaker failures",
		},
		[]string{"service"},
	)
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitBreaker implements the circuit breaker pattern with generics and sliding window
type CircuitBreaker[T any] struct {
	state             CircuitState
	failureTimestamps []time.Time
	failureThreshold  int
	failureWindow     time.Duration
	successThreshold  int
	timeout           time.Duration
	lastFailureTime   time.Time
	mu                sync.RWMutex
	serviceName       string
}

// NewCircuitBreaker creates a new circuit breaker with sliding window
// failureThreshold: number of failures in the window before opening (e.g., 5)
// timeout: how long to wait before half-open (e.g., 30s)
func NewCircuitBreaker[T any](serviceName string, failureThreshold int, timeout time.Duration) *CircuitBreaker[T] {
	return &CircuitBreaker[T]{
		state:             StateClosed,
		failureTimestamps: make([]time.Time, 0),
		failureThreshold:  failureThreshold,
		failureWindow:     10 * time.Second, // 10 second sliding window
		successThreshold:  1,                // One success in half-open moves to closed
		timeout:           timeout,
		serviceName:       serviceName,
	}
}

// Execute runs the provided function with circuit breaker protection
func (cb *CircuitBreaker[T]) Execute(fn func() (T, error)) (T, error) {
	var zero T

	// Check if circuit is open
	if !cb.canAttempt() {
		circuitBreakerFailures.WithLabelValues(cb.serviceName).Inc()
		return zero, ErrCircuitOpen
	}

	// Execute the function
	result, err := fn()

	// Record result
	if err != nil {
		cb.recordFailure()
		return zero, err
	}

	cb.recordSuccess()
	return result, nil
}

// canAttempt checks if a request can be attempted
func (cb *CircuitBreaker[T]) canAttempt() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		// Check if timeout has elapsed
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.setState(StateHalfOpen)
			return true
		}
		return false

	case StateHalfOpen:
		return true

	default:
		return false
	}
}

// recordFailure records a failed attempt with timestamp
func (cb *CircuitBreaker[T]) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	cb.failureTimestamps = append(cb.failureTimestamps, now)
	cb.lastFailureTime = now

	circuitBreakerFailures.WithLabelValues(cb.serviceName).Inc()

	// Clean up old failures outside the window
	cb.cleanupOldFailures(now)

	// Count failures in the current window
	failuresInWindow := cb.countFailuresInWindow(now)

	switch cb.state {
	case StateClosed:
		if failuresInWindow >= cb.failureThreshold {
			cb.setState(StateOpen)
		}

	case StateHalfOpen:
		// Any failure in half-open state reopens the circuit
		cb.setState(StateOpen)
	}
}

// recordSuccess records a successful attempt
func (cb *CircuitBreaker[T]) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateHalfOpen:
		// Success in half-open moves to closed
		cb.failureTimestamps = make([]time.Time, 0) // Reset failure history
		cb.setState(StateClosed)

	case StateClosed:
		// Success in closed state - no action needed
		// Don't reset failures as we use sliding window
	}
}

// countFailuresInWindow counts failures within the sliding window
func (cb *CircuitBreaker[T]) countFailuresInWindow(now time.Time) int {
	windowStart := now.Add(-cb.failureWindow)
	count := 0
	for _, ts := range cb.failureTimestamps {
		if ts.After(windowStart) {
			count++
		}
	}
	return count
}

// cleanupOldFailures removes failure timestamps outside the sliding window
func (cb *CircuitBreaker[T]) cleanupOldFailures(now time.Time) {
	windowStart := now.Add(-cb.failureWindow)
	validFailures := make([]time.Time, 0)
	for _, ts := range cb.failureTimestamps {
		if ts.After(windowStart) {
			validFailures = append(validFailures, ts)
		}
	}
	cb.failureTimestamps = validFailures
}

// setState updates the circuit breaker state and metrics
func (cb *CircuitBreaker[T]) setState(state CircuitState) {
	cb.state = state
	circuitBreakerState.WithLabelValues(cb.serviceName).Set(float64(state))
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker[T]) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

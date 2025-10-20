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

// CircuitBreaker implements the circuit breaker pattern with generics
type CircuitBreaker[T any] struct {
	state            CircuitState
	failureCount     int
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	lastFailureTime  time.Time
	mu               sync.RWMutex
	serviceName      string
}

// NewCircuitBreaker creates a new circuit breaker
// failureThreshold: number of failures before opening
// timeout: how long to wait before half-open
func NewCircuitBreaker[T any](serviceName string, failureThreshold int, timeout time.Duration) *CircuitBreaker[T] {
	return &CircuitBreaker[T]{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: 1, // One success in half-open moves to closed
		timeout:          timeout,
		serviceName:      serviceName,
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

// recordFailure records a failed attempt
func (cb *CircuitBreaker[T]) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	circuitBreakerFailures.WithLabelValues(cb.serviceName).Inc()

	switch cb.state {
	case StateClosed:
		if cb.failureCount >= cb.failureThreshold {
			cb.setState(StateOpen)
		}

	case StateHalfOpen:
		cb.setState(StateOpen)
	}
}

// recordSuccess records a successful attempt
func (cb *CircuitBreaker[T]) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateHalfOpen:
		cb.failureCount = 0
		cb.setState(StateClosed)

	case StateClosed:
		cb.failureCount = 0
	}
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

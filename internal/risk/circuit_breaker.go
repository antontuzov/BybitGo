package risk

import (
	"sync"
	"time"
)

// CircuitBreaker implements the circuit breaker pattern for API calls
type CircuitBreaker struct {
	mutex            sync.RWMutex
	state            string // "closed", "open", "half-open"
	failureCount     int
	lastFailure      time.Time
	timeout          time.Duration
	failureThreshold int
}

// NewCircuitBreaker creates a new CircuitBreaker
func NewCircuitBreaker(timeout time.Duration, failureThreshold int) *CircuitBreaker {
	return &CircuitBreaker{
		state:            "closed",
		failureCount:     0,
		timeout:          timeout,
		failureThreshold: failureThreshold,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Check if circuit is open
	if cb.state == "open" {
		// Check if timeout has passed
		if time.Since(cb.lastFailure) > cb.timeout {
			// Move to half-open state
			cb.state = "half-open"
		} else {
			return &CircuitBreakerOpenError{}
		}
	}

	// Execute the function
	err := fn()

	// Handle result based on current state
	if cb.state == "half-open" {
		if err != nil {
			// Failed again, open circuit
			cb.state = "open"
			cb.lastFailure = time.Now()
			return err
		} else {
			// Success, close circuit
			cb.state = "closed"
			cb.failureCount = 0
			return nil
		}
	}

	// Handle result in closed state
	if err != nil {
		cb.failureCount++
		cb.lastFailure = time.Now()

		// Check if we should open the circuit
		if cb.failureCount >= cb.failureThreshold {
			cb.state = "open"
		}

		return err
	} else {
		// Success, reset failure count
		cb.failureCount = 0
		return nil
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() string {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// CircuitBreakerOpenError represents an error when the circuit breaker is open
type CircuitBreakerOpenError struct{}

func (e *CircuitBreakerOpenError) Error() string {
	return "circuit breaker is open"
}

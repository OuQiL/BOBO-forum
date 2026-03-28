package repository

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_New(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{})
	if cb == nil {
		t.Fatal("Expected circuit breaker to be created")
	}
	if cb.State() != StateClosed {
		t.Errorf("Expected initial state to be closed, got %s", cb.State())
	}
}

func TestCircuitBreaker_Allow(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{})

	if !cb.Allow() {
		t.Error("Expected Allow() to return true when circuit is closed")
	}
}

func TestCircuitBreaker_OpenAfterFailures(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 3,
	})

	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.State() != StateOpen {
		t.Errorf("Expected state to be open after %d failures, got %s", 3, cb.State())
	}

	if cb.Allow() {
		t.Error("Expected Allow() to return false when circuit is open")
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 2,
		Timeout:          100 * time.Millisecond,
	})

	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Fatalf("Expected state to be open, got %s", cb.State())
	}

	time.Sleep(150 * time.Millisecond)

	if !cb.Allow() {
		t.Error("Expected Allow() to return true after timeout (half-open state)")
	}

	if cb.State() != StateHalfOpen {
		t.Errorf("Expected state to be half-open, got %s", cb.State())
	}
}

func TestCircuitBreaker_CloseAfterSuccessInHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          100 * time.Millisecond,
	})

	cb.RecordFailure()
	cb.RecordFailure()

	time.Sleep(150 * time.Millisecond)

	cb.Allow()
	cb.RecordSuccess()
	cb.RecordSuccess()

	if cb.State() != StateClosed {
		t.Errorf("Expected state to be closed after successes in half-open, got %s", cb.State())
	}
}

func TestCircuitBreaker_OpenAfterFailureInHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 2,
		Timeout:          100 * time.Millisecond,
	})

	cb.RecordFailure()
	cb.RecordFailure()

	time.Sleep(150 * time.Millisecond)

	cb.Allow()
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Errorf("Expected state to be open after failure in half-open, got %s", cb.State())
	}
}

func TestCircuitBreaker_Call(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 3,
	})

	err := cb.Call(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	callErr := errors.New("call error")
	err = cb.Call(func() error {
		return callErr
	})

	if err == nil {
		t.Error("Expected error from call")
	}
}

func TestCircuitBreaker_CallOpenCircuit(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 2,
	})

	cb.RecordFailure()
	cb.RecordFailure()

	err := cb.Call(func() error {
		return nil
	})

	if err != ErrCircuitOpen {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 2,
	})

	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Fatalf("Expected state to be open, got %s", cb.State())
	}

	cb.Reset()

	if cb.State() != StateClosed {
		t.Errorf("Expected state to be closed after reset, got %s", cb.State())
	}
}

func TestCircuitBreaker_Stats(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 10,
	})

	cb.RecordSuccess()
	cb.RecordSuccess()
	cb.RecordFailure()

	stats := cb.Stats()

	if stats.Successes != 2 {
		t.Errorf("Expected 2 successes, got %d", stats.Successes)
	}

	if stats.Failures != 1 {
		t.Errorf("Expected 1 failure, got %d", stats.Failures)
	}

	if stats.State != StateClosed {
		t.Errorf("Expected state to be closed, got %s", stats.State)
	}
}

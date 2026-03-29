package repository

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrCircuitOpen     = errors.New("circuit breaker is open")
	ErrTooManyRequests = errors.New("too many requests")
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

type CircuitBreakerConfig struct {
	FailureThreshold int           `json:",default=5"`
	SuccessThreshold int           `json:",default=3"`
	Timeout          time.Duration `json:",default=30s"`
	MaxHalfOpenReqs  int           `json:",default=1"`
	FailureRate      float64       `json:",default=0.5"`
}

type CircuitBreaker struct {
	mu           sync.RWMutex
	state        State
	failures     int
	successes    int
	requests     int
	lastFailTime time.Time
	halfOpenReqs int
	config       CircuitBreakerConfig
}

func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 5
	}
	if config.SuccessThreshold == 0 {
		config.SuccessThreshold = 3
	}
	if config.MaxHalfOpenReqs == 0 {
		config.MaxHalfOpenReqs = 1
	}
	if config.FailureRate == 0 {
		config.FailureRate = 0.5
	}

	return &CircuitBreaker{
		state:  StateClosed,
		config: config,
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailTime) > cb.config.Timeout {
			cb.state = StateHalfOpen
			cb.halfOpenReqs = 0
			cb.successes = 0
			return true
		}
		return false
	case StateHalfOpen:
		if cb.halfOpenReqs < cb.config.MaxHalfOpenReqs {
			cb.halfOpenReqs++
			return true
		}
		return false
	}
	return false
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.successes++
		cb.failures = 0
		cb.requests = 0
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
			cb.requests = 0
		}
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailTime = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failures++
		cb.requests++
		if cb.failures >= cb.config.FailureThreshold {
			cb.state = StateOpen
			return
		}
		if cb.requests >= cb.config.FailureThreshold && float64(cb.failures)/float64(cb.requests) >= cb.config.FailureRate {
			cb.state = StateOpen
		}
	case StateHalfOpen:
		cb.failures++
		cb.state = StateOpen
		cb.halfOpenReqs = 0
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	if !cb.Allow() {
		return ErrCircuitOpen
	}

	err := fn()
	if err != nil {
		cb.RecordFailure()
		return err
	}

	cb.RecordSuccess()
	return nil
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.requests = 0
	cb.halfOpenReqs = 0
}

func (cb *CircuitBreaker) Stats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return CircuitBreakerStats{
		State:        cb.state,
		Failures:     cb.failures,
		Successes:    cb.successes,
		Requests:     cb.requests,
		HalfOpenReqs: cb.halfOpenReqs,
	}
}

type CircuitBreakerStats struct {
	State        State
	Failures     int
	Successes    int
	Requests     int
	HalfOpenReqs int
}

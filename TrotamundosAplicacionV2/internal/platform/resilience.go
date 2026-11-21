package platform

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
	"trotamundos/internal/domain"
)

type ProviderState string

const (
	ProviderHealthy    ProviderState = "HEALTHY"
	ProviderDegraded   ProviderState = "DEGRADED"
	ProviderRecovering ProviderState = "RECOVERING"

	StateClosed   ProviderState = ProviderHealthy
	StateOpen     ProviderState = ProviderDegraded
	StateHalfOpen ProviderState = ProviderRecovering
)

// ProviderCircuitBreaker isolates transient external provider failures transparently
// using bounded exponential backoff retries and automatic recovery windows.
type ProviderCircuitBreaker struct {
	mu               sync.Mutex
	state            ProviderState
	window           []bool
	windowSize       int
	failureThreshold float64
	openUntil        time.Time
	cooldown         time.Duration
	maxAttempts      int
	baseBackoff      time.Duration
}

// ProviderGuard is an alias for ProviderCircuitBreaker.
type ProviderGuard = ProviderCircuitBreaker

func NewProviderCircuitBreaker() *ProviderCircuitBreaker {
	return &ProviderCircuitBreaker{
		state:            ProviderHealthy,
		windowSize:       20,
		failureThreshold: 0.50,
		cooldown:         20 * time.Second,
		maxAttempts:      3,
		baseBackoff:      80 * time.Millisecond,
	}
}

func NewProviderGuard(cooldown, baseBackoff time.Duration) *ProviderCircuitBreaker {
	cb := NewProviderCircuitBreaker()
	if cooldown > 0 {
		cb.cooldown = cooldown
	}
	if baseBackoff > 0 {
		cb.baseBackoff = baseBackoff
	}
	return cb
}

func (cb *ProviderCircuitBreaker) State() ProviderState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state == ProviderDegraded && time.Now().After(cb.openUntil) {
		cb.state = ProviderRecovering
	}
	return cb.state
}

func (cb *ProviderCircuitBreaker) FailureRatePercent() float64 {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if len(cb.window) == 0 {
		return 0
	}
	failures := 0
	for _, ok := range cb.window {
		if !ok {
			failures++
		}
	}
	return (float64(failures) / float64(len(cb.window))) * 100.0
}

func (cb *ProviderCircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = ProviderHealthy
	cb.window = nil
	cb.openUntil = time.Time{}
}

// Execute runs an external operation with automatic retries and transparent circuit protection.
func (cb *ProviderCircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	now := time.Now()
	if cb.state == ProviderDegraded {
		if now.Before(cb.openUntil) {
			cb.mu.Unlock()
			return fmt.Errorf("%w: reintento programado automáticamente", domain.ErrProviderUnavailable)
		}
		cb.state = ProviderRecovering
	}
	attempts := cb.maxAttempts
	baseBackoff := cb.baseBackoff
	cb.mu.Unlock()

	var lastErr error
	for i := 0; i < attempts; i++ {
		lastErr = fn()
		if lastErr == nil {
			cb.record(true)
			return nil
		}
		if i < attempts-1 {
			jitter := 0.85 + rand.Float64()*0.30
			delay := time.Duration(float64(baseBackoff) * float64(int(1)<<i) * jitter)
			time.Sleep(delay)
		}
	}

	cb.record(false)
	return lastErr
}

func (cb *ProviderCircuitBreaker) record(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.window = append(cb.window, success)
	if len(cb.window) > cb.windowSize {
		cb.window = cb.window[len(cb.window)-cb.windowSize:]
	}

	if success {
		if cb.state == ProviderRecovering {
			cb.state = ProviderHealthy
			cb.window = []bool{true}
		}
		return
	}

	failures := 0
	for _, ok := range cb.window {
		if !ok {
			failures++
		}
	}
	rate := float64(failures) / float64(len(cb.window))
	if len(cb.window) >= 4 && rate >= cb.failureThreshold {
		cb.state = ProviderDegraded
		cb.openUntil = time.Now().Add(cb.cooldown)
	}
}

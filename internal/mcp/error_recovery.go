package mcp

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// Phase 4.2: Advanced Error Handling & Recovery

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// Circuit Breaker Configuration
const (
	DefaultFailureThreshold  = 5
	DefaultTimeoutDuration   = 30 * time.Second
	DefaultRetryInterval     = 10 * time.Second
	DefaultMaxRetries        = 3
	DefaultInitialBackoff    = 100 * time.Millisecond
	DefaultMaxBackoff        = 5 * time.Second
	DefaultBackoffMultiplier = 2.0
	DefaultJitterFactor      = 0.1
)

// ErrorContext provides detailed error information with context preservation
type ErrorContext struct {
	Operation      string                 `json:"operation"`
	ToolName       string                 `json:"tool_name"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	OriginalError  string                 `json:"original_error"`
	ErrorCode      string                 `json:"error_code"`
	Timestamp      time.Time              `json:"timestamp"`
	RetryAttempt   int                    `json:"retry_attempt,omitempty"`
	RetryCount     int                    `json:"retry_count,omitempty"`
	RecoveryAction string                 `json:"recovery_action,omitempty"`
	ContextData    map[string]string      `json:"context_data,omitempty"`
}

// CircuitBreaker implements the circuit breaker pattern for failing operations
type CircuitBreaker struct {
	name             string
	state            CircuitBreakerState
	failureCount     int
	failureThreshold int
	timeout          time.Duration
	retryInterval    time.Duration
	lastFailureTime  time.Time
	mutex            sync.RWMutex
}

// RetryConfig holds configuration for retry mechanisms
type RetryConfig struct {
	MaxRetries        int
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64
	JitterFactor      float64
	RetryableErrors   []string
}

// ErrorRecoveryManager manages circuit breakers and retry mechanisms
type ErrorRecoveryManager struct {
	circuitBreakers map[string]*CircuitBreaker
	retryConfig     *RetryConfig
	mutex           sync.RWMutex
}

// NewErrorRecoveryManager creates a new error recovery manager
func NewErrorRecoveryManager() *ErrorRecoveryManager {
	return &ErrorRecoveryManager{
		circuitBreakers: make(map[string]*CircuitBreaker),
		retryConfig: &RetryConfig{
			MaxRetries:        DefaultMaxRetries,
			InitialBackoff:    DefaultInitialBackoff,
			MaxBackoff:        DefaultMaxBackoff,
			BackoffMultiplier: DefaultBackoffMultiplier,
			JitterFactor:      DefaultJitterFactor,
			RetryableErrors:   []string{"query_error", "storage_error", "network_error", "timeout_error"},
		},
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		state:            CircuitBreakerClosed,
		failureCount:     0,
		failureThreshold: DefaultFailureThreshold,
		timeout:          DefaultTimeoutDuration,
		retryInterval:    DefaultRetryInterval,
	}
}

// GetCircuitBreaker gets or creates a circuit breaker for an operation
func (erm *ErrorRecoveryManager) GetCircuitBreaker(operationName string) *CircuitBreaker {
	erm.mutex.Lock()
	defer erm.mutex.Unlock()

	if cb, exists := erm.circuitBreakers[operationName]; exists {
		return cb
	}

	cb := NewCircuitBreaker(operationName)
	erm.circuitBreakers[operationName] = cb
	return cb
}

// CanExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		if time.Since(cb.lastFailureTime) > cb.timeout {
			return true // Transition to half-open will be handled in execution
		}
		return false
	case CircuitBreakerHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount = 0
	cb.state = CircuitBreakerClosed
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.failureCount >= cb.failureThreshold {
		cb.state = CircuitBreakerOpen
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetFailureCount returns the current failure count
func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.failureCount
}

// TransitionToHalfOpen transitions the circuit breaker to half-open state
func (cb *CircuitBreaker) TransitionToHalfOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.state == CircuitBreakerOpen && time.Since(cb.lastFailureTime) > cb.timeout {
		cb.state = CircuitBreakerHalfOpen
	}
}

// NewErrorContext creates a new error context
func NewErrorContext(operation, toolName, originalError string) *ErrorContext {
	return &ErrorContext{
		Operation:     operation,
		ToolName:      toolName,
		OriginalError: originalError,
		Timestamp:     time.Now(),
		ContextData:   make(map[string]string),
	}
}

// WithParameters adds parameters to the error context
func (ec *ErrorContext) WithParameters(params map[string]interface{}) *ErrorContext {
	ec.Parameters = params
	return ec
}

// WithErrorCode adds an error code to the error context
func (ec *ErrorContext) WithErrorCode(code string) *ErrorContext {
	ec.ErrorCode = code
	return ec
}

// WithRetryInfo adds retry information to the error context
func (ec *ErrorContext) WithRetryInfo(attempt, count int) *ErrorContext {
	ec.RetryAttempt = attempt
	ec.RetryCount = count
	return ec
}

// WithRecoveryAction adds recovery action information
func (ec *ErrorContext) WithRecoveryAction(action string) *ErrorContext {
	ec.RecoveryAction = action
	return ec
}

// WithContextData adds additional context data
func (ec *ErrorContext) WithContextData(key, value string) *ErrorContext {
	if ec.ContextData == nil {
		ec.ContextData = make(map[string]string)
	}
	ec.ContextData[key] = value
	return ec
}

// ToError converts the error context to a standard error
func (ec *ErrorContext) ToError() error {
	return fmt.Errorf("operation '%s' failed: %s (error_code: %s, tool: %s, timestamp: %s)",
		ec.Operation, ec.OriginalError, ec.ErrorCode, ec.ToolName, ec.Timestamp.Format(time.RFC3339))
}

// calculateBackoff calculates the backoff duration with exponential backoff and jitter
func (erm *ErrorRecoveryManager) calculateBackoff(attempt int) time.Duration {
	backoff := float64(erm.retryConfig.InitialBackoff) *
		math.Pow(erm.retryConfig.BackoffMultiplier, float64(attempt))

	// Apply maximum backoff limit
	if backoff > float64(erm.retryConfig.MaxBackoff) {
		backoff = float64(erm.retryConfig.MaxBackoff)
	}

	// Add jitter to prevent thundering herd
	// #nosec G404 - Using math/rand for jitter is appropriate here, crypto/rand is overkill for retry timing variance
	jitter := backoff * erm.retryConfig.JitterFactor * (rand.Float64()*2 - 1)
	backoff += jitter

	// Ensure minimum backoff
	if backoff < float64(erm.retryConfig.InitialBackoff) {
		backoff = float64(erm.retryConfig.InitialBackoff)
	}

	return time.Duration(backoff)
}

// IsRetryableError checks if an error is retryable
func (erm *ErrorRecoveryManager) IsRetryableError(err error, errorCode string) bool {
	if err == nil {
		return false
	}

	// Check if error code is in retryable list
	for _, retryableCode := range erm.retryConfig.RetryableErrors {
		if errorCode == retryableCode {
			return true
		}
	}

	return false
}

// ExecuteWithRecovery executes an operation with circuit breaker and retry mechanisms
func (erm *ErrorRecoveryManager) ExecuteWithRecovery(
	ctx context.Context,
	operationName string,
	operation func() (*mcp.CallToolResult, error),
) (*mcp.CallToolResult, error) {
	cb := erm.GetCircuitBreaker(operationName)

	// Check circuit breaker
	if !cb.CanExecute() {
		errorCtx := NewErrorContext(operationName, "", "circuit breaker open").
			WithErrorCode("circuit_breaker_open").
			WithRecoveryAction("wait for circuit breaker timeout")
		return nil, errorCtx.ToError()
	}

	// Transition to half-open if needed
	if cb.GetState() == CircuitBreakerOpen {
		cb.TransitionToHalfOpen()
	}

	var lastErr error
	var lastResult *mcp.CallToolResult

	// Retry loop with exponential backoff
	for attempt := 0; attempt <= erm.retryConfig.MaxRetries; attempt++ {
		// Execute operation
		result, err := operation()

		if err == nil {
			cb.RecordSuccess()
			return result, nil
		}

		lastErr = err
		lastResult = result

		// Record failure for circuit breaker
		cb.RecordFailure()

		// Don't retry on last attempt
		if attempt == erm.retryConfig.MaxRetries {
			break
		}

		// Check if error is retryable
		errorCode := erm.extractErrorCode(err)
		if !erm.IsRetryableError(err, errorCode) {
			break
		}

		// Calculate backoff and wait
		backoffDuration := erm.calculateBackoff(attempt)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoffDuration):
			// Continue to next retry
		}
	}

	// All retries exhausted
	errorCtx := NewErrorContext(operationName, "", lastErr.Error()).
		WithErrorCode("max_retries_exhausted").
		WithRetryInfo(erm.retryConfig.MaxRetries, erm.retryConfig.MaxRetries).
		WithRecoveryAction("consider checking system resources or configuration")

	return lastResult, errorCtx.ToError()
}

// extractErrorCode extracts error code from error message
func (erm *ErrorRecoveryManager) extractErrorCode(err error) string {
	if err == nil {
		return ""
	}

	// Simple error code extraction - could be enhanced
	errMsg := err.Error()

	// Handle empty error message
	if errMsg == "" {
		return "unknown_empty_error"
	}

	// Check for common error patterns
	if contains(errMsg, "query") {
		return "query_error"
	}
	if contains(errMsg, "storage") || contains(errMsg, "database") {
		return "storage_error"
	}
	if contains(errMsg, "network") || contains(errMsg, "connection") {
		return "network_error"
	}
	if contains(errMsg, "timeout") {
		return "timeout_error"
	}
	if contains(errMsg, "validation") {
		return "validation_error"
	}

	return "unknown_error"
}

// contains is a helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr ||
		(len(str) > len(substr) &&
			(str[:len(substr)] == substr ||
				str[len(str)-len(substr):] == substr ||
				findSubstring(str, substr))))
}

// findSubstring is a simple substring search helper
func findSubstring(str, substr string) bool {
	if substr == "" {
		return true
	}
	if len(str) < len(substr) {
		return false
	}

	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetRecoveryStats returns statistics about error recovery
func (erm *ErrorRecoveryManager) GetRecoveryStats() map[string]interface{} {
	erm.mutex.RLock()
	defer erm.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_circuit_breakers"] = len(erm.circuitBreakers)

	circuitBreakerStats := make(map[string]interface{})
	for name, cb := range erm.circuitBreakers {
		cbStats := map[string]interface{}{
			"state":         cb.GetState(),
			"failure_count": cb.GetFailureCount(),
		}
		circuitBreakerStats[name] = cbStats
	}
	stats["circuit_breakers"] = circuitBreakerStats

	retryStats := map[string]interface{}{
		"max_retries":      erm.retryConfig.MaxRetries,
		"initial_backoff":  erm.retryConfig.InitialBackoff.String(),
		"max_backoff":      erm.retryConfig.MaxBackoff.String(),
		"retryable_errors": erm.retryConfig.RetryableErrors,
	}
	stats["retry_config"] = retryStats

	return stats
}

package mcp

import (
	"fmt"
	"testing"
	"time"
)

// ============================================================================
// Phase 4.2: Circuit Breaker Tests
// ============================================================================

func TestNewCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker("test_operation")

	if cb == nil {
		t.Fatal("NewCircuitBreaker should not return nil")
	}

	if cb.name != "test_operation" {
		t.Errorf("Expected name 'test_operation', got '%s'", cb.name)
	}

	if cb.GetState() != CircuitBreakerClosed {
		t.Errorf("Expected initial state to be Closed, got %v", cb.GetState())
	}

	if cb.GetFailureCount() != 0 {
		t.Errorf("Expected initial failure count to be 0, got %d", cb.GetFailureCount())
	}

	if cb.failureThreshold != DefaultFailureThreshold {
		t.Errorf("Expected failure threshold to be %d, got %d", DefaultFailureThreshold, cb.failureThreshold)
	}
}

func TestCircuitBreaker_CanExecute_Closed(t *testing.T) {
	cb := NewCircuitBreaker("test_operation")

	// Circuit breaker should allow execution when closed
	if !cb.CanExecute() {
		t.Error("Circuit breaker should allow execution when closed")
	}
}

func TestCircuitBreaker_RecordSuccess(t *testing.T) {
	cb := NewCircuitBreaker("test_operation")

	// Record some failures first
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.GetFailureCount() != 2 {
		t.Errorf("Expected failure count 2, got %d", cb.GetFailureCount())
	}

	// Record success should reset failure count and ensure closed state
	cb.RecordSuccess()

	if cb.GetFailureCount() != 0 {
		t.Errorf("Expected failure count to be reset to 0, got %d", cb.GetFailureCount())
	}

	if cb.GetState() != CircuitBreakerClosed {
		t.Errorf("Expected state to be Closed after success, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_RecordFailure_IncreasesCount(t *testing.T) {
	cb := NewCircuitBreaker("test_operation")

	// Record first failure
	cb.RecordFailure()

	if cb.GetFailureCount() != 1 {
		t.Errorf("Expected failure count 1, got %d", cb.GetFailureCount())
	}

	if cb.GetState() != CircuitBreakerClosed {
		t.Errorf("Expected state to remain Closed after single failure, got %v", cb.GetState())
	}

	// Record second failure
	cb.RecordFailure()

	if cb.GetFailureCount() != 2 {
		t.Errorf("Expected failure count 2, got %d", cb.GetFailureCount())
	}

	if cb.GetState() != CircuitBreakerClosed {
		t.Errorf("Expected state to remain Closed after two failures, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_RecordFailure_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker("test_operation")

	// Record failures up to threshold
	for i := 0; i < DefaultFailureThreshold; i++ {
		cb.RecordFailure()
	}

	if cb.GetState() != CircuitBreakerOpen {
		t.Errorf("Expected state to be Open after %d failures, got %v", DefaultFailureThreshold, cb.GetState())
	}

	if cb.GetFailureCount() != DefaultFailureThreshold {
		t.Errorf("Expected failure count %d, got %d", DefaultFailureThreshold, cb.GetFailureCount())
	}
}

func TestCircuitBreaker_CanExecute_Open(t *testing.T) {
	cb := NewCircuitBreaker("test_operation")

	// Force circuit breaker to open
	for i := 0; i < DefaultFailureThreshold; i++ {
		cb.RecordFailure()
	}

	if cb.GetState() != CircuitBreakerOpen {
		t.Fatalf("Circuit breaker should be open after %d failures", DefaultFailureThreshold)
	}

	// Should not allow execution when open
	if cb.CanExecute() {
		t.Error("Circuit breaker should not allow execution when open")
	}
}

func TestCircuitBreaker_CanExecute_OpenWithTimeout(t *testing.T) {
	cb := NewCircuitBreaker("test_operation")

	// Set a very short timeout for testing
	cb.timeout = 10 * time.Millisecond

	// Force circuit breaker to open
	for i := 0; i < DefaultFailureThreshold; i++ {
		cb.RecordFailure()
	}

	if cb.GetState() != CircuitBreakerOpen {
		t.Fatalf("Circuit breaker should be open after %d failures", DefaultFailureThreshold)
	}

	// Should not allow execution immediately
	if cb.CanExecute() {
		t.Error("Circuit breaker should not allow execution immediately when open")
	}

	// Wait for timeout
	time.Sleep(15 * time.Millisecond)

	// Should allow execution after timeout
	if !cb.CanExecute() {
		t.Error("Circuit breaker should allow execution after timeout")
	}
}

func TestCircuitBreaker_TransitionToHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker("test_operation")

	// Set a very short timeout for testing
	cb.timeout = 10 * time.Millisecond

	// Force circuit breaker to open
	for i := 0; i < DefaultFailureThreshold; i++ {
		cb.RecordFailure()
	}

	if cb.GetState() != CircuitBreakerOpen {
		t.Fatalf("Circuit breaker should be open after %d failures", DefaultFailureThreshold)
	}

	// Wait for timeout
	time.Sleep(15 * time.Millisecond)

	// Transition to half-open
	cb.TransitionToHalfOpen()

	if cb.GetState() != CircuitBreakerHalfOpen {
		t.Errorf("Expected state to be HalfOpen after timeout, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_CanExecute_HalfOpen(t *testing.T) {
	cb := NewCircuitBreaker("test_operation")

	// Set a very short timeout for testing
	cb.timeout = 10 * time.Millisecond

	// Force circuit breaker to open
	for i := 0; i < DefaultFailureThreshold; i++ {
		cb.RecordFailure()
	}

	// Wait for timeout and transition to half-open
	time.Sleep(15 * time.Millisecond)
	cb.TransitionToHalfOpen()

	if cb.GetState() != CircuitBreakerHalfOpen {
		t.Fatalf("Circuit breaker should be half-open")
	}

	// Should allow execution when half-open
	if !cb.CanExecute() {
		t.Error("Circuit breaker should allow execution when half-open")
	}
}

func TestCircuitBreaker_HalfOpenToClosedOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker("test_operation")

	// Set a very short timeout for testing
	cb.timeout = 10 * time.Millisecond

	// Force circuit breaker to open
	for i := 0; i < DefaultFailureThreshold; i++ {
		cb.RecordFailure()
	}

	// Wait for timeout and transition to half-open
	time.Sleep(15 * time.Millisecond)
	cb.TransitionToHalfOpen()

	if cb.GetState() != CircuitBreakerHalfOpen {
		t.Fatalf("Circuit breaker should be half-open")
	}

	// Record success should transition to closed
	cb.RecordSuccess()

	if cb.GetState() != CircuitBreakerClosed {
		t.Errorf("Expected state to be Closed after success in half-open, got %v", cb.GetState())
	}

	if cb.GetFailureCount() != 0 {
		t.Errorf("Expected failure count to be reset to 0, got %d", cb.GetFailureCount())
	}
}

func TestCircuitBreaker_HalfOpenToOpenOnFailure(t *testing.T) {
	cb := NewCircuitBreaker("test_operation")

	// Set a very short timeout for testing
	cb.timeout = 10 * time.Millisecond

	// Force circuit breaker to open
	for i := 0; i < DefaultFailureThreshold; i++ {
		cb.RecordFailure()
	}

	// Wait for timeout and transition to half-open
	time.Sleep(15 * time.Millisecond)
	cb.TransitionToHalfOpen()

	if cb.GetState() != CircuitBreakerHalfOpen {
		t.Fatalf("Circuit breaker should be half-open")
	}

	// Record failure should transition back to open
	cb.RecordFailure()

	if cb.GetState() != CircuitBreakerOpen {
		t.Errorf("Expected state to be Open after failure in half-open, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := NewCircuitBreaker("test_operation")

	// Test concurrent access to circuit breaker
	done := make(chan bool, 10)

	// Launch multiple goroutines to access circuit breaker concurrently
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Perform various operations
			cb.CanExecute()
			cb.RecordFailure()
			cb.GetState()
			cb.GetFailureCount()
			cb.RecordSuccess()
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Circuit breaker should still be functional
	if cb.GetState() != CircuitBreakerClosed {
		t.Errorf("Expected state to be Closed after concurrent access, got %v", cb.GetState())
	}
}

// ============================================================================
// Phase 4.2: Error Recovery Manager Tests
// ============================================================================

func TestNewErrorRecoveryManager(t *testing.T) {
	erm := NewErrorRecoveryManager()

	if erm == nil {
		t.Fatal("NewErrorRecoveryManager should not return nil")
	}

	if erm.circuitBreakers == nil {
		t.Error("Circuit breakers map should be initialized")
	}

	if erm.retryConfig == nil {
		t.Error("Retry config should be initialized")
	}

	if erm.retryConfig.MaxRetries != DefaultMaxRetries {
		t.Errorf("Expected max retries %d, got %d", DefaultMaxRetries, erm.retryConfig.MaxRetries)
	}

	if erm.retryConfig.InitialBackoff != DefaultInitialBackoff {
		t.Errorf("Expected initial backoff %v, got %v", DefaultInitialBackoff, erm.retryConfig.InitialBackoff)
	}
}

func TestErrorRecoveryManager_GetCircuitBreaker(t *testing.T) {
	erm := NewErrorRecoveryManager()

	// Get circuit breaker for new operation
	cb1 := erm.GetCircuitBreaker("operation1")
	if cb1 == nil {
		t.Fatal("GetCircuitBreaker should not return nil")
	}

	// Get circuit breaker for same operation should return same instance
	cb2 := erm.GetCircuitBreaker("operation1")
	if cb1 != cb2 {
		t.Error("GetCircuitBreaker should return same instance for same operation")
	}

	// Get circuit breaker for different operation should return different instance
	cb3 := erm.GetCircuitBreaker("operation2")
	if cb1 == cb3 {
		t.Error("GetCircuitBreaker should return different instances for different operations")
	}
}

func TestErrorRecoveryManager_CalculateBackoff(t *testing.T) {
	erm := NewErrorRecoveryManager()

	// Test backoff calculation for different attempts
	backoff0 := erm.calculateBackoff(0)
	backoff1 := erm.calculateBackoff(1)
	backoff2 := erm.calculateBackoff(2)

	// Backoff should increase with attempts
	if backoff1 <= backoff0 {
		t.Errorf("Expected backoff to increase with attempts, got %v <= %v", backoff1, backoff0)
	}

	if backoff2 <= backoff1 {
		t.Errorf("Expected backoff to increase with attempts, got %v <= %v", backoff2, backoff1)
	}

	// Test maximum backoff limit
	backoff10 := erm.calculateBackoff(10)
	if backoff10 > erm.retryConfig.MaxBackoff*2 {
		t.Errorf("Backoff should not exceed reasonable limit, got %v", backoff10)
	}
}

func TestErrorRecoveryManager_IsRetryableError(t *testing.T) {
	erm := NewErrorRecoveryManager()

	testCases := []struct {
		errorCode string
		expected  bool
	}{
		{"query_error", true},
		{"storage_error", true},
		{"network_error", true},
		{"timeout_error", true},
		{"validation_error", false},
		{"unknown_error", false},
		{"", false},
	}

	for _, tc := range testCases {
		err := fmt.Errorf("test error")
		result := erm.IsRetryableError(err, tc.errorCode)
		if result != tc.expected {
			t.Errorf("Expected IsRetryableError(%s) to be %v, got %v", tc.errorCode, tc.expected, result)
		}
	}

	// Test with nil error
	if erm.IsRetryableError(nil, "query_error") {
		t.Error("IsRetryableError should return false for nil error")
	}
}

func TestErrorRecoveryManager_ExtractErrorCode(t *testing.T) {
	erm := NewErrorRecoveryManager()

	testCases := []struct {
		errorMsg     string
		expectedCode string
	}{
		{"query failed", "query_error"},
		{"storage connection failed", "storage_error"},
		{"database error occurred", "storage_error"},
		{"network timeout", "network_error"},
		{"connection refused", "network_error"},
		{"operation timeout", "timeout_error"},
		{"validation failed", "validation_error"},
		{"unknown issue", "unknown_error"},
		{"", "unknown_empty_error"},
	}

	for _, tc := range testCases {
		// Always create an error for non-empty test cases
		err := fmt.Errorf("%s", tc.errorMsg)

		result := erm.extractErrorCode(err)
		if result != tc.expectedCode {
			t.Errorf("Expected extractErrorCode(%s) to be %s, got %s", tc.errorMsg, tc.expectedCode, result)
		}
	}

	// Test with nil error
	if erm.extractErrorCode(nil) != "" {
		t.Error("extractErrorCode should return empty string for nil error")
	}
}

// ============================================================================
// Phase 4.2: Error Context Tests
// ============================================================================

func TestNewErrorContext(t *testing.T) {
	operation := "test_operation"
	toolName := "test_tool"
	originalError := "test error"

	ec := NewErrorContext(operation, toolName, originalError)

	if ec == nil {
		t.Fatal("NewErrorContext should not return nil")
	}

	if ec.Operation != operation {
		t.Errorf("Expected operation '%s', got '%s'", operation, ec.Operation)
	}

	if ec.ToolName != toolName {
		t.Errorf("Expected tool name '%s', got '%s'", toolName, ec.ToolName)
	}

	if ec.OriginalError != originalError {
		t.Errorf("Expected original error '%s', got '%s'", originalError, ec.OriginalError)
	}

	if ec.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}

	if ec.ContextData == nil {
		t.Error("ContextData should be initialized")
	}
}

func TestErrorContext_WithParameters(t *testing.T) {
	ec := NewErrorContext("test_op", "test_tool", "test error")

	params := map[string]interface{}{
		"param1": "value1",
		"param2": 42,
	}

	result := ec.WithParameters(params)

	if result != ec {
		t.Error("WithParameters should return same instance for chaining")
	}

	if len(ec.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(ec.Parameters))
	}

	if ec.Parameters["param1"] != "value1" {
		t.Errorf("Expected param1 to be 'value1', got %v", ec.Parameters["param1"])
	}

	if ec.Parameters["param2"] != 42 {
		t.Errorf("Expected param2 to be 42, got %v", ec.Parameters["param2"])
	}
}

func TestErrorContext_WithErrorCode(t *testing.T) {
	ec := NewErrorContext("test_op", "test_tool", "test error")

	errorCode := "test_code"
	result := ec.WithErrorCode(errorCode)

	if result != ec {
		t.Error("WithErrorCode should return same instance for chaining")
	}

	if ec.ErrorCode != errorCode {
		t.Errorf("Expected error code '%s', got '%s'", errorCode, ec.ErrorCode)
	}
}

func TestErrorContext_WithRetryInfo(t *testing.T) {
	ec := NewErrorContext("test_op", "test_tool", "test error")

	attempt := 2
	count := 3
	result := ec.WithRetryInfo(attempt, count)

	if result != ec {
		t.Error("WithRetryInfo should return same instance for chaining")
	}

	if ec.RetryAttempt != attempt {
		t.Errorf("Expected retry attempt %d, got %d", attempt, ec.RetryAttempt)
	}

	if ec.RetryCount != count {
		t.Errorf("Expected retry count %d, got %d", count, ec.RetryCount)
	}
}

func TestErrorContext_WithRecoveryAction(t *testing.T) {
	ec := NewErrorContext("test_op", "test_tool", "test error")

	action := "retry operation"
	result := ec.WithRecoveryAction(action)

	if result != ec {
		t.Error("WithRecoveryAction should return same instance for chaining")
	}

	if ec.RecoveryAction != action {
		t.Errorf("Expected recovery action '%s', got '%s'", action, ec.RecoveryAction)
	}
}

func TestErrorContext_WithContextData(t *testing.T) {
	ec := NewErrorContext("test_op", "test_tool", "test error")

	key := "test_key"
	value := "test_value"
	result := ec.WithContextData(key, value)

	if result != ec {
		t.Error("WithContextData should return same instance for chaining")
	}

	if ec.ContextData[key] != value {
		t.Errorf("Expected context data[%s] to be '%s', got '%s'", key, value, ec.ContextData[key])
	}

	// Test adding multiple context data
	ec.WithContextData("key2", "value2")
	if len(ec.ContextData) != 2 {
		t.Errorf("Expected 2 context data entries, got %d", len(ec.ContextData))
	}
}

func TestErrorContext_ToError(t *testing.T) {
	ec := NewErrorContext("test_operation", "test_tool", "original error message").
		WithErrorCode("test_code").
		WithRetryInfo(1, 3).
		WithRecoveryAction("retry").
		WithContextData("context_key", "context_value")

	err := ec.ToError()

	if err == nil {
		t.Fatal("ToError should not return nil")
	}

	errMsg := err.Error()

	// Check that error message contains expected components
	if !contains(errMsg, "test_operation") {
		t.Error("Error message should contain operation name")
	}

	if !contains(errMsg, "original error message") {
		t.Error("Error message should contain original error")
	}

	if !contains(errMsg, "test_code") {
		t.Error("Error message should contain error code")
	}

	if !contains(errMsg, "test_tool") {
		t.Error("Error message should contain tool name")
	}
}

func TestErrorContext_ChainedOperations(t *testing.T) {
	// Test method chaining
	ec := NewErrorContext("test_op", "test_tool", "test error").
		WithErrorCode("test_code").
		WithParameters(map[string]interface{}{"param": "value"}).
		WithRetryInfo(1, 3).
		WithRecoveryAction("retry").
		WithContextData("key", "value")

	if ec.ErrorCode != "test_code" {
		t.Error("Error code should be set through chaining")
	}

	if ec.Parameters["param"] != "value" {
		t.Error("Parameters should be set through chaining")
	}

	if ec.RetryAttempt != 1 || ec.RetryCount != 3 {
		t.Error("Retry info should be set through chaining")
	}

	if ec.RecoveryAction != "retry" {
		t.Error("Recovery action should be set through chaining")
	}

	if ec.ContextData["key"] != "value" {
		t.Error("Context data should be set through chaining")
	}
}

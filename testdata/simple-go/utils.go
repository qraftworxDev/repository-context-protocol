package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Utility constants for testing
const (
	UtilVersion    = "2.1.0"
	MaxRetries     = 3
	RetryDelay     = 100 * time.Millisecond
	CacheKeyPrefix = "cache:"
)

// Utility variables
var (
	emailRegex    *regexp.Regexp
	phoneRegex    *regexp.Regexp
	utilsLogger   Logger
	defaultConfig map[string]interface{}
)

// Logger interface for utility logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// StringUtils provides string utility functions
type StringUtils struct{}

// NewStringUtils creates a new string utilities instance
func NewStringUtils() *StringUtils {
	return &StringUtils{}
}

// FormatName formats a name with proper capitalization
func (su *StringUtils) FormatName(name string) string {
	// This function calls TrimWhitespace and CapitalizeWords
	trimmed := su.TrimWhitespace(name)
	return su.CapitalizeWords(trimmed)
}

// TrimWhitespace removes extra whitespace
func (su *StringUtils) TrimWhitespace(input string) string {
	return strings.TrimSpace(input)
}

// CapitalizeWords capitalizes the first letter of each word
func (su *StringUtils) CapitalizeWords(input string) string {
	words := strings.Fields(input)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

// ValidateEmail validates email format using regex
func ValidateEmail(email string) bool {
	// This function calls GetEmailRegex
	regex := GetEmailRegex()
	return regex.MatchString(email)
}

// GetEmailRegex returns the compiled email regex
func GetEmailRegex() *regexp.Regexp {
	if emailRegex == nil {
		emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	}
	return emailRegex
}

// ValidatePhone validates phone number format
func ValidatePhone(phone string) bool {
	// This function calls GetPhoneRegex
	regex := GetPhoneRegex()
	return regex.MatchString(phone)
}

// GetPhoneRegex returns the compiled phone regex
func GetPhoneRegex() *regexp.Regexp {
	if phoneRegex == nil {
		phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	}
	return phoneRegex
}

// GenerateID generates a random ID
func GenerateID() string {
	// This function calls generateRandomBytes
	bytes := generateRandomBytes(8)
	return hex.EncodeToString(bytes)
}

// generateRandomBytes generates random bytes
func generateRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return bytes
}

// FormatTime formats time with a standard format
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ParseTime parses time from string
func ParseTime(timeStr string) (time.Time, error) {
	// This function calls FormatTime indirectly via validation
	parsed, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format: %w", err)
	}
	return parsed, nil
}

// SortUsers sorts users by name
func SortUsers(users []*User) []*User {
	// This function calls compareUsers for sorting
	sorted := make([]*User, len(users))
	copy(sorted, users)

	sort.Slice(sorted, func(i, j int) bool {
		return compareUsers(sorted[i], sorted[j])
	})

	return sorted
}

// compareUsers compares two users for sorting
func compareUsers(a, b *User) bool {
	return strings.ToLower(a.Name) < strings.ToLower(b.Name)
}

// FilterUsers filters users by active status
func FilterUsers(users []*User, activeOnly bool) []*User {
	// This function calls isUserActive
	filtered := make([]*User, 0)

	for _, user := range users {
		if !activeOnly || isUserActive(user) {
			filtered = append(filtered, user)
		}
	}

	return filtered
}

// isUserActive checks if user is active
func isUserActive(user *User) bool {
	return user != nil && user.IsActive
}

// MapUsers transforms users using a mapper function
func MapUsers(users []*User, mapper func(*User) *User) []*User {
	mapped := make([]*User, len(users))
	for i, user := range users {
		mapped[i] = mapper(user)
	}
	return mapped
}

// ConvertToJSON attempts to convert an object to JSON-like representation
func ConvertToJSON(obj interface{}) map[string]interface{} {
	// This function calls reflectObject for conversion
	return reflectObject(obj)
}

// reflectObject uses reflection to convert object to map
func reflectObject(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	if obj == nil {
		return result
	}

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return result
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return result
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.IsExported() {
			result[field.Name] = val.Field(i).Interface()
		}
	}

	return result
}

// RetryOperation retries an operation with exponential backoff
func RetryOperation(operation func() error) error {
	// This function calls calculateRetryDelay
	var lastErr error

	for attempt := 0; attempt < MaxRetries; attempt++ {
		if err := operation(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if attempt < MaxRetries-1 {
			delay := calculateRetryDelay(attempt)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", MaxRetries, lastErr)
}

// calculateRetryDelay calculates the delay for retry attempts
func calculateRetryDelay(attempt int) time.Duration {
	base := RetryDelay
	multiplier := time.Duration(1 << attempt) // Exponential backoff
	return base * multiplier
}

// InitializeUtilities initializes utility components
func InitializeUtilities() {
	// This function calls multiple initialization functions
	initializeRegex()
	initializeDefaultConfig()
	initializeLogger()
}

// initializeRegex initializes regex patterns
func initializeRegex() {
	GetEmailRegex() // Force initialization
	GetPhoneRegex() // Force initialization
}

// initializeDefaultConfig initializes default configuration
func initializeDefaultConfig() {
	defaultConfig = map[string]interface{}{
		"timeout":   DefaultTimeout,
		"maxUsers":  MaxUsers,
		"version":   UtilVersion,
		"debugMode": DEBUG_MODE,
	}
}

// initializeLogger initializes the logger (placeholder)
func initializeLogger() {
	// In a real implementation, this would set up logging
	// For testing purposes, we'll leave it as a placeholder
}

// GetDefaultConfig returns the default configuration
func GetDefaultConfig() map[string]interface{} {
	if defaultConfig == nil {
		initializeDefaultConfig()
	}
	return defaultConfig
}

// CleanupUtilities cleans up utility resources
func CleanupUtilities() {
	emailRegex = nil
	phoneRegex = nil
	utilsLogger = nil
	defaultConfig = nil
}

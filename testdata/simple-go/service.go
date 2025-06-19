package main

import (
	"fmt"
	"log"
	"strings"
)

// UserManager provides high-level user management operations
type UserManager struct {
	service      UserService
	cache        CacheManager
	notification NotificationService
	config       *Config
}

// NewUserManager creates a new user manager
func NewUserManager(service UserService) *UserManager {
	return &UserManager{
		service: service,
		config:  NewConfig(),
	}
}

// ProcessUser processes a user creation request with validation
func (um *UserManager) ProcessUser(name, email string) (*User, error) {
	// This function calls ValidateUser and CreateUser - good for call graph testing
	if err := um.ValidateUser(name, email); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	user, err := um.service.CreateUser(name, email)
	if err != nil {
		return nil, fmt.Errorf("creation failed: %w", err)
	}

	// Send welcome notification
	um.SendWelcomeNotification(user)

	return user, nil
}

// ValidateUser validates user input
func (um *UserManager) ValidateUser(name, email string) error {
	// This function calls IsValidEmail - creates call chain
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if !um.IsValidEmail(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// IsValidEmail checks if email format is valid
func (um *UserManager) IsValidEmail(email string) bool {
	// Simple email validation for testing
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// SendWelcomeNotification sends a welcome notification to the user
func (um *UserManager) SendWelcomeNotification(user *User) {
	// This would call notification service if it existed
	log.Printf("Welcome notification sent to %s", user.Email)
}

// GetUserWithProfile retrieves user and creates/returns profile
func (um *UserManager) GetUserWithProfile(userID int) (*User, *Profile, error) {
	// This function calls multiple other functions
	user, err := um.service.GetUser(userID)
	if err != nil {
		return nil, nil, err
	}

	profile := um.CreateUserProfile(user)
	return user, profile, nil
}

// CreateUserProfile creates a profile for a user
func (um *UserManager) CreateUserProfile(user *User) *Profile {
	// This function calls NewProfile
	profile := NewProfile(user.ID)
	profile.Bio = fmt.Sprintf("Profile for %s", user.Name)
	return profile
}

// UpdateUserStatus updates user status with logging
func (um *UserManager) UpdateUserStatus(userID int, active bool) error {
	// This function calls GetUser and UpdateUser
	user, err := um.service.GetUser(userID)
	if err != nil {
		return err
	}

	user.IsActive = active
	err = um.service.UpdateUser(user)
	if err != nil {
		return err
	}

	um.LogUserActivity(user, "status_update")
	return nil
}

// LogUserActivity logs user activity
func (um *UserManager) LogUserActivity(user *User, activity string) {
	log.Printf("User %d (%s) performed activity: %s", user.ID, user.Name, activity)
}

// BatchProcessUsers processes multiple users
func (um *UserManager) BatchProcessUsers(requests []UserRequest) []UserResult {
	// This function calls ProcessUser multiple times
	results := make([]UserResult, len(requests))

	for i, req := range requests {
		user, err := um.ProcessUser(req.Name, req.Email)
		results[i] = UserResult{
			User:  user,
			Error: err,
		}
	}

	return results
}

// UserRequest represents a user creation request
type UserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserResult represents the result of user processing
type UserResult struct {
	User  *User `json:"user,omitempty"`
	Error error `json:"error,omitempty"`
}

// SearchUsers searches for users by name pattern
func (um *UserManager) SearchUsers(pattern string) ([]*User, error) {
	// For testing purposes, this would search through users
	// This function calls helper functions
	normalizedPattern := um.NormalizeSearchPattern(pattern)
	log.Printf("Searching users with pattern: %s", normalizedPattern)

	// In real implementation, this would search through actual users
	return []*User{}, nil
}

// NormalizeSearchPattern normalizes the search pattern
func (um *UserManager) NormalizeSearchPattern(pattern string) string {
	return strings.ToLower(strings.TrimSpace(pattern))
}

// InitializeServices initializes all services
func InitializeServices() *UserManager {
	// This function calls multiple constructors
	userService := NewInMemoryUserService()
	manager := NewUserManager(userService)

	// Initialize global variables
	globalConfig = NewConfig()
	serviceRegistry = make(map[string]interface{})
	isInitialized = true

	return manager
}

// CleanupServices cleans up all services
func CleanupServices() {
	globalConfig = nil
	serviceRegistry = nil
	isInitialized = false
	log.Println("Services cleaned up")
}

package main

import (
	"fmt"
	"strings"
)

// User represents a user in the system
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

// String returns a string representation of the user
func (u User) String() string {
	return fmt.Sprintf("User{ID: %d, Name: %s}", u.ID, u.Name)
}

// Activate sets the user as active
func (u *User) Activate() {
	u.IsActive = true
}

// UserService provides user-related operations
type UserService interface {
	GetUser(id int) (*User, error)
	CreateUser(name, email string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id int) error
}

// InMemoryUserService is an in-memory implementation of UserService
type InMemoryUserService struct {
	users  map[int]*User
	nextID int
}

// NewInMemoryUserService creates a new in-memory user service
func NewInMemoryUserService() *InMemoryUserService {
	return &InMemoryUserService{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

// GetUser retrieves a user by ID
func (s *InMemoryUserService) GetUser(id int) (*User, error) {
	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user with ID %d not found", id)
	}
	return user, nil
}

// CreateUser creates a new user
func (s *InMemoryUserService) CreateUser(name, email string) (*User, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	user := &User{
		ID:       s.nextID,
		Name:     name,
		Email:    email,
		IsActive: true,
	}

	s.users[s.nextID] = user
	s.nextID++

	return user, nil
}

// UpdateUser updates an existing user
func (s *InMemoryUserService) UpdateUser(user *User) error {
	if _, exists := s.users[user.ID]; !exists {
		return fmt.Errorf("user with ID %d not found", user.ID)
	}

	s.users[user.ID] = user
	return nil
}

// DeleteUser deletes a user by ID
func (s *InMemoryUserService) DeleteUser(id int) error {
	if _, exists := s.users[id]; !exists {
		return fmt.Errorf("user with ID %d not found", id)
	}

	delete(s.users, id)
	return nil
}

// GetAllUsers returns all users (added for testing)
func (s *InMemoryUserService) GetAllUsers() []*User {
	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}

// GetUserCount returns the number of users
func (s *InMemoryUserService) GetUserCount() int {
	return len(s.users)
}

func main() {
	// Initialize utilities first
	InitializeUtilities()

	// Initialize services
	manager := InitializeServices()

	// Demo basic user operations
	fmt.Println("=== Basic User Operations Demo ===")
	demoBasicOperations(manager)

	// Demo advanced user operations
	fmt.Println("\n=== Advanced User Operations Demo ===")
	demoAdvancedOperations(manager)

	// Demo utility functions
	fmt.Println("\n=== Utility Functions Demo ===")
	demoUtilityFunctions()

	// Run our new tests
	fmt.Println("\n=== Random Bytes Tests ===")
	TestGenerateRandomBytes()
	TestGenerateID()

	// Cleanup
	CleanupServices()
	CleanupUtilities()
}

// demoBasicOperations demonstrates basic user operations
func demoBasicOperations(manager *UserManager) {
	// This function calls ProcessUser which has a call chain
	user, err := manager.ProcessUser("John Doe", "john@example.com")
	if err != nil {
		fmt.Printf("Error processing user: %v\n", err)
		return
	}

	fmt.Printf("Processed user: %s\n", user.String())

	// Get the user back
	retrievedUser, profile, err := manager.GetUserWithProfile(user.ID)
	if err != nil {
		fmt.Printf("Error retrieving user: %v\n", err)
		return
	}

	fmt.Printf("Retrieved user: %s\n", retrievedUser.String())
	fmt.Printf("User profile: UserID=%d, Skills=%d\n", profile.UserID, profile.GetSkillCount())
}

// demoAdvancedOperations demonstrates advanced user operations
func demoAdvancedOperations(manager *UserManager) {
	// This function calls multiple service methods

	// Create multiple users for batch processing
	requests := []UserRequest{
		{Name: "Alice Smith", Email: "alice@example.com"},
		{Name: "Bob Johnson", Email: "bob@example.com"},
		{Name: "Charlie Brown", Email: "charlie@example.com"},
	}

	results := manager.BatchProcessUsers(requests)
	fmt.Printf("Batch processed %d users\n", len(results))

	// Update user status
	if len(results) > 0 && results[0].User != nil {
		err := manager.UpdateUserStatus(results[0].User.ID, false)
		if err != nil {
			fmt.Printf("Error updating user status: %v\n", err)
		} else {
			fmt.Printf("Updated user %d status\n", results[0].User.ID)
		}
	}

	// Search users (demonstrates pattern matching)
	users, err := manager.SearchUsers("Alice")
	if err != nil {
		fmt.Printf("Error searching users: %v\n", err)
	} else {
		fmt.Printf("Found %d users matching pattern\n", len(users))
	}
}

// demoUtilityFunctions demonstrates utility function usage
func demoUtilityFunctions() {
	// This function calls various utility functions with call chains

	// String utilities
	stringUtils := NewStringUtils()
	formatted := stringUtils.FormatName("  john   doe  ")
	fmt.Printf("Formatted name: %s\n", formatted)

	// Email validation (calls GetEmailRegex)
	isValid := ValidateEmail("test@example.com")
	fmt.Printf("Email validation result: %t\n", isValid)

	// Phone validation (calls GetPhoneRegex)
	phoneValid := ValidatePhone("+1234567890")
	fmt.Printf("Phone validation result: %t\n", phoneValid)

	// ID generation (calls generateRandomBytes)
	id := GenerateID()
	fmt.Printf("Generated ID: %s\n", id)

	// Configuration
	config := GetDefaultConfig()
	fmt.Printf("Default config loaded with %d keys\n", len(config))
}

// TestFunction is specifically for testing function search
func TestFunction() string {
	return "This is a test function"
}

// AnotherTestFunction for pattern matching tests
func AnotherTestFunction() int {
	return 42
}

// QueryEngineFunction for specific function name tests
func QueryEngineFunction() {
	// Empty function for testing specific queries
}

// SearchByName for specific function name tests
func SearchByName(name string) []string {
	return []string{name}
}

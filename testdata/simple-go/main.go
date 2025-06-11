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

func main() {
	service := NewInMemoryUserService()

	// Create a user
	user, err := service.CreateUser("John Doe", "john@example.com")
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return
	}

	fmt.Printf("Created user: %s\n", user.String())

	// Get the user
	retrievedUser, err := service.GetUser(user.ID)
	if err != nil {
		fmt.Printf("Error retrieving user: %v\n", err)
		return
	}

	fmt.Printf("Retrieved user: %s\n", retrievedUser.String())
}

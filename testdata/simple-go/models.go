package main

import (
	"fmt"
	"time"
)

// Constants for testing constant search
const (
	MaxUsers       = 100
	DefaultTimeout = 30 * time.Second
	ServiceVersion = "1.0.0"
	DEBUG_MODE     = true
)

// Package-level variables for testing variable search
var (
	globalConfig    *Config
	userCount       int
	serviceRegistry map[string]interface{}
	isInitialized   bool = false
)

// Config represents application configuration
type Config struct {
	DatabaseURL string          `json:"database_url"`
	Port        int             `json:"port"`
	LogLevel    string          `json:"log_level"`
	Features    map[string]bool `json:"features"`
}

// NewConfig creates a new configuration
func NewConfig() *Config {
	return &Config{
		DatabaseURL: "localhost:5432",
		Port:        8080,
		LogLevel:    "info",
		Features:    make(map[string]bool),
	}
}

// Address represents a user's address
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
}

// String returns a formatted address string
func (a Address) String() string {
	return fmt.Sprintf("%s, %s, %s %s, %s", a.Street, a.City, a.State, a.ZipCode, a.Country)
}

// Validate checks if the address is valid
func (a *Address) Validate() error {
	if a.Street == "" {
		return fmt.Errorf("street cannot be empty")
	}
	if a.City == "" {
		return fmt.Errorf("city cannot be empty")
	}
	return nil
}

// Profile represents a user profile with extended information
type Profile struct {
	UserID    int       `json:"user_id"`
	Bio       string    `json:"bio"`
	Address   *Address  `json:"address"`
	Skills    []string  `json:"skills"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewProfile creates a new user profile
func NewProfile(userID int) *Profile {
	return &Profile{
		UserID:    userID,
		Skills:    make([]string, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// AddSkill adds a skill to the profile
func (p *Profile) AddSkill(skill string) {
	p.Skills = append(p.Skills, skill)
	p.UpdatedAt = time.Now()
}

// GetSkillCount returns the number of skills
func (p *Profile) GetSkillCount() int {
	return len(p.Skills)
}

// Repository interface for data access
type Repository interface {
	Save(entity interface{}) error
	FindByID(id int) (interface{}, error)
	FindAll() ([]interface{}, error)
	Delete(id int) error
}

// CacheManager handles caching operations
type CacheManager interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
}

// NotificationService handles notifications
type NotificationService interface {
	SendEmail(to, subject, body string) error
	SendSMS(to, message string) error
	SendPush(userID int, message string) error
}

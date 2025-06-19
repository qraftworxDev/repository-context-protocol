package main

import (
	"fmt"
	"time"
)

// Additional constants for testing constant searches
const (
	TestConstant    = "test_value"
	MAX_BUFFER_SIZE = 1024
	API_VERSION     = "v1.2.3"
	ENABLE_LOGGING  = true
)

// Additional variables for testing variable searches
var (
	testVariable      string
	bufferSize        int    = MAX_BUFFER_SIZE
	apiEndpoint       string = "https://api.example.com"
	featureFlags      map[string]bool
	connectionTimeout time.Duration = 5 * time.Second
)

// QueryEngine type for testing type searches
type QueryEngine struct {
	Name    string
	Version string
	Config  *Config
}

// NewQueryEngine creates a new query engine
func NewQueryEngine(name string) *QueryEngine {
	return &QueryEngine{
		Name:    name,
		Version: API_VERSION,
		Config:  NewConfig(),
	}
}

// SearchResult represents search results
type SearchResult struct {
	ID          int                    `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Score       float64                `json:"score"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewSearchResult creates a new search result
func NewSearchResult(id int, title string) *SearchResult {
	return &SearchResult{
		ID:        id,
		Title:     title,
		Score:     0.0,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// TestService interface for testing
type TestService interface {
	TestMethod(input string) (string, error)
	ProcessData(data []byte) error
	GetResults() ([]*SearchResult, error)
}

// MockTestService implements TestService for testing
type MockTestService struct {
	results []*SearchResult
}

// NewMockTestService creates a mock test service
func NewMockTestService() *MockTestService {
	return &MockTestService{
		results: make([]*SearchResult, 0),
	}
}

// TestMethod implements TestService.TestMethod
func (m *MockTestService) TestMethod(input string) (string, error) {
	return fmt.Sprintf("processed: %s", input), nil
}

// ProcessData implements TestService.ProcessData
func (m *MockTestService) ProcessData(data []byte) error {
	// This function calls CreateTestResult
	result := m.CreateTestResult(string(data))
	m.results = append(m.results, result)
	return nil
}

// GetResults implements TestService.GetResults
func (m *MockTestService) GetResults() ([]*SearchResult, error) {
	return m.results, nil
}

// CreateTestResult creates a test result (for call graph testing)
func (m *MockTestService) CreateTestResult(data string) *SearchResult {
	// This function calls NewSearchResult
	result := NewSearchResult(len(m.results)+1, data)
	result.Description = fmt.Sprintf("Test result for: %s", data)
	return result
}

// Pattern matching functions for testing search patterns

// QueryFunction1 - matches Query* pattern
func QueryFunction1() string {
	return "query function 1"
}

// QueryFunction2 - matches Query* pattern
func QueryFunction2() int {
	return 2
}

// ProcessQuery - matches *Query pattern
func ProcessQuery(query string) error {
	return nil
}

// ExecuteQuery - matches *Query pattern
func ExecuteQuery() *SearchResult {
	return NewSearchResult(999, "executed query")
}

// TestEngine - matches Test* pattern
type TestEngine struct {
	Name string
}

// TestRunner - matches Test* pattern
type TestRunner struct {
	Engine *TestEngine
}

// NewTestEngine creates a test engine
func NewTestEngine() *TestEngine {
	return &TestEngine{Name: "test"}
}

// NewTestRunner creates a test runner
func NewTestRunner() *TestRunner {
	// This function calls NewTestEngine
	return &TestRunner{
		Engine: NewTestEngine(),
	}
}

// Functions with clear call relationships for depth testing

// Level1Function calls Level2Function
func Level1Function() string {
	return Level2Function()
}

// Level2Function calls Level3Function
func Level2Function() string {
	return Level3Function()
}

// Level3Function calls Level4Function
func Level3Function() string {
	return Level4Function()
}

// Level4Function calls Level5Function
func Level4Function() string {
	return Level5Function()
}

// Level5Function - end of chain
func Level5Function() string {
	return "level 5 reached"
}

// Functions that call each other for complex call graph testing

// CircularA calls CircularB (creates circular reference)
func CircularA() {
	// In real code this would create infinite recursion
	// but for testing call graphs it's useful
	CircularB()
}

// CircularB calls CircularA
func CircularB() {
	// Circular reference for testing
	CircularA()
}

// UtilityFunctions - functions that are called by many others

// LogMessage is called by many functions
func LogMessage(message string) {
	fmt.Printf("[LOG] %s\n", message)
}

// ValidateInput is called by many functions
func ValidateInput(input string) bool {
	return input != ""
}

// Functions that use the utility functions

// ProcessInput processes input with validation and logging
func ProcessInput(input string) error {
	// This function calls ValidateInput and LogMessage
	LogMessage("Processing input")

	if !ValidateInput(input) {
		LogMessage("Invalid input")
		return fmt.Errorf("invalid input")
	}

	LogMessage("Input processed successfully")
	return nil
}

// HandleRequest handles a request with logging
func HandleRequest(request string) string {
	// This function calls LogMessage and ValidateInput
	LogMessage("Handling request")

	if ValidateInput(request) {
		LogMessage("Request is valid")
		return "handled: " + request
	}

	LogMessage("Request is invalid")
	return "error: invalid request"
}

// Integration function that ties everything together
func RunDemo() {
	// This function calls many other functions, good for testing include-callees
	LogMessage("Starting demo")

	// Test query engine
	engine := NewQueryEngine("demo-engine")
	LogMessage(fmt.Sprintf("Created engine: %s", engine.Name))

	// Test service
	service := NewMockTestService()
	result, _ := service.TestMethod("demo data")
	LogMessage(result)

	// Test levels
	levelResult := Level1Function()
	LogMessage(levelResult)

	// Test processing
	err := ProcessInput("test input")
	if err != nil {
		LogMessage(fmt.Sprintf("Process error: %v", err))
	}

	// Test request handling
	response := HandleRequest("test request")
	LogMessage(response)

	LogMessage("Demo completed")
}

// Additional types for comprehensive type testing

// DataProcessor interface
type DataProcessor interface {
	Process(data interface{}) error
	GetStatus() string
}

// SimpleProcessor implements DataProcessor
type SimpleProcessor struct {
	Status string
}

// Process implements DataProcessor.Process
func (p *SimpleProcessor) Process(data interface{}) error {
	p.Status = "processed"
	return nil
}

// GetStatus implements DataProcessor.GetStatus
func (p *SimpleProcessor) GetStatus() string {
	return p.Status
}

// Complex nested type for testing
type ComplexType struct {
	ID       int
	Data     map[string]interface{}
	Children []*ComplexType
	Parent   *ComplexType
	Metadata struct {
		Created time.Time
		Tags    []string
	}
}

// NewComplexType creates a complex type
func NewComplexType(id int) *ComplexType {
	return &ComplexType{
		ID:       id,
		Data:     make(map[string]interface{}),
		Children: make([]*ComplexType, 0),
	}
}

// Test functions for generateRandomBytes and GenerateID

// TestGenerateRandomBytes tests the generateRandomBytes function
func TestGenerateRandomBytes() {
	fmt.Println("Testing generateRandomBytes...")

	// Test with different lengths
	testCases := []int{0, 1, 8, 16, 32, 64}

	for _, length := range testCases {
		bytes := generateRandomBytes(length)
		if len(bytes) != length {
			fmt.Printf("FAIL: generateRandomBytes(%d) returned %d bytes, expected %d\n", length, len(bytes), length)
			return
		}
		fmt.Printf("PASS: generateRandomBytes(%d) returned %d bytes\n", length, len(bytes))
	}

	// Test that different calls produce different results (for non-zero lengths)
	if len(testCases) > 1 {
		bytes1 := generateRandomBytes(8)
		bytes2 := generateRandomBytes(8)

		// Check they're different (very high probability)
		same := true
		for i := 0; i < len(bytes1); i++ {
			if bytes1[i] != bytes2[i] {
				same = false
				break
			}
		}

		if same {
			fmt.Println("WARN: generateRandomBytes(8) produced identical results (possible but unlikely)")
		} else {
			fmt.Println("PASS: generateRandomBytes(8) produced different results")
		}
	}
}

// TestGenerateID tests the GenerateID function
func TestGenerateID() {
	fmt.Println("Testing GenerateID...")

	// Test that GenerateID returns non-empty string
	id1 := GenerateID()
	if id1 == "" {
		fmt.Println("FAIL: GenerateID() returned empty string")
		return
	}
	fmt.Printf("PASS: GenerateID() returned: %s\n", id1)

	// Test that different calls produce different IDs
	id2 := GenerateID()
	if id1 == id2 {
		fmt.Println("WARN: GenerateID() produced identical results (possible but unlikely)")
	} else {
		fmt.Println("PASS: GenerateID() produced different results")
	}

	// Test expected length (8 bytes = 16 hex characters)
	expectedLength := 16
	if len(id1) != expectedLength {
		fmt.Printf("FAIL: GenerateID() returned %d characters, expected %d\n", len(id1), expectedLength)
		return
	}
	fmt.Printf("PASS: GenerateID() returned expected length: %d\n", len(id1))
}

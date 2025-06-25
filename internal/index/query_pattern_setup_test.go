package index

import (
	"fmt"
	"testing"
	"time"

	"repository-context-protocol/internal/models"
)

// setupComplexPatternTestData creates comprehensive test data for complex pattern matching scenarios
func SetupComplexPatternTestData(t *testing.T, storage *HybridStorage) {
	// Create a comprehensive file context with various naming patterns
	fileContext := &models.FileContext{
		Path:     "complex_patterns.go",
		Language: "go",
		Checksum: "complex123",
		ModTime:  time.Now(),
		Functions: []models.Function{
			// HTTP Handler functions
			{
				Name:      "HandleUserLogin",
				Signature: "func HandleUserLogin(w http.ResponseWriter, r *http.Request)",
				StartLine: 10,
				EndLine:   20,
				Calls:     []string{"ValidateCredentials", "GenerateToken"},
			},
			{
				Name:      "HandleUserLogout",
				Signature: "func HandleUserLogout(w http.ResponseWriter, r *http.Request)",
				StartLine: 25,
				EndLine:   35,
				Calls:     []string{"InvalidateToken"},
			},
			{
				Name:      "HandleAPIRequest",
				Signature: "func HandleAPIRequest(w http.ResponseWriter, r *http.Request)",
				StartLine: 40,
				EndLine:   50,
				Calls:     []string{"AuthenticateRequest", "ProcessRequest"},
			},

			// Service layer functions
			{
				Name:      "ProcessUserData",
				Signature: "func ProcessUserData(data *UserData) error",
				StartLine: 55,
				EndLine:   65,
				Calls:     []string{"ValidateUserData", "SaveUserData"},
			},
			{
				Name:      "ProcessPaymentData",
				Signature: "func ProcessPaymentData(payment *Payment) error",
				StartLine: 70,
				EndLine:   80,
				Calls:     []string{"ValidatePayment", "ChargePayment"},
			},

			// Validation functions
			{
				Name:      "ValidateCredentials",
				Signature: "func ValidateCredentials(username, password string) bool",
				StartLine: 85,
				EndLine:   95,
				CalledBy:  []string{"HandleUserLogin"},
			},
			{
				Name:      "ValidateUserData",
				Signature: "func ValidateUserData(data *UserData) error",
				StartLine: 100,
				EndLine:   110,
				CalledBy:  []string{"ProcessUserData"},
			},
			{
				Name:      "ValidatePayment",
				Signature: "func ValidatePayment(payment *Payment) error",
				StartLine: 115,
				EndLine:   125,
				CalledBy:  []string{"ProcessPaymentData"},
			},

			// Utility functions with different naming patterns
			{
				Name:      "generateUUID",
				Signature: "func generateUUID() string",
				StartLine: 130,
				EndLine:   135,
			},
			{
				Name:      "parseJSONData",
				Signature: "func parseJSONData(data []byte) (map[string]interface{}, error)",
				StartLine: 140,
				EndLine:   150,
			},
			{
				Name:      "convertToXML",
				Signature: "func convertToXML(data interface{}) ([]byte, error)",
				StartLine: 155,
				EndLine:   165,
			},

			// Database functions
			{
				Name:      "ConnectDB",
				Signature: "func ConnectDB() (*sql.DB, error)",
				StartLine: 170,
				EndLine:   180,
			},
			{
				Name:      "QueryUsers",
				Signature: "func QueryUsers(filter UserFilter) ([]User, error)",
				StartLine: 185,
				EndLine:   195,
			},
			{
				Name:      "QueryPayments",
				Signature: "func QueryPayments(userID int) ([]Payment, error)",
				StartLine: 200,
				EndLine:   210,
			},
		},
		Types: []models.TypeDef{
			{
				Name:      "UserData",
				Kind:      "struct",
				StartLine: 215,
				EndLine:   225,
			},
			{
				Name:      "PaymentData",
				Kind:      "struct",
				StartLine: 230,
				EndLine:   240,
			},
			{
				Name:      "APIResponse",
				Kind:      "struct",
				StartLine: 245,
				EndLine:   255,
			},
			{
				Name:      "UserValidator",
				Kind:      "interface",
				StartLine: 260,
				EndLine:   265,
			},
			{
				Name:      "PaymentProcessor",
				Kind:      "interface",
				StartLine: 270,
				EndLine:   275,
			},
			{
				Name:      "ResponseWriter",
				Kind:      "interface",
				StartLine: 280,
				EndLine:   285,
			},
		},
		Variables: []models.Variable{
			{
				Name:      "defaultTimeout",
				Type:      "time.Duration",
				StartLine: 290,
				EndLine:   290,
			},
			{
				Name:      "maxRetryCount",
				Type:      "int",
				StartLine: 291,
				EndLine:   291,
			},
			{
				Name:      "dbConnectionString",
				Type:      "string",
				StartLine: 292,
				EndLine:   292,
			},
		},
		Constants: []models.Constant{
			{
				Name:      "APIVersion",
				Type:      "string",
				Value:     "v1.0",
				StartLine: 295,
				EndLine:   295,
			},
			{
				Name:      "MaxPaymentAmount",
				Type:      "float64",
				Value:     "10000.0",
				StartLine: 296,
				EndLine:   296,
			},
			{
				Name:      "DefaultUserRole",
				Type:      "string",
				Value:     "user",
				StartLine: 297,
				EndLine:   297,
			},
		},
	}

	err := storage.StoreFileContext(fileContext)
	if err != nil {
		t.Fatalf("Failed to store complex pattern test data: %v", err)
	}
}

// setupRegexTestCases creates test data specifically for regex edge cases and complex patterns
func SetupRegexTestCases(t *testing.T, storage *HybridStorage) {
	fileContext := &models.FileContext{
		Path:     "regex_edge_cases.go",
		Language: "go",
		Checksum: "regex456",
		ModTime:  time.Now(),
		Functions: []models.Function{
			// Functions with special characters and patterns
			{
				Name:      "init",
				Signature: "func init()",
				StartLine: 5,
				EndLine:   10,
			},
			{
				Name:      "String",
				Signature: "func (u User) String() string",
				StartLine: 15,
				EndLine:   20,
			},
			{
				Name:      "GoString",
				Signature: "func (u User) GoString() string",
				StartLine: 25,
				EndLine:   30,
			},
			{
				Name:      "MarshalJSON",
				Signature: "func (u User) MarshalJSON() ([]byte, error)",
				StartLine: 35,
				EndLine:   45,
			},
			{
				Name:      "UnmarshalJSON",
				Signature: "func (u *User) UnmarshalJSON(data []byte) error",
				StartLine: 50,
				EndLine:   60,
			},

			// Functions with numbers and underscores
			{
				Name:      "processV1Data",
				Signature: "func processV1Data(data []byte) error",
				StartLine: 65,
				EndLine:   75,
			},
			{
				Name:      "processV2Data",
				Signature: "func processV2Data(data []byte) error",
				StartLine: 80,
				EndLine:   90,
			},
			{
				Name:      "handle_legacy_format",
				Signature: "func handle_legacy_format(input string) string",
				StartLine: 95,
				EndLine:   105,
			},
			{
				Name:      "parse_config_file",
				Signature: "func parse_config_file(filename string) (*Config, error)",
				StartLine: 110,
				EndLine:   120,
			},
		},
		Types: []models.TypeDef{
			{
				Name:      "User",
				Kind:      "struct",
				StartLine: 120,
				EndLine:   124,
			},
			{
				Name:      "HTTPClient",
				Kind:      "struct",
				StartLine: 125,
				EndLine:   135,
			},
			{
				Name:      "XMLParser",
				Kind:      "struct",
				StartLine: 140,
				EndLine:   150,
			},
			{
				Name:      "JSONEncoder",
				Kind:      "struct",
				StartLine: 155,
				EndLine:   165,
			},
		},
	}

	err := storage.StoreFileContext(fileContext)
	if err != nil {
		t.Fatalf("Failed to store regex test case data: %v", err)
	}
}

// setupPerformanceTestData creates a large dataset for performance testing of pattern matching
func SetupPerformanceTestData(t *testing.T, storage *HybridStorage) {
	fileContext := &models.FileContext{
		Path:      "performance_test.go",
		Language:  "go",
		Checksum:  "perf789",
		ModTime:   time.Now(),
		Functions: make([]models.Function, 0, 100),
		Types:     make([]models.TypeDef, 0, 50),
		Variables: make([]models.Variable, 0, 30),
		Constants: make([]models.Constant, 0, 20),
	}

	// Generate 100 functions with various naming patterns
	for i := 0; i < 100; i++ {
		var name string
		switch i % 5 {
		case 0:
			name = fmt.Sprintf("HandleRequest%d", i)
		case 1:
			name = fmt.Sprintf("ProcessData%d", i)
		case 2:
			name = fmt.Sprintf("ValidateInput%d", i)
		case 3:
			name = fmt.Sprintf("GenerateReport%d", i)
		case 4:
			name = fmt.Sprintf("ExecuteQuery%d", i)
		}

		fileContext.Functions = append(fileContext.Functions, models.Function{
			Name:      name,
			Signature: fmt.Sprintf("func %s() error", name),
			StartLine: i*10 + 1,
			EndLine:   i*10 + 5,
		})
	}

	// Generate 50 types
	for i := 0; i < 50; i++ {
		var name string
		switch i % 3 {
		case 0:
			name = fmt.Sprintf("DataModel%d", i)
		case 1:
			name = fmt.Sprintf("ServiceClient%d", i)
		case 2:
			name = fmt.Sprintf("RequestHandler%d", i)
		}

		fileContext.Types = append(fileContext.Types, models.TypeDef{
			Name:      name,
			Kind:      "struct",
			StartLine: 1000 + i*10,
			EndLine:   1000 + i*10 + 5,
		})
	}

	err := storage.StoreFileContext(fileContext)
	if err != nil {
		t.Fatalf("Failed to store performance test data: %v", err)
	}
}

// setupConcurrencyTestData creates test data for testing concurrent regex compilation and caching
func SetupConcurrencyTestData(t *testing.T, storage *HybridStorage) {
	// Create multiple files with overlapping function names for concurrency testing
	for fileIndex := 0; fileIndex < 5; fileIndex++ {
		fileContext := &models.FileContext{
			Path:      fmt.Sprintf("concurrent_test_%d.go", fileIndex),
			Language:  "go",
			Checksum:  fmt.Sprintf("concurrent%d", fileIndex),
			ModTime:   time.Now(),
			Functions: make([]models.Function, 0, 20),
		}

		for funcIndex := 0; funcIndex < 20; funcIndex++ {
			name := fmt.Sprintf("ConcurrentFunction_%d_%d", fileIndex, funcIndex)
			fileContext.Functions = append(fileContext.Functions, models.Function{
				Name:      name,
				Signature: fmt.Sprintf("func %s() error", name),
				StartLine: funcIndex*5 + 1,
				EndLine:   funcIndex*5 + 3,
			})
		}

		err := storage.StoreFileContext(fileContext)
		if err != nil {
			t.Fatalf("Failed to store concurrency test data for file %d: %v", fileIndex, err)
		}
	}
}

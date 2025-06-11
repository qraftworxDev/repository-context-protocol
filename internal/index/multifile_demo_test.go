package index

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"repository-context-protocol/internal/models"
)

func TestMultiFileDemo_GlobalCallGraphEnrichment(t *testing.T) {
	// Create a realistic multi-file scenario showing global call graph capabilities
	fileContexts := []models.FileContext{
		{
			Path:     "cmd/main.go",
			Language: "go",
			Checksum: "main_abc123",
			ModTime:  time.Now(),
			Functions: []models.Function{
				{
					Name:      "main",
					Signature: "func main()",
					StartLine: 10,
					EndLine:   20,
					Calls:     []string{"StartApp", "fmt.Println"},
					CalledBy:  []string{},
				},
			},
		},
		{
			Path:     "internal/service/app.go",
			Language: "go",
			Checksum: "service_def456",
			ModTime:  time.Now(),
			Functions: []models.Function{
				{
					Name:      "StartApp",
					Signature: "func StartApp() error",
					StartLine: 15,
					EndLine:   25,
					Calls:     []string{"InitDatabase", "StartWebServer"},
					CalledBy:  []string{"main"},
				},
				{
					Name:      "InitDatabase",
					Signature: "func InitDatabase() error",
					StartLine: 30,
					EndLine:   40,
					Calls:     []string{"db.Connect", "db.Migrate"},
					CalledBy:  []string{"StartApp"},
				},
			},
		},
		{
			Path:     "internal/service/web.go",
			Language: "go",
			Checksum: "web_ghi789",
			ModTime:  time.Now(),
			Functions: []models.Function{
				{
					Name:      "StartWebServer",
					Signature: "func StartWebServer() error",
					StartLine: 10,
					EndLine:   20,
					Calls:     []string{"RegisterRoutes", "server.Listen"},
					CalledBy:  []string{"StartApp"},
				},
				{
					Name:      "RegisterRoutes",
					Signature: "func RegisterRoutes()",
					StartLine: 25,
					EndLine:   35,
					Calls:     []string{"HandleUser", "HandleHealth"},
					CalledBy:  []string{"StartWebServer"},
				},
				{
					Name:      "HandleUser",
					Signature: "func HandleUser(w http.ResponseWriter, r *http.Request)",
					StartLine: 40,
					EndLine:   50,
					Calls:     []string{"CreateUser", "ValidateUser"},
					CalledBy:  []string{"RegisterRoutes"},
				},
				{
					Name:      "HandleHealth",
					Signature: "func HandleHealth(w http.ResponseWriter, r *http.Request)",
					StartLine: 55,
					EndLine:   60,
					Calls:     []string{"json.NewEncoder"},
					CalledBy:  []string{"RegisterRoutes"},
				},
			},
		},
		{
			Path:     "internal/user/user.go",
			Language: "go",
			Checksum: "user_jkl012",
			ModTime:  time.Now(),
			Functions: []models.Function{
				{
					Name:      "CreateUser",
					Signature: "func CreateUser(name string) *User",
					StartLine: 20,
					EndLine:   25,
					Calls:     []string{"ValidateUser"},
					CalledBy:  []string{"HandleUser"},
				},
				{
					Name:      "ValidateUser",
					Signature: "func ValidateUser(user *User) bool",
					StartLine: 30,
					EndLine:   40,
					Calls:     []string{"strings.TrimSpace"},
					CalledBy:  []string{"CreateUser", "HandleUser"},
				},
			},
		},
	}

	// Create enrichment processor and enrich the contexts
	enrichment := NewGlobalEnrichment()
	enrichedContexts, err := enrichment.EnrichFileContexts(fileContexts)
	if err != nil {
		t.Fatalf("Failed to enrich file contexts: %v", err)
	}

	// Convert to JSON for readable output showing the enriched global relationships
	for i, context := range enrichedContexts {
		jsonData, err := json.MarshalIndent(context, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal context %d to JSON: %v", i, err)
		}
		fmt.Printf("\n=== ENRICHED FILE CONTEXT: %s ===\n%s\n", context.Path, string(jsonData))
	}

	// Test 1: main function should have cross-file calls only
	mainFile := findFileContext(enrichedContexts, "cmd/main.go")
	if mainFile == nil {
		t.Fatal("Could not find cmd/main.go")
	}
	mainFunc := findFunction(mainFile.Functions, "main")
	if mainFunc == nil {
		t.Fatal("Could not find main function")
	}

	// Main should have no local calls (all calls go to other files)
	if len(mainFunc.LocalCalls) != 0 {
		t.Errorf("Expected main to have 0 local calls, got %d", len(mainFunc.LocalCalls))
	}

	// Main should have 2 cross-file calls: StartApp and fmt.Println
	if len(mainFunc.CrossFileCalls) != 2 {
		t.Errorf("Expected main to have 2 cross-file calls, got %d", len(mainFunc.CrossFileCalls))
	}

	// Verify specific cross-file call to StartApp
	startAppCall := findCallReference(mainFunc.CrossFileCalls, "StartApp")
	if startAppCall == nil {
		t.Error("Expected main to have cross-file call to StartApp")
	} else if startAppCall.File != "internal/service/app.go" {
		t.Errorf("Expected StartApp call to reference internal/service/app.go, got %s", startAppCall.File)
	}

	// Test 2: StartApp should have mixed local and cross-file calls
	serviceFile := findFileContext(enrichedContexts, "internal/service/app.go")
	if serviceFile == nil {
		t.Fatal("Could not find internal/service/app.go")
	}
	startAppFunc := findFunction(serviceFile.Functions, "StartApp")
	if startAppFunc == nil {
		t.Fatal("Could not find StartApp function")
	}

	// StartApp should have 1 local call (InitDatabase) and 1 cross-file call (StartWebServer)
	if len(startAppFunc.LocalCalls) != 1 {
		t.Errorf("Expected StartApp to have 1 local call, got %d", len(startAppFunc.LocalCalls))
	}
	if startAppFunc.LocalCalls[0] != "InitDatabase" {
		t.Errorf("Expected local call to InitDatabase, got %s", startAppFunc.LocalCalls[0])
	}

	if len(startAppFunc.CrossFileCalls) != 1 {
		t.Errorf("Expected StartApp to have 1 cross-file call, got %d", len(startAppFunc.CrossFileCalls))
	}
	webServerCall := findCallReference(startAppFunc.CrossFileCalls, "StartWebServer")
	if webServerCall == nil {
		t.Error("Expected StartApp to have cross-file call to StartWebServer")
	} else if webServerCall.File != "internal/service/web.go" {
		t.Errorf("Expected StartWebServer call to reference internal/service/web.go, got %s", webServerCall.File)
	}

	// Test 3: ValidateUser should be called from multiple files
	userFile := findFileContext(enrichedContexts, "internal/user/user.go")
	if userFile == nil {
		t.Fatal("Could not find internal/user/user.go")
	}
	validateUserFunc := findFunction(userFile.Functions, "ValidateUser")
	if validateUserFunc == nil {
		t.Fatal("Could not find ValidateUser function")
	}

	// ValidateUser should have 1 local caller (CreateUser) and 1 cross-file caller (HandleUser)
	if len(validateUserFunc.LocalCallers) != 1 {
		t.Errorf("Expected ValidateUser to have 1 local caller, got %d", len(validateUserFunc.LocalCallers))
	}
	if validateUserFunc.LocalCallers[0] != "CreateUser" {
		t.Errorf("Expected local caller CreateUser, got %s", validateUserFunc.LocalCallers[0])
	}

	if len(validateUserFunc.CrossFileCallers) != 1 {
		t.Errorf("Expected ValidateUser to have 1 cross-file caller, got %d", len(validateUserFunc.CrossFileCallers))
	}

	if len(validateUserFunc.CrossFileCallers) > 0 {
		handleUserCaller := validateUserFunc.CrossFileCallers[0]
		if handleUserCaller.FunctionName != "HandleUser" {
			t.Errorf("Expected cross-file caller HandleUser, got %s", handleUserCaller.FunctionName)
		}
		if handleUserCaller.File != "internal/service/web.go" {
			t.Errorf("Expected cross-file caller from internal/service/web.go, got %s", handleUserCaller.File)
		}
	}

	// Test 4: RegisterRoutes should have only local calls within web.go
	webFile := findFileContext(enrichedContexts, "internal/service/web.go")
	if webFile == nil {
		t.Fatal("Could not find internal/service/web.go")
	}
	registerRoutesFunc := findFunction(webFile.Functions, "RegisterRoutes")
	if registerRoutesFunc == nil {
		t.Fatal("Could not find RegisterRoutes function")
	}

	// RegisterRoutes should have 2 local calls (HandleUser, HandleHealth)
	if len(registerRoutesFunc.LocalCalls) != 2 {
		t.Errorf("Expected RegisterRoutes to have 2 local calls, got %d", len(registerRoutesFunc.LocalCalls))
	}

	// Test 5: Global call graph statistics
	globalCallGraph := enrichment.GetGlobalCallGraph()
	stats := globalCallGraph.GetStatistics()

	fmt.Printf("\n=== GLOBAL CALL GRAPH STATISTICS ===\n")
	fmt.Printf("Total Functions: %d\n", stats.TotalFunctions)
	fmt.Printf("Total Call Relations: %d\n", stats.TotalCallRelations)
	fmt.Printf("Max Call Depth: %d\n", stats.MaxCallDepth)

	// Verify we have all expected functions
	// Expected: main, StartApp, InitDatabase, StartWebServer, RegisterRoutes, HandleUser, HandleHealth, CreateUser, ValidateUser
	expectedMinFunctions := 8
	if stats.TotalFunctions < expectedMinFunctions {
		t.Errorf("Expected at least %d functions in global call graph, got %d", expectedMinFunctions, stats.TotalFunctions)
	}

	// Test 6: Call path analysis
	callPath := globalCallGraph.GetCallPath("main", "ValidateUser")
	fmt.Printf("\nCall path from main to ValidateUser: %v\n", callPath)

	if len(callPath) == 0 {
		t.Error("Expected to find a call path from main to ValidateUser")
	}

	// The path should start with main and end with ValidateUser
	if len(callPath) > 0 && callPath[0] != "main" {
		t.Errorf("Expected call path to start with 'main', got %s", callPath[0])
	}
	if len(callPath) > 0 && callPath[len(callPath)-1] != "ValidateUser" {
		t.Errorf("Expected call path to end with 'ValidateUser', got %s", callPath[len(callPath)-1])
	}
}

func TestMultiFileDemo_CompareWithSingleFile(t *testing.T) {
	// Create a scenario that shows the difference between single-file parsing and global enrichment

	// Single file context (what parser would produce alone)
	singleFileContext := models.FileContext{
		Path:     "main.go",
		Language: "go",
		Checksum: "single123",
		ModTime:  time.Now(),
		Functions: []models.Function{
			{
				Name:      "main",
				Signature: "func main()",
				StartLine: 5,
				EndLine:   10,
				Calls:     []string{"ProcessData"}, // Parser can see the call but not that it's cross-file
				CalledBy:  []string{},
			},
		},
	}

	// Multi-file contexts for enrichment
	multiFileContexts := []models.FileContext{
		singleFileContext,
		{
			Path:     "processor.go",
			Language: "go",
			Checksum: "processor123",
			ModTime:  time.Now(),
			Functions: []models.Function{
				{
					Name:      "ProcessData",
					Signature: "func ProcessData(data string) error",
					StartLine: 10,
					EndLine:   20,
					Calls:     []string{},
					CalledBy:  []string{"main"},
				},
			},
		},
	}

	// Test single-file context (baseline)
	mainFunc := singleFileContext.Functions[0]

	fmt.Printf("\n=== BEFORE ENRICHMENT (Single File Parser View) ===\n")
	fmt.Printf("main.Calls: %v\n", mainFunc.Calls)
	fmt.Printf("main.LocalCalls: %v (empty - not yet categorized)\n", mainFunc.LocalCalls)
	fmt.Printf("main.CrossFileCalls: %v (empty - not yet categorized)\n", mainFunc.CrossFileCalls)

	// Test enriched contexts
	enrichment := NewGlobalEnrichment()
	enrichedContexts, err := enrichment.EnrichFileContexts(multiFileContexts)
	if err != nil {
		t.Fatalf("Failed to enrich contexts: %v", err)
	}

	enrichedMainFile := findFileContext(enrichedContexts, "main.go")
	if enrichedMainFile == nil {
		t.Fatal("Could not find enriched main.go")
	}
	enrichedMainFunc := findFunction(enrichedMainFile.Functions, "main")
	if enrichedMainFunc == nil {
		t.Fatal("Could not find enriched main function")
	}

	fmt.Printf("\n=== AFTER ENRICHMENT (Global Analysis) ===\n")
	fmt.Printf("main.Calls: %v\n", enrichedMainFunc.Calls)
	fmt.Printf("main.LocalCalls: %v\n", enrichedMainFunc.LocalCalls)
	fmt.Printf("main.CrossFileCalls: %v\n", enrichedMainFunc.CrossFileCalls)

	// Verify the enrichment correctly identified the cross-file call
	if len(enrichedMainFunc.LocalCalls) != 0 {
		t.Errorf("Expected 0 local calls after enrichment, got %d", len(enrichedMainFunc.LocalCalls))
	}
	if len(enrichedMainFunc.CrossFileCalls) != 1 {
		t.Errorf("Expected 1 cross-file call after enrichment, got %d", len(enrichedMainFunc.CrossFileCalls))
	}

	if len(enrichedMainFunc.CrossFileCalls) > 0 {
		crossFileCall := enrichedMainFunc.CrossFileCalls[0]
		if crossFileCall.FunctionName != "ProcessData" {
			t.Errorf("Expected cross-file call to ProcessData, got %s", crossFileCall.FunctionName)
		}
		if crossFileCall.File != "processor.go" {
			t.Errorf("Expected cross-file call to reference processor.go, got %s", crossFileCall.File)
		}
	}
}

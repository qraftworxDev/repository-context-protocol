package models

import "testing"

func TestFunction_Creation(t *testing.T) {
	fn := &Function{
		Name:      "greet",
		Signature: "func greet(name string) string",
		Parameters: []Parameter{
			{Name: "name", Type: "string"},
		},
		Returns: []Type{
			{Name: "string", Kind: "basic"},
		},
		StartLine: 10,
		EndLine:   12,
		Calls:     []string{"fmt.Sprintf"},
		CalledBy:  []string{"main"},
	}

	if fn.Name != "greet" {
		t.Errorf("Expected name 'greet', got %s", fn.Name)
	}
	if len(fn.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(fn.Parameters))
	}
	if fn.Parameters[0].Name != "name" {
		t.Errorf("Expected parameter name 'name', got %s", fn.Parameters[0].Name)
	}
	if len(fn.Returns) != 1 {
		t.Errorf("Expected 1 return type, got %d", len(fn.Returns))
	}
	if fn.StartLine != 10 {
		t.Errorf("Expected start line 10, got %d", fn.StartLine)
	}
}

func TestParameter_Creation(t *testing.T) {
	param := Parameter{
		Name: "ctx",
		Type: "context.Context",
	}

	if param.Name != "ctx" {
		t.Errorf("Expected name 'ctx', got %s", param.Name)
	}
	if param.Type != "context.Context" {
		t.Errorf("Expected type 'context.Context', got %s", param.Type)
	}
}

func TestVariable_Creation(t *testing.T) {
	variable := Variable{
		Name:      "result",
		Type:      "string",
		StartLine: 5,
		EndLine:   5,
	}

	if variable.Name != "result" {
		t.Errorf("Expected name 'result', got %s", variable.Name)
	}
	if variable.Type != "string" {
		t.Errorf("Expected type 'string', got %s", variable.Type)
	}
}

func TestImport_Creation(t *testing.T) {
	imp := Import{
		Path:  "fmt",
		Alias: "",
	}

	if imp.Path != "fmt" {
		t.Errorf("Expected path 'fmt', got %s", imp.Path)
	}

	// Test with alias
	impWithAlias := Import{
		Path:  "github.com/stretchr/testify/assert",
		Alias: "assert",
	}

	if impWithAlias.Alias != "assert" {
		t.Errorf("Expected alias 'assert', got %s", impWithAlias.Alias)
	}
}

func TestExport_Creation(t *testing.T) {
	export := Export{
		Name: "ProcessUser",
		Type: "func(User) error",
		Kind: "function",
	}

	if export.Name != "ProcessUser" {
		t.Errorf("Expected name 'ProcessUser', got %s", export.Name)
	}
	if export.Kind != "function" {
		t.Errorf("Expected kind 'function', got %s", export.Kind)
	}
}

func TestCallReference_Creation(t *testing.T) {
	ref := CallReference{
		FunctionName: "ProcessUser",
		File:         "user.go",
		Line:         15,
	}

	if ref.FunctionName != "ProcessUser" {
		t.Errorf("Expected function name 'ProcessUser', got %s", ref.FunctionName)
	}
	if ref.File != "user.go" {
		t.Errorf("Expected file 'user.go', got %s", ref.File)
	}
	if ref.Line != 15 {
		t.Errorf("Expected line 15, got %d", ref.Line)
	}
}

func TestFunction_EnhancedCallTracking(t *testing.T) {
	fn := &Function{
		Name:      "main",
		Signature: "func main()",
		StartLine: 10,
		EndLine:   15,
		// All calls (local + cross-file)
		Calls:    []string{"localHelper", "user.CreateUser", "fmt.Println"},
		CalledBy: []string{}, // main is not called by anyone
		// Local calls only
		LocalCalls:   []string{"localHelper"},
		LocalCallers: []string{},
		// Cross-file calls with metadata
		CrossFileCalls: []CallReference{
			{FunctionName: "CreateUser", File: "user.go", Line: 12},
			{FunctionName: "fmt.Println", File: "fmt", Line: 13},
		},
		CrossFileCallers: []CallReference{}, // main is not called by other files
	}

	// Test all calls include both local and cross-file
	if len(fn.Calls) != 3 {
		t.Errorf("Expected 3 total calls, got %d", len(fn.Calls))
	}

	// Test local calls separation
	if len(fn.LocalCalls) != 1 {
		t.Errorf("Expected 1 local call, got %d", len(fn.LocalCalls))
	}
	if fn.LocalCalls[0] != "localHelper" {
		t.Errorf("Expected local call 'localHelper', got %s", fn.LocalCalls[0])
	}

	// Test cross-file calls with metadata
	if len(fn.CrossFileCalls) != 2 {
		t.Errorf("Expected 2 cross-file calls, got %d", len(fn.CrossFileCalls))
	}

	createUserCall := fn.CrossFileCalls[0]
	if createUserCall.FunctionName != "CreateUser" {
		t.Errorf("Expected cross-file call 'CreateUser', got %s", createUserCall.FunctionName)
	}
	if createUserCall.File != "user.go" {
		t.Errorf("Expected cross-file call file 'user.go', got %s", createUserCall.File)
	}
	if createUserCall.Line != 12 {
		t.Errorf("Expected cross-file call line 12, got %d", createUserCall.Line)
	}
}

func TestFunction_CrossFileCallers(t *testing.T) {
	fn := &Function{
		Name:      "CreateUser",
		Signature: "func CreateUser(name string) *User",
		StartLine: 5,
		EndLine:   8,
		// Called by functions in other files
		CalledBy:     []string{"main", "TestCreateUser"},
		LocalCallers: []string{"TestCreateUser"}, // Local test function
		CrossFileCallers: []CallReference{
			{FunctionName: "main", File: "main.go", Line: 12},
		},
	}

	// Test cross-file callers
	if len(fn.CrossFileCallers) != 1 {
		t.Errorf("Expected 1 cross-file caller, got %d", len(fn.CrossFileCallers))
	}

	mainCaller := fn.CrossFileCallers[0]
	if mainCaller.FunctionName != "main" {
		t.Errorf("Expected cross-file caller 'main', got %s", mainCaller.FunctionName)
	}
	if mainCaller.File != "main.go" {
		t.Errorf("Expected cross-file caller file 'main.go', got %s", mainCaller.File)
	}
}

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

// Enhanced tests for new CallReference validation and backward compatibility

func TestCallReference_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		cr       CallReference
		expected bool
	}{
		{
			name: "valid with all fields",
			cr: CallReference{
				FunctionName: "TestFunc",
				File:         "/path/to/file.go",
				Line:         42,
				CallType:     CallTypeFunction,
			},
			expected: true,
		},
		{
			name: "valid without CallType",
			cr: CallReference{
				FunctionName: "TestFunc",
				File:         "/path/to/file.go",
				Line:         42,
			},
			expected: true,
		},
		{
			name: "valid with method CallType",
			cr: CallReference{
				FunctionName: "obj.Method",
				File:         "/path/to/file.go",
				Line:         10,
				CallType:     CallTypeMethod,
			},
			expected: true,
		},
		{
			name: "valid external call",
			cr: CallReference{
				FunctionName: "fmt.Println",
				File:         "external",
				Line:         5,
				CallType:     CallTypeExternal,
			},
			expected: true,
		},
		{
			name: "invalid - empty function name",
			cr: CallReference{
				FunctionName: "",
				File:         "/path/to/file.go",
				Line:         42,
				CallType:     CallTypeFunction,
			},
			expected: false,
		},
		{
			name: "invalid - empty file",
			cr: CallReference{
				FunctionName: "TestFunc",
				File:         "",
				Line:         42,
				CallType:     CallTypeFunction,
			},
			expected: false,
		},
		{
			name: "invalid CallType",
			cr: CallReference{
				FunctionName: "TestFunc",
				File:         "/path/to/file.go",
				Line:         42,
				CallType:     "invalid_type",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cr.IsValid()
			if result != tt.expected {
				t.Errorf("CallReference.IsValid() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestFunction_GetAllCalls_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		function Function
		expected []string
	}{
		{
			name: "uses deprecated field when available",
			function: Function{
				Calls:      []string{"oldCall1", "oldCall2"},
				LocalCalls: []string{"newLocal1"},
				CrossFileCalls: []CallReference{
					{FunctionName: "newCross1", File: "other.go"},
				},
			},
			expected: []string{"oldCall1", "oldCall2"},
		},
		{
			name: "builds from enhanced fields when deprecated empty",
			function: Function{
				Calls:      []string{},
				LocalCalls: []string{"local1", "local2"},
				CrossFileCalls: []CallReference{
					{FunctionName: "cross1", File: "file1.go"},
					{FunctionName: "cross2", File: "file2.go"},
				},
			},
			expected: []string{"local1", "local2", "cross1", "cross2"},
		},
		{
			name: "empty when no calls",
			function: Function{
				Calls:          []string{},
				LocalCalls:     []string{},
				CrossFileCalls: []CallReference{},
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function.GetAllCalls()
			if len(result) != len(tt.expected) {
				t.Errorf("GetAllCalls() length = %d, expected %d", len(result), len(tt.expected))
				return
			}
			for i, call := range result {
				if call != tt.expected[i] {
					t.Errorf("GetAllCalls()[%d] = %s, expected %s", i, call, tt.expected[i])
				}
			}
		})
	}
}

func TestFunction_GetAllCallers_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		function Function
		expected []string
	}{
		{
			name: "uses deprecated field when available",
			function: Function{
				CalledBy:     []string{"oldCaller1", "oldCaller2"},
				LocalCallers: []string{"newLocal1"},
				CrossFileCallers: []CallReference{
					{FunctionName: "newCross1", File: "other.go"},
				},
			},
			expected: []string{"oldCaller1", "oldCaller2"},
		},
		{
			name: "builds from enhanced fields when deprecated empty",
			function: Function{
				CalledBy:     []string{},
				LocalCallers: []string{"caller1", "caller2"},
				CrossFileCallers: []CallReference{
					{FunctionName: "crossCaller1", File: "file1.go"},
					{FunctionName: "crossCaller2", File: "file2.go"},
				},
			},
			expected: []string{"caller1", "caller2", "crossCaller1", "crossCaller2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function.GetAllCallers()
			if len(result) != len(tt.expected) {
				t.Errorf("GetAllCallers() length = %d, expected %d", len(result), len(tt.expected))
				return
			}
			for i, caller := range result {
				if caller != tt.expected[i] {
					t.Errorf("GetAllCallers()[%d] = %s, expected %s", i, caller, tt.expected[i])
				}
			}
		})
	}
}

func TestFunction_GetCallsInFile(t *testing.T) {
	function := Function{
		CrossFileCalls: []CallReference{
			{FunctionName: "func1", File: "file1.go", Line: 10},
			{FunctionName: "func2", File: "file2.go", Line: 20},
			{FunctionName: "func3", File: "file1.go", Line: 30},
		},
	}

	result := function.GetCallsInFile("file1.go")
	expected := []CallReference{
		{FunctionName: "func1", File: "file1.go", Line: 10},
		{FunctionName: "func3", File: "file1.go", Line: 30},
	}

	if len(result) != len(expected) {
		t.Errorf("GetCallsInFile() length = %d, expected %d", len(result), len(expected))
		return
	}

	for i, call := range result {
		if call.FunctionName != expected[i].FunctionName || call.File != expected[i].File {
			t.Errorf("GetCallsInFile()[%d] = %+v, expected %+v", i, call, expected[i])
		}
	}
}

// Benchmark tests for performance validation
func BenchmarkFunction_GetAllCalls(b *testing.B) {
	function := Function{
		LocalCalls:     make([]string, 100),
		CrossFileCalls: make([]CallReference, 100),
	}

	// Initialize test data
	for i := 0; i < 100; i++ {
		function.LocalCalls[i] = "localFunc" + string(rune(i))
		function.CrossFileCalls[i] = CallReference{
			FunctionName: "crossFunc" + string(rune(i)),
			File:         "file.go",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = function.GetAllCalls()
	}
}

func BenchmarkCallReference_IsValid(b *testing.B) {
	cr := CallReference{
		FunctionName: "TestFunc",
		File:         "/path/to/file.go",
		Line:         42,
		CallType:     CallTypeFunction,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cr.IsValid()
	}
}

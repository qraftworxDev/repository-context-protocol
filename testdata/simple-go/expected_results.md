# Simple-Go Testdata Expected Results

This document provides expected results for all test commands in the comprehensive CLI test suite when run against the simple-go testdata.

## File Structure Overview

The simple-go testdata now contains:
- `main.go` - Main application with User types and demo functions
- `models.go` - Additional types, constants, and interfaces
- `service.go` - Service layer with call relationships
- `utils.go` - Utility functions with clear call chains
- `test_demo.go` - Additional test functions and patterns
- `go.mod` - Module definition

## Function Search Tests

### Basic Function Searches
```bash
./bin/repocontext query --function "NewQueryEngine"
```
**Expected Result:** Should find the `NewQueryEngine` function in `test_demo.go` at line ~36

```bash
./bin/repocontext query --function "SearchByName"
```
**Expected Result:** Should find the `SearchByName` function in `main.go` at line ~178

```bash
./bin/repocontext query --function "ProcessUser"
```
**Expected Result:** Should find the `ProcessUser` function in `service.go` at line ~28

```bash
./bin/repocontext query --function "nonexistent"
```
**Expected Result:** Should return "No results found" or empty result set

### Function Search with Call Graph Options

```bash
./bin/repocontext query --function "ProcessUser" --include-callees
```
**Expected Callees:**
- `ValidateUser` (called by ProcessUser)
- `IsValidEmail` (called by ValidateUser)
- `CreateUser` (called by ProcessUser)
- `SendWelcomeNotification` (called by ProcessUser)

```bash
./bin/repocontext query --function "ValidateInput" --include-callers
```
**Expected Callers:**
- `ProcessInput` (calls ValidateInput)
- `HandleRequest` (calls ValidateInput)

```bash
./bin/repocontext query --function "Level1Function" --include-callees --depth 3
```
**Expected Call Chain (depth 3):**
- Level1Function → Level2Function → Level3Function → Level4Function

## Type Search Tests

### Basic Type Searches
```bash
./bin/repocontext query --type "QueryEngine"
```
**Expected Result:** Should find the `QueryEngine` struct in `test_demo.go`

```bash
./bin/repocontext query --type "User"
```
**Expected Result:** Should find the `User` struct in `main.go`

```bash
./bin/repocontext query --type "Config"
```
**Expected Result:** Should find the `Config` struct in `models.go`

```bash
./bin/repocontext query --type "UserService"
```
**Expected Result:** Should find the `UserService` interface in `main.go`

### Type Search with Related Functions

```bash
./bin/repocontext query --type "User" --include-types
```
**Expected Related Types:**
- `UserService` (interface that works with User)
- `UserManager` (struct that manages Users)
- `Profile` (related to User)

## Variable Search Tests

### Basic Variable Searches
```bash
./bin/repocontext query --variable "globalConfig"
```
**Expected Result:** Should find the `globalConfig` variable in `models.go`

```bash
./bin/repocontext query --variable "testVariable"
```
**Expected Result:** Should find the `testVariable` in `test_demo.go`

```bash
./bin/repocontext query --variable "bufferSize"
```
**Expected Result:** Should find the `bufferSize` variable in `test_demo.go`

```bash
./bin/repocontext query --variable "emailRegex"
```
**Expected Result:** Should find the `emailRegex` variable in `utils.go`

## Constant Search Tests

### Basic Constant Searches
```bash
./bin/repocontext query --variable "MaxUsers"
```
**Expected Result:** Should find the `MaxUsers` constant in `models.go` (value: 100)

```bash
./bin/repocontext query --variable "API_VERSION"
```
**Expected Result:** Should find the `API_VERSION` constant in `test_demo.go` (value: "v1.2.3")

```bash
./bin/repocontext query --variable "DEBUG_MODE"
```
**Expected Result:** Should find the `DEBUG_MODE` constant in `models.go` (value: true)

## File Search Tests

### Basic File Searches
```bash
./bin/repocontext query --file "main.go"
```
**Expected Result:** Should return contents of `main.go` with main function and User types

```bash
./bin/repocontext query --file "service.go"
```
**Expected Result:** Should return contents of `service.go` with UserManager and service functions

```bash
./bin/repocontext query --file "utils.go"
```
**Expected Result:** Should return contents of `utils.go` with utility functions

```bash
./bin/repocontext query --file "models.go"
```
**Expected Result:** Should return contents of `models.go` with Config, Address, Profile types

## Pattern Search Tests

### Search Pattern Tests (Query* pattern)
```bash
./bin/repocontext query --search "Query*"
```
**Expected Matches:**
- `QueryEngine` (type in test_demo.go)
- `QueryEngineFunction` (function in main.go)
- `QueryFunction1` (function in test_demo.go)
- `QueryFunction2` (function in test_demo.go)

### Search Pattern Tests (*Engine pattern)
```bash
./bin/repocontext query --search "*Engine"
```
**Expected Matches:**
- `QueryEngine` (type in test_demo.go)
- `TestEngine` (type in test_demo.go)

### Search Pattern Tests (Test* pattern)
```bash
./bin/repocontext query --search "Test*"
```
**Expected Matches:**
- `TestFunction` (function in main.go)
- `TestConstant` (constant in test_demo.go)
- `TestService` (interface in test_demo.go)
- `TestEngine` (type in test_demo.go)
- `TestRunner` (type in test_demo.go)
- `TestMethod` (function in test_demo.go)

## Entity Type Search Tests

### Function Entity Search
```bash
./bin/repocontext query --entity-type "function"
```
**Expected Functions (partial list):**
- All functions from main.go (NewInMemoryUserService, GetUser, CreateUser, etc.)
- All functions from service.go (ProcessUser, ValidateUser, SendWelcomeNotification, etc.)
- All functions from utils.go (FormatName, ValidateEmail, GenerateID, etc.)
- All functions from test_demo.go (NewQueryEngine, Level1Function, RunDemo, etc.)

### Type Entity Search
```bash
./bin/repocontext query --entity-type "type"
```
**Expected Types:**
- `User`, `UserService`, `InMemoryUserService` (from main.go)
- `Config`, `Address`, `Profile`, `Repository`, `CacheManager`, `NotificationService` (from models.go)
- `UserManager`, `UserRequest`, `UserResult` (from service.go)
- `Logger`, `StringUtils` (from utils.go)
- `QueryEngine`, `SearchResult`, `TestService`, `TestEngine`, etc. (from test_demo.go)

### Variable Entity Search
```bash
./bin/repocontext query --entity-type "variable"
```
**Expected Variables:**
- `globalConfig`, `userCount`, `serviceRegistry`, `isInitialized` (from models.go)
- `emailRegex`, `phoneRegex`, `utilsLogger`, `defaultConfig` (from utils.go)
- `testVariable`, `bufferSize`, `apiEndpoint`, `featureFlags`, `connectionTimeout` (from test_demo.go)

### Constant Entity Search
```bash
./bin/repocontext query --entity-type "constant"
```
**Expected Constants:**
- `MaxUsers`, `DefaultTimeout`, `ServiceVersion`, `DEBUG_MODE` (from models.go)
- `UtilVersion`, `MaxRetries`, `RetryDelay`, `CacheKeyPrefix` (from utils.go)
- `TestConstant`, `MAX_BUFFER_SIZE`, `API_VERSION`, `ENABLE_LOGGING` (from test_demo.go)

## Call Graph Depth Tests

### Depth 1 (Direct calls only)
```bash
./bin/repocontext query --function "ProcessUser" --include-callees --depth 1
```
**Expected Callees (depth 1):**
- `ValidateUser`
- `CreateUser` (via service)
- `SendWelcomeNotification`

### Depth 2 (Two levels deep)
```bash
./bin/repocontext query --function "ProcessUser" --include-callees --depth 2
```
**Expected Callees (depth 2):**
- All depth 1 callees plus:
- `IsValidEmail` (called by ValidateUser)
- Any functions called by CreateUser implementation

### Multi-level Chain Test
```bash
./bin/repocontext query --function "Level1Function" --include-callees --depth 5
```
**Expected Call Chain:**
- Level1Function → Level2Function → Level3Function → Level4Function → Level5Function

## Token Limit Tests

### Low Token Limit
```bash
./bin/repocontext query --function "ProcessUser" --max-tokens 100
```
**Expected Result:** Should truncate output at ~100 tokens, showing partial function definition

### High Token Limit
```bash
./bin/repocontext query --function "ProcessUser" --max-tokens 1000
```
**Expected Result:** Should show complete function definition with more context

## Format Tests

### JSON Format
```bash
./bin/repocontext query --function "ProcessUser" --json
```
**Expected JSON Structure:**
```json
{
  "results": [
    {
      "name": "ProcessUser",
      "type": "function",
      "file": "service.go",
      "line": 28,
      "content": "...",
      "call_graph": {
        "callees": [...],
        "callers": [...]
      }
    }
  ]
}
```

## Complex Integration Tests

### Full Options Test
```bash
./bin/repocontext query --function "RunDemo" --include-callers --include-callees --include-types --max-tokens 2000 --depth 2 --json
```
**Expected Result:**
- Should find `RunDemo` function in test_demo.go
- Should include all direct callees (NewQueryEngine, NewMockTestService, Level1Function, ProcessInput, HandleRequest, LogMessage)
- Should include functions up to depth 2
- Should include related types (QueryEngine, MockTestService, etc.)
- Should format as JSON
- Should respect 2000 token limit

### High Depth Integration Test
```bash
./bin/repocontext query --function "InitializeServices" --include-callees --depth 3
```
**Expected Call Chain:**
- InitializeServices → NewInMemoryUserService
- InitializeServices → NewUserManager → NewConfig
- And other initialization chains up to depth 3

## Error Cases Expected Results

### Multiple Criteria Error
```bash
./bin/repocontext query --function "test" --type "test"
```
**Expected Result:** Should return error message about conflicting search criteria

### Non-existent Function
```bash
./bin/repocontext query --function "CompletelyNonExistent"
```
**Expected Result:** Should return "No results found" or empty result set

### Invalid Entity Type
```bash
./bin/repocontext query --entity-type "invalid"
```
**Expected Result:** Should return error about invalid entity type

## Performance Expectations

### Large Query Performance
```bash
./bin/repocontext query --entity-type "function" --include-callers --include-callees --max-tokens 10000
```
**Expected Result:** Should complete in reasonable time (< 5 seconds) and return comprehensive function list with call graphs

### Wildcard Search Performance
```bash
./bin/repocontext query --search "*" --max-tokens 5000
```
**Expected Result:** Should match many entities but be limited by token count, completing in reasonable time

## Key Validation Points

1. **Call Graph Accuracy:** Verify that call relationships match actual function calls in the code
2. **Pattern Matching:** Confirm that wildcard patterns match expected entities
3. **Depth Control:** Ensure depth limits are respected in call graph traversal
4. **Token Limits:** Verify output is truncated appropriately when token limits are exceeded
5. **JSON Format:** Ensure JSON output is valid and contains expected fields
6. **Error Handling:** Confirm appropriate errors for invalid inputs
7. **Performance:** Ensure reasonable response times for complex queries

## Summary Statistics (Approximate)

- **Total Functions:** ~60+ functions across all files
- **Total Types:** ~20+ types (structs, interfaces)
- **Total Variables:** ~15+ package-level variables
- **Total Constants:** ~10+ constants
- **Call Relationships:** Complex web with 5+ level deep chains
- **Pattern Matches:** Multiple entities matching Query*, Test*, *Engine patterns

This testdata provides comprehensive coverage for all CLI test scenarios with predictable, verifiable results.

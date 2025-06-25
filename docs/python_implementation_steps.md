Phase 1: Core Python Parser Foundation (TDD)
Step 1: Set up Test Infrastructure
Create test directory structure:
testdata/python-simple/ with basic Python files
testdata/python-simple/main.py (simple functions, no classes)
testdata/python-simple/models.py (basic class with methods)
testdata/python-simple/expected_results.md (expected parsing output)
Step 2: Write Failing Tests First
Create internal/ast/python/parser_test.go:
Test NewPythonParser() constructor
Test GetSupportedExtensions() returns [".py"]
Test GetLanguageName() returns "python"
Test ParseFile() with simple function (should fail initially)
Test ParseFile() with basic class (should fail initially)
Test error handling for invalid Python syntax
Run tests to confirm they fail:
Apply
Run
v
Step 3: Implement Minimal Go Parser
Create internal/ast/python/parser.go:
Implement PythonParser struct
Implement interface methods (return errors for ParseFile() initially)
Add basic Python executable detection
Register parser in the parser registry
Run tests - some should pass, ParseFile should still fail
Step 4: Create Python AST Extractor
Create internal/ast/python/extractor.py:
Basic script that accepts file path as argument
Parse Python AST using ast module
Extract minimal function information
Output basic JSON structure matching Go models
Update Go parser to call Python script:
Implement executeExtractor() method
Implement parseJSON() method
Make basic ParseFile() work for simple functions
Run tests - basic function parsing should now pass
Step 5: Extend for Basic Classes
Add class parsing to extractor.py:
Extract class definitions
Extract method definitions within classes
Map to TypeDef and Method models
Update tests to validate class parsing results
Run all tests - should pass for basic scenarios
Phase 2: Advanced Features (TDD)
Step 6: Type Hint Support
Add type hint test cases:
Functions with type annotations
Variables with type hints
Generic types and complex annotations
Implement type hint analysis in extractor.py:
Parse type annotations from AST
Map Python types to Go model types
Handle complex types (List, Dict, Union, Optional)
Run tests - type information should be correctly extracted
Step 7: Import System
Add import test cases:
Absolute imports (import module)
Relative imports (from .module import func)
Aliased imports (import numpy as np)
Star imports (from module import *)
Implement import extraction in extractor.py:
Parse different import statement types
Map to Import model structure
Run tests - import information should be correctly parsed
Step 8: Call Graph Generation
Add call graph test cases:
Function calls within same file
Method calls on objects
Cross-module function calls
Implement call graph analysis in extractor.py:
Identify function call sites
Track method calls with receiver information
Build call relationship data
Run tests - call relationships should be correctly identified
Phase 3: Integration & Polish (TDD)
Step 9: Error Handling
Add error handling test cases:
Invalid Python syntax
Missing Python interpreter
Script execution failures
Implement robust error handling:
Graceful syntax error handling
Clear error messages for setup issues
Fallback behaviors where appropriate
Run tests - error scenarios should be handled gracefully
Step 10: Integration Testing
Create integration tests:
internal/ast/python/integration_test.go
Test parsing of complex Python projects
Test performance with larger codebases
Cross-validate results with expected outputs
Create comprehensive test data:
testdata/python-complex/ with inheritance, decorators
Real-world Python code samples
Edge cases and corner scenarios
Step 11: Validation & Optimization
Run comprehensive test suite:
Apply
Run
works
Performance testing:
Benchmark parsing speed
Memory usage validation
Optimize bottlenecks
Final validation:
Run linting: go vet ./...
Run full test suite: make test
Validate against real Python projects
Continuous TDD Cycle
For each step:
Red: Write tests that fail
Green: Implement minimal code to make tests pass
Refactor: Clean up implementation while keeping tests green
Validate: Run linting and full test suite
Success Criteria
All unit tests pass
Integration tests pass with real Python code
Parser correctly handles 90%+ of common Python patterns
Error handling provides clear, actionable messages
Performance meets specified benchmarks (parse 1000+ files in <30s)
This approach ensures each component is thoroughly tested before implementation and maintains high code quality throughout the development process.

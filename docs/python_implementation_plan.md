# Python Language Support Implementation Plan

## Overview

This document outlines the comprehensive implementation plan for adding Python language support to the repository-context-protocol tool. The implementation will extend the existing parser registry architecture to support Python AST parsing, semantic analysis, and cross-file relationship tracking.

## Current Architecture Analysis

### Existing Parser Interface
The tool uses a clean `LanguageParser` interface that Python must implement:

```go
type LanguageParser interface {
    ParseFile(path string, content []byte) (*models.FileContext, error)
    GetSupportedExtensions() []string
    GetLanguageName() string
}
```

### Data Models to Support
Python parser must populate the following models:

- **FileContext**: Main container for all parsed entities
- **Function**: Function definitions with parameters, returns, and call relationships
- **TypeDef**: Class definitions, inheritance, and method relationships
- **Variable**: Module-level and class-level variables
- **Constant**: Module constants and class constants
- **Import**: Import statements and module dependencies
- **Export**: Public API elements

## Implementation Strategy

### 1. Python AST Parsing Approach

**Selected Approach: External Python Script + JSON Communication**

**Rationale:**
- Native Python `ast` module provides complete AST access
- JSON bridge maintains type safety and simplicity
- Avoids CGo complexity and cross-compilation issues
- Leverages Python's rich ecosystem for analysis

**Architecture:**
```
Go Parser → Python Script → JSON Output → Go Models
```

### 2. Python AST Extraction Script

**Location:** `internal/ast/python/extractor.py`

**Core Responsibilities:**
- Parse Python files using `ast` module
- Extract functions, classes, variables, imports
- Perform static analysis for type hints
- Generate call graphs and relationships
- Output structured JSON for Go consumption

**Key Features:**
- Type hint analysis (PEP 484, 526, 563)
- Decorator extraction and analysis
- Class inheritance mapping
- Method resolution order (MRO) tracking
- Import dependency analysis

### 3. Python-Specific Challenges & Solutions

#### 3.1 Dynamic Typing Handling

**Challenge:** Python's dynamic typing vs Go's static type system

**Solution Strategy:**
- **Type Hints**: Extract and parse type annotations
- **Inference**: Basic type inference from literals and context
- **Fallback**: Use `Any` or `dynamic` for untyped elements
- **Documentation**: Extract type info from docstrings

**Type Mapping:**
```
Python Type     → Go Model Type
int             → "int"
str             → "string"
float           → "float64"
bool            → "bool"
List[T]         → "[]T"
Dict[K,V]       → "map[K]V"
Optional[T]     → "*T"
Union[A,B]      → "A|B"
```

#### 3.2 Import System Complexity

**Challenge:** Python's flexible import system

**Solution:**
- **Absolute Imports**: `import module` → `{Path: "module", Alias: ""}`
- **Relative Imports**: `from .module import func` → `{Path: ".module", Alias: "func"}`
- **Aliased Imports**: `import numpy as np` → `{Path: "numpy", Alias: "np"}`
- **Star Imports**: `from module import *` → `{Path: "module", Alias: "*"}`

#### 3.3 Class vs Function Distinction

**Challenge:** Python classes vs Go structs+methods

**Solution:**
- **Classes → TypeDef**: Map Python classes to `TypeDef` with `kind: "class"`
- **Methods → Methods**: Instance methods become `Method` entries
- **Static/Class Methods**: Distinguished by decorator analysis
- **Properties**: Treated as computed fields

#### 3.4 Call Graph Analysis

**Challenge:** Dynamic method dispatch and late binding

**Solution:**
- **Static Analysis**: Extract visible function calls
- **Method Calls**: `obj.method()` → track receiver type when possible
- **Dynamic Calls**: `getattr(obj, 'method')()` → mark as dynamic
- **Inheritance**: Track method overrides and super() calls

### 4. Implementation Components

#### 4.1 Python Parser (Go)

**File:** `internal/ast/python/parser.go`

```go
type PythonParser struct {
    extractorPath string
    pythonPath    string
}

func NewPythonParser() *PythonParser
func (p *PythonParser) ParseFile(path string, content []byte) (*models.FileContext, error)
func (p *PythonParser) GetSupportedExtensions() []string
func (p *PythonParser) GetLanguageName() string
```

**Key Methods:**
- `executeExtractor()`: Run Python script with file content
- `parseJSON()`: Convert JSON output to Go models
- `validatePythonSetup()`: Check Python availability and dependencies

#### 4.2 Python AST Extractor

**File:** `internal/ast/python/extractor.py`

**Core Classes:**
- `PythonASTExtractor`: Main extraction orchestrator
- `FunctionExtractor`: Function and method analysis
- `ClassExtractor`: Class hierarchy and member analysis
- `TypeAnalyzer`: Type hint parsing and inference
- `CallGraphBuilder`: Function call relationship tracking

**Output Format:**
```json
{
  "path": "file.py",
  "language": "python",
  "functions": [...],
  "types": [...],
  "variables": [...],
  "constants": [...],
  "imports": [...],
  "exports": [...]
}
```

#### 4.3 Error Handling Strategy

**Python Script Errors:**
- Syntax errors → Return partial results with error annotation
- Import errors → Continue with available information
- Type analysis errors → Fallback to `Any` type

**Go Integration Errors:**
- Python not found → Clear error message with setup instructions
- Script execution failure → Detailed error with debugging info
- JSON parsing errors → Structured error with context

### 5. Type System Mapping

#### 5.1 Function Signatures

**Python:**
```python
def process_data(items: List[str], count: int = 10) -> Dict[str, int]:
    pass
```

**Go Model:**
```go
Function{
    Name: "process_data",
    Parameters: []Parameter{
        {Name: "items", Type: "List[str]"},
        {Name: "count", Type: "int"},
    },
    Returns: []Type{
        {Name: "Dict[str, int]", Kind: "composite"},
    },
}
```

#### 5.2 Class Definitions

**Python:**
```python
class DataProcessor(BaseProcessor):
    count: int = 0

    def process(self, data: str) -> str:
        return data.upper()
```

**Go Model:**
```go
TypeDef{
    Name: "DataProcessor",
    Kind: "class",
    Fields: []Field{
        {Name: "count", Type: "int"},
    },
    Methods: []Method{
        {Name: "process", Parameters: [...], Returns: [...]},
    },
    Embedded: []string{"BaseProcessor"},
}
```

### 6. Call Graph Implementation

#### 6.1 Function Calls

**Extraction Strategy:**
- `ast.Call` nodes → Function name extraction
- Method calls → Receiver type + method name
- Built-in functions → Mark as external
- Lambda calls → Anonymous function handling

#### 6.2 Cross-File Relationships

**Implementation:**
- Import analysis → Build module dependency graph
- Function usage tracking → Cross-module call detection
- Class inheritance → Multi-file class hierarchies
- Global variable access → Module-level dependencies

### 7. Testing Strategy

#### 7.1 Unit Tests

**Test Files:**
- `parser_test.go`: Core parsing functionality
- `extractor_test.py`: Python script validation
- `integration_test.go`: End-to-end parsing tests

**Test Cases:**
- Simple functions with type hints
- Complex class hierarchies
- Import variations (absolute, relative, aliased)
- Error conditions and edge cases

#### 7.2 Integration Tests

**Test Data:**
- `testdata/python-simple/`: Basic Python project
- `testdata/python-complex/`: Advanced features (inheritance, decorators)
- `testdata/python-mixed/`: Mixed typing scenarios

**Validation:**
- JSON output format correctness
- Model field population accuracy
- Call graph relationship accuracy
- Performance benchmarking

### 8. Performance Considerations

#### 8.1 Execution Overhead

**Python Script Startup:**
- Cache Python interpreter instance
- Reuse extractor script for multiple files
- Batch processing for large codebases

**JSON Processing:**
- Stream processing for large files
- Incremental parsing for memory efficiency
- Compression for large datasets

#### 8.2 Memory Management

**Python Side:**
- AST node cleanup after processing
- Streaming output for large files
- Memory-efficient data structures

**Go Side:**
- Bounded memory usage during parsing
- Efficient JSON unmarshaling
- Garbage collection optimization

### 9. Deployment Considerations

#### 9.1 Python Dependency Management

**Requirements:**
- Python 3.8+ (for comprehensive type hint support)
- Standard library only (no external dependencies)
- Cross-platform compatibility

**Distribution:**
- Embed Python script in Go binary
- Runtime Python detection and validation
- Fallback options for missing Python

#### 9.2 Error Messaging

**User-Friendly Errors:**
- Clear Python setup instructions
- Specific error codes for different failure modes
- Debugging information for development

### 10. Future Enhancements

#### 10.1 Advanced Analysis

**Potential Features:**
- Dynamic type inference using execution traces
- Decorator behavior analysis
- Async/await pattern detection
- Context manager usage tracking

#### 10.2 Integration Improvements

**Optimization Opportunities:**
- Native Python extension (C/Go integration)
- Persistent Python process for better performance
- Language server protocol integration
- Real-time analysis for editor integration

### 11. Implementation

**Phase 1: Core Parser
- Basic Python AST extraction script
- Go parser implementation
- Simple function and class parsing
- Basic test suite

**Phase 2: Advanced Features
- Type hint analysis and mapping
- Import system handling
- Call graph generation
- Comprehensive test coverage

**Phase 3: Integration & Polish
- Error handling improvements
- Performance optimization
- Documentation and examples
- Production readiness testing

### 12. Success Metrics

**Functionality:**
- Parse 95%+ of real-world Python codebases
- Accurate type information extraction
- Complete call graph generation
- Robust error handling

**Performance:**
- Parse medium Python projects (1000+ files) in <30 seconds
- Memory usage under 500MB for large codebases
- Graceful handling of syntax errors

**Quality:**
- 90%+ test coverage
- Zero critical bugs in common scenarios
- Clear documentation and examples
- Positive developer experience

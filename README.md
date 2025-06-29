# Repository Context Protocol

[![codecov](https://codecov.io/gh/qraftworxDev/repository-context-protocol/graph/badge.svg?token=IV7364HI6X)](https://codecov.io/gh/qraftworxDev/repository-context-protocol)

A comprehensive Go CLI tool for semantic code repository analysis and indexing. RepoContext extracts AST-level information, builds global call graphs, and provides fast, context-aware queries optimized for LLM agent consumption.

## 🚀 Features

### Core Capabilities
- **AST-Level Parsing**: Deep semantic analysis of Go codebases with function signatures, type definitions, variables, constants, and relationships
- **Global Call Graph**: Cross-file function call analysis with caller/callee relationship tracking
- **Hybrid Storage**: SQLite index for fast lookups + MessagePack chunks for detailed semantic data
- **Semantic Chunking**: Intelligent grouping of related code for token-efficient LLM consumption
- **Rich Metadata**: Checksums, timestamps, line numbers, and signature validation
- **Multiple Output Formats**: JSON and text output with token counting for LLM integration

### Query Operations
- Search by function name, type, variable, or file
- Pattern-based search with wildcards and regex patterns
- Call graph traversal (callers/callees) with depth control
- Token-aware result limiting for LLM context windows
- Cross-file relationship analysis

> **Note**: Regex patterns with unsupported features (lookbehind, lookahead) are automatically converted to supported alternatives with warnings. See [docs/regex_limitations.md](docs/regex_limitations.md) for details.

## 🏗️ Architecture

```
.repocontext/
├── index.db           # SQLite: fast lookups, relationships, metadata
├── chunks/            # MessagePack: detailed semantic data
│   ├── auth_001.msgpack
│   └── api_001.msgpack
└── manifest.json      # Chunk directory, metadata
```

### 3-Phase Processing Pipeline
1. **Parse**: AST extraction from source files using go/ast
2. **Enrich**: Global call graph analysis and cross-file relationships
3. **Store**: Hybrid storage with indexed metadata and chunked semantic data

## 📦 Installation

```bash
# Build from source
git clone https://github.com/qraftworxDev/repository-context-protocol
cd repository-context-protocol
make build

# Install binaries
make install
```

## 🔧 Usage

### Basic Commands

```bash
# Initialize a repository for indexing
repocontext init

# Build the semantic index
repocontext build

# Query the index
repocontext query --function "ProcessUser" --include-callers --json
repocontext query --type "UserService" --include-callees
repocontext query --search "authentication" --max-tokens 2000
repocontext query --file "user.go" --format json
```

### Advanced Queries

```bash
# Search with call graph analysis
repocontext query --function "main" --include-callees --depth 3

# Token-limited queries for LLM consumption
repocontext query --search "payment" --max-tokens 1500 --json

# Pattern matching
repocontext query --function "*User*" --include-types

# Regex patterns (with automatic conversion of unsupported features)
repocontext query --function "/Handle.*User/" --include-callers
```

### Example Output

```json
{
  "query": "ProcessUser",
  "search_type": "function",
  "entries": [
    {
      "index_entry": {
        "name": "ProcessUser",
        "type": "function",
        "file": "user.go",
        "start_line": 25,
        "signature": "func ProcessUser(user *User) error"
      },
      "chunk_data": {
        "id": "user_001",
        "files": ["user.go"],
        "file_data": [...],
        "token_count": 850
      }
    }
  ],
  "call_graph": {
    "function": "ProcessUser",
    "callers": [{"function": "main", "file": "main.go", "line": 15}],
    "callees": [{"function": "ValidateUser", "file": "validation.go", "line": 28}]
  },
  "token_count": 1245,
  "executed_at": "2024-01-15T10:30:00Z"
}
```

## 🧪 Implementation Status

### ✅ Completed Components

#### **Models** (`internal/models/`)
- Complete data structures: `FileContext`, `Function`, `TypeDef`, `Variable`, `Constant`
- Global indexing models: `Reference`, `CallRelation`, `IndexEntry`
- Semantic chunking: `SemanticChunk`, `Manifest`, `ChunkInfo`
- Storage abstractions with full validation
- **38 comprehensive test files** covering all models

#### **AST Parsing** (`internal/ast/`)
- **Go Parser**: Full AST analysis using `go/ast` and `go/token`
  - Function extraction with parameters, returns, signatures
  - Type definitions (structs, interfaces) with methods and fields
  - Variable and constant declarations with type inference
  - Import/export analysis
  - Method binding to types
  - Call graph extraction from function bodies
- **Parser Registry**: Extensible architecture for multiple languages
- **Metadata Generation**: SHA-256 checksums, modification times, line numbers
- **Export Detection**: Go visibility rules (uppercase = exported)

#### **CLI Interface** (`internal/cli/`)
- **Commands**: `init`, `build`, `query` with comprehensive flag support
- **Query Operations**: Function, type, variable, file, and pattern search
- **Output Formats**: JSON and text with token counting
- **Call Graph Options**: Include callers/callees, depth control
- **Validation**: Repository state, search criteria, and parameter validation
- **Error Handling**: Comprehensive error messages and status codes

#### **Index System** (`internal/index/`)
- **Hybrid Storage**: SQLite for fast lookups + MessagePack for semantic data
- **Query Engine**: Multiple search types with relationship traversal
- **Global Enrichment**: Cross-file call analysis and relationship categorization
- **Chunking Strategy**: File-based semantic chunking with token estimation
- **Statistics**: Build metrics, performance tracking, coverage analysis
- **SQLite Schema**: Optimized indexes for name, type, file, and call relationships

#### **Main Binary** (`cmd/repocontext/`)
- **CLI Entrypoint**: Proper command execution and error handling
- **Integration Tests**: End-to-end testing with real Go projects
- **Build System**: Make targets for build, test, install, and coverage

### 🎯 Key Achievements

1. **Comprehensive Test Coverage**: 38 test files covering all components
2. **Production-Ready CLI**: Full command structure with proper validation
3. **Semantic Analysis**: Deep AST parsing with call graph analysis
4. **LLM Integration**: Token-aware queries with JSON output
5. **Performance**: Hybrid storage for fast queries and efficient chunking
6. **Extensible**: Parser registry supports future language additions

### 📊 Test Coverage

This project maintains **comprehensive test coverage** across all components:

- **AST Parsing**: 15 test files covering Go parser, call graphs, exports, variables, constants
- **Index System**: 12 test files covering storage, queries, enrichment, chunking
- **Models**: 8 test files validating all data structures and serialization
- **CLI**: 6 test files covering commands, validation, and integration
- **Integration**: Multi-file demos and end-to-end pipeline validation

```bash
# Generate coverage reports
make coverage-report

# View HTML coverage
open coverage.html

# Run all tests
make test
```

### 🔮 Future Enhancements

- **Language Support**: Python and TypeScript parsers
- **LSP Server**: Enhanced editor integration with semantic navigation
- **HTTP API**: REST endpoints for LLM agent integration
- **Advanced Chunking**: Semantic similarity-based chunk optimization
- **Call Graph Visualization**: Interactive dependency graphs

## 🏷️ Project Structure

```
repository-context-protocol/
├── cmd/
│   ├── repocontext/           # Main CLI binary ✅
│   └── lsp/                   # LSP server (future)
├── internal/
│   ├── ast/                   # Language parsers ✅
│   │   ├── golang/            # Go AST parser ✅
│   │   ├── python/            # Python parser (future)
│   │   └── typescript/        # TypeScript parser (future)
│   ├── index/                 # Core indexing ✅
│   │   ├── builder.go         # 3-phase index builder ✅
│   │   ├── hybrid.go          # Hybrid storage ✅
│   │   ├── enrichment.go      # Global call graph ✅
│   │   └── query.go           # Query engine ✅
│   ├── models/                # Data structures ✅
│   │   ├── context.go         # Core context types ✅
│   │   ├── function.go        # Function representation ✅
│   │   ├── types.go           # Type definitions ✅
│   │   └── storage.go         # Storage models ✅
│   └── cli/                   # CLI commands ✅
│       ├── commands.go        # Command registry ✅
│       ├── init.go           # Repository initialization ✅
│       ├── build.go          # Index building ✅
│       └── query.go          # Query interface ✅
├── testdata/                  # Test repositories ✅
└── docs/                      # Documentation ✅
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add comprehensive tests
4. Ensure all tests pass: `make test`
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**RepoContext** - Semantic code analysis for the AI era 🤖

### Todo:
1. precommit integration
1. mcp server creation
1. testing with LLM
1. extend to support Python

# Go bugs
1. repocontext query --search "main*" --include-callers --json
    - callers: null i.s.o. callers: []
    - callees: [] i.s.o. omitted if empty
1. query --entity-type function --include-callers --include-callees
    - callers: (none)
    - callees: populated
1. repocontext query --entity-type function --include-types
    - doesn't differ from "repocontext query --entity-type function"

# Python bugs
1. Function signature maps to go-based types, should stick to Python
  {
    "name": "extract",
    "signature": "() -\u003e map[string]interface{}",
    "parameters": null,
    "returns": [
      {
        "name": "map[string]interface{}",
        "kind": "builtin"
      }
    ],
    "start_line": 32,
    "end_line": 66
  }
1. --include-callers and --include-callees don't seem to have any effect. Could be broken.
1. "kind" is empty in exports
  {
    "name": "PythonASTExtractor",
    "type": "class",
    "kind": ""
  }
1. imports are incomplete
  {
    "path": "typing"
  }
  should be Dict or Any, etc. as per the file content

# TODO:
1. test the rest of the functionality using depth, tokens limits, and other search functionality.
   - tokens limit works
   -
1. Test different outputs.
1. Fix lookup issue - when using nested "repos" (e.g. there's a .repocontext folder at root and inside another folder) the product returns the data from the root folder's content
    likely related to the lookup initialising by first going to root, and then doing the query lookup against the .repocontext. Should recursively step up the chain of paths to find the first instance of the folder and default to root.
1. Getting this warning at start up:
```text
Warning: Repository initialization failed: failed to initialize query engine: repository not initialized - .repocontext directory not found
Server will continue with limited functionality
{"jsonrpc":"2.0","id":3,"result":{"content":[{"type":"text","text":"{\n  \"path\": \"/Users/q/Development/Go/repository-context-protocol/repository-context-protocol\",\n  \"repo_context_path\": \"/Users/q/Development/Go/repository-context-protocol/repository-context-protocol/.repocontext\",\n  \"already_initialized\": false,\n  \"message\": \"Repository initialized successfully\",\n  \"created_directories\": [\n    \"/Users/q/Development/Go/repository-context-protocol/repository-context-protocol/.repocontext\",\n    \"/Users/q/Development/Go/repository-context-protocol/repository-context-protocol/.repocontext/chunks\"\n  ],\n  \"created_files\": [\n    \"/Users/q/Development/Go/repository-context-protocol/repository-context-protocol/.repocontext/manifest.json\"\n  ]\n}"}]}}
```
    need to look into this and resolve - it's expected that the directory not exist on first run
1. return format needs to change. The output is "prettified" with \n and white space. This should be removed, it'll waste space. Need to see if it's stored like this - fix it at storage layer if yes.

# MCP server testing
## Test tools systematically
Note: we can expand this list to test all functions available
```bash
./bin/repocontext-mcp << 'EOF'
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}}
EOF
```

```bash
./bin/repocontext-mcp << 'EOF'
{"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}}
EOF
```

```bash
./bin/repocontext-mcp << 'EOF'
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "initialize_repository", "arguments": {}}}
EOF
```

```bash
./bin/repocontext-mcp << 'EOF'
{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "build_index", "arguments": {}}}
EOF
```

```bash
./bin/repocontext-mcp << 'EOF'
{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "query_by_name", "arguments": {"name": "main"}}}
EOF
```

```bash
./bin/repocontext-mcp << 'EOF'
{"jsonrpc": "2.0", "id": 6, "method": "tools/call", "params": {"name": "list_functions", "arguments": {}}}
EOF
```

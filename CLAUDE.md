# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Repository Context Protocol is a Go CLI tool for semantic code repository analysis and indexing. It extracts AST-level information, builds global call graphs, and provides fast, context-aware queries optimized for LLM agent consumption.

## Build and Development Commands

### Build
- `make build` - Build both repocontext and repocontext-mcp binaries
- `make install` - Install binaries to /usr/local/bin/

### Testing
- `make test` - Run all tests (MCP tests run sequentially to avoid SQLite locks)
- `make test-mcp` - Run MCP tests only (sequentially)
- `make test-parallel` - Run all tests in parallel (may cause SQLite locks)
- `make coverage` - Run tests with coverage report
- `make integration-test` - Run integration tests (requires build first)

### Code Quality
- `make lint` - Run golangci-lint (auto-installs if missing)
- `make fmt` - Format code with gofmt and goimports
- `make vet` - Run go vet
- `make pre-commit` - Run fmt, vet, lint, and test in sequence

### Development Setup
- `make tools` - Install development tools (golangci-lint, goimports, goreleaser)
- `make deps` - Download dependencies
- `make dev-setup` - Setup development environment with test data
- `make setup` - Install tools and dependencies

### Other
- `make clean` - Remove build artifacts
- `make tidy` - Tidy Go modules

## Architecture

### Core Components

1. **AST Parsing** (`internal/ast/`):
   - Language-agnostic parser interface with registry system
   - Go parser (`golang/`) - complete AST analysis using go/ast
   - Python parser (`python/`) - Python AST analysis
   - TypeScript parser (`typescript/`) - planned

2. **Index System** (`internal/index/`):
   - `builder.go` - 3-phase indexing pipeline (Parse → Enrich → Store)
   - `hybrid.go` - Hybrid storage (SQLite for fast lookups + MessagePack for semantic data)
   - `enrichment.go` - Global call graph analysis and cross-file relationships
   - `query.go` - Query engine with multiple search types

3. **Data Models** (`internal/models/`):
   - `context.go` - Core data structures (RepoContext, FileContext, GlobalIndex)
   - `function.go` - Function representation with parameters/returns
   - `types.go` - Type definitions, variables, constants
   - `storage.go` - Storage models for hybrid system

4. **CLI Interface** (`internal/cli/`):
   - `commands.go` - Command registry and validation
   - `init.go` - Repository initialization
   - `build.go` - Index building orchestration
   - `query.go` - Query interface with output formatting

5. **MCP Server** (`internal/mcp/`):
   - Model Context Protocol server implementation
   - Tools for repository querying and call graph analysis

### Storage Architecture

```
.repocontext/
├── index.db           # SQLite: fast lookups, relationships, metadata
├── chunks/            # MessagePack: detailed semantic data
│   ├── auth_001.msgpack
│   └── api_001.msgpack
└── manifest.json      # Chunk directory, metadata
```

### 3-Phase Processing Pipeline
1. **Parse**: AST extraction using language-specific parsers
2. **Enrich**: Global call graph analysis and cross-file relationships
3. **Store**: Hybrid storage with indexed metadata and chunked semantic data

## Testing Strategy

- MCP tests use SQLite databases which can cause locking issues during parallel execution
- MCP tests run sequentially (`-p 1`) while other tests run in parallel
- 38+ comprehensive test files covering all components
- Integration tests with real Go projects in `testdata/`

## Key Usage Patterns

### CLI Usage
```bash
repocontext init              # Initialize repository
repocontext build             # Build semantic index
repocontext query --function "ProcessUser" --include-callers --json
repocontext query --type "UserService" --include-callees
repocontext query --search "authentication" --max-tokens 2000
```

### MCP Server Testing
```bash
./bin/repocontext-mcp << 'EOF'
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}}
EOF
```

## Language Support

- **Go**: Full AST parsing with call graphs, exports, type analysis
- **Python**: AST analysis with call tracking (some known bugs with local vs cross-file calls)
- **TypeScript**: Planned future support

## Important Notes

- Regex patterns with unsupported features (lookbehind, lookahead) are automatically converted
- Files removed from repository should be removed from index (TODO item)
- Nested .repocontext folders may cause lookup issues (recursively steps up to find first instance)
- Token-aware result limiting for LLM context windows
- Comprehensive error recovery and validation throughout

# Ways of work
## Phase 1 - User-driven:
This is handled from the user's side. The user prompts you and starts working through the process with you. The process is roughyl outlined below.
1. Explore problem/solution
   1. The user builds context with you by describing the exact problem at hand.
       - if it's a bug, the user provides the error trace and any details on the bug
       - if it's a new feature, as far as possible, the user provides a full spec write up with expected outcomes/validation measures
   1. The user works with you to come up with a solution best suited to your needs & constraints (either time, complexity, scope, etc.)

## Phase 2 - AI-driven:
This is handled primarily by you. You are prompted once and you start working through the approach outlined below.

1. create a new git branch to work on
1. Produce the plan and document the execution approach. Break it down into small incremental steps. Plan it according to Test Driven Development best practices.
    1. Produce the plan, and present it to the user. Await confirmation from the user.
    1. The user can ask you to change the plan, or add more details to the plan.
1. commit the generated document(s) to git
1. Start with building the tests (Red Phase)
1. create a git commit for each test added (use partial commit) grouped by logically associated change/addition
1. Build code to make tests pass (Green Phase)
1. create a commit for each logically associated change/addition
1. Validate tests pass
    1. Iterate through fixing any issues, running tests after each change
    1. Commit each change that solves a bug in the tests
1. Review the code noting any necessary optimisations or proposed refactors that would improve the solution (Refactor Phase)
    1. Prompt the user asking if the proposal should be implemented or documented for later
    1. Do what the user says
1. Lint code
    1. Fix any linting issues. Go through it one by one
    1. after each fix, check if the error is fixed
    1. once the error is fixed, commit the change and continue until all errors are fixed
1. Analyse the code for security vulnerabilities noting any necessary changes needed to ensure the changes don't introduce attack vectors
    1. Prompt the user to asking if the necessary changes should be implemented or documented for later
    1. Do what the user says
1. Concisely and completely document the changes made in /docs

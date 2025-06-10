# RepoContext: Semantic Code Indexing Tool

## OBJECTIVE
Build a Go CLI tool that creates semantic indexes of code repositories for LLM agent consumption. Tool extracts AST-level information (functions, types, call graphs) and provides fast, context-aware queries with minimal token overhead.

## CORE ARCHITECTURE

### Data Flow
1. Parse source files → Extract AST semantic data → Build cross-references → Store in SQLite + MessagePack chunks
2. Query: Search index → Load relevant chunks → Return focused context (JSON/HTTP)
3. Integration: CLI commands, LSP server, HTTP API for LLM agents

### Storage Schema
```
.repocontext/
├── index.db           # SQLite: fast lookups, relationships  
├── chunks/            # MessagePack: detailed semantic data
│   ├── auth_001.msgpack
│   └── api_001.msgpack
└── manifest.json      # Chunk directory, metadata
```

### Core Data Types
```go
type FileContext struct {
    Path, Language, Checksum string
    Functions []Function
    Types []TypeDef  
    Imports []Import
    CallGraph map[string][]string
}

type Function struct {
    Name, Signature string
    Parameters []Parameter
    Returns []Type
    StartLine, EndLine int
    Calls []string      // Functions this calls
    CalledBy []string   // Functions that call this
}

type Reference struct {
    File, Entity, Type string
    StartLine, EndLine int
    Signature string
}
```

## IMPLEMENTATION PHASES

### Phase 1: Go Parser + Basic Index
- `internal/ast/golang/parser.go`: Go AST extraction using go/ast, go/parser
- `internal/index/builder.go`: Build index from parsed files
- `internal/index/sqlite.go`: SQLite storage with indexes on name, file, type
- `cmd/repocontext/main.go`: CLI with init, build, query commands
- Target: `repocontext build && repocontext query --function "main"`

### Phase 2: Query System
- `internal/index/query.go`: Search by name, type, file, relationships
- `internal/cli/query.go`: JSON output for machine consumption
- Token-aware context building (max_tokens parameter)
- Target: `repocontext query --json --function "ProcessUser" --include-callers --max-tokens 2000`

### Phase 3: LSP Integration  
- `cmd/lsp/main.go`: LSP server binary
- `internal/lsp/server.go`: LSP protocol implementation
- Handlers: textDocument/definition, textDocument/hover, workspace/symbol
- Enhanced "go to definition" using semantic index
- Rich hover tooltips with call relationships

### Phase 4: LLM Agent Integration
- `internal/cli/serve.go`: HTTP server mode
- Custom LSP commands for agent context requests
- Chunked context delivery to fit token limits
- API endpoints: /query, /context, /related

## LANGUAGE SUPPORT PROGRESSION
1. **Go**: Built-in go/ast, go/types for full semantic analysis
2. **Python**: Python ast module via subprocess, handle dynamic typing
3. **TypeScript**: TypeScript compiler API via Node.js subprocess

## INTEGRATION METHODS

### CLI Tool Calling
```bash
repocontext query --json --search "authentication" --max-tokens 3000
# Returns: {functions: [...], types: [...], relationships: {...}, token_count: 2847}
```

### LSP Server  
```go
// Enhanced editor features
textDocument/definition → precise cross-file jumps using semantic index
textDocument/hover → rich context: signature + callers + callees + docs
workspace/symbol → semantic search across entire codebase
```

### HTTP API
```bash
repocontext serve --port 8080
curl "localhost:8080/context?q=ProcessPayment&max_tokens=2000"
```

## KEY TECHNICAL DECISIONS

### AST vs Compiler APIs
- **Start with AST**: Faster implementation, works on broken code, universal availability
- **Upgrade path**: Add go/types for semantic analysis in Phase 2+
- **AST provides**: Function signatures, imports, structure, call sites
- **Compiler APIs add**: Resolved types, cross-file references, semantic validation

### Storage Strategy
- **SQLite**: Fast indexed queries, relationships, metadata
- **MessagePack**: Compact binary storage for detailed AST data  
- **Chunking**: Group related code, load only needed context
- **Incremental**: Hash-based change detection, update only modified files

### Query Optimization
- **Hierarchical addressing**: file/entity/type/line for precise lookup
- **Token-aware**: Fit responses within LLM context windows
- **Relationship prioritization**: direct matches > callers > callees > types
- **Semantic chunking**: Keep related functions together

## VALIDATION CRITERIA
- **Speed**: Index 10k+ file repo in <30s, queries <200ms
- **Accuracy**: AST extraction captures 95%+ of semantic relationships  
- **Integration**: Works with VS Code/Cursor via LSP, callable by LLM agents
- **Scalability**: Handles repos up to 100k+ files via chunking strategy
- **Context Quality**: Returns focused, relevant context within token limits

## REPOSITORY STRUCTURE
```
repocontext/
├── cmd/{repocontext,lsp}/     # CLI and LSP binaries
├── internal/
│   ├── ast/{golang,python,typescript}/    # Language parsers
│   ├── index/{builder,storage,query}/     # Core indexing
│   ├── models/                            # Data structures  
│   ├── cli/                               # CLI commands
│   └── lsp/                               # LSP implementation
├── testdata/                              # Test repositories
└── scripts/                               # Build/test automation
```

## SUCCESS METRICS
- LLM agent receives precise code context in <2000 tokens
- Editor integration provides enhanced navigation/search
- Tool indexes real codebases (5k-50k files) performantly
- Query accuracy enables effective code assistance
repocontext/
├── cmd/
│   ├── repocontext/           # Main CLI binary
│   │   └── main.go
│   └── lsp/                   # LSP server binary
│       └── main.go
├── internal/
│   ├── ast/                   # Language parsers
│   │   ├── parser.go          # Common parser interface
│   │   ├── golang/
│   │   │   └── parser.go      # Go AST parser
│   │   ├── python/
│   │   │   └── parser.go      # Python parser (future)
│   │   └── typescript/
│   │       └── parser.go      # TypeScript parser (future)
│   ├── index/                 # Core indexing logic
│   │   ├── builder.go         # Index builder
│   │   ├── storage.go         # Storage interface
│   │   ├── sqlite.go          # SQLite implementation
│   │   └── query.go           # Query engine
│   ├── models/                # Data structures
│   │   ├── context.go         # Core context types
│   │   ├── function.go        # Function representation
│   │   ├── types.go           # Type definitions
│   │   └── reference.go       # Cross-references
│   ├── cli/                   # CLI commands
│   │   ├── commands.go        # Command definitions
│   │   ├── init.go           # Initialize repo
│   │   ├── build.go          # Build index
│   │   ├── query.go          # Query interface
│   │   └── serve.go          # HTTP server mode
│   └── lsp/                   # LSP implementation
│       ├── server.go          # LSP server
│       ├── handlers.go        # LSP method handlers
│       └── protocol.go        # LSP protocol helpers
├── pkg/                       # Public APIs (if needed)
│   └── client/
│       └── client.go          # Go client library
├── testdata/                  # Test repositories
│   ├── simple-go/             # Basic Go project for testing
│   ├── complex-go/            # Complex Go project
│   └── multi-lang/            # Mixed language project
├── scripts/
│   ├── install.sh             # Installation script
│   └── test-integration.sh    # Integration tests
├── docs/
│   ├── README.md
│   ├── ARCHITECTURE.md        # This design doc
│   └── INTEGRATION.md         # How to integrate with editors
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
├── go.mod
├── go.sum
├── Makefile                   # Build automation
└── README.md

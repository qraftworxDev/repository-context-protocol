```go
// internal/models/context.go
package models

import "time"

// Core data structures
type RepoContext struct {
    Path         string                 `json:"path"`
    Language     string                 `json:"language"`
    LastUpdate   time.Time             `json:"last_update"`
    Files        map[string]*FileContext `json:"files"`
    GlobalIndex  *GlobalIndex          `json:"global_index"`
}

type FileContext struct {
    Path        string      `json:"path"`
    Language    string      `json:"language"`
    Checksum    string      `json:"checksum"`
    ModTime     time.Time   `json:"mod_time"`
    Functions   []Function  `json:"functions"`
    Types       []TypeDef   `json:"types"`
    Variables   []Variable  `json:"variables"`
    Imports     []Import    `json:"imports"`
    Exports     []Export    `json:"exports"`
}

type GlobalIndex struct {
    ByName      map[string][]Reference `json:"by_name"`
    ByFile      map[string][]Reference `json:"by_file"`
    ByType      map[string][]Reference `json:"by_type"`
    CallGraph   map[string][]string    `json:"call_graph"`
    TypeGraph   map[string][]string    `json:"type_graph"`
}
```

```go
// internal/ast/parser.go
package ast

// Language-agnostic parser interface
type LanguageParser interface {
    ParseFile(path string, content []byte) (*models.FileContext, error)
    GetSupportedExtensions() []string
    GetLanguageName() string
}

type ParserRegistry struct {
    parsers map[string]LanguageParser
}

func NewParserRegistry() *ParserRegistry {
    return &ParserRegistry{
        parsers: make(map[string]LanguageParser),
    }
}

func (r *ParserRegistry) Register(parser LanguageParser) {
    for _, ext := range parser.GetSupportedExtensions() {
        r.parsers[ext] = parser
    }
}

func (r *ParserRegistry) GetParser(fileExt string) (LanguageParser, bool) {
    parser, exists := r.parsers[fileExt]
    return parser, exists
}
```

```go
// cmd/repocontext/main.go
package main

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/yourorg/repocontext/internal/cli"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "repocontext",
        Short: "Semantic code repository indexing and querying tool",
    }

    rootCmd.AddCommand(cli.NewInitCommand())
    rootCmd.AddCommand(cli.NewBuildCommand())
    rootCmd.AddCommand(cli.NewQueryCommand())
    rootCmd.AddCommand(cli.NewServeCommand())

    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

```go
// internal/ast/golang/parser.go

package golang

import (
    "go/ast"
    "go/parser"
    "go/token"

    "github.com/yourorg/repocontext/internal/models"
)

type GoParser struct {
    fset *token.FileSet
}

func NewGoParser() *GoParser {
    return &GoParser{
        fset: token.NewFileSet(),
    }
}

func (p *GoParser) ParseFile(path string, content []byte) (*models.FileContext, error) {
    // Parse Go AST and extract functions, types, imports
    file, err := parser.ParseFile(p.fset, path, content, parser.ParseComments)
    if err != nil {
        return nil, err
    }

    ctx := &models.FileContext{
        Path:     path,
        Language: "go",
    }

    // Extract functions, types, etc. from AST
    ast.Inspect(file, func(n ast.Node) bool {
        switch node := n.(type) {
        case *ast.FuncDecl:
            ctx.Functions = append(ctx.Functions, p.extractFunction(node))
        case *ast.TypeSpec:
            ctx.Types = append(ctx.Types, p.extractType(node))
        }
        return true
    })

    return ctx, nil
}
```

```makefile
.PHONY: build test install clean

BINARY_NAME=repocontext
LSP_BINARY_NAME=repocontext-lsp

build:
	go build -o bin/$(BINARY_NAME) cmd/repocontext/main.go
	go build -o bin/$(LSP_BINARY_NAME) cmd/lsp/main.go

test:
	go test ./...

integration-test: build
	./scripts/test-integration.sh

install: build
	cp bin/$(BINARY_NAME) /usr/local/bin/
	cp bin/$(LSP_BINARY_NAME) /usr/local/bin/

clean:
	rm -rf bin/

dev-setup:
	cd testdata/simple-go && ../../bin/$(BINARY_NAME) init
	cd testdata/simple-go && ../../bin/$(BINARY_NAME) build
```

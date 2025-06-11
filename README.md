# Repository Context Protocol

[![codecov](https://codecov.io/gh/qraftworxDev/repository-context-protocol/graph/badge.svg?token=IV7364HI6X)](https://codecov.io/gh/qraftworxDev/repository-context-protocol)
[![Tests](https://github.com/qraftworxDev/repository-context-protocol/workflows/Tests/badge.svg)](https://github.com/qraftworxDev/repository-context-protocol/actions)

A comprehensive tool for analyzing and indexing code repositories with rich context extraction and global call graph analysis.

## Test Coverage

This project maintains high test coverage across all components:

- **Index Builder**: Comprehensive model population and validation
- **Global Enrichment**: Cross-file call analysis and relationship tracking
- **Metadata Validation**: Checksums, timestamps, and signature verification
- **Integration Tests**: End-to-end pipeline validation

### Coverage Reports

To generate coverage reports locally:

```bash
# Generate coverage report
make coverage-report

# View HTML coverage report
open coverage.html

# View terminal coverage summary
make coverage
```

### Coverage Breakdown

Key components with test coverage:
- `internal/index/builder.go` - IndexBuilder with 3-phase pipeline
- `internal/index/enrichment.go` - Global call graph enrichment
- `internal/index/hybrid.go` - Hybrid storage system
- `internal/models/` - All model structures and validation
- `internal/ast/` - Language-specific parsers

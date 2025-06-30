# Test Strategy

## Overview

The repository uses a hybrid testing approach to balance performance with reliability, particularly for tests that use SQLite databases.

## Test Organization

### Standard Tests
- **Location**: All packages except `internal/mcp`
- **Execution**: Parallel (`-race` flag enabled)
- **Performance**: Fast execution using all available CPU cores

### MCP Tests
- **Location**: `internal/mcp` package
- **Execution**: Sequential (`-p 1` flag)
- **Reason**: SQLite database locking issues during parallel execution
- **Trade-off**: Slightly slower but 100% reliable

## Make Targets

### `make test` (Default)
Runs comprehensive test suite with optimal strategy:
1. All non-MCP tests in parallel
2. MCP tests sequentially
3. Both with race detection enabled

### `make test-mcp`
Runs only MCP tests sequentially:
- Useful for debugging MCP-specific issues
- Faster than full test suite when only testing MCP functionality

### `make test-parallel`
Runs all tests in parallel (legacy behavior):
- May cause SQLite locking issues in MCP tests
- Useful for performance testing or when SQLite locks are not a concern

### `make coverage`
Runs all tests with coverage analysis:
- Uses sequential execution to avoid SQLite issues
- Generates HTML coverage report

## SQLite Database Locking Issue

The MCP tests create temporary SQLite databases for integration testing. When run in parallel, multiple test processes can attempt to access the same database files simultaneously, causing:

- `database is locked` errors
- `disk I/O error: no such file or directory` errors
- Intermittent test failures

## Solution Benefits

1. **Reliability**: 100% test pass rate by eliminating race conditions
2. **Performance**: Non-MCP tests still run in parallel for speed
3. **Maintainability**: Clear separation of concerns
4. **CI/CD Friendly**: Consistent behavior across different environments

## Usage Examples

```bash
# Run all tests (recommended)
make test

# Run only MCP tests
make test-mcp

# Run tests with coverage
make coverage

# Run all tests in parallel (may fail)
make test-parallel
```

## CI/CD Integration

For continuous integration, use `make test` to ensure reliable test execution while maintaining good performance for the majority of the test suite.

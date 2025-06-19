# RepoContext CLI Comprehensive Test Suite

This document contains all possible combinations of repocontext CLI commands for comprehensive testing.

## Setup Commands

```bash
# Build the CLI
go build -o bin/repocontext cmd/repocontext/main.go

# Initialize repository (run once)
./bin/repocontext init

# Build index (run once)
./bin/repocontext build
```

## Basic Commands

```bash
# Help commands
./bin/repocontext --help > outputs/help_main.txt
./bin/repocontext help > outputs/help_command.txt
./bin/repocontext query --help > outputs/help_query.txt
./bin/repocontext init --help > outputs/help_init.txt
./bin/repocontext build --help > outputs/help_build.txt

# Version
./bin/repocontext --version > outputs/version.txt
```

## Search Type Tests - Basic (No QueryOptions)

### Function Search
```bash
./bin/repocontext query --function "NewQueryEngine" > outputs/function_basic.txt
./bin/repocontext query --function "SearchByName" > outputs/function_basic2.txt
./bin/repocontext query --function "nonexistent" > outputs/function_notfound.txt
```

### Type Search
```bash
./bin/repocontext query --type "QueryEngine" > outputs/type_basic.txt
./bin/repocontext query --type "SearchResult" > outputs/type_basic2.txt
./bin/repocontext query --type "nonexistent" > outputs/type_notfound.txt
```

### Variable Search
```bash
./bin/repocontext query --variable "storage" > outputs/variable_basic.txt
./bin/repocontext query --variable "engine" > outputs/variable_basic2.txt
./bin/repocontext query --variable "nonexistent" > outputs/variable_notfound.txt
```

### File Search
```bash
./bin/repocontext query --file "query.go" > outputs/file_basic.txt
./bin/repocontext query --file "main.go" > outputs/file_basic2.txt
./bin/repocontext query --file "nonexistent.go" > outputs/file_notfound.txt
```

### Pattern Search (NEW - should now support queryOptions)
```bash
./bin/repocontext query --search "Query*" > outputs/search_basic.txt
./bin/repocontext query --search "*Engine" > outputs/search_basic2.txt
./bin/repocontext query --search "Test*" > outputs/search_basic3.txt
./bin/repocontext query --search "nonexistent*" > outputs/search_notfound.txt
```

### Entity Type Search (NEW - should now support queryOptions)
```bash
./bin/repocontext query --entity-type "function" > outputs/entity_function.txt
./bin/repocontext query --entity-type "type" > outputs/entity_type.txt
./bin/repocontext query --entity-type "variable" > outputs/entity_variable.txt
./bin/repocontext query --entity-type "constant" > outputs/entity_constant.txt
```

## QueryOptions Tests - Single Options

### Include Callers Only
```bash
./bin/repocontext query --function "NewQueryEngine" --include-callers > outputs/function_callers.txt
./bin/repocontext query --search "Query*" --include-callers > outputs/search_callers.txt
./bin/repocontext query --entity-type "function" --include-callers > outputs/entity_callers.txt
```

### Include Callees Only
```bash
./bin/repocontext query --function "NewQueryEngine" --include-callees > outputs/function_callees.txt
./bin/repocontext query --search "Query*" --include-callees > outputs/search_callees.txt
./bin/repocontext query --entity-type "function" --include-callees > outputs/entity_callees.txt
```

### Include Types Only
```bash
./bin/repocontext query --function "NewQueryEngine" --include-types > outputs/function_types.txt
./bin/repocontext query --search "Query*" --include-types > outputs/search_types.txt
./bin/repocontext query --entity-type "function" --include-types > outputs/entity_types.txt
```

### Max Tokens Only
```bash
./bin/repocontext query --function "NewQueryEngine" --max-tokens 100 > outputs/function_tokens100.txt
./bin/repocontext query --function "NewQueryEngine" --max-tokens 500 > outputs/function_tokens500.txt
./bin/repocontext query --search "Query*" --max-tokens 100 > outputs/search_tokens100.txt
./bin/repocontext query --entity-type "function" --max-tokens 200 > outputs/entity_tokens200.txt
```

### Depth Control Only
```bash
./bin/repocontext query --function "NewQueryEngine" --include-callees --depth 1 > outputs/function_depth1.txt
./bin/repocontext query --function "NewQueryEngine" --include-callees --depth 2 > outputs/function_depth2.txt
./bin/repocontext query --function "NewQueryEngine" --include-callees --depth 3 > outputs/function_depth3.txt
./bin/repocontext query --search "Query*" --include-callees --depth 1 > outputs/search_depth1.txt
./bin/repocontext query --search "Query*" --include-callees --depth 2 > outputs/search_depth2.txt
./bin/repocontext query --entity-type "function" --include-callees --depth 1 > outputs/entity_depth1.txt
./bin/repocontext query --entity-type "function" --include-callees --depth 2 > outputs/entity_depth2.txt
```

## Format Options

### JSON Format
```bash
./bin/repocontext query --function "NewQueryEngine" --json > outputs/function_json.txt
./bin/repocontext query --search "Query*" --json > outputs/search_json.txt
./bin/repocontext query --entity-type "function" --json > outputs/entity_json.txt
./bin/repocontext query --function "NewQueryEngine" --format json > outputs/function_format_json.txt
```

### Text Format (Explicit)
```bash
./bin/repocontext query --function "NewQueryEngine" --format text > outputs/function_format_text.txt
./bin/repocontext query --search "Query*" --format text > outputs/search_format_text.txt
```

### Verbose and Compact
```bash
./bin/repocontext query --function "NewQueryEngine" --verbose > outputs/function_verbose.txt
./bin/repocontext query --function "NewQueryEngine" --compact > outputs/function_compact.txt
./bin/repocontext query --search "Query*" --verbose > outputs/search_verbose.txt
./bin/repocontext query --search "Query*" --compact > outputs/search_compact.txt
```

## QueryOptions Combinations

### Full Options (All Search Types)
```bash
./bin/repocontext query --function "NewQueryEngine" --include-callers --include-callees --include-types --max-tokens 1000 --depth 2 --json > outputs/function_full.txt

./bin/repocontext query --type "QueryEngine" --include-callers --include-callees --include-types --max-tokens 1000 --depth 2 --json > outputs/type_full.txt

./bin/repocontext query --variable "storage" --include-callers --include-callees --include-types --max-tokens 1000 --depth 2 --json > outputs/variable_full.txt

./bin/repocontext query --file "query.go" --include-callers --include-callees --include-types --max-tokens 1000 --depth 2 --json > outputs/file_full.txt

# These should now work with all options (NEW functionality)
./bin/repocontext query --search "Query*" --include-callers --include-callees --include-types --max-tokens 1000 --depth 2 --json > outputs/search_full.txt

./bin/repocontext query --entity-type "function" --include-callers --include-callees --include-types --max-tokens 1000 --depth 2 --json > outputs/entity_full.txt
```

### Call Graph Combinations
```bash
./bin/repocontext query --function "NewQueryEngine" --include-callers --include-callees > outputs/function_both_graphs.txt
./bin/repocontext query --search "Query*" --include-callers --include-callees > outputs/search_both_graphs.txt
./bin/repocontext query --entity-type "function" --include-callers --include-callees > outputs/entity_both_graphs.txt
```

### Token Limits with Other Options
```bash
./bin/repocontext query --function "NewQueryEngine" --include-callees --max-tokens 50 > outputs/function_low_tokens.txt
./bin/repocontext query --search "Query*" --include-callees --max-tokens 50 > outputs/search_low_tokens.txt
./bin/repocontext query --entity-type "function" --include-callees --max-tokens 100 > outputs/entity_low_tokens.txt
```

## Error Cases

### Invalid Combinations
```bash
# Multiple search criteria (should fail)
./bin/repocontext query --function "test" --type "test" 2> outputs/error_multiple_criteria.txt

# Invalid entity type
./bin/repocontext query --entity-type "invalid" 2> outputs/error_invalid_entity.txt

# Invalid format
./bin/repocontext query --function "test" --format "invalid" 2> outputs/error_invalid_format.txt

# Negative values
./bin/repocontext query --function "test" --max-tokens -1 2> outputs/error_negative_tokens.txt
./bin/repocontext query --function "test" --depth -1 2> outputs/error_negative_depth.txt

# No search criteria
./bin/repocontext query 2> outputs/error_no_criteria.txt
```

### Repository Errors
```bash
# Test without initialization (run from temp directory)
cd /tmp && /path/to/repocontext query --function "test" 2> outputs/error_no_init.txt
cd - # return to original directory
```

## Edge Cases

### Empty Results
```bash
./bin/repocontext query --function "CompletelyNonExistent" --include-callers --include-callees > outputs/edge_empty_with_options.txt
./bin/repocontext query --search "ZZZ*" --include-callers > outputs/edge_empty_search.txt
```

### Large Queries
```bash
./bin/repocontext query --entity-type "function" --include-callers --include-callees --max-tokens 10000 > outputs/edge_large_query.txt
./bin/repocontext query --search "*" --max-tokens 5000 > outputs/edge_wildcard_search.txt
```

### Special Characters
```bash
./bin/repocontext query --search "*" > outputs/edge_wildcard_only.txt
./bin/repocontext query --file "*.go" > outputs/edge_file_wildcard.txt
```

## Performance Tests

### High Depth
```bash
./bin/repocontext query --entity-type "function" --include-callees --depth 5 > outputs/perf_high_depth.txt
```

### No Token Limits
```bash
./bin/repocontext query --entity-type "function" --include-callers --include-callees --max-tokens 0 > outputs/perf_no_limit.txt
```

## Comparison Tests (Before vs After Fix)

These commands specifically test the NEW functionality that was added:

### Pattern Search with QueryOptions (Should now work)
```bash
./bin/repocontext query --search "Query*" --include-callers --include-callees --max-tokens 500 --depth 2 > outputs/new_search_with_options.txt
```

### Entity Type Search with QueryOptions (Should now work)
```bash
./bin/repocontext query --entity-type "function" --include-callers --include-callees --max-tokens 500 --depth 2 > outputs/new_entity_with_options.txt
```

## Summary Files to Generate

After running all tests, create these summary files:

```bash
# Count successful vs failed commands
echo "=== TEST SUMMARY ===" > outputs/test_summary.txt
echo "Total files generated: $(ls outputs/*.txt | wc -l)" >> outputs/test_summary.txt
echo "Error files: $(ls outputs/error_*.txt 2>/dev/null | wc -l)" >> outputs/test_summary.txt
echo "Success files: $(ls outputs/*.txt | grep -v error_ | wc -l)" >> outputs/test_summary.txt

# Check for call graph presence in new functionality
echo "=== CALL GRAPH CHECK ===" >> outputs/test_summary.txt
echo "Search commands with call graphs: $(grep -l "Call Graph:" outputs/search_*.txt | wc -l)" >> outputs/test_summary.txt
echo "Entity commands with call graphs: $(grep -l "Call Graph:" outputs/entity_*.txt | wc -l)" >> outputs/test_summary.txt
```

## Key Things to Verify

1. **NEW Functionality**: `outputs/search_*.txt` and `outputs/entity_*.txt` should now have call graph sections
2. **Consistency**: All search types should behave similarly with same options
3. **Token Limits**: Files with `tokens` in name should respect limits
4. **Depth Control**: `depth1.txt` < `depth2.txt` < `depth3.txt` in terms of results
5. **Error Handling**: `error_*.txt` files should contain appropriate error messages
6. **JSON Format**: Files ending in `_json.txt` should contain valid JSON with expected fields

Run all these commands and review the output files to ensure the queryOptions functionality works correctly across all search types.

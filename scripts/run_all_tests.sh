#!/bin/bash

# RepoContext CLI Comprehensive Test Runner
# This script runs all possible command combinations and saves outputs

set -e

echo "=== Starting RepoContext CLI Comprehensive Test ==="

# Create outputs directory
mkdir -p outputs

echo "=== Building CLI ==="
go build -o bin/repocontext cmd/repocontext/main.go

echo "=== Initializing Repository ==="
./bin/repocontext init || echo "Repository already initialized"

echo "=== Building Index ==="
./bin/repocontext build

echo "=== Running Basic Commands ==="
./bin/repocontext --help > outputs/help_main.txt
./bin/repocontext help > outputs/help_command.txt
./bin/repocontext query --help > outputs/help_query.txt
./bin/repocontext init --help > outputs/help_init.txt
./bin/repocontext build --help > outputs/help_build.txt
./bin/repocontext --version > outputs/version.txt

echo "=== Running Basic Search Tests ==="
# Function Search
./bin/repocontext query --function "NewQueryEngine" > outputs/function_basic.txt
./bin/repocontext query --function "SearchByName" > outputs/function_basic2.txt
./bin/repocontext query --function "nonexistent" > outputs/function_notfound.txt

# Type Search
./bin/repocontext query --type "QueryEngine" > outputs/type_basic.txt
./bin/repocontext query --type "SearchResult" > outputs/type_basic2.txt
./bin/repocontext query --type "nonexistent" > outputs/type_notfound.txt

# Variable Search
./bin/repocontext query --variable "storage" > outputs/variable_basic.txt
./bin/repocontext query --variable "engine" > outputs/variable_basic2.txt
./bin/repocontext query --variable "nonexistent" > outputs/variable_notfound.txt

# File Search
./bin/repocontext query --file "query.go" > outputs/file_basic.txt
./bin/repocontext query --file "main.go" > outputs/file_basic2.txt
./bin/repocontext query --file "nonexistent.go" > outputs/file_notfound.txt

echo "=== Running NEW Pattern Search Tests ==="
./bin/repocontext query --search "Query*" > outputs/search_basic.txt
./bin/repocontext query --search "*Engine" > outputs/search_basic2.txt
./bin/repocontext query --search "Test*" > outputs/search_basic3.txt
./bin/repocontext query --search "nonexistent*" > outputs/search_notfound.txt

echo "=== Running NEW Entity Type Search Tests ==="
./bin/repocontext query --entity-type "function" > outputs/entity_function.txt
./bin/repocontext query --entity-type "type" > outputs/entity_type.txt
./bin/repocontext query --entity-type "variable" > outputs/entity_variable.txt
./bin/repocontext query --entity-type "constant" > outputs/entity_constant.txt

echo "=== Running QueryOptions Tests - Include Callers ==="
./bin/repocontext query --function "NewQueryEngine" --include-callers > outputs/function_callers.txt
./bin/repocontext query --search "Query*" --include-callers > outputs/search_callers.txt
./bin/repocontext query --entity-type "function" --include-callers > outputs/entity_callers.txt

echo "=== Running QueryOptions Tests - Include Callees ==="
./bin/repocontext query --function "NewQueryEngine" --include-callees > outputs/function_callees.txt
./bin/repocontext query --search "Query*" --include-callees > outputs/search_callees.txt
./bin/repocontext query --entity-type "function" --include-callees > outputs/entity_callees.txt

echo "=== Running QueryOptions Tests - Include Types ==="
./bin/repocontext query --function "NewQueryEngine" --include-types > outputs/function_types.txt
./bin/repocontext query --search "Query*" --include-types > outputs/search_types.txt
./bin/repocontext query --entity-type "function" --include-types > outputs/entity_types.txt

echo "=== Running QueryOptions Tests - Max Tokens ==="
./bin/repocontext query --function "NewQueryEngine" --max-tokens 100 > outputs/function_tokens100.txt
./bin/repocontext query --function "NewQueryEngine" --max-tokens 500 > outputs/function_tokens500.txt
./bin/repocontext query --search "Query*" --max-tokens 100 > outputs/search_tokens100.txt
./bin/repocontext query --entity-type "function" --max-tokens 200 > outputs/entity_tokens200.txt

echo "=== Running QueryOptions Tests - Depth Control ==="
./bin/repocontext query --function "NewQueryEngine" --include-callees --depth 1 > outputs/function_depth1.txt
./bin/repocontext query --function "NewQueryEngine" --include-callees --depth 2 > outputs/function_depth2.txt
./bin/repocontext query --function "NewQueryEngine" --include-callees --depth 3 > outputs/function_depth3.txt
./bin/repocontext query --search "Query*" --include-callees --depth 1 > outputs/search_depth1.txt
./bin/repocontext query --search "Query*" --include-callees --depth 2 > outputs/search_depth2.txt
./bin/repocontext query --entity-type "function" --include-callees --depth 1 > outputs/entity_depth1.txt
./bin/repocontext query --entity-type "function" --include-callees --depth 2 > outputs/entity_depth2.txt

echo "=== Running Format Options Tests ==="
./bin/repocontext query --function "NewQueryEngine" --json > outputs/function_json.txt
./bin/repocontext query --search "Query*" --json > outputs/search_json.txt
./bin/repocontext query --entity-type "function" --json > outputs/entity_json.txt
./bin/repocontext query --function "NewQueryEngine" --format json > outputs/function_format_json.txt
./bin/repocontext query --function "NewQueryEngine" --format text > outputs/function_format_text.txt
./bin/repocontext query --search "Query*" --format text > outputs/search_format_text.txt

./bin/repocontext query --function "NewQueryEngine" --verbose > outputs/function_verbose.txt
./bin/repocontext query --function "NewQueryEngine" --compact > outputs/function_compact.txt
./bin/repocontext query --search "Query*" --verbose > outputs/search_verbose.txt
./bin/repocontext query --search "Query*" --compact > outputs/search_compact.txt

echo "=== Running Full QueryOptions Combinations ==="
./bin/repocontext query --function "NewQueryEngine" --include-callers --include-callees --include-types --max-tokens 1000 --depth 2 --json > outputs/function_full.txt
./bin/repocontext query --type "QueryEngine" --include-callers --include-callees --include-types --max-tokens 1000 --depth 2 --json > outputs/type_full.txt
./bin/repocontext query --variable "storage" --include-callers --include-callees --include-types --max-tokens 1000 --depth 2 --json > outputs/variable_full.txt
./bin/repocontext query --file "query.go" --include-callers --include-callees --include-types --max-tokens 1000 --depth 2 --json > outputs/file_full.txt

# NEW functionality - these should now work with all options
./bin/repocontext query --search "Query*" --include-callers --include-callees --include-types --max-tokens 1000 --depth 2 --json > outputs/search_full.txt
./bin/repocontext query --entity-type "function" --include-callers --include-callees --include-types --max-tokens 1000 --depth 2 --json > outputs/entity_full.txt

echo "=== Running Call Graph Combinations ==="
./bin/repocontext query --function "NewQueryEngine" --include-callers --include-callees > outputs/function_both_graphs.txt
./bin/repocontext query --search "Query*" --include-callers --include-callees > outputs/search_both_graphs.txt
./bin/repocontext query --entity-type "function" --include-callers --include-callees > outputs/entity_both_graphs.txt

echo "=== Running Token Limits with Other Options ==="
./bin/repocontext query --function "NewQueryEngine" --include-callees --max-tokens 50 > outputs/function_low_tokens.txt
./bin/repocontext query --search "Query*" --include-callees --max-tokens 50 > outputs/search_low_tokens.txt
./bin/repocontext query --entity-type "function" --include-callees --max-tokens 100 > outputs/entity_low_tokens.txt

echo "=== Running Error Cases ==="
./bin/repocontext query --function "test" --type "test" 2> outputs/error_multiple_criteria.txt || true
./bin/repocontext query --entity-type "invalid" 2> outputs/error_invalid_entity.txt || true
./bin/repocontext query --function "test" --format "invalid" 2> outputs/error_invalid_format.txt || true
./bin/repocontext query --function "test" --max-tokens -1 2> outputs/error_negative_tokens.txt || true
./bin/repocontext query --function "test" --depth -1 2> outputs/error_negative_depth.txt || true
./bin/repocontext query 2> outputs/error_no_criteria.txt || true

echo "=== Running Edge Cases ==="
./bin/repocontext query --function "CompletelyNonExistent" --include-callers --include-callees > outputs/edge_empty_with_options.txt
./bin/repocontext query --search "ZZZ*" --include-callers > outputs/edge_empty_search.txt
./bin/repocontext query --entity-type "function" --include-callers --include-callees --max-tokens 10000 > outputs/edge_large_query.txt
./bin/repocontext query --search "*" --max-tokens 5000 > outputs/edge_wildcard_search.txt
./bin/repocontext query --search "*" > outputs/edge_wildcard_only.txt

echo "=== Running Performance Tests ==="
./bin/repocontext query --entity-type "function" --include-callees --depth 5 > outputs/perf_high_depth.txt
./bin/repocontext query --entity-type "function" --include-callers --include-callees --max-tokens 0 > outputs/perf_no_limit.txt

echo "=== Running NEW Functionality Tests ==="
./bin/repocontext query --search "Query*" --include-callers --include-callees --max-tokens 500 --depth 2 > outputs/new_search_with_options.txt
./bin/repocontext query --entity-type "function" --include-callers --include-callees --max-tokens 500 --depth 2 > outputs/new_entity_with_options.txt

echo "=== Generating Summary ==="
echo "=== TEST SUMMARY ===" > outputs/test_summary.txt
echo "Total files generated: $(ls outputs/*.txt | wc -l)" >> outputs/test_summary.txt
echo "Error files: $(ls outputs/error_*.txt 2>/dev/null | wc -l)" >> outputs/test_summary.txt
echo "Success files: $(ls outputs/*.txt | grep -v error_ | wc -l)" >> outputs/test_summary.txt

echo "=== CALL GRAPH CHECK ===" >> outputs/test_summary.txt
echo "Search commands with call graphs: $(grep -l "Call Graph:" outputs/search_*.txt 2>/dev/null | wc -l)" >> outputs/test_summary.txt
echo "Entity commands with call graphs: $(grep -l "Call Graph:" outputs/entity_*.txt 2>/dev/null | wc -l)" >> outputs/test_summary.txt

echo "=== Test Complete! ==="
echo "Check the outputs/ directory for all test results."
echo "Review outputs/test_summary.txt for a summary of results."
echo ""
echo "Key files to check:"
echo "- outputs/search_*.txt (should now have call graphs)"
echo "- outputs/entity_*.txt (should now have call graphs)"
echo "- outputs/*_full.txt (complete queryOptions functionality)"
echo "- outputs/error_*.txt (error handling)"

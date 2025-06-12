package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"repository-context-protocol/internal/index"

	"github.com/spf13/cobra"
)

// Constants for the query command
const (
	DefaultDepth = 2
)

// QueryFlags holds all the flags for the query command
type QueryFlags struct {
	// Search criteria flags
	Function   string
	Type       string
	Variable   string
	File       string
	Search     string
	EntityType string

	// Context flags
	IncludeCallers bool
	IncludeCallees bool
	IncludeTypes   bool
	Depth          int
	MaxTokens      int

	// Output flags
	Format  string
	JSON    bool
	Verbose bool
	Compact bool

	// Repository flags
	Path string
}

// NewQueryCommand creates the query command for searching indexed repositories
func NewQueryCommand() *cobra.Command {
	flags := &QueryFlags{}

	cmd := createQueryCommand(flags)
	addQueryFlags(cmd, flags)
	return cmd
}

func createQueryCommand(flags *QueryFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "query",
		Short: "Search and query indexed repository data",
		Long: `Query the semantic index to search for code entities and retrieve contextual information.

This command provides powerful search capabilities over the indexed repository data,
allowing you to find functions, types, variables, and other code entities by name,
type, or pattern. It can also include call graph information and format output
for both human consumption and LLM integration.

Search Types:
  --function, -f    Search for a specific function by name
  --type, -t        Search for a specific type by name
  --variable, -v    Search for a specific variable by name
  --file            Search for all entities within a specific file
  --search, -s      Search using patterns (supports wildcards)
  --entity-type     Search for all entities of a specific type

Context Options:
  --include-callers Include functions that call the target
  --include-callees Include functions called by the target
  --include-types   Include related type definitions
  --depth           Maximum depth for relationship traversal (default: 2)
  --max-tokens      Maximum tokens for LLM consumption (0 = no limit)

Output Options:
  --format          Output format: text, json (default: text)
  --json            Shorthand for --format json
  --verbose         Include detailed information
  --compact         Minimal output

Examples:
  # Search for a function by name
  repocontext query --function main

  # Search for all functions with call graph
  repocontext query --entity-type function --include-callers --include-callees

  # Search with JSON output for LLM consumption
  repocontext query --function processData --json --max-tokens 1000

  # Search in a specific file
  repocontext query --file main.go

  # Pattern search
  repocontext query --search "Test*" --compact`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuery(flags)
		},
	}
}

func addQueryFlags(cmd *cobra.Command, flags *QueryFlags) {
	// Search criteria flags
	cmd.Flags().StringVarP(&flags.Function, "function", "f", "", "Search for a specific function by name")
	cmd.Flags().StringVarP(&flags.Type, "type", "t", "", "Search for a specific type by name")
	cmd.Flags().StringVarP(&flags.Variable, "variable", "v", "", "Search for a specific variable by name")
	cmd.Flags().StringVar(&flags.File, "file", "", "Search for all entities within a specific file")
	cmd.Flags().StringVarP(&flags.Search, "search", "s", "", "Search using patterns (supports wildcards)")
	cmd.Flags().StringVar(&flags.EntityType, "entity-type", "",
		"Search for all entities of a specific type (function, type, variable, constant)")

	// Context flags
	cmd.Flags().BoolVar(&flags.IncludeCallers, "include-callers", false, "Include functions that call the target")
	cmd.Flags().BoolVar(&flags.IncludeCallees, "include-callees", false, "Include functions called by the target")
	cmd.Flags().BoolVar(&flags.IncludeTypes, "include-types", false, "Include related type definitions")
	cmd.Flags().IntVar(&flags.Depth, "depth", DefaultDepth, "Maximum depth for relationship traversal")
	cmd.Flags().IntVar(&flags.MaxTokens, "max-tokens", 0, "Maximum tokens for LLM consumption (0 = no limit)")

	// Output flags
	cmd.Flags().StringVar(&flags.Format, "format", "text", "Output format: text, json")
	cmd.Flags().BoolVar(&flags.JSON, "json", false, "Output in JSON format (shorthand for --format json)")
	cmd.Flags().BoolVar(&flags.Verbose, "verbose", false, "Include detailed information")
	cmd.Flags().BoolVar(&flags.Compact, "compact", false, "Minimal output")

	// Repository flags
	cmd.Flags().StringVarP(&flags.Path, "path", "p", ".", "Path to the repository (defaults to current directory)")
}

// runQuery executes the query command
func runQuery(flags *QueryFlags) error {
	// Validate inputs
	if err := validateQueryInputs(flags); err != nil {
		return err
	}

	// Handle JSON flag
	if flags.JSON {
		flags.Format = "json"
	}

	// Execute search
	result, err := executeSearch(flags)
	if err != nil {
		return err
	}

	// Output results
	return outputResults(result, flags)
}

func validateQueryInputs(flags *QueryFlags) error {
	// Validate repository exists
	if err := validateRepository(flags.Path); err != nil {
		return err
	}

	// Validate search criteria
	if err := validateSearchCriteria(flags); err != nil {
		return err
	}

	// Validate other flags
	return validateFlags(flags)
}

func executeSearch(flags *QueryFlags) (*index.SearchResult, error) {
	// Initialize storage and query engine
	storage := index.NewHybridStorage(filepath.Join(flags.Path, ".repocontext"))
	if err := storage.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer storage.Close()

	queryEngine := index.NewQueryEngine(storage)

	// Build query options
	queryOptions := index.QueryOptions{
		IncludeCallers: flags.IncludeCallers,
		IncludeCallees: flags.IncludeCallees,
		IncludeTypes:   flags.IncludeTypes,
		MaxDepth:       flags.Depth,
		MaxTokens:      flags.MaxTokens,
		Format:         flags.Format,
	}

	// Execute search based on criteria
	var result *index.SearchResult
	var err error

	switch {
	case flags.Function != "":
		result, err = queryEngine.SearchByNameWithOptions(flags.Function, queryOptions)
	case flags.Type != "":
		result, err = queryEngine.SearchByNameWithOptions(flags.Type, queryOptions)
	case flags.Variable != "":
		result, err = queryEngine.SearchByNameWithOptions(flags.Variable, queryOptions)
	case flags.File != "":
		result, err = queryEngine.SearchInFile(flags.File)
	case flags.Search != "":
		result, err = queryEngine.SearchByPattern(flags.Search)
	case flags.EntityType != "":
		result, err = queryEngine.SearchByType(flags.EntityType)
	default:
		return nil, fmt.Errorf("no search criteria specified")
	}

	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return result, nil
}

func outputResults(result *index.SearchResult, flags *QueryFlags) error {
	// Initialize storage for formatting (if needed)
	storage := index.NewHybridStorage(filepath.Join(flags.Path, ".repocontext"))
	if err := storage.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize storage for formatting: %w", err)
	}
	defer storage.Close()

	queryEngine := index.NewQueryEngine(storage)

	// Format and output results
	output, err := queryEngine.FormatResults(result, flags.Format)
	if err != nil {
		return fmt.Errorf("failed to format results: %w", err)
	}

	fmt.Print(string(output))
	return nil
}

// validateRepository checks if the repository is initialized
func validateRepository(path string) error {
	repoContextPath := filepath.Join(path, ".repocontext")
	if _, err := os.Stat(repoContextPath); os.IsNotExist(err) {
		return fmt.Errorf("repository not initialized at %s (run 'repocontext init' first)", path)
	}
	return nil
}

// validateSearchCriteria ensures at least one search criterion is provided
func validateSearchCriteria(flags *QueryFlags) error {
	if flags.Function == "" && flags.Type == "" && flags.Variable == "" &&
		flags.File == "" && flags.Search == "" && flags.EntityType == "" {
		return fmt.Errorf("at least one search criterion must be specified")
	}
	return nil
}

// validateFlags validates flag values
func validateFlags(flags *QueryFlags) error {
	// Validate format
	validFormats := []string{"text", "json"}
	if !contains(validFormats, flags.Format) {
		return fmt.Errorf("invalid format '%s', must be one of: %s", flags.Format, strings.Join(validFormats, ", "))
	}

	// Validate depth
	if flags.Depth < 0 {
		return fmt.Errorf("depth must be non-negative, got %d", flags.Depth)
	}

	// Validate max-tokens
	if flags.MaxTokens < 0 {
		return fmt.Errorf("max-tokens must be non-negative, got %d", flags.MaxTokens)
	}

	// Validate entity-type if specified
	if flags.EntityType != "" {
		validEntityTypes := []string{"function", "type", "variable", "constant"}
		if !contains(validEntityTypes, flags.EntityType) {
			return fmt.Errorf("invalid entity-type '%s', must be one of: %s", flags.EntityType, strings.Join(validEntityTypes, ", "))
		}
	}

	return nil
}

// contains checks if a string slice contains a given string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

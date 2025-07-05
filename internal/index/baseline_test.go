package index

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"repository-context-protocol/internal/models"
)

// BenchmarkCurrentPerformance establishes baseline performance metrics
func BenchmarkCurrentPerformance(b *testing.B) {
	testCases := []struct {
		name      string
		queryType string
		pattern   string
		dataSize  string
	}{
		{"Simple Name Query", "name", "ProcessUser", "small"},
		{"Type Query", "type", "function", "medium"},
		{"Pattern Query", "pattern", "Process*", "medium"},
		{"Complex Pattern", "pattern", "*User*Process*", "large"},
		{"File Query", "file", "internal/index/query.go", "medium"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			qe := setupQueryEngineForBenchmark(tc.dataSize, b)
			defer cleanupQueryEngine(qe)

			b.ResetTimer()
			b.ReportAllocs()

			var memStatsBefore, memStatsAfter runtime.MemStats
			runtime.GC() // Clean up before measuring
			runtime.ReadMemStats(&memStatsBefore)

			start := time.Now()

			for i := 0; i < b.N; i++ {
				var result *SearchResult
				var err error

				switch tc.queryType {
				case "name":
					result, err = qe.SearchByName(tc.pattern)
				case "type":
					result, err = qe.SearchByType(tc.pattern)
				case "pattern":
					result, err = qe.SearchByPattern(tc.pattern)
				case "file":
					result, err = qe.SearchInFile(tc.pattern)
				}

				if err != nil {
					b.Fatal(err)
				}

				// Ensure we access the result to prevent optimization
				_ = len(result.Entries)
			}

			elapsed := time.Since(start)
			runtime.ReadMemStats(&memStatsAfter)

			// Calculate per-operation metrics
			avgLatency := elapsed / time.Duration(b.N)
			memoryDelta := memStatsAfter.TotalAlloc - memStatsBefore.TotalAlloc
			allocsDelta := memStatsAfter.Mallocs - memStatsBefore.Mallocs

			b.ReportMetric(float64(avgLatency.Nanoseconds()), "ns/op")
			b.ReportMetric(float64(memoryDelta), "B/op")
			b.ReportMetric(float64(allocsDelta), "allocs/op")
		})
	}
}

// BenchmarkPatternMatchingMemory measures memory usage specifically for pattern matching
func BenchmarkPatternMatchingMemory(b *testing.B) {
	qe := setupQueryEngineForBenchmark("large", b)
	defer cleanupQueryEngine(qe)

	patterns := []string{
		"Process*",
		"*User*",
		"Handle*Request*",
		"*Service",
	}

	for _, pattern := range patterns {
		b.Run(fmt.Sprintf("Pattern_%s", pattern), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				result, err := qe.SearchByPattern(pattern)
				if err != nil {
					b.Fatal(err)
				}
				// Access result to prevent optimization
				_ = len(result.Entries)
			}
		})
	}
}

// BenchmarkMultipleTypeQueries measures N+1 query pattern performance
func BenchmarkMultipleTypeQueries(b *testing.B) {
	qe := setupQueryEngineForBenchmark("medium", b)
	defer cleanupQueryEngine(qe)

	entityTypes := []string{"function", "variable", "constant", "struct", "interface"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		totalEntries := 0

		for _, entityType := range entityTypes {
			result, err := qe.SearchByType(entityType)
			if err != nil {
				b.Fatal(err)
			}
			totalEntries += len(result.Entries)
		}

		// Ensure we use the result
		_ = totalEntries
	}
}

// setupQueryEngineForBenchmark creates a test query engine with data
func setupQueryEngineForBenchmark(dataSize string, b *testing.B) *QueryEngine {
	tmpDir, err := os.MkdirTemp("", "benchmark_test_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}

	storage := NewHybridStorage(tmpDir)
	if err := storage.Initialize(); err != nil {
		b.Fatalf("Failed to initialize storage: %v", err)
	}

	// Generate test data based on size
	var numFiles, itemsPerFile int
	switch dataSize {
	case "small":
		numFiles, itemsPerFile = 5, 10
	case "medium":
		numFiles, itemsPerFile = 20, 25
	case "large":
		numFiles, itemsPerFile = 50, 50
	default:
		numFiles, itemsPerFile = 10, 20
	}

	// Create test data
	for f := 0; f < numFiles; f++ {
		fileContext := createTestFileContext(f, itemsPerFile)
		if err := storage.StoreFileContext(fileContext); err != nil {
			b.Fatalf("Failed to store test data: %v", err)
		}
	}

	return NewQueryEngine(storage)
}

// createTestFileContext creates a test file context with various entities
func createTestFileContext(fileNum, itemsPerFile int) *models.FileContext {
	fileName := fmt.Sprintf("test_file_%d.go", fileNum)

	fileContext := &models.FileContext{
		Path:      fileName,
		Functions: make([]models.Function, 0),
		Variables: make([]models.Variable, 0),
		Constants: make([]models.Constant, 0),
		Types:     make([]models.TypeDef, 0),
	}

	// Add functions
	for i := 0; i < itemsPerFile; i++ {
		function := models.Function{
			Name:      fmt.Sprintf("ProcessUser%d", i),
			Signature: fmt.Sprintf("func ProcessUser%d() error", i),
			StartLine: i*10 + 1,
			EndLine:   i*10 + 5,
			Calls:     []string{fmt.Sprintf("HandleRequest%d", i)},
		}
		fileContext.Functions = append(fileContext.Functions, function)
	}

	// Add variables
	for i := 0; i < itemsPerFile/2; i++ {
		variable := models.Variable{
			Name:      fmt.Sprintf("userVar%d", i),
			Type:      "string",
			StartLine: i*5 + 100,
			EndLine:   i*5 + 100,
		}
		fileContext.Variables = append(fileContext.Variables, variable)
	}

	// Add constants
	for i := 0; i < itemsPerFile/3; i++ {
		constant := models.Constant{
			Name:      fmt.Sprintf("UserConstant%d", i),
			Type:      "int",
			Value:     fmt.Sprintf("%d", i),
			StartLine: i*3 + 200,
			EndLine:   i*3 + 200,
		}
		fileContext.Constants = append(fileContext.Constants, constant)
	}

	// Add types
	for i := 0; i < itemsPerFile/4; i++ {
		typeDef := models.TypeDef{
			Name:      fmt.Sprintf("UserService%d", i),
			Kind:      "struct",
			StartLine: i*8 + 300,
			EndLine:   i*8 + 307,
		}
		fileContext.Types = append(fileContext.Types, typeDef)
	}

	return fileContext
}

// cleanupQueryEngine cleans up test resources
func cleanupQueryEngine(qe *QueryEngine) {
	if qe != nil && qe.storage != nil {
		_ = qe.storage.Close()
		// Clean up temp directory
		if baseDir := qe.storage.baseDir; baseDir != "" {
			_ = os.RemoveAll(baseDir)
		}
	}
}

package index

import (
	"runtime"
	"testing"
	"time"
)

// BenchmarkRealRepo tests performance against the actual repository data
func BenchmarkRealRepo(b *testing.B) {
	// Use the existing index for this repository
	storage := NewHybridStorage(".repocontext")
	if err := storage.Initialize(); err != nil {
		b.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	qe := NewQueryEngine(storage)

	testCases := []struct {
		name        string
		queryType   string
		query       string
		description string
	}{
		{"NameQuery_SearchByName", "name", "SearchByName", "Find specific function"},
		{"NameQuery_QueryEngine", "name", "QueryEngine", "Find struct type"},
		{"TypeQuery_Function", "type", "function", "All functions (N+1 pattern test)"},
		{"TypeQuery_Variable", "type", "variable", "All variables"},
		{"PatternQuery_Process", "pattern", "Process*", "Pattern matching"},
		{"PatternQuery_Query", "pattern", "*Query*", "Complex pattern"},
		{"PatternQuery_Index", "pattern", "Index*", "Index-related functions"},
		{"FileQuery_QueryGo", "file", "internal/index/query.go", "Single file entities"},
		{"FileQuery_SqliteGo", "file", "internal/index/sqlite.go", "Database file entities"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			var memStatsBefore, memStatsAfter runtime.MemStats
			runtime.GC() // Clean up before measuring
			runtime.ReadMemStats(&memStatsBefore)

			b.ResetTimer()
			b.ReportAllocs()

			start := time.Now()
			var totalResults int

			for i := 0; i < b.N; i++ {
				var result *SearchResult
				var err error

				switch tc.queryType {
				case "name":
					result, err = qe.SearchByName(tc.query)
				case "type":
					result, err = qe.SearchByType(tc.query)
				case "pattern":
					result, err = qe.SearchByPattern(tc.query)
				case "file":
					result, err = qe.SearchInFile(tc.query)
				}

				if err != nil {
					b.Fatal(err)
				}

				totalResults += len(result.Entries)
			}

			elapsed := time.Since(start)
			runtime.ReadMemStats(&memStatsAfter)

			// Report metrics
			avgLatency := elapsed / time.Duration(b.N)
			memoryDelta := memStatsAfter.TotalAlloc - memStatsBefore.TotalAlloc
			allocsDelta := memStatsAfter.Mallocs - memStatsBefore.Mallocs

			b.ReportMetric(float64(avgLatency.Nanoseconds()), "ns/op")
			b.ReportMetric(float64(memoryDelta)/float64(b.N), "B/op")
			b.ReportMetric(float64(allocsDelta)/float64(b.N), "allocs/op")
			b.ReportMetric(float64(totalResults)/float64(b.N), "results/op")

			b.Logf("%s: %d results, %v avg latency", tc.description, totalResults/b.N, avgLatency)
		})
	}
}

// BenchmarkRealRepoWithCache tests cache effectiveness with repeated queries
func BenchmarkRealRepoWithCache(b *testing.B) {
	storage := NewHybridStorage(".repocontext")
	if err := storage.Initialize(); err != nil {
		b.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	qe := NewQueryEngine(storage)

	// Pre-populate cache with a query
	_, _ = qe.SearchByPattern("Process*")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := qe.SearchByPattern("Process*")
		if err != nil {
			b.Fatal(err)
		}
		_ = len(result.Entries) // Use result
	}
}

// BenchmarkRealRepoN1Pattern specifically tests the N+1 elimination
func BenchmarkRealRepoN1Pattern(b *testing.B) {
	storage := NewHybridStorage(".repocontext")
	if err := storage.Initialize(); err != nil {
		b.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	qe := NewQueryEngine(storage)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// This query hits multiple entity types - tests batch query effectiveness
		result, err := qe.SearchByPatternWithOptions("*", QueryOptions{IncludeTypes: true})
		if err != nil {
			b.Fatal(err)
		}
		_ = len(result.Entries) // Use result
	}
}

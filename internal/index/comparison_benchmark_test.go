package index

import (
	"testing"
)

// BenchmarkComparison compares optimized vs non-optimized approaches
func BenchmarkComparison(b *testing.B) {
	storage := NewHybridStorage(".repocontext")
	if err := storage.Initialize(); err != nil {
		b.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	qe := NewQueryEngine(storage)

	// Test the pattern query that showed the most benefit from batch operations
	b.Run("Optimized_PatternQuery", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			result, err := qe.SearchByPatternWithOptions("*Query*", QueryOptions{IncludeTypes: true})
			if err != nil {
				b.Fatal(err)
			}
			_ = len(result.Entries)
		}
	})

	// Test cache effectiveness by running the same query multiple times
	b.Run("Optimized_CacheTest", func(b *testing.B) {
		// Prime the cache
		_, _ = qe.SearchByPattern("Process*")

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			result, err := qe.SearchByPattern("Process*")
			if err != nil {
				b.Fatal(err)
			}
			_ = len(result.Entries)
		}
	})

	// Test name lookups (should benefit from prepared statements)
	b.Run("Optimized_NameLookup", func(b *testing.B) {
		names := []string{"SearchByName", "QueryEngine", "Initialize", "NewQueryEngine", "SearchByPattern"}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			name := names[i%len(names)]
			result, err := qe.SearchByName(name)
			if err != nil {
				b.Fatal(err)
			}
			_ = len(result.Entries)
		}
	})
}

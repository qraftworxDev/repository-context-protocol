package index

import (
	"fmt"
	"os"
	"sync"
	"testing"
)

// BenchmarkQueryEngine_BasicPatternMatching benchmarks basic pattern matching operations
func BenchmarkQueryEngine_BasicPatternMatching(b *testing.B) {
	tempDir, storage := setupTestStorage(&testing.T{})
	defer os.RemoveAll(tempDir)

	// Setup comprehensive test data
	SetupComplexPatternTestData(&testing.T{}, storage)
	SetupPerformanceTestData(&testing.T{}, storage)

	engine := NewQueryEngine(storage)

	patterns := []struct {
		name    string
		pattern string
	}{
		{"exact_match", "HandleUserLogin"},
		{"prefix_wildcard", "Handle*"},
		{"suffix_wildcard", "*Data"},
		{"middle_wildcard", "Process*Data"},
		{"multiple_wildcards", "*User*"},
		{"character_class", "[HP]*"},
		{"brace_expansion", "{Handle,Process}*"},
		{"complex_glob", "*{User,Payment}[DV]*"},
	}

	for _, p := range patterns {
		b.Run(p.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results, err := engine.SearchByPattern(p.pattern)
				if err != nil {
					b.Fatalf("Search failed: %v", err)
				}
				_ = results // Prevent optimization
			}
		})
	}
}

// BenchmarkQueryEngine_RegexPatternMatching benchmarks regex pattern matching operations
func BenchmarkQueryEngine_RegexPatternMatching(b *testing.B) {
	tempDir, storage := setupTestStorage(&testing.T{})
	defer os.RemoveAll(tempDir)

	SetupComplexPatternTestData(&testing.T{}, storage)
	SetupRegexTestCases(&testing.T{}, storage)
	SetupPerformanceTestData(&testing.T{}, storage)

	engine := NewQueryEngine(storage)

	regexPatterns := []struct {
		name    string
		pattern string
	}{
		{"word_boundary", "/\\bHandle/"},
		{"complex_alternation", "/(Handle|Process|Validate).*(User|Payment)/"},
		{"lookahead", "/Handle(?=User)/"},
		{"character_class_range", "/[A-Z][a-z]+/"},
		{"quantifiers", "/^[A-Z][a-z]{2,8}[A-Z]/"},
		{"non_greedy", "/Process.*?Data/"},
		{"case_insensitive", "/(?i)json/"},
		{"unicode_categories", "/\\p{Lu}\\p{Ll}+/"},
	}

	for _, p := range regexPatterns {
		b.Run(p.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results, err := engine.SearchByPattern(p.pattern)
				if err != nil {
					b.Fatalf("Regex search failed: %v", err)
				}
				_ = results
			}
		})
	}
}

// BenchmarkQueryEngine_LargeDatasetPerformance benchmarks performance with large datasets
func BenchmarkQueryEngine_LargeDatasetPerformance(b *testing.B) {
	tempDir, storage := setupTestStorage(&testing.T{})
	defer os.RemoveAll(tempDir)

	// Create multiple large datasets
	for i := 0; i < 10; i++ {
		SetupPerformanceTestData(&testing.T{}, storage)
	}

	engine := NewQueryEngine(storage)

	datasetPatterns := []struct {
		name    string
		pattern string
	}{
		{"broad_wildcard", "*Request*"},
		{"narrow_exact", "HandleRequest42"},
		{"complex_brace", "{Handle,Process,Validate,Generate,Execute}*"},
		{"character_range", "*[0-9]*"},
		{"multiple_conditions", "*Data*{Model,Client,Handler}*"},
	}

	for _, p := range datasetPatterns {
		b.Run(p.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results, err := engine.SearchByPattern(p.pattern)
				if err != nil {
					b.Fatalf("Large dataset search failed: %v", err)
				}
				_ = results
			}
		})
	}
}

// BenchmarkQueryEngine_ConcurrentPatternMatching benchmarks concurrent pattern matching
func BenchmarkQueryEngine_ConcurrentPatternMatching(b *testing.B) {
	tempDir, storage := setupTestStorage(&testing.T{})
	defer os.RemoveAll(tempDir)

	SetupComplexPatternTestData(&testing.T{}, storage)
	SetupConcurrencyTestData(&testing.T{}, storage)

	engine := NewQueryEngine(storage)

	patterns := []string{
		"Handle*",
		"*Data",
		"Process*",
		"/.*User.*/",
		"{Handle,Process}*",
		"*[A-Z]*",
	}

	concurrencyLevels := []int{1, 2, 4, 8, 16}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("concurrency_%d", concurrency), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				wg.Add(concurrency)

				for j := 0; j < concurrency; j++ {
					go func(patternIndex int) {
						defer wg.Done()
						pattern := patterns[patternIndex%len(patterns)]
						results, err := engine.SearchByPattern(pattern)
						if err != nil {
							b.Errorf("Concurrent search failed: %v", err)
						}
						_ = results
					}(j)
				}
				wg.Wait()
			}
		})
	}
}

// BenchmarkQueryEngine_RegexCachePerformance benchmarks regex compilation caching
func BenchmarkQueryEngine_RegexCachePerformance(b *testing.B) {
	tempDir, storage := setupTestStorage(&testing.T{})
	defer os.RemoveAll(tempDir)

	SetupComplexPatternTestData(&testing.T{}, storage)

	engine := NewQueryEngine(storage)

	// Complex regex patterns that benefit from caching
	complexPatterns := []string{
		"/^[A-Z][a-z]+(?:[A-Z][a-z]+)*$/",
		"/(Handle|Process|Validate|Query).*(?:User|Payment|Data|Request)/",
		"/\\b(?:get|set|is|has)[A-Z][a-zA-Z]*\\b/",
		"/^(?:[a-z]+[A-Z]){2,}[a-z]*$/",
		"/(?i)(?:json|xml|http|api|db|sql).*(?:parser|client|handler|service)/",
	}

	// Test cold cache performance (first time compilation)
	b.Run("cold_cache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Create new engine to ensure cold cache
			freshEngine := NewQueryEngine(storage)
			pattern := complexPatterns[i%len(complexPatterns)]
			results, err := freshEngine.SearchByPattern(pattern)
			if err != nil {
				b.Fatalf("Cold cache search failed: %v", err)
			}
			_ = results
		}
	})

	// Test warm cache performance (reusing compiled regex)
	b.Run("warm_cache", func(b *testing.B) {
		// Pre-warm the cache
		for _, pattern := range complexPatterns {
			_, _ = engine.SearchByPattern(pattern)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pattern := complexPatterns[i%len(complexPatterns)]
			results, err := engine.SearchByPattern(pattern)
			if err != nil {
				b.Fatalf("Warm cache search failed: %v", err)
			}
			_ = results
		}
	})
}

// BenchmarkQueryEngine_PatternComplexity benchmarks different pattern complexity levels
func BenchmarkQueryEngine_PatternComplexity(b *testing.B) {
	tempDir, storage := setupTestStorage(&testing.T{})
	defer os.RemoveAll(tempDir)

	SetupComplexPatternTestData(&testing.T{}, storage)
	SetupPerformanceTestData(&testing.T{}, storage)

	engine := NewQueryEngine(storage)

	complexityLevels := []struct {
		name     string
		patterns []string
	}{
		{
			name: "simple",
			patterns: []string{
				"Handle*",
				"*Data",
				"Process*",
			},
		},
		{
			name: "medium",
			patterns: []string{
				"Handle*User*",
				"*{User,Payment}*",
				"[HP]*Data",
			},
		},
		{
			name: "complex",
			patterns: []string{
				"{Handle,Process}*{User,Payment,API}*",
				"*{User,Payment}[DV]*",
				"[A-H]*{Login,Logout,Request}*",
			},
		},
		{
			name: "very_complex",
			patterns: []string{
				"{Handle{User,API},Process*Data}*",
				"*[VCP]*{User,Payment}[DV]*",
				"{Query,Validate}????*[A-Z]*",
			},
		},
	}

	for _, level := range complexityLevels {
		b.Run(level.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pattern := level.patterns[i%len(level.patterns)]
				results, err := engine.SearchByPattern(pattern)
				if err != nil {
					b.Fatalf("Pattern search failed: %v", err)
				}
				_ = results
			}
		})
	}
}

// BenchmarkQueryEngine_MemoryEfficiency benchmarks memory usage during pattern matching
func BenchmarkQueryEngine_MemoryEfficiency(b *testing.B) {
	tempDir, storage := setupTestStorage(&testing.T{})
	defer os.RemoveAll(tempDir)

	// Create substantial dataset
	for i := 0; i < 20; i++ {
		SetupPerformanceTestData(&testing.T{}, storage)
	}

	engine := NewQueryEngine(storage)

	patterns := []string{
		"*Request*",     // High result count
		"HandleRequest", // Low result count
		"*[0-9]*",       // Medium result count
	}

	for _, pattern := range patterns {
		b.Run(fmt.Sprintf("pattern_%s", pattern), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results, err := engine.SearchByPattern(pattern)
				if err != nil {
					b.Fatalf("Memory efficiency test failed: %v", err)
				}
				// Process results to simulate real usage
				for _, entry := range results.Entries {
					_ = entry.IndexEntry.Name
				}
			}
		})
	}
}

// BenchmarkQueryEngine_ScalabilityTest benchmarks scalability with increasing data sizes
func BenchmarkQueryEngine_ScalabilityTest(b *testing.B) {
	dataSizes := []int{1, 10, 50, 100, 500}

	for _, size := range dataSizes {
		b.Run(fmt.Sprintf("data_size_%d", size), func(b *testing.B) {
			tempDir, storage := setupTestStorage(&testing.T{})
			defer os.RemoveAll(tempDir)

			// Setup data proportional to size
			for i := 0; i < size; i++ {
				SetupPerformanceTestData(&testing.T{}, storage)
			}

			engine := NewQueryEngine(storage)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Use a pattern that will scale with data size
				results, err := engine.SearchByPattern("*Request*")
				if err != nil {
					b.Fatalf("Scalability test failed: %v", err)
				}
				_ = results
			}
		})
	}
}

// BenchmarkQueryEngine_PatternTypeComparison compares performance between pattern types
func BenchmarkQueryEngine_PatternTypeComparison(b *testing.B) {
	tempDir, storage := setupTestStorage(&testing.T{})
	defer os.RemoveAll(tempDir)

	SetupComplexPatternTestData(&testing.T{}, storage)
	SetupRegexTestCases(&testing.T{}, storage)

	engine := NewQueryEngine(storage)

	patternTypes := []struct {
		name    string
		pattern string
		ptype   string
	}{
		{"exact_match", "HandleUserLogin", "exact"},
		{"glob_simple", "Handle*", "glob"},
		{"glob_complex", "{Handle,Process}*{User,Payment}*", "glob"},
		{"regex_simple", "/Handle.*/", "regex"},
		{"regex_complex", "/(Handle|Process).*(User|Payment)/", "regex"},
	}

	for _, pt := range patternTypes {
		b.Run(fmt.Sprintf("%s_%s", pt.ptype, pt.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results, err := engine.SearchByPattern(pt.pattern)
				if err != nil {
					b.Fatalf("Pattern type comparison failed: %v", err)
				}
				_ = results
			}
		})
	}
}

package index

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

// TestQueryEngine_ConcurrentRegexCaching tests thread safety of regex compilation and caching
func TestQueryEngine_ConcurrentRegexCaching(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	// Setup test data using the setup function
	SetupConcurrencyTestData(t, storage)

	engine := NewQueryEngine(storage)

	// Test concurrent access to the same regex pattern
	pattern := "/ConcurrentFunction_.*/"
	numGoroutines := 10
	numIterations := 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numIterations)
	results := make(chan int, numGoroutines*numIterations)

	// Launch multiple goroutines using the same pattern
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				searchResults, err := engine.SearchByPattern(pattern)
				if err != nil {
					errors <- err
					return
				}
				results <- len(searchResults.Entries)
			}
		}(i)
	}

	wg.Wait()
	close(errors)
	close(results)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent search error: %v", err)
	}

	// Verify all results are consistent
	var resultCounts []int
	for count := range results {
		resultCounts = append(resultCounts, count)
	}

	if len(resultCounts) > 0 {
		expectedCount := resultCounts[0]
		for i, count := range resultCounts {
			if count != expectedCount {
				t.Errorf("Inconsistent results: iteration %d got %d results, expected %d",
					i, count, expectedCount)
			}
		}
		t.Logf("All %d concurrent searches returned consistent results: %d matches",
			len(resultCounts), expectedCount)
	}
}

// TestQueryEngine_ConcurrentDifferentPatterns tests concurrent searches with different regex patterns
func TestQueryEngine_ConcurrentDifferentPatterns(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	SetupConcurrencyTestData(t, storage)

	engine := NewQueryEngine(storage)

	patterns := []string{
		"/ConcurrentFunction_0_.*/",
		"/ConcurrentFunction_1_.*/",
		"/ConcurrentFunction_2_.*/",
		"/ConcurrentFunction_3_.*/",
		"/ConcurrentFunction_4_.*/",
	}

	numGoroutines := len(patterns)
	numIterations := 10

	var wg sync.WaitGroup
	results := make(chan struct {
		pattern string
		count   int
		err     error
	}, numGoroutines*numIterations)

	// Launch goroutines with different patterns
	for i, pattern := range patterns {
		wg.Add(1)
		go func(_ int, p string) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				searchResults, err := engine.SearchByPattern(p)
				count := 0
				if err == nil {
					count = len(searchResults.Entries)
				}
				results <- struct {
					pattern string
					count   int
					err     error
				}{p, count, err}
			}
		}(i, pattern)
	}

	wg.Wait()
	close(results)

	// Collect and verify results
	patternResults := make(map[string][]int)
	for result := range results {
		if result.err != nil {
			t.Errorf("Error searching pattern %s: %v", result.pattern, result.err)
			continue
		}
		patternResults[result.pattern] = append(patternResults[result.pattern], result.count)
	}

	// Verify consistency within each pattern
	for pattern, counts := range patternResults {
		if len(counts) > 0 {
			expectedCount := counts[0]
			for i, count := range counts {
				if count != expectedCount {
					t.Errorf("Pattern %s: inconsistent results at iteration %d: got %d, expected %d",
						pattern, i, count, expectedCount)
				}
			}
			t.Logf("Pattern %s: all %d searches returned consistent %d results",
				pattern, len(counts), expectedCount)
		}
	}
}

// TestQueryEngine_ConcurrentRegexCacheGrowth tests cache behavior under concurrent access
func TestQueryEngine_ConcurrentRegexCacheGrowth(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	SetupConcurrencyTestData(t, storage)

	engine := NewQueryEngine(storage)

	// Generate many unique patterns to test cache growth
	patterns := make([]string, 50)
	for i := 0; i < 50; i++ {
		patterns[i] = fmt.Sprintf("/ConcurrentFunction_%d_.*/", i%5)
	}

	numGoroutines := 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*len(patterns))

	// Launch goroutines that use different patterns
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()
			for j, pattern := range patterns {
				_, err := engine.SearchByPattern(pattern)
				if err != nil {
					errors <- err
					return
				}
				// Add small delay to increase chance of race conditions
				if j%10 == 0 {
					time.Sleep(time.Microsecond)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent cache growth error: %v", err)
	}

	// Verify cache contains patterns (this access should be thread-safe)
	engine.regexMutex.RLock()
	cacheSize := len(engine.regexCache)
	engine.regexMutex.RUnlock()

	t.Logf("Regex cache contains %d compiled patterns after concurrent access", cacheSize)

	if cacheSize == 0 {
		t.Error("Expected regex cache to contain compiled patterns")
	}
}

// TestQueryEngine_ConcurrentMixedOperations tests mixed concurrent operations
func TestQueryEngine_ConcurrentMixedOperations(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	// Setup both complex and concurrency test data
	SetupComplexPatternTestData(t, storage)
	SetupConcurrencyTestData(t, storage)

	engine := NewQueryEngine(storage)

	operations := []func() (int, error){
		// Regex pattern searches
		func() (int, error) {
			results, err := engine.SearchByPattern("/Handle.*/")
			if err != nil {
				return 0, err
			}
			return len(results.Entries), nil
		},
		func() (int, error) {
			results, err := engine.SearchByPattern("/ConcurrentFunction.*/")
			if err != nil {
				return 0, err
			}
			return len(results.Entries), nil
		},
		// Glob pattern searches
		func() (int, error) {
			results, err := engine.SearchByPattern("Process*")
			if err != nil {
				return 0, err
			}
			return len(results.Entries), nil
		},
		// Exact searches
		func() (int, error) {
			results, err := engine.SearchByPattern("HandleUserLogin")
			if err != nil {
				return 0, err
			}
			return len(results.Entries), nil
		},
	}

	numGoroutines := 20
	numIterations := 5

	var wg sync.WaitGroup
	results := make(chan struct {
		opIndex int
		count   int
		err     error
	}, numGoroutines*numIterations)

	// Launch goroutines with mixed operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(gID int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				opIndex := (gID + j) % len(operations)
				operation := operations[opIndex]
				count, err := operation()
				results <- struct {
					opIndex int
					count   int
					err     error
				}{opIndex, count, err}
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Collect results by operation type
	operationResults := make(map[int][]int)
	for result := range results {
		if result.err != nil {
			t.Errorf("Error in operation %d: %v", result.opIndex, result.err)
			continue
		}
		operationResults[result.opIndex] = append(operationResults[result.opIndex], result.count)
	}

	// Verify consistency within each operation type
	for opIndex, counts := range operationResults {
		if len(counts) > 0 {
			expectedCount := counts[0]
			for i, count := range counts {
				if count != expectedCount {
					t.Errorf("Operation %d: inconsistent results at iteration %d: got %d, expected %d",
						opIndex, i, count, expectedCount)
				}
			}
			t.Logf("Operation %d: all %d executions returned consistent %d results",
				opIndex, len(counts), expectedCount)
		}
	}
}

// TestQueryEngine_ConcurrentRegexCompilationRaces tests for race conditions in regex compilation
func TestQueryEngine_ConcurrentRegexCompilationRaces(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	SetupConcurrencyTestData(t, storage)

	engine := NewQueryEngine(storage)

	// Use a complex pattern that takes time to compile
	complexPattern := "/^(?:[A-Za-z_][A-Za-z0-9_]*)?(?:ConcurrentFunction|HandleRequest|ProcessData)(?:_\\d+_\\d+|[A-Z][a-z]*)*(?:\\w*)?$/"

	numGoroutines := 15
	var wg sync.WaitGroup
	compilationResults := make(chan struct {
		goroutineID int
		duration    time.Duration
		count       int
		err         error
	}, numGoroutines)

	// Launch multiple goroutines that will trigger regex compilation simultaneously
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(gID int) {
			defer wg.Done()
			start := time.Now()
			results, err := engine.SearchByPattern(complexPattern)
			duration := time.Since(start)

			count := 0
			if err == nil {
				count = len(results.Entries)
			}

			compilationResults <- struct {
				goroutineID int
				duration    time.Duration
				count       int
				err         error
			}{gID, duration, count, err}
		}(i)
	}

	wg.Wait()
	close(compilationResults)

	// Analyze results
	var durations []time.Duration
	var counts []int
	for result := range compilationResults {
		if result.err != nil {
			t.Errorf("Goroutine %d failed: %v", result.goroutineID, result.err)
			continue
		}
		durations = append(durations, result.duration)
		counts = append(counts, result.count)
	}

	// Verify all counts are identical (no race condition corruption)
	if len(counts) > 0 {
		expectedCount := counts[0]
		for i, count := range counts {
			if count != expectedCount {
				t.Errorf("Race condition detected: goroutine %d got %d results, expected %d",
					i, count, expectedCount)
			}
		}
		t.Logf("All %d concurrent compilations returned consistent %d results",
			len(counts), expectedCount)
	}

	// Log timing information for analysis
	if len(durations) > 0 {
		var totalDuration time.Duration
		for _, d := range durations {
			totalDuration += d
		}
		avgDuration := totalDuration / time.Duration(len(durations))
		t.Logf("Average regex compilation and search duration: %v", avgDuration)
	}
}

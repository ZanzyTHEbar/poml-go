package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	poml "github.com/ZanzyTHEbar/poml/sdk"
	"github.com/ZanzyTHEbar/poml/sdk/writer"
)

type BenchmarkResult struct {
	FileName    string
	Iterations  int
	TotalTime   time.Duration
	AvgTime     time.Duration
	MinTime     time.Duration
	MaxTime     time.Duration
	SuccessRate float64
}

func runBenchmark(fileName string, iterations int) BenchmarkResult {
	data, err := os.ReadFile(fileName)
	if err != nil {
		panic(fmt.Sprintf("Failed to read %s: %v", fileName, err))
	}
	content := string(data)

	var totalTime time.Duration
	minTime := time.Hour
	var maxTime time.Duration
	successCount := 0

	fmt.Printf("Running benchmark for %s (%d iterations)...\n", fileName, iterations)

	for i := 0; i < iterations; i++ {
		start := time.Now()

		// Test regular parsing and rendering
		doc, err := poml.ParseWithParticiple(content)
		if err != nil {
			fmt.Printf("  Iteration %d: Parse failed: %v\n", i+1, err)
			continue
		}

		opts := &writer.RenderOptions{
			Context: make(map[string]interface{}),
		}

		_, err = writer.RenderAST(doc, opts)
		if err != nil {
			fmt.Printf("  Iteration %d: Render failed: %v\n", i+1, err)
			continue
		}

		elapsed := time.Since(start)
		totalTime += elapsed

		if elapsed < minTime {
			minTime = elapsed
		}
		if elapsed > maxTime {
			maxTime = elapsed
		}

		successCount++

		if (i+1)%10 == 0 {
			fmt.Printf("  Completed %d/%d iterations\n", i+1, iterations)
		}
	}

	avgTime := totalTime / time.Duration(iterations)
	successRate := float64(successCount) / float64(iterations) * 100

	return BenchmarkResult{
		FileName:    fileName,
		Iterations:  iterations,
		TotalTime:   totalTime,
		AvgTime:     avgTime,
		MinTime:     minTime,
		MaxTime:     maxTime,
		SuccessRate: successRate,
	}
}

func runSpeakerModeBenchmark(fileName string, iterations int) BenchmarkResult {
	data, err := os.ReadFile(fileName)
	if err != nil {
		panic(fmt.Sprintf("Failed to read %s: %v", fileName, err))
	}
	content := string(data)

	var totalTime time.Duration
	minTime := time.Hour
	var maxTime time.Duration
	successCount := 0

	fmt.Printf("Running speaker mode benchmark for %s (%d iterations)...\n", fileName, iterations)

	for i := 0; i < iterations; i++ {
		start := time.Now()

		// Test speaker mode parsing and rendering
		doc, err := poml.ParseWithParticiple(content)
		if err != nil {
			fmt.Printf("  Iteration %d: Parse failed: %v\n", i+1, err)
			continue
		}

		opts := &writer.RenderOptions{
			Context:     make(map[string]interface{}),
			SpeakerMode: true,
		}

		_, err = writer.RenderAST(doc, opts)
		if err != nil {
			fmt.Printf("  Iteration %d: Speaker render failed: %v\n", i+1, err)
			continue
		}

		elapsed := time.Since(start)
		totalTime += elapsed

		if elapsed < minTime {
			minTime = elapsed
		}
		if elapsed > maxTime {
			maxTime = elapsed
		}

		successCount++

		if (i+1)%10 == 0 {
			fmt.Printf("  Completed %d/%d iterations\n", i+1, iterations)
		}
	}

	avgTime := totalTime / time.Duration(iterations)
	successRate := float64(successCount) / float64(iterations) * 100

	return BenchmarkResult{
		FileName:    fileName + " (speaker mode)",
		Iterations:  iterations,
		TotalTime:   totalTime,
		AvgTime:     avgTime,
		MinTime:     minTime,
		MaxTime:     maxTime,
		SuccessRate: successRate,
	}
}

func printResults(results []BenchmarkResult) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("PERFORMANCE BENCHMARK RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("%-30s %-8s %-12s %-12s %-12s %-12s %-8s\n",
		"File", "Iters", "Total", "Avg", "Min", "Max", "Success")
	fmt.Println(strings.Repeat("-", 100))

	for _, result := range results {
		fmt.Printf("%-30s %-8d %-12v %-12v %-12v %-12v %6.1f%%\n",
			result.FileName,
			result.Iterations,
			result.TotalTime,
			result.AvgTime,
			result.MinTime,
			result.MaxTime,
			result.SuccessRate)
	}
}

func main() {
	fmt.Println("Go POML Performance Benchmarks")
	fmt.Println("Testing parsing and rendering performance across different complexity levels")

	iterations := 100 // Number of iterations per benchmark

	benchmarkFiles := []string{
		"benchmarks/simple.poml",
		"benchmarks/medium.poml",
		"benchmarks/complex.poml",
	}

	var results []BenchmarkResult

	// Run regular mode benchmarks
	for _, file := range benchmarkFiles {
		result := runBenchmark(file, iterations)
		results = append(results, result)
	}

	// Run speaker mode benchmarks for files that support it
	for _, file := range benchmarkFiles {
		if file == "benchmarks/simple.poml" {
			continue // Skip simple file for speaker mode
		}
		result := runSpeakerModeBenchmark(file, iterations)
		results = append(results, result)
	}

	printResults(results)

	fmt.Println("\nBenchmark completed!")
}

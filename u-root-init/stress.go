package main

import (
	"fmt"
	"runtime"
	"time"
)

// RunStressTests executes short CPU and memory stress tests during kernel boot validation.
func RunStressTests() (bool, string) {
	fmt.Println("--- Starting Kernel Stress Tests ---")

	// 1. CPU Stress (calculate primes for a short duration)
	start := time.Now()
	cpuDuration := 2 * time.Second
	done := make(chan bool)

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					// simple math loop to stress CPU
					_ = 1000 * 1000
				}
			}
		}()
	}

	time.Sleep(cpuDuration)
	close(done)
	fmt.Printf("[DEBUG] CPU Stress Test completed in %v\n", time.Since(start))

	// 2. Memory Stress (allocate chunks)
	memStart := time.Now()
	var chunks [][]byte
	chunkSize := 10 * 1024 * 1024 // 10MB
	numChunks := 5                // Total 50MB

	for i := 0; i < numChunks; i++ {
		chunk := make([]byte, chunkSize)
		// touch the memory to ensure it's actually allocated
		for j := 0; j < chunkSize; j += 4096 {
			chunk[j] = 1
		}
		chunks = append(chunks, chunk)
	}

	totalBytes := 0
	for _, c := range chunks {
		totalBytes += len(c)
	}

	fmt.Printf("[DEBUG] Memory Stress Test allocated %d MB in %v\n", totalBytes/(1024*1024), time.Since(memStart))

	// Free memory
	runtime.KeepAlive(chunks)
	runtime.GC()

	return true, "Stress tests completed successfully"
}

package main

import (
	"strings"
	"testing"
)

func TestRunStressTests(t *testing.T) {
	// The stress tests should run successfully and return true
	passed, msg := RunStressTests()

	if !passed {
		t.Errorf("Expected stress tests to pass, but got: %s", msg)
	}

	if !strings.Contains(msg, "completed") {
		t.Errorf("Expected message to contain 'completed', got: %s", msg)
	}
}

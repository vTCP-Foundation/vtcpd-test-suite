package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {

	testing.Init()

	// Setup code before running tests
	println("Setting up test environment...")

	// You can initialize resources, set up test databases,
	// create temporary files, etc.

	// Run all tests and store the exit code
	exitCode := m.Run()

	// Cleanup code after tests finish
	println("Cleaning up test environment...")

	// You can clean up resources, close connections,
	// remove temporary files, etc.

	// Exit with the exit code from the tests
	os.Exit(exitCode)
}

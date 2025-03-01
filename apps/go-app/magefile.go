//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Generate GraphQL schema
func Generate() error {
	fmt.Println("Generating from GraphQL schema files...")
	cmd := exec.Command("go", "run", "github.com/99designs/gqlgen", "generate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Generate a migration file
func MigrationGenerate() error {
	fmt.Println("Generate migration in ./migrations")
	timestamp := time.Now().Format("2006-01-02_150405")
	filename := fmt.Sprintf("migrations/%s_migration.sql", timestamp)
	_, err := os.Create(filename)
	return err
}

// Build the application
func Build() error {
	fmt.Println("Building...")
	cmd := exec.Command("go", "build", "-o", "tmp/main", "cmd/api/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run the application
func Run() error {
	cmd := exec.Command("go", "run", "cmd/api/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Test the application
func Test() error {
	fmt.Println("Testing...")
	cmd := exec.Command("go", "test", "./tests/...", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Clean build artifacts
func Clean() error {
	fmt.Println("Cleaning...")
	return os.Remove("main")
}

// Update dependencies
func Update() error {
	fmt.Println("Updating dependencies...")
	cmd := exec.Command("go", "get", "-u", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Watch for file changes and reload
func Watch() error {
	if _, err := exec.LookPath("air"); err == nil {
		fmt.Println("Watching...")
		cmd := exec.Command("air")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	var choice string
	fmt.Print("Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] ")
	fmt.Scanln(&choice)
	if choice != "n" && choice != "N" {
		cmd := exec.Command("go", "install", "github.com/cosmtrek/air@latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
		cmd = exec.Command("air")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	fmt.Println("You chose not to install air. Exiting...")
	return nil
}

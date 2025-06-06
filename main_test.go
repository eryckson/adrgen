package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"DATABASE_CHOICE", "database-choice"},
		{"microservice architecture", "microservice-architecture"},
		{"", ""},
		{"Already-Kebab-Case", "already-kebab-case"},
		{"Multiple   Spaces", "multiple---spaces"},
	}

	for _, test := range tests {
		result := toKebabCase(test.input)
		if result != test.expected {
			t.Errorf("toKebabCase(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestEnsureDir(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test", "nested", "dir")

	err := ensureDir(testPath)
	if err != nil {
		t.Errorf("ensureDir(%q) failed: %v", testPath, err)
	}

	// Check if directory exists
	info, err := os.Stat(testPath)
	if err != nil {
		t.Errorf("Directory %q was not created", testPath)
	}
	if !info.IsDir() {
		t.Errorf("%q is not a directory", testPath)
	}
}

func TestWriteFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "test content"

	err := writeFile(testFile, testContent)
	if err != nil {
		t.Errorf("writeFile(%q, %q) failed: %v", testFile, testContent, err)
	}

	// Read and verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Failed to read test file: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("File content = %q, want %q", string(content), testContent)
	}
}

func TestExtractTitleFromFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"001-database-choice.md", "Database Choice"},
		{"002-adr-template.md", "ADR Template"},
		{"simple.md", "simple.md"},
		{"003-multiple-word-title.md", "Multiple Word Title"},
	}

	for _, test := range tests {
		result := extractTitleFromFilename(test.input)
		if result != test.expected {
			t.Errorf("extractTitleFromFilename(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestUpdateIndex(t *testing.T) {
	// Create temporary ADR directory
	tempDir := t.TempDir()
	originalAdrDir := adrDir
	adrDir = tempDir
	defer func() { adrDir = originalAdrDir }()

	// Create test ADR files
	testFiles := []string{
		"001-first-decision.md",
		"002-second-decision.md",
		"template.md", // Should be ignored
	}

	for _, file := range testFiles {
		err := writeFile(filepath.Join(tempDir, file), "test content")
		if err != nil {
			t.Fatalf("Failed to create test file %q: %v", file, err)
		}
	}

	err := updateIndex()
	if err != nil {
		t.Errorf("updateIndex() failed: %v", err)
	}

	// Verify index content
	content, err := os.ReadFile(filepath.Join(tempDir, indexFile))
	if err != nil {
		t.Fatalf("Failed to read index file: %v", err)
	}

	expectedContent := "# 📄 Architecture Decision Records\n\n" +
		"- [First Decision](001-first-decision.md)\n" +
		"- [Second Decision](002-second-decision.md)\n"

	if string(content) != expectedContent {
		t.Errorf("Index content = %q, want %q", string(content), expectedContent)
	}
}

func TestLoadTemplateOrDefault(t *testing.T) {
	// Test with non-existent template
	tempDir := t.TempDir()
	originalAdrDir := adrDir
	adrDir = tempDir
	defer func() { adrDir = originalAdrDir }()

	result := loadTemplateOrDefault()
	if !strings.Contains(result, "# ADR {{number}}: {{title}}") {
		t.Error("Default template not returned when template file doesn't exist")
	}

	// Test with existing template
	customTemplate := "Custom template {{number}} {{title}} {{status}} {{date}}"
	err := writeFile(filepath.Join(tempDir, templateFile), customTemplate)
	if err != nil {
		t.Fatalf("Failed to create test template file: %v", err)
	}

	result = loadTemplateOrDefault()
	if result != customTemplate {
		t.Errorf("loadTemplateOrDefault() = %q, want %q", result, customTemplate)
	}
}

func TestRenderTemplate(t *testing.T) {
	template := "ADR {{number}}: {{title}} ({{status}}) - {{date}}"
	number := "001"
	status := "Accepted"
	title := "Test Decision"
	date := "2024-03-20"

	expected := "ADR 001: Test Decision (Accepted) - 2024-03-20"
	result := renderTemplate(template, number, status, title, date)

	if result != expected {
		t.Errorf("renderTemplate() = %q, want %q", result, expected)
	}
}

func TestAdrExists(t *testing.T) {
	// Create temporary ADR directory
	tempDir := t.TempDir()
	originalAdrDir := adrDir
	adrDir = tempDir
	defer func() { adrDir = originalAdrDir }()

	// Test non-existent ADR
	if adrExists("001") {
		t.Error("adrExists() returned true for non-existent ADR")
	}

	// Create test ADR file
	err := writeFile(filepath.Join(tempDir, "adr-001-test.md"), "test content")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test existing ADR
	if !adrExists("001") {
		t.Error("adrExists() returned false for existing ADR")
	}
}

func TestUpdateIndexError(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	originalAdrDir := adrDir
	adrDir = tempDir
	defer func() { adrDir = originalAdrDir }()

	// Make the directory read-only to cause permission error
	err := os.Chmod(tempDir, 0444)
	if err != nil {
		t.Fatalf("Failed to change directory permissions: %v", err)
	}

	// Try to update index in read-only directory
	err = updateIndex()
	if err == nil {
		t.Error("Expected error when writing to read-only directory")
	}
}

func TestWriteFileError(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Create file and make it read-only
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	err = os.Chmod(testFile, 0444)
	if err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}

	// Try to write to read-only file
	err = writeFile(testFile, "new content")
	if err == nil {
		t.Error("Expected error when writing to read-only file")
	}
}

func TestMainWithDirectoryError(t *testing.T) {
	// Save original args and restore them after the test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Create temporary directory
	tempDir := t.TempDir()
	originalAdrDir := adrDir
	adrDir = filepath.Join(tempDir, "nonexistent")
	defer func() { adrDir = originalAdrDir }()

	// Make parent directory read-only to prevent creation of new directory
	err := os.Chmod(tempDir, 0444)
	if err != nil {
		t.Fatalf("Failed to change directory permissions: %v", err)
	}

	// Set up command line arguments
	os.Args = []string{"cmd", "--number", "001", "--status", "Accepted", "--title", "Test Decision"}

	// Save original stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Run main
	main()

	// Restore stdout and get output
	w.Close()
	output := make([]byte, 1024)
	n, _ := r.Read(output)

	// Check if error message was printed
	if !strings.Contains(string(output[:n]), "Error creating directory") {
		t.Error("Expected directory creation error message")
	}
}

// func TestMainWithReadDirError(t *testing.T) {
// 	// Save original args and restore them after the test
// 	oldArgs := os.Args
// 	defer func() { os.Args = oldArgs }()

// 	// Reset flags to avoid redefinition
// 	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

// 	// Create temporary directory
// 	tempDir := t.TempDir()
// 	originalAdrDir := adrDir
// 	adrDir = tempDir
// 	defer func() {
// 		// Restore permissions before cleanup
// 		_ = os.Chmod(tempDir, 0755)
// 		adrDir = originalAdrDir
// 	}()

// 	// Create test ADR file with proper name format
// 	adrFile := filepath.Join(tempDir, "adr-001-test-decision.md")
// 	err := writeFile(adrFile, "# ADR 001: Test Decision\n\n**Status**: Proposed\n\nTest content")
// 	if err != nil {
// 		t.Fatalf("Failed to create test file: %v", err)
// 	}

// 	// Set up command line arguments for updating existing ADR
// 	os.Args = []string{"cmd", "--number", "001", "--status", "Superseded"}

// 	// Save original stdout
// 	oldStdout := os.Stdout
// 	r, w, _ := os.Pipe()
// 	os.Stdout = w
// 	defer func() { os.Stdout = oldStdout }()

// 	// Make directory unreadable but executable (so we can still access files by name)
// 	err = os.Chmod(tempDir, 0111)
// 	if err != nil {
// 		t.Fatalf("Failed to change directory permissions: %v", err)
// 	}

// 	// Run main
// 	main()

// 	// Restore stdout and get output
// 	w.Close()
// 	output := make([]byte, 1024)
// 	n, _ := r.Read(output)

// 	// Check if error message was printed
// 	if !strings.Contains(string(output[:n]), "Error reading directory:") {
// 		t.Errorf("Expected 'Error reading directory:' message, got: %s", string(output[:n]))
// 	}
// }

func TestMainFunction(t *testing.T) {
	// Save original args and restore them after the test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Save original stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create temporary directory for test
	tempDir := t.TempDir()
	originalAdrDir := adrDir
	adrDir = tempDir
	defer func() { adrDir = originalAdrDir }()

	// Create a null writer to discard flag usage output
	nullWriter := os.NewFile(0, os.DevNull)

	testCases := []struct {
		name     string
		args     []string
		wantErr  bool
		checkDir bool
		setup    func() error
	}{
		{
			name:     "Missing required flags",
			args:     []string{"cmd"},
			wantErr:  true,
			checkDir: false,
		},
		{
			name:     "Missing status flag",
			args:     []string{"cmd", "--number", "001"},
			wantErr:  true,
			checkDir: false,
		},
		{
			name:     "New ADR without title",
			args:     []string{"cmd", "--number", "001", "--status", "Accepted"},
			wantErr:  true,
			checkDir: false,
		},
		{
			name:     "Valid new ADR",
			args:     []string{"cmd", "--number", "001", "--status", "Accepted", "--title", "Test Decision"},
			wantErr:  false,
			checkDir: true,
		},
		{
			name: "Update existing ADR",
			args: []string{"cmd", "--number", "002", "--status", "Superseded"},
			setup: func() error {
				return writeFile(filepath.Join(tempDir, "adr-002-existing.md"),
					"# ADR 002: Existing\n\n**Status**: Accepted\n\nTest content")
			},
			wantErr:  false,
			checkDir: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test environment if needed
			if tc.setup != nil {
				err := tc.setup()
				if err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Set command line arguments
			os.Args = tc.args

			// Reset flags
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			flag.CommandLine.SetOutput(nullWriter)

			// Redirect stdout to capture output or discard it
			var r, w *os.File
			var output []byte
			if tc.wantErr {
				// If we expect an error, capture the output to check it
				r, w, _ = os.Pipe()
				os.Stdout = w
			} else {
				// If we don't expect an error, discard the output
				os.Stdout = nullWriter
			}

			// Run main
			main()

			// Restore stdout and get output if needed
			if tc.wantErr {
				w.Close()
				os.Stdout = oldStdout
				output = make([]byte, 1024)
				n, _ := r.Read(output)
				hasError := strings.Contains(string(output[:n]), "Error") ||
					strings.Contains(string(output[:n]), "Required flags")

				if !hasError {
					t.Errorf("Expected error output but got none\nOutput: %s", string(output[:n]))
				}
			}

			// Check if directory was created when expected
			if tc.checkDir {
				if _, err := os.Stat(tempDir); os.IsNotExist(err) {
					t.Error("ADR directory was not created")
				}
			}
		})
	}
}

func TestMainWithUpdateError(t *testing.T) {
	// Save original args and restore them after the test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Reset flags to avoid redefinition
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Create temporary directory
	tempDir := t.TempDir()
	originalAdrDir := adrDir
	adrDir = tempDir
	defer func() { adrDir = originalAdrDir }()

	// Create test ADR file and make it read-only
	adrFile := filepath.Join(tempDir, "adr-001-test.md")
	err := writeFile(adrFile, "test content")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	err = os.Chmod(adrFile, 0444)
	if err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}
	defer func() {
		// Restore permissions before cleanup
		_ = os.Chmod(adrFile, 0644)
	}()

	// Set up command line arguments for updating existing ADR
	os.Args = []string{"cmd", "--number", "001", "--status", "Superseded"}

	// Save original stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Run main
	main()

	// Restore stdout and get output
	w.Close()
	output := make([]byte, 1024)
	n, _ := r.Read(output)

	// Check if error message was printed
	if !strings.Contains(string(output[:n]), "Error writing ADR") {
		t.Error("Expected error when writing to read-only ADR file")
	}
}

func TestMainWithIndexUpdateError(t *testing.T) {
	// Save original args and restore them after the test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Reset flags to avoid redefinition
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Create temporary directory
	tempDir := t.TempDir()
	originalAdrDir := adrDir
	adrDir = tempDir
	defer func() { adrDir = originalAdrDir }()

	// Create test ADR file
	err := writeFile(filepath.Join(tempDir, "adr-001-test.md"), "test content")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create index file and make it read-only
	indexPath := filepath.Join(tempDir, indexFile)
	err = writeFile(indexPath, "# Test Index")
	if err != nil {
		t.Fatalf("Failed to create index file: %v", err)
	}
	err = os.Chmod(indexPath, 0444)
	if err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}
	defer func() {
		// Restore permissions before cleanup
		_ = os.Chmod(indexPath, 0644)
	}()

	// Set up command line arguments for new ADR
	os.Args = []string{"cmd", "--number", "002", "--status", "Accepted", "--title", "Test Decision"}

	// Save original stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Run main
	main()

	// Restore stdout and get output
	w.Close()
	output := make([]byte, 1024)
	n, _ := r.Read(output)

	// Check if error message was printed
	if !strings.Contains(string(output[:n]), "Error updating index") {
		t.Error("Expected error when updating read-only index file")
	}
}

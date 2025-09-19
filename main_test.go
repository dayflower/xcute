package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// MockCommandExecutor is a mock implementation of CommandExecutor for testing
type MockCommandExecutor struct {
	shellCommands  []string
	directCommands [][]string
	exitCode       int
}

func NewMockCommandExecutor(exitCode int) *MockCommandExecutor {
	return &MockCommandExecutor{
		exitCode: exitCode,
	}
}

func (m *MockCommandExecutor) ExecuteShell(command string) int {
	m.shellCommands = append(m.shellCommands, command)
	return m.exitCode
}

func (m *MockCommandExecutor) ExecuteDirect(args []string) int {
	// Create a copy of args to avoid slice mutation issues
	argsCopy := make([]string, len(args))
	copy(argsCopy, args)
	m.directCommands = append(m.directCommands, argsCopy)
	return m.exitCode
}

func TestReplacePlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		template string
		input    string
		expected string
	}{
		{
			name:     "single placeholder",
			template: "echo {}",
			input:    "hello",
			expected: "echo hello",
		},
		{
			name:     "multiple placeholders",
			template: "cp {} backup/{}",
			input:    "file.txt",
			expected: "cp file.txt backup/file.txt",
		},
		{
			name:     "no placeholder",
			template: "ls -la",
			input:    "ignored",
			expected: "ls -la",
		},
		{
			name:     "placeholder with spaces",
			template: "echo {}",
			input:    "hello world",
			expected: "echo hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replacePlaceholders(tt.template, tt.input)
			if result != tt.expected {
				t.Errorf("replacePlaceholders(%q, %q) = %q, want %q", tt.template, tt.input, result, tt.expected)
			}
		})
	}
}

func TestProcessStdin_DirectMode_DryRun(t *testing.T) {
	stdin := strings.NewReader("file1.txt\nfile2.txt\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	executor := NewMockCommandExecutor(0)

	app := &App{
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		executor: executor,
		useColor: false,
	}

	opts := Options{
		dryRun:    true,
		shellMode: false,
	}
	args := []string{"echo", "{}"}

	err := app.processStdin(opts, args)
	if err != nil {
		t.Fatalf("processStdin failed: %v", err)
	}

	expected := "echo file1.txt\necho file2.txt\n"
	if stdout.String() != expected {
		t.Errorf("stdout = %q, want %q", stdout.String(), expected)
	}

	// No commands should be executed in dry run mode
	if len(executor.directCommands) != 0 {
		t.Errorf("Expected no commands to be executed in dry run mode, got %d", len(executor.directCommands))
	}
}

func TestProcessStdin_DirectMode_Execution(t *testing.T) {
	stdin := strings.NewReader("file1.txt\nfile2.txt\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	executor := NewMockCommandExecutor(0)

	app := &App{
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		executor: executor,
		useColor: false,
	}

	opts := Options{
		dryRun:    false,
		shellMode: false,
	}
	args := []string{"echo", "{}"}

	err := app.processStdin(opts, args)
	if err != nil {
		t.Fatalf("processStdin failed: %v", err)
	}

	// Check that commands were executed
	expectedCommands := [][]string{
		{"echo", "file1.txt"},
		{"echo", "file2.txt"},
	}

	if len(executor.directCommands) != len(expectedCommands) {
		t.Fatalf("Expected %d commands, got %d", len(expectedCommands), len(executor.directCommands))
	}

	for i, expected := range expectedCommands {
		actual := executor.directCommands[i]
		if len(actual) != len(expected) {
			t.Errorf("Command %d: expected %d args, got %d", i, len(expected), len(actual))
			continue
		}
		for j, expectedArg := range expected {
			if actual[j] != expectedArg {
				t.Errorf("Command %d, arg %d: expected %q, got %q", i, j, expectedArg, actual[j])
			}
		}
	}
}

func TestProcessStdin_ShellMode_DryRun(t *testing.T) {
	stdin := strings.NewReader("file1.txt\nfile2.txt\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	executor := NewMockCommandExecutor(0)

	app := &App{
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		executor: executor,
		useColor: false,
	}

	opts := Options{
		dryRun:    true,
		shellMode: true,
	}
	args := []string{"echo hello {} && echo processed {}"}

	err := app.processStdin(opts, args)
	if err != nil {
		t.Fatalf("processStdin failed: %v", err)
	}

	expected := "echo hello file1.txt && echo processed file1.txt\necho hello file2.txt && echo processed file2.txt\n"
	if stdout.String() != expected {
		t.Errorf("stdout = %q, want %q", stdout.String(), expected)
	}

	// No commands should be executed in dry run mode
	if len(executor.shellCommands) != 0 {
		t.Errorf("Expected no commands to be executed in dry run mode, got %d", len(executor.shellCommands))
	}
}

func TestProcessStdin_ShellMode_Execution(t *testing.T) {
	stdin := strings.NewReader("hello\nworld\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	executor := NewMockCommandExecutor(0)

	app := &App{
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		executor: executor,
		useColor: false,
	}

	opts := Options{
		dryRun:    false,
		shellMode: true,
	}
	args := []string{"echo hello {}"}

	err := app.processStdin(opts, args)
	if err != nil {
		t.Fatalf("processStdin failed: %v", err)
	}

	// Check that shell commands were executed
	expectedCommands := []string{
		"echo hello hello",
		"echo hello world",
	}

	if len(executor.shellCommands) != len(expectedCommands) {
		t.Fatalf("Expected %d shell commands, got %d", len(expectedCommands), len(executor.shellCommands))
	}

	for i, expected := range expectedCommands {
		if executor.shellCommands[i] != expected {
			t.Errorf("Shell command %d: expected %q, got %q", i, expected, executor.shellCommands[i])
		}
	}
}

func TestProcessStdin_EmptyLines(t *testing.T) {
	stdin := strings.NewReader("line1\n\nline2\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	executor := NewMockCommandExecutor(0)

	app := &App{
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		executor: executor,
		useColor: false,
	}

	opts := Options{
		dryRun:      false,
		shellMode:   false,
		showWhat:    true, // Enable to see empty line notification
		showCommand: false,
	}
	args := []string{"echo", "{}"}

	err := app.processStdin(opts, args)
	if err != nil {
		t.Fatalf("processStdin failed: %v", err)
	}

	// Should only execute 2 commands (skip empty line)
	expectedCommands := [][]string{
		{"echo", "line1"},
		{"echo", "line2"},
	}

	if len(executor.directCommands) != len(expectedCommands) {
		t.Errorf("Expected %d commands, got %d", len(expectedCommands), len(executor.directCommands))
	}

	// Check stderr contains empty line notification
	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "[empty line]") {
		t.Errorf("Expected empty line notification in stderr, got: %q", stderrStr)
	}
}

func TestProcessStdin_ErrorHandling(t *testing.T) {
	stdin := strings.NewReader("file1\nfile2\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	executor := NewMockCommandExecutor(1) // Return error exit code

	app := &App{
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		executor: executor,
		useColor: false,
	}

	opts := Options{
		dryRun:        false,
		shellMode:     false,
		forceContinue: false, // Stop on first error
	}
	args := []string{"echo", "{}"}

	err := app.processStdin(opts, args)
	if err == nil {
		t.Fatal("Expected error due to command failure, got nil")
	}

	// Should only execute first command before stopping
	if len(executor.directCommands) != 1 {
		t.Errorf("Expected 1 command to be executed before stopping, got %d", len(executor.directCommands))
	}
}

func TestProcessStdin_ForceContinueOnError(t *testing.T) {
	stdin := strings.NewReader("file1\nfile2\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	executor := NewMockCommandExecutor(1) // Return error exit code

	app := &App{
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		executor: executor,
		useColor: false,
	}

	opts := Options{
		dryRun:        false,
		shellMode:     false,
		forceContinue: true, // Continue on errors
	}
	args := []string{"echo", "{}"}

	err := app.processStdin(opts, args)
	// In force continue mode, we expect an error indicating there were command failures
	// but processing continued
	if err == nil {
		t.Fatal("Expected error indicating command failures occurred")
	}
	if !strings.Contains(err.Error(), "commands completed with errors") {
		t.Fatalf("Expected error about command failures, got: %v", err)
	}

	// Should execute both commands despite errors
	if len(executor.directCommands) != 2 {
		t.Errorf("Expected 2 commands to be executed with force continue, got %d", len(executor.directCommands))
	}
}

func TestShouldUseColor(t *testing.T) {
	tests := []struct {
		name        string
		colorOption string
		noColor     string
		expected    bool
	}{
		{
			name:        "NO_COLOR set overrides everything",
			colorOption: "always",
			noColor:     "1",
			expected:    false,
		},
		{
			name:        "NO_COLOR empty string doesn't disable",
			colorOption: "always",
			noColor:     "",
			expected:    true,
		},
		{
			name:        "color never",
			colorOption: "never",
			noColor:     "",
			expected:    false,
		},
		{
			name:        "color always",
			colorOption: "always",
			noColor:     "",
			expected:    true,
		},
		{
			name:        "color auto with buffer (non-terminal)",
			colorOption: "auto",
			noColor:     "",
			expected:    false,
		},
		{
			name:        "invalid color option",
			colorOption: "invalid",
			noColor:     "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set NO_COLOR environment variable
			originalNoColor := os.Getenv("NO_COLOR")
			if tt.noColor != "" {
				os.Setenv("NO_COLOR", tt.noColor)
			} else {
				os.Unsetenv("NO_COLOR")
			}
			defer func() {
				if originalNoColor != "" {
					os.Setenv("NO_COLOR", originalNoColor)
				} else {
					os.Unsetenv("NO_COLOR")
				}
			}()

			// Use buffer as stderr (non-terminal) for consistent testing
			stderr := &bytes.Buffer{}

			result := shouldUseColor(tt.colorOption, stderr)
			if result != tt.expected {
				t.Errorf("shouldUseColor(%q) = %v, want %v", tt.colorOption, result, tt.expected)
			}
		})
	}
}

func TestColorOutput(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	executor := NewMockCommandExecutor(0)

	// Test with color enabled
	appWithColor := &App{
		stdin:    strings.NewReader("test\n"),
		stdout:   stdout,
		stderr:   stderr,
		executor: executor,
		useColor: true,
	}

	// Test with color disabled
	appWithoutColor := &App{
		stdin:    strings.NewReader("test\n"),
		stdout:   stdout,
		stderr:   stderr,
		executor: executor,
		useColor: false,
	}

	// Test error output
	appWithColor.printError("error message")
	appWithoutColor.printError("error message")

	// The exact output will depend on fatih/color implementation
	// We just verify that the methods don't crash and produce some output
	if stderr.Len() == 0 {
		t.Error("Expected some output to stderr")
	}
}
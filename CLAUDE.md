# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go project for implementing `xcute`, a CLI tool that executes commands with placeholders replaced by input from stdin. The project currently contains only the specification (SPEC.ja.md) and needs to be implemented from scratch.

## Architecture

Based on SPEC.ja.md, the tool should implement:

- **Input Processing**: Read lines from stdin and replace `{}` placeholders in command templates
- **Execution Modes**: 
  - Direct execution (default): Execute commands directly
  - Shell execution (`-c` flag): Execute commands through shell
- **Options Support**: `-n` (dry run), `-i` (interactive), `-w` (show what), `-l` (show command line), `-f` (force continue), `-t` (interval)
- **Error Handling**: Stop on first error by default, continue with `-f` flag
- **Output**: Use ANSI colors for enhanced visibility with `-w` and `-l` options

## Development Commands

Since this is a new Go project, typical commands would be:

```bash
# Initialize Go module (if not done)
go mod init github.com/dayflower/xcute

# Build the project
go build -o xcute ./cmd/xcute

# Run tests
go test ./...

# Format code
go fmt ./...

# Install dependencies
go mod tidy
```

## Implementation Structure

Suggested file structure for implementation:
- `main.go` or `cmd/xcute/main.go`: Entry point and CLI argument parsing
- `internal/executor/`: Core execution logic for different modes
- `internal/options/`: Option handling and validation
- `internal/formatter/`: ANSI color formatting for output
# xcute

A CLI tool that executes commands with placeholders replaced by input from stdin.

## Overview

`xcute` reads data from stdin line by line and executes specified command templates with `{}` placeholders replaced by each line's content.

## Basic Usage

```bash
# Echo each line from files.txt
cat files.txt | xcute echo {}

# Use multiple placeholders
echo -e "apple\nbanana" | xcute echo "Processing: {} -> backup/{}"
```

## Execution Modes

### Direct Execution Mode (Default)

Commands are executed directly. Arguments are preserved as-is, making it safe for filenames with spaces.

```bash
cat input.txt | xcute echo {}
cat input.txt | xcute echo hello {}
cat input.txt | xcute cp {} backup/
```

### Shell Execution Mode (`-c` flag)

Commands are executed through the shell, allowing complex shell operations.

```bash
cat words.txt | xcute -c 'echo hello {} && echo thanks {}'
cat files.txt | xcute -c 'wc -l {} | head -1'
```

**Note**: In shell mode, only one argument is allowed after `-c`. Multiple arguments will result in an error.

## Options

- `-n`: **Dry run** - Show commands that would be executed without running them
- `-i`: **Interactive** - Prompt for confirmation before each command execution
- `-w`: **Show target** - Display input lines with color highlighting
- `-l`: **Show command line** - Display commands before execution and exit codes after
- `-f`: **Force continue** - Continue execution even if errors occur
- `-t <seconds>`: **Interval** - Wait specified seconds between command executions
- `-c`: **Shell mode** - Execute commands through shell
- `--color <when>`: **Color control** - Control colored output (never/always/auto)

## Examples

```bash
# Basic usage
echo -e "file1.txt\nfile2.txt" | xcute cat {}

# Multiple placeholders
echo -e "apple\nbanana" | xcute echo "File: {} -> backup/{}"

# Shell command execution
find . -name "*.txt" | xcute -c 'wc -l {} && echo "processed {}"'

# Dry run
cat filelist.txt | xcute -n rm {}

# Interactive execution
cat filelist.txt | xcute -i rm {}

# Detailed logging
cat filelist.txt | xcute -l -w cp {} backup/

# Continue on errors
cat filelist.txt | xcute -f rm {}

# With interval
cat urls.txt | xcute -t 0.5 curl -s {}

# Color control examples
cat files.txt | xcute -l -w --color always cp {} backup/
NO_COLOR=1 cat files.txt | xcute -l -w cp {} backup/
```

## Error Handling

- **Default behavior**: Stop on first error and exit with that error's status code
- **With `-f` flag**: Continue execution on errors and exit with the last error's status code (or 0 if no errors)

## Empty Line Handling

- Empty input lines are skipped (no command execution)
- When `-w` or `-l` options are used, empty lines are reported to stderr

## Color Output Control

Color output can be controlled through the `--color` option and environment variables:

### Color Options

- `--color never`: Disable colored output
- `--color always`: Enable colored output regardless of output destination
- `--color auto`: Enable colored output only when stderr is a terminal (default)

### Environment Variables

- `NO_COLOR`: When set to any non-empty value, disables colored output (takes precedence over `--color`)

### Color Usage

When enabled, colors are used for:
- **Red**: Error messages and non-zero exit codes
- **Green**: Success messages and zero exit codes  
- **Yellow**: Warning messages and empty line notifications
- **Blue**: Command line display (with `-l` option)
- **Cyan**: Input line display (with `-w` option)

### Examples

```bash
# Force color output even when redirecting
cat files.txt | xcute -l -w --color always cp {} backup/ 2>&1 | less

# Disable color output
cat files.txt | xcute -l -w --color never cp {} backup/

# Use NO_COLOR environment variable
NO_COLOR=1 cat files.txt | xcute -l -w cp {} backup/
```

## Installation

### Homebrew (macOS/Linux)

```bash
brew install dayflower/tap/xcute
```

### GitHub Releases

Download pre-built binaries from the [releases page](https://github.com/dayflower/xcute/releases) for your platform:

- Linux (x86_64, ARM64)
- Windows (x86_64, ARM64) 
- macOS (x86_64, ARM64)

### Go Install

```bash
go install github.com/dayflower/xcute@latest
```

### Build from Source

```bash
git clone https://github.com/dayflower/xcute.git
cd xcute
make build
# or
go build -o xcute
```

## License

MIT License - see [LICENSE](LICENSE) file for details.
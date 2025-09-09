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
```

## Error Handling

- **Default behavior**: Stop on first error and exit with that error's status code
- **With `-f` flag**: Continue execution on errors and exit with the last error's status code (or 0 if no errors)

## Empty Line Handling

- Empty input lines are skipped (no command execution)
- When `-w` or `-l` options are used, empty lines are reported to stderr

## Installation

```bash
go install github.com/dayflower/xcute@latest
```

Or build from source:

```bash
git clone https://github.com/dayflower/xcute.git
cd xcute
go build -o xcute
```

## License

MIT License - see [LICENSE](LICENSE) file for details.
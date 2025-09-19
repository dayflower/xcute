package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

type Options struct {
	dryRun      bool
	interactive bool
	showWhat    bool
	showCommand bool
	forceContinue bool
	interval    float64
	shellMode   bool
	color       string
}

// CommandExecutor interface for command execution abstraction
type CommandExecutor interface {
	ExecuteShell(command string) int
	ExecuteDirect(args []string) int
}

// RealCommandExecutor implements CommandExecutor using actual system commands
type RealCommandExecutor struct {
	stdout io.Writer
	stderr io.Writer
}

func NewRealCommandExecutor(stdout, stderr io.Writer) *RealCommandExecutor {
	return &RealCommandExecutor{
		stdout: stdout,
		stderr: stderr,
	}
}

func (e *RealCommandExecutor) ExecuteShell(command string) int {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = e.stdout
	cmd.Stderr = e.stderr
	
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		}
		return 1
	}
	return 0
}

func (e *RealCommandExecutor) ExecuteDirect(args []string) int {
	if len(args) == 0 {
		return 0
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = e.stdout
	cmd.Stderr = e.stderr
	
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		}
		return 1
	}
	return 0
}

// App holds the application state and dependencies
type App struct {
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
	executor CommandExecutor
	useColor bool
}

// shouldUseColor determines whether to use color output based on options and environment
func shouldUseColor(colorOption string, stderr io.Writer) bool {
	// NO_COLOR environment variable takes precedence
	if noColor := os.Getenv("NO_COLOR"); noColor != "" {
		return false
	}

	switch colorOption {
	case "never":
		return false
	case "always":
		return true
	case "auto":
		// Check if stderr is a terminal
		if f, ok := stderr.(*os.File); ok {
			return isatty.IsTerminal(f.Fd())
		}
		return false
	default:
		return false
	}
}

// Color functions for output
var (
	colorError     = color.New(color.FgRed)
	colorSuccess   = color.New(color.FgGreen)
	colorWarning   = color.New(color.FgYellow)
	colorCommand   = color.New(color.FgBlue)
	colorTarget    = color.New(color.FgCyan)
)

// Colored output methods for App
func (app *App) printError(format string, args ...interface{}) {
	if app.useColor {
		colorError.Fprintf(app.stderr, format, args...)
	} else {
		fmt.Fprintf(app.stderr, format, args...)
	}
}

func (app *App) printSuccess(format string, args ...interface{}) {
	if app.useColor {
		colorSuccess.Fprintf(app.stderr, format, args...)
	} else {
		fmt.Fprintf(app.stderr, format, args...)
	}
}

func (app *App) printWarning(format string, args ...interface{}) {
	if app.useColor {
		colorWarning.Fprintf(app.stderr, format, args...)
	} else {
		fmt.Fprintf(app.stderr, format, args...)
	}
}

func (app *App) printCommand(format string, args ...interface{}) {
	if app.useColor {
		colorCommand.Fprintf(app.stderr, format, args...)
	} else {
		fmt.Fprintf(app.stderr, format, args...)
	}
}

func (app *App) printTarget(format string, args ...interface{}) {
	if app.useColor {
		colorTarget.Fprintf(app.stderr, format, args...)
	} else {
		fmt.Fprintf(app.stderr, format, args...)
	}
}

func main() {
	var opts Options

	flag.BoolVar(&opts.dryRun, "n", false, "dry run mode")
	flag.BoolVar(&opts.interactive, "i", false, "interactive mode")
	flag.BoolVar(&opts.showWhat, "w", false, "show target (input line)")
	flag.BoolVar(&opts.showCommand, "l", false, "show command line")
	flag.BoolVar(&opts.forceContinue, "f", false, "force continue on errors")
	flag.Float64Var(&opts.interval, "t", 0, "interval between commands in seconds")
	flag.BoolVar(&opts.shellMode, "c", false, "shell mode")
	flag.StringVar(&opts.color, "color", "auto", "color output (never/always/auto)")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] command_template\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Validate color option
	if opts.color != "never" && opts.color != "always" && opts.color != "auto" {
		fmt.Fprintf(os.Stderr, "Error: invalid color option '%s', must be never/always/auto\n", opts.color)
		os.Exit(1)
	}

	// Validate arguments based on execution mode
	if opts.shellMode {
		if len(args) != 1 {
			fmt.Fprintf(os.Stderr, "Error: shell mode (-c) requires exactly one argument\n")
			fmt.Fprintf(os.Stderr, "Usage: %s -c 'shell_command'\n", os.Args[0])
			os.Exit(1)
		}
	}

	app := &App{
		stdin:    os.Stdin,
		stdout:   os.Stdout,
		stderr:   os.Stderr,
		executor: NewRealCommandExecutor(os.Stdout, os.Stderr),
		useColor: shouldUseColor(opts.color, os.Stderr),
	}

	err := app.processStdin(opts, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func (app *App) processStdin(opts Options, args []string) error {
	scanner := bufio.NewScanner(app.stdin)
	lastErrorCode := 0

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\n\r")
		
		// Handle empty lines
		if line == "" {
			if opts.showWhat || opts.showCommand {
				app.printWarning("[empty line]\n")
			}
			continue
		}

		if opts.showWhat {
			app.printTarget("%s\n", line)
		}

		// Replace placeholders and prepare command
		var commandDisplay string
		var exitCode int
		
		if opts.shellMode {
			// Shell mode: single string command
			command := replacePlaceholders(args[0], line)
			commandDisplay = command
			
			if opts.showCommand {
				app.printCommand("> %s\n", command)
			}

			if opts.dryRun {
				fmt.Fprintf(app.stdout, "%s\n", command)
				continue
			}

			if opts.interactive {
				fmt.Fprintf(app.stderr, "Execute: %s [y/N] ", command)
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					continue
				}
			}

			exitCode = app.executor.ExecuteShell(command)
		} else {
			// Direct mode: replace placeholders in each argument
			commandArgs := make([]string, len(args))
			for i, arg := range args {
				commandArgs[i] = replacePlaceholders(arg, line)
			}
			commandDisplay = strings.Join(commandArgs, " ")
			
			if opts.showCommand {
				app.printCommand("> %s\n", commandDisplay)
			}

			if opts.dryRun {
				fmt.Fprintf(app.stdout, "%s\n", commandDisplay)
				continue
			}

			if opts.interactive {
				fmt.Fprintf(app.stderr, "Execute: %s [y/N] ", commandDisplay)
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					continue
				}
			}

			exitCode = app.executor.ExecuteDirect(commandArgs)
		}

		if opts.showCommand {
			if exitCode == 0 {
				app.printSuccess("[exit: %d]\n", exitCode)
			} else {
				app.printError("[exit: %d]\n", exitCode)
			}
		}

		if exitCode != 0 {
			lastErrorCode = exitCode
			if !opts.forceContinue {
				return fmt.Errorf("command failed with exit code %d", exitCode)
			}
		}

		// Wait interval if specified
		if opts.interval > 0 {
			time.Sleep(time.Duration(opts.interval * float64(time.Second)))
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stdin: %v", err)
	}

	if lastErrorCode != 0 {
		return fmt.Errorf("commands completed with errors, last exit code: %d", lastErrorCode)
	}

	return nil
}

func replacePlaceholders(template, input string) string {
	return strings.ReplaceAll(template, "{}", input)
}



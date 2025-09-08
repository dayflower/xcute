package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
)

type Options struct {
	dryRun      bool
	interactive bool
	showWhat    bool
	showCommand bool
	forceContinue bool
	interval    float64
	shellMode   bool
}

// Color functions for output
var (
	colorError     = color.New(color.FgRed)
	colorSuccess   = color.New(color.FgGreen)
	colorWarning   = color.New(color.FgYellow)
	colorCommand   = color.New(color.FgBlue)
	colorTarget    = color.New(color.FgCyan)
)

func main() {
	var opts Options

	flag.BoolVar(&opts.dryRun, "n", false, "dry run mode")
	flag.BoolVar(&opts.interactive, "i", false, "interactive mode")
	flag.BoolVar(&opts.showWhat, "w", false, "show target (input line)")
	flag.BoolVar(&opts.showCommand, "l", false, "show command line")
	flag.BoolVar(&opts.forceContinue, "f", false, "force continue on errors")
	flag.Float64Var(&opts.interval, "t", 0, "interval between commands in seconds")
	flag.BoolVar(&opts.shellMode, "c", false, "shell mode")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] command_template\n", os.Args[0])
		flag.PrintDefaults()
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

	err := processStdin(opts, args)
	if err != nil {
		os.Exit(1)
	}
}

func processStdin(opts Options, args []string) error {
	scanner := bufio.NewScanner(os.Stdin)
	lastErrorCode := 0

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\n\r")
		
		// Handle empty lines
		if line == "" {
			if opts.showWhat || opts.showCommand {
				colorWarning.Fprintf(os.Stderr, "[empty line]\n")
			}
			continue
		}

		if opts.showWhat {
			colorTarget.Fprintf(os.Stderr, "%s\n", line)
		}

		// Replace placeholders and prepare command
		var commandDisplay string
		var exitCode int
		
		if opts.shellMode {
			// Shell mode: single string command
			command := replacePlaceholders(args[0], line)
			commandDisplay = command
			
			if opts.showCommand {
				colorCommand.Fprintf(os.Stderr, "> %s\n", command)
			}

			if opts.dryRun {
				fmt.Println(command)
				continue
			}

			if opts.interactive {
				fmt.Fprintf(os.Stderr, "Execute: %s [y/N] ", command)
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					continue
				}
			}

			exitCode = executeShellCommand(command)
		} else {
			// Direct mode: replace placeholders in each argument
			commandArgs := make([]string, len(args))
			for i, arg := range args {
				commandArgs[i] = replacePlaceholders(arg, line)
			}
			commandDisplay = strings.Join(commandArgs, " ")
			
			if opts.showCommand {
				colorCommand.Fprintf(os.Stderr, "> %s\n", commandDisplay)
			}

			if opts.dryRun {
				fmt.Println(commandDisplay)
				continue
			}

			if opts.interactive {
				fmt.Fprintf(os.Stderr, "Execute: %s [y/N] ", commandDisplay)
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					continue
				}
			}

			exitCode = executeDirectCommand(commandArgs)
		}

		if opts.showCommand {
			if exitCode == 0 {
				colorSuccess.Fprintf(os.Stderr, "[exit: %d]\n", exitCode)
			} else {
				colorError.Fprintf(os.Stderr, "[exit: %d]\n", exitCode)
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
		os.Exit(lastErrorCode)
	}

	return nil
}

func replacePlaceholders(template, input string) string {
	return strings.ReplaceAll(template, "{}", input)
}

func executeShellCommand(command string) int {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		}
		return 1
	}
	return 0
}

func executeDirectCommand(args []string) int {
	if len(args) == 0 {
		return 0
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		}
		return 1
	}
	return 0
}


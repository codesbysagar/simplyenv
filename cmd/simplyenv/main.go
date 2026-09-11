package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"simplyenv/internal/core"
	"simplyenv/internal/server"
	"simplyenv/internal/shell"
)

const version = "1.0.0"

func printHelp() {
	helpText := `simplyenv - The developer's environment manager & modern direnv alternative

USAGE:
  simplyenv [command] [options]

COMMANDS:
  eval                 Evaluate directory environment and output shell export/unset (default)
  hook <shell>         Generate shell hook (bash, zsh, fish, pwsh)
  install-hook [shell] Automatically add simplyenv hook to your shell profile (.bashrc, .zshrc, etc.)
  uninstall-hook [sh]  Safely remove simplyenv hook from your shell profile
  ui                   Launch the graphical environment manager in your browser
  list [dir]           List environment variables for current or target directory
  set <KEY=VAL> [dir]  Set or update an environment variable
  unset <KEY> [dir]    Remove an environment variable
  import <file> [dir]  Bulk import environment variables from a file (.env or JSON)
  export [options]     Export environment variables (default: .env format)
  version              Print simplyenv version
  help                 Print this help message

EXAMPLES:
  # Configure shell integration
  eval "$(simplyenv hook bash)"

  # Open the graphical user interface
  simplyenv ui

  # Manage variables from terminal
  simplyenv set DATABASE_URL="postgres://localhost:5432/mydb"
  simplyenv list
  simplyenv unset DATABASE_URL

  # Bulk import & export
  simplyenv import .env.production
  simplyenv export --format json
`
	fmt.Print(helpText)
}

func main() {
	args := os.Args[1:]

	// Default behavior when invoked without arguments: eval (for shell hook)
	if len(args) == 0 {
		runEval()
		return
	}

	command := args[0]
	cmdArgs := args[1:]

	switch command {
	case "eval":
		runEval()

	case "hook":
		runHook(cmdArgs)

	case "install-hook":
		runInstallHook(cmdArgs)

	case "uninstall-hook":
		runUninstallHook(cmdArgs)

	case "ui":
		runUI(cmdArgs)

	case "list", "ls":
		runList(cmdArgs)

	case "set":
		runSet(cmdArgs)

	case "unset", "rm":
		runUnset(cmdArgs)

	case "import":
		runImport(cmdArgs)

	case "export":
		runExport(cmdArgs)

	case "version", "-v", "--version":
		fmt.Printf("simplyenv version %s\n", version)

	case "help", "-h", "--help":
		printHelp()

	default:
		// If argument looks like KEY=VALUE, treat as set shortcut
		if strings.Contains(command, "=") {
			runSet(args)
			return
		}

		fmt.Fprintf(os.Stderr, "Unknown command: %s\nRun 'simplyenv help' for usage.\n", command)
		os.Exit(1)
	}
}

func runEval() {
	// 1. Get previous state from environment variable
	prevState := core.DecodeState(os.Getenv("SIMPLYENV_STATE"))

	// 2. Find the config file for the current directory
	wd, _ := os.Getwd()
	configPath, err := core.FindConfig(wd)

	// --- Calculate Differences ---
	varsToUnset := make(map[string]bool)
	for _, key := range prevState.VarKeys {
		if key != "" {
			varsToUnset[key] = true
		}
	}

	var newEnvVars map[string]string
	var newStateStr string

	if err == nil { // A config file was found
		newEnvVars, _ = core.ParseEnvFile(configPath)
		newStateStr = core.EncodeState(configPath, newEnvVars)

		// Don't unset variables that are present in the new environment
		for key := range newEnvVars {
			delete(varsToUnset, key)
		}

		// Auto-register project in registry
		_, _ = core.RegisterProject(filepath.Dir(configPath))
	}

	// 3. Generate the shell commands
	output := shell.FormatUnset(varsToUnset)
	output += shell.FormatForShell(newEnvVars)
	output += shell.FormatState(newStateStr)

	fmt.Print(output)
}

func runHook(args []string) {
	shellType := "bash"
	if len(args) > 0 {
		shellType = args[0]
	}
	fmt.Print(shell.GenerateHook(shellType))
}

func runInstallHook(args []string) {
	shellType := ""
	if len(args) > 0 {
		shellType = args[0]
	} else {
		shellType = shell.DetectUserShell()
	}

	profilePath, err := shell.InstallHook(shellType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error installing hook: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Shell hook automatically installed into %s\n", profilePath)
	fmt.Println("  To activate in your current session, run: source " + profilePath)
}

func runUninstallHook(args []string) {
	shellType := ""
	if len(args) > 0 {
		shellType = args[0]
	} else {
		shellType = shell.DetectUserShell()
	}

	profilePath, err := shell.UninstallHook(shellType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error removing hook: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Shell hook removed from %s\n", profilePath)
}

func runUI(args []string) {
	uiCmd := flag.NewFlagSet("ui", flag.ExitOnError)
	port := uiCmd.Int("port", 8085, "Port to run the UI server on")
	noBrowser := uiCmd.Bool("no-browser", false, "Do not automatically launch browser")
	_ = uiCmd.Parse(args)

	// Register current working directory automatically
	if wd, err := os.Getwd(); err == nil {
		_, _ = core.RegisterProject(wd)
	}

	err := server.StartServer(*port, !*noBrowser)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running simplyenv UI: %v\n", err)
		os.Exit(1)
	}
}

func runList(args []string) {
	targetDir, _ := os.Getwd()
	if len(args) > 0 {
		targetDir = args[0]
	}

	configPath, err := core.FindConfig(targetDir)
	if err != nil {
		fmt.Println("No .simplyenv or .envrc found for this directory.")
		return
	}

	vars, err := core.ParseEnvFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", configPath, err)
		os.Exit(1)
	}

	if len(vars) == 0 {
		fmt.Printf("Config file %s is empty.\n", configPath)
		return
	}

	fmt.Printf("Variables configured in %s:\n\n", configPath)
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("  %-24s = %s\n", k, vars[k])
	}
}

func runSet(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: simplyenv set KEY=VALUE [directory]")
		os.Exit(1)
	}

	pair := args[0]
	parts := strings.SplitN(pair, "=", 2)
	if len(parts) != 2 {
		fmt.Fprintln(os.Stderr, "Error: Must specify variable as KEY=VALUE")
		os.Exit(1)
	}

	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])

	targetDir, _ := os.Getwd()
	if len(args) > 1 {
		targetDir = args[1]
	}

	configPath, err := core.FindConfig(targetDir)
	if err != nil {
		configPath = filepath.Join(targetDir, ".simplyenv")
	}

	if err := core.SetEnvVar(configPath, key, val); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting variable: %v\n", err)
		os.Exit(1)
	}

	_, _ = core.RegisterProject(targetDir)
	fmt.Printf("✓ Set %s in %s\n", key, configPath)
}

func runUnset(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: simplyenv unset KEY [directory]")
		os.Exit(1)
	}

	key := strings.TrimSpace(args[0])
	targetDir, _ := os.Getwd()
	if len(args) > 1 {
		targetDir = args[1]
	}

	configPath, err := core.FindConfig(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "No config file found in %s\n", targetDir)
		os.Exit(1)
	}

	if err := core.DeleteEnvVar(configPath, key); err != nil {
		fmt.Fprintf(os.Stderr, "Error removing variable: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Removed %s from %s\n", key, configPath)
}

func runImport(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: simplyenv import <filepath> [directory]")
		os.Exit(1)
	}

	importFile := args[0]
	targetDir, _ := os.Getwd()
	if len(args) > 1 {
		targetDir = args[1]
	}

	data, err := os.ReadFile(importFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", importFile, err)
		os.Exit(1)
	}

	importedVars, err := core.ParseRawImport(string(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing import: %v\n", err)
		os.Exit(1)
	}

	configPath, err := core.FindConfig(targetDir)
	if err != nil {
		configPath = filepath.Join(targetDir, ".simplyenv")
	}

	existingVars, _ := core.ParseEnvFile(configPath)
	if existingVars == nil {
		existingVars = make(map[string]string)
	}

	for k, v := range importedVars {
		existingVars[k] = v
	}

	if err := core.WriteEnvFile(configPath, existingVars); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", configPath, err)
		os.Exit(1)
	}

	_, _ = core.RegisterProject(targetDir)
	fmt.Printf("✓ Successfully imported %d variables into %s\n", len(importedVars), configPath)
}

func runExport(args []string) {
	exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
	formatStr := exportCmd.String("format", "env", "Export format: env, json, shell, docker")
	_ = exportCmd.Parse(args)

	targetDir, _ := os.Getwd()
	remaining := exportCmd.Args()
	if len(remaining) > 0 {
		targetDir = remaining[0]
	}

	configPath, err := core.FindConfig(targetDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "No config file found.")
		os.Exit(1)
	}

	vars, err := core.ParseEnvFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		os.Exit(1)
	}

	output, err := core.FormatExport(vars, core.ExportFormat(*formatStr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting export: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)
}

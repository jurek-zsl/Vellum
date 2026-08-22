package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jurekzsl/vellum/internal/config"
	"github.com/jurekzsl/vellum/internal/model"
	"github.com/jurekzsl/vellum/internal/runner"
	"github.com/jurekzsl/vellum/internal/store"
	"github.com/jurekzsl/vellum/internal/tui"
)

const Version = "0.2.0"

func main() {
	if len(os.Args) > 1 {
		handleCLI(os.Args[1:])
		return
	}

	p := tea.NewProgram(
		tui.InitialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Vellum encountered an error: %v\n", err)
		os.Exit(1)
	}
}

func handleCLI(args []string) {
	cmd := args[0]

	switch cmd {
	case "version", "-v", "--version":
		fmt.Printf("Vellum v%s\n", Version)
		return

	case "help", "-h", "--help":
		printHelp()
		return

	case "list", "ls":
		handleList(args[1:])
		return

	case "run", "exec":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: script or alias name required. Usage: vellum run <name> [args...]")
			os.Exit(1)
		}
		handleRun(args[1], args[2:])
		return

	case "export":
		handleExport(args[1:])
		return

	default:
		// If user typed `vellum <scriptName> [args...]` directly
		handleRun(cmd, args[1:])
		return
	}
}

func printHelp() {
	fmt.Printf(`Vellum v%s - Modern Script & Shell Alias Manager

Usage:
  vellum                    Launch interactive TUI
  vellum run <name> [args]  Execute a script or alias headlessly
  vellum list [--json]      List all configured scripts and aliases
  vellum export [md|json]   Export scripts and metadata
  vellum version            Print version information
  vellum help               Show this help message

Options:
  -h, --help                Show help
  -v, --version             Show version
`, Version)
}

func handleList(args []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	s := store.NewStore(cfg)
	items, err := s.ListItems()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing items: %v\n", err)
		os.Exit(1)
	}

	for _, a := range args {
		if a == "--json" || a == "-j" {
			data, _ := json.MarshalIndent(items, "", "  ")
			fmt.Println(string(data))
			return
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tLAST RUN\tEXIT\tRUNS\tDESCRIPTION")
	fmt.Fprintln(w, "----\t----\t--------\t----\t----\t-----------")

	for _, item := range items {
		lastRun := "Never"
		if !item.LastRunAt.IsZero() {
			lastRun = item.LastRunAt.Format("2006-01-02 15:04")
		}
		itemType := string(item.Type)
		if item.Type == model.TypeScript && item.ScriptType != "" {
			itemType = fmt.Sprintf("script(%s)", item.ScriptType)
		}
		desc := item.Desc
		if desc == "" && item.Type == model.TypeAlias {
			desc = item.Command
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\n",
			item.Name,
			itemType,
			lastRun,
			item.LastExitCode,
			item.RunCount,
			desc,
		)
	}
	w.Flush()
}

func handleRun(name string, params []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	s := store.NewStore(cfg)
	items, err := s.ListItems()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing items: %v\n", err)
		os.Exit(1)
	}

	var target *model.Metadata
	for _, item := range items {
		if strings.EqualFold(item.Name, name) {
			target = &item
			break
		}
	}

	if target == nil {
		fmt.Fprintf(os.Stderr, "Error: item '%s' not found\n", name)
		os.Exit(1)
	}

	logFile, err := s.CreateLogFile(*target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create log file: %v\n", err)
	} else {
		defer logFile.Close()
	}

	var stdoutWriter = os.Stdout
	var stderrWriter = os.Stderr

	exitCode, durationMs, runErr := runner.ExecuteHeadless(*target, params, cfg.DefaultShell, "", stdoutWriter, stderrWriter)
	_ = s.UpdateExecution(*target, exitCode, durationMs)

	if runErr != nil && exitCode != 0 {
		os.Exit(exitCode)
	}
}

func handleExport(args []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	s := store.NewStore(cfg)
	format := "md"
	if len(args) > 0 {
		format = args[0]
	}

	if format == "json" {
		items, err := s.ListItems()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		data, _ := json.MarshalIndent(items, "", "  ")
		fmt.Println(string(data))
		return
	}

	out, err := s.ExportMarkdown()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting markdown: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(out)
}


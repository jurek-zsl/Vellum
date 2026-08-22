package runner

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jurekzsl/vellum/internal/model"
)

// PrepareCommand builds an interactive *exec.Cmd with stdout/stderr piped to terminal and logfile
func PrepareCommand(item model.Metadata, params []string, logFile *os.File, shell string, user string) (*exec.Cmd, error) {
	if shell == "" {
		shell = "/bin/sh"
	}

	var fullCmd string

	if item.Type == model.TypeAlias {
		// For aliases: run the alias command with params appended
		aliasStr := item.Command
		if len(params) > 0 {
			var quotedParams []string
			for _, p := range params {
				quotedParams = append(quotedParams, quoteShellArg(p))
			}
			aliasStr += " " + strings.Join(quotedParams, " ")
		}
		fullCmd = aliasStr
	} else {
		var parts []string
		interpreter := getInterpreter(item.ScriptType)
		if interpreter != "" {
			parts = append(parts, strings.Fields(interpreter)...)
			parts = append(parts, item.FilePath)
		} else {
			parts = []string{item.FilePath}
		}

		parts = append(parts, params...)

		var quotedParts []string
		for _, p := range parts {
			quotedParts = append(quotedParts, quoteShellArg(p))
		}
		fullCmd = strings.Join(quotedParts, " ")
	}

	if item.RequiresSudo && user == "" {
		fullCmd = "sudo " + fullCmd
	}

	// Interactive wrapper: pause on completion so output is readable
	wrapper := fmt.Sprintf("%s; echo ''; echo 'Press Enter to return to Vellum...'; read line", fullCmd)

	var cmd *exec.Cmd
	if user != "" {
		cmd = exec.Command("sudo", "-u", user, shell, "-c", wrapper)
	} else {
		cmd = exec.Command(shell, "-c", wrapper)
	}

	setSysProcAttr(cmd)

	if logFile != nil {
		cmd.Stdout = io.MultiWriter(os.Stdout, logFile)
		cmd.Stderr = io.MultiWriter(os.Stderr, logFile)
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.Stdin = os.Stdin

	return cmd, nil
}

// ExecuteHeadless runs a script or alias non-interactively, capturing exit code and duration
func ExecuteHeadless(item model.Metadata, params []string, shell string, user string, stdout io.Writer, stderr io.Writer) (int, int64, error) {
	if shell == "" {
		shell = "/bin/sh"
	}

	var fullCmd string
	if item.Type == model.TypeAlias {
		fullCmd = item.Command
		if len(params) > 0 {
			var quotedParams []string
			for _, p := range params {
				quotedParams = append(quotedParams, quoteShellArg(p))
			}
			fullCmd += " " + strings.Join(quotedParams, " ")
		}
	} else {
		var parts []string
		interpreter := getInterpreter(item.ScriptType)
		if interpreter != "" {
			parts = append(parts, strings.Fields(interpreter)...)
			parts = append(parts, item.FilePath)
		} else {
			parts = []string{item.FilePath}
		}
		parts = append(parts, params...)

		var quotedParts []string
		for _, p := range parts {
			quotedParts = append(quotedParts, quoteShellArg(p))
		}
		fullCmd = strings.Join(quotedParts, " ")
	}

	if item.RequiresSudo && user == "" {
		fullCmd = "sudo " + fullCmd
	}

	var cmd *exec.Cmd
	if user != "" {
		cmd = exec.Command("sudo", "-u", user, shell, "-c", fullCmd)
	} else {
		cmd = exec.Command(shell, "-c", fullCmd)
	}

	setSysProcAttr(cmd)

	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}

	start := time.Now()
	err := cmd.Run()
	durationMs := time.Since(start).Milliseconds()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return exitCode, durationMs, err
}

func quoteShellArg(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func getInterpreter(t model.ScriptType) string {
	switch t {
	// Coding
	case model.ScriptTypePython:
		return "python3"
	case model.ScriptTypeJavascript:
		return "node"
	case model.ScriptTypeTypescript:
		if _, err := exec.LookPath("ts-node"); err == nil {
			return "ts-node"
		}
		if _, err := exec.LookPath("bun"); err == nil {
			return "bun run"
		}
		if _, err := exec.LookPath("deno"); err == nil {
			return "deno run"
		}
		return "node"
	case model.ScriptTypeJSX:
		return "node"
	case model.ScriptTypeTSX:
		if _, err := exec.LookPath("ts-node"); err == nil {
			return "ts-node"
		}
		if _, err := exec.LookPath("bun"); err == nil {
			return "bun run"
		}
		return "node"
	case model.ScriptTypeRuby:
		return "ruby"
	case model.ScriptTypePerl:
		return "perl"
	case model.ScriptTypePHP:
		return "php"
	case model.ScriptTypeLua:
		return "lua"
	case model.ScriptTypeTcl:
		if _, err := exec.LookPath("tclsh"); err == nil {
			return "tclsh"
		}
		return "tcl"

	// Shell
	case model.ScriptTypeBash, model.ScriptTypeShell:
		return "bash"
	case model.ScriptTypeZsh:
		return "zsh"
	case model.ScriptTypeFish:
		return "fish"
	case model.ScriptTypePowershell:
		if _, err := exec.LookPath("pwsh"); err == nil {
			return "pwsh -File"
		}
		return "powershell -File"

	// Compiled
	case model.ScriptTypeGo:
		return "go run"
	case model.ScriptTypeRust:
		if _, err := exec.LookPath("rust-script"); err == nil {
			return "rust-script"
		}
		return "cargo run --quiet --manifest-path"
	case model.ScriptTypeJava:
		return "java"
	case model.ScriptTypeKotlin:
		if _, err := exec.LookPath("kotlinc"); err == nil {
			return "kotlinc -script"
		}
		return "kotlin"
	case model.ScriptTypeSwift:
		return "swift"

	// System
	case model.ScriptTypeAppleScript:
		return "osascript"
	case model.ScriptTypeBat, model.ScriptTypeCmd:
		return "cmd.exe /c"
	case model.ScriptTypeVBS:
		return "cscript //nologo"

	default:
		return ""
	}
}


package runner

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/jurekzsl/vellum/internal/model"
)

func PrepareCommand(item model.Metadata, params []string, logFile *os.File, shell string) (*exec.Cmd, error) {
	if shell == "" {
		shell = "/bin/sh"
	}
	var parts []string

	if item.Type == model.TypeAlias {
		// For aliases, run directly in user's shell
		parts = []string{shell, "-c", item.Command}
	} else {
		interpreter := getInterpreter(item.ScriptType)
		if interpreter != "" {
			// Handle "go run" case
			interpreterParts := strings.Split(interpreter, " ")
			parts = append(parts, interpreterParts...)
			parts = append(parts, item.FilePath)
		} else {
			parts = []string{item.FilePath}
		}
	}

	parts = append(parts, params...)

	// Always use Sudo as requested
	if item.RequiresSudo {
		parts = append([]string{"sudo"}, parts...)
	}

	// Quote parts for shell wrapper
	var quotedParts []string
	for _, p := range parts {
		// Single quote escaping for shell
		escaped := strings.ReplaceAll(p, "'", "'\\''")
		quotedParts = append(quotedParts, fmt.Sprintf("'%s'", escaped))
	}
	fullCmd := strings.Join(quotedParts, " ")

	// Wrap in shell to wait for Enter
	wrapper := fmt.Sprintf("%s; echo ''; echo 'Press Enter to return to Vellum...'; read line", fullCmd)
	cmd := exec.Command(shell, "-c", wrapper)

	mwOut := io.MultiWriter(os.Stdout, logFile)
	mwErr := io.MultiWriter(os.Stderr, logFile)

	cmd.Stdout = mwOut
	cmd.Stderr = mwErr
	cmd.Stdin = os.Stdin

	return cmd, nil
}

func getInterpreter(t model.ScriptType) string {
	switch t {
	case model.ScriptTypePython:
		return "python3"
	case model.ScriptTypeJavascript, model.ScriptTypeTypescript, model.ScriptTypeJSX, model.ScriptTypeTSX:
		// For TS/JSX we might need ts-node or similar, but let's stick to node for JS or assume user has setup
		// For simple node scripts:
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
		return "tcl"
	case model.ScriptTypeBash, model.ScriptTypeShell:
		return "bash"
	case model.ScriptTypeZsh:
		return "zsh"
	case model.ScriptTypeFish:
		return "fish"
	case model.ScriptTypePowershell:
		return "pwsh" // or powershell
	case model.ScriptTypeGo:
		return "go run" // This needs splitting
	default:
		return ""
	}
}

// Special handling for multi-word interpreters like "go run"
func PrepareCommandWithInterpreter(item model.Metadata, params []string, logFile *os.File) (*exec.Cmd, error) {
	// Re-implementing simplified logic to handle "go run" split
	var parts []string

	interpreter := getInterpreter(item.ScriptType)
	if interpreter == "go run" {
		parts = []string{"go", "run"}
	} else if interpreter != "" {
		parts = []string{interpreter}
	}

	if item.Type == model.TypeAlias {
		parts = []string{"/bin/sh", "-c", item.Command}
	} else {
		if len(parts) > 0 {
			parts = append(parts, item.FilePath)
		} else {
			parts = []string{item.FilePath}
		}
	}

	parts = append(parts, params...)

	if item.RequiresSudo {
		parts = append([]string{"sudo"}, parts...)
	}

	name := parts[0]
	args := parts[1:]

	cmd := exec.Command(name, args...)

	mwOut := io.MultiWriter(os.Stdout, logFile)
	mwErr := io.MultiWriter(os.Stderr, logFile)

	cmd.Stdout = mwOut
	cmd.Stderr = mwErr
	cmd.Stdin = os.Stdin

	return cmd, nil
}

package model

import (
	"fmt"
	"time"
)

type ItemType string

const (
	TypeScript ItemType = "script"
	TypeAlias  ItemType = "alias"
)

type ScriptType string

const (
	// Coding
	ScriptTypePython     ScriptType = ".py"
	ScriptTypeJavascript ScriptType = ".js"
	ScriptTypeRuby       ScriptType = ".rb"
	ScriptTypePerl       ScriptType = ".pl"
	ScriptTypePHP        ScriptType = ".php"
	ScriptTypeLua        ScriptType = ".lua"
	ScriptTypeTcl        ScriptType = ".tcl"
	ScriptTypeTypescript ScriptType = ".ts"
	ScriptTypeJSX        ScriptType = ".jsx"
	ScriptTypeTSX        ScriptType = ".tsx"

	// Shell
	ScriptTypeShell      ScriptType = ".sh"
	ScriptTypeBash       ScriptType = ".bash"
	ScriptTypeZsh        ScriptType = ".zsh"
	ScriptTypeFish       ScriptType = ".fish"
	ScriptTypePowershell ScriptType = ".ps1"

	// Compiled (Source)
	ScriptTypeGo    ScriptType = ".go"
	ScriptTypeRust  ScriptType = ".rs"
	ScriptTypeJava  ScriptType = ".java"
	ScriptTypeKotlin ScriptType = ".kt"
	ScriptTypeSwift ScriptType = ".swift"

	// System
	ScriptTypeBat         ScriptType = ".bat"
	ScriptTypeCmd         ScriptType = ".cmd"
	ScriptTypeVBS         ScriptType = ".vbs"
	ScriptTypeAppleScript ScriptType = ".applescript"
)

// Metadata represents the JSON structure stored for each item
type Metadata struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Desc        string     `json:"description"`
	Type        ItemType   `json:"type"` // script or alias
	ScriptType  ScriptType `json:"script_type,omitempty"`
	Command     string     `json:"command"` // The command to run or the alias expansion
	RequiresSudo bool       `json:"requires_sudo"`
	UsesParams   bool       `json:"uses_params"`
	ParamDesc    string     `json:"param_desc,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	LastRunAt    time.Time  `json:"last_run_at,omitempty"`
	FilePath     string     `json:"-"` // Internal use, not stored in JSON
}

func (m Metadata) Title() string { 
	symbol := "s"
	if m.Type == TypeAlias {
		symbol = "l"
	}
	return fmt.Sprintf("[%s] %s", symbol, m.Name) 
}

func (m Metadata) Description() string { return m.Desc }

func (m Metadata) FilterValue() string { 
	symbol := "s"
	if m.Type == TypeAlias {
		symbol = "l"
	}
	return fmt.Sprintf("*%s %s", symbol, m.Name) 
}

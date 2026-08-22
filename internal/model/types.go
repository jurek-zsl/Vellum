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
	LastRunAt      time.Time  `json:"last_run_at,omitempty"`
	LastExitCode   int        `json:"last_exit_code,omitempty"`
	LastDurationMs int64      `json:"last_duration_ms,omitempty"`
	RunCount       int        `json:"run_count,omitempty"`
	FilePath       string     `json:"-"` // Internal use, not stored in JSON
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

func GetDefaultTemplate(t ScriptType) string {
	switch t {
	case ScriptTypePython:
		return "#!/usr/bin/env python3\n\ndef main():\n    print(\"Hello from Vellum Python script!\")\n\nif __name__ == \"__main__\":\n    main()\n"
	case ScriptTypeJavascript:
		return "#!/usr/bin/env node\n\nconsole.log(\"Hello from Vellum JavaScript script!\");\n"
	case ScriptTypeTypescript:
		return "#!/usr/bin/env ts-node\n\nconst greeting: string = \"Hello from Vellum TypeScript script!\";\nconsole.log(greeting);\n"
	case ScriptTypeJSX:
		return "// JSX Script\nconsole.log(\"Hello from Vellum JSX script!\");\n"
	case ScriptTypeTSX:
		return "// TSX Script\nconst greeting: string = \"Hello from Vellum TSX script!\";\nconsole.log(greeting);\n"
	case ScriptTypeRuby:
		return "#!/usr/bin/env ruby\n\nputs \"Hello from Vellum Ruby script!\"\n"
	case ScriptTypePerl:
		return "#!/usr/bin/env perl\nuse strict;\nuse warnings;\n\nprint \"Hello from Vellum Perl script!\\n\";\n"
	case ScriptTypePHP:
		return "<?php\n\necho \"Hello from Vellum PHP script!\\n\";\n"
	case ScriptTypeLua:
		return "#!/usr/bin/env lua\n\nprint(\"Hello from Vellum Lua script!\")\n"
	case ScriptTypeTcl:
		return "#!/usr/bin/env tclsh\n\nputs \"Hello from Vellum Tcl script!\"\n"
	case ScriptTypeShell, ScriptTypeBash:
		return "#!/usr/bin/env bash\nset -euo pipefail\n\necho \"Hello from Vellum Bash script!\"\n"
	case ScriptTypeZsh:
		return "#!/usr/bin/env zsh\n\necho \"Hello from Vellum Zsh script!\"\n"
	case ScriptTypeFish:
		return "#!/usr/bin/env fish\n\necho \"Hello from Vellum Fish script!\"\n"
	case ScriptTypePowershell:
		return "# PowerShell Script\nWrite-Host \"Hello from Vellum PowerShell script!\"\n"
	case ScriptTypeGo:
		return "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello from Vellum Go script!\")\n}\n"
	case ScriptTypeRust:
		return "fn main() {\n    println!(\"Hello from Vellum Rust script!\");\n}\n"
	case ScriptTypeJava:
		return "public class Main {\n    public static void main(String[] args) {\n        System.out.println(\"Hello from Vellum Java script!\");\n    }\n}\n"
	case ScriptTypeKotlin:
		return "fun main() {\n    println(\"Hello from Vellum Kotlin script!\")\n}\n"
	case ScriptTypeSwift:
		return "#!/usr/bin/env swift\nimport Foundation\n\nprint(\"Hello from Vellum Swift script!\")\n"
	case ScriptTypeAppleScript:
		return "display notification \"Hello from Vellum!\" with title \"Vellum\"\n"
	case ScriptTypeBat, ScriptTypeCmd:
		return "@echo off\necho Hello from Vellum Batch script!\n"
	case ScriptTypeVBS:
		return "WScript.Echo \"Hello from Vellum VBScript!\"\n"
	default:
		return "#!/usr/bin/env bash\n\necho \"Hello from Vellum!\"\n"
	}
}

func ScriptTypeDisplayName(t ScriptType) string {
	switch t {
	case ScriptTypePython:
		return "Python (.py)"
	case ScriptTypeJavascript:
		return "JavaScript (.js)"
	case ScriptTypeTypescript:
		return "TypeScript (.ts)"
	case ScriptTypeJSX:
		return "React JSX (.jsx)"
	case ScriptTypeTSX:
		return "React TSX (.tsx)"
	case ScriptTypeRuby:
		return "Ruby (.rb)"
	case ScriptTypePerl:
		return "Perl (.pl)"
	case ScriptTypePHP:
		return "PHP (.php)"
	case ScriptTypeLua:
		return "Lua (.lua)"
	case ScriptTypeTcl:
		return "Tcl (.tcl)"
	case ScriptTypeShell:
		return "Shell (.sh)"
	case ScriptTypeBash:
		return "Bash (.bash)"
	case ScriptTypeZsh:
		return "Zsh (.zsh)"
	case ScriptTypeFish:
		return "Fish (.fish)"
	case ScriptTypePowershell:
		return "PowerShell (.ps1)"
	case ScriptTypeGo:
		return "Go (.go)"
	case ScriptTypeRust:
		return "Rust (.rs)"
	case ScriptTypeJava:
		return "Java (.java)"
	case ScriptTypeKotlin:
		return "Kotlin (.kt)"
	case ScriptTypeSwift:
		return "Swift (.swift)"
	case ScriptTypeAppleScript:
		return "AppleScript (.applescript)"
	case ScriptTypeBat:
		return "Batch (.bat)"
	case ScriptTypeCmd:
		return "Windows Cmd (.cmd)"
	case ScriptTypeVBS:
		return "VBScript (.vbs)"
	default:
		return string(t)
	}
}


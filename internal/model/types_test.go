package model

import (
	"testing"
)

func TestScriptTypeHelpers(t *testing.T) {
	allTypes := []ScriptType{
		ScriptTypePython,
		ScriptTypeJavascript,
		ScriptTypeTypescript,
		ScriptTypeJSX,
		ScriptTypeTSX,
		ScriptTypeRuby,
		ScriptTypePerl,
		ScriptTypePHP,
		ScriptTypeLua,
		ScriptTypeTcl,
		ScriptTypeShell,
		ScriptTypeBash,
		ScriptTypeZsh,
		ScriptTypeFish,
		ScriptTypePowershell,
		ScriptTypeGo,
		ScriptTypeRust,
		ScriptTypeJava,
		ScriptTypeKotlin,
		ScriptTypeSwift,
		ScriptTypeAppleScript,
		ScriptTypeBat,
		ScriptTypeCmd,
		ScriptTypeVBS,
	}

	for _, st := range allTypes {
		tmpl := GetDefaultTemplate(st)
		if tmpl == "" {
			t.Errorf("expected default template for %s, got empty", st)
		}

		name := ScriptTypeDisplayName(st)
		if name == "" {
			t.Errorf("expected display name for %s, got empty", st)
		}
	}
}

func TestMetadataModel(t *testing.T) {
	metaScript := Metadata{
		Name: "test-script",
		Desc: "test description",
		Type: TypeScript,
	}

	if metaScript.Title() != "[s] test-script" {
		t.Errorf("unexpected script title: %s", metaScript.Title())
	}
	if metaScript.Description() != "test description" {
		t.Errorf("unexpected description: %s", metaScript.Description())
	}
	if metaScript.FilterValue() != "*s test-script" {
		t.Errorf("unexpected filter value: %s", metaScript.FilterValue())
	}

	metaAlias := Metadata{
		Name: "test-alias",
		Type: TypeAlias,
	}
	if metaAlias.Title() != "[l] test-alias" {
		t.Errorf("unexpected alias title: %s", metaAlias.Title())
	}
}

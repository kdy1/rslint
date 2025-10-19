package parser_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/typescript-estree/parser"
)

func TestParseOptions(t *testing.T) {
	t.Parallel()

	opts := &parser.ParseOptions{
		SourceType: "module",
		JSX:        true,
		FilePath:   "test.tsx",
		Loc:        true,
		Range:      true,
		Tokens:     true,
		Comment:    true,
	}

	if opts.SourceType != "module" {
		t.Errorf("Expected source type 'module', got '%s'", opts.SourceType)
	}

	if opts.JSX != true {
		t.Error("Expected JSX to be enabled")
	}

	if opts.Loc != true {
		t.Error("Expected Loc to be enabled")
	}

	if opts.Range != true {
		t.Error("Expected Range to be enabled")
	}
}

func TestJSDocParsingMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode parser.JSDocParsingMode
		want parser.JSDocParsingMode
	}{
		{"All mode", parser.JSDocParsingModeAll, parser.JSDocParsingModeAll},
		{"None mode", parser.JSDocParsingModeNone, parser.JSDocParsingModeNone},
		{"Type-info mode", parser.JSDocParsingModeTypeInfo, parser.JSDocParsingModeTypeInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &parser.ParseOptions{
				JSDocParsingMode: tt.mode,
			}

			if opts.JSDocParsingMode != tt.want {
				t.Errorf("Expected JSDocParsingMode %v, got %v", tt.want, opts.JSDocParsingMode)
			}
		})
	}
}

func TestParse_SimpleScript(t *testing.T) {
	t.Parallel()
	t.Skip("Standalone parsing not yet fully implemented - requires TypeScript source file creation API")

	source := "const x = 42;"
	ast, err := parser.Parse(source, &parser.ParseOptions{
		FilePath:   "test.ts",
		SourceType: "script",
		Loc:        true,
		Range:      true,
	})

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if ast == nil {
		t.Fatal("Expected AST, got nil")
	}

	if ast.Type() != "Program" {
		t.Errorf("Expected Program node, got %s", ast.Type())
	}

	if ast.SourceType != "script" {
		t.Errorf("Expected source type 'script', got '%s'", ast.SourceType)
	}
}

func TestParse_SimpleModule(t *testing.T) {
	t.Parallel()
	t.Skip("Standalone parsing not yet fully implemented - requires TypeScript source file creation API")
}

func TestParse_WithJSX(t *testing.T) {
	t.Parallel()
	t.Skip("Standalone parsing not yet fully implemented - requires TypeScript source file creation API")
}

func TestParse_DefaultOptions(t *testing.T) {
	t.Parallel()
	t.Skip("Standalone parsing not yet fully implemented - requires TypeScript source file creation API")
}

func TestParse_InvalidSyntax(t *testing.T) {
	t.Parallel()
	t.Skip("Standalone parsing not yet fully implemented - requires TypeScript source file creation API")
}

func TestParse_InvalidSyntaxAllowed(t *testing.T) {
	t.Parallel()
	t.Skip("Standalone parsing not yet fully implemented - requires TypeScript source file creation API")
}

func TestGetSupportedTypeScriptVersion(t *testing.T) {
	t.Parallel()

	version := parser.GetSupportedTypeScriptVersion()
	if version == "" {
		t.Error("Expected non-empty TypeScript version string")
	}

	// Should follow semver range format
	if len(version) < 5 {
		t.Errorf("Version string seems too short: %s", version)
	}
}

func TestValidateTypeScriptVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"Valid version 4.7.0", "4.7.0", false},
		{"Valid version 5.0.0", "5.0.0", false},
		{"Valid version 5.3.0", "5.3.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateTypeScriptVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTypeScriptVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

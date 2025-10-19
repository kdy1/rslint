package converter_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/typescript-estree/converter"
)

func TestConvertOptions(t *testing.T) {
	t.Parallel()

	opts := &converter.ConvertOptions{
		FilePath:   "test.ts",
		SourceType: "module",
		Loc:        true,
		Range:      true,
		Tokens:     true,
		Comment:    true,
	}

	if opts.FilePath != "test.ts" {
		t.Errorf("Expected FilePath to be 'test.ts', got '%s'", opts.FilePath)
	}

	if opts.SourceType != "module" {
		t.Errorf("Expected SourceType to be 'module', got '%s'", opts.SourceType)
	}

	if !opts.Loc {
		t.Error("Expected Loc to be true")
	}

	if !opts.Range {
		t.Error("Expected Range to be true")
	}

	if !opts.Tokens {
		t.Error("Expected Tokens to be true")
	}

	if !opts.Comment {
		t.Error("Expected Comment to be true")
	}
}

func TestConvertProgram_NilSourceFile(t *testing.T) {
	t.Parallel()

	opts := &converter.ConvertOptions{
		FilePath: "test.ts",
	}

	_, err := converter.ConvertProgram(nil, nil, opts)
	if err == nil {
		t.Error("Expected error for nil source file, got nil")
	}
}

// TODO: Add more comprehensive converter tests once TypeScript source file creation is available
// For now, the actual conversion will be tested through the parser integration tests

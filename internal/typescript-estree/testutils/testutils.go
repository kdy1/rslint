// Package testutils provides testing utilities for the typescript-estree module.
package testutils

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/typescript-estree/types"
)

// AssertNodeType checks that a node has the expected type.
func AssertNodeType(t *testing.T, node types.Node, expectedType string) {
	t.Helper()
	if node == nil {
		t.Fatal("Expected non-nil node")
	}
	if node.Type() != expectedType {
		t.Errorf("Expected node type '%s', got '%s'", expectedType, node.Type())
	}
}

// AssertLocation checks that a node has valid location information.
func AssertLocation(t *testing.T, node types.Node) {
	t.Helper()
	if node == nil {
		t.Fatal("Expected non-nil node")
	}
	loc := node.Loc()
	if loc == nil {
		t.Error("Expected non-nil location")
		return
	}
	if loc.Start.Line <= 0 {
		t.Errorf("Expected valid start line, got %d", loc.Start.Line)
	}
	if loc.End.Line <= 0 {
		t.Errorf("Expected valid end line, got %d", loc.End.Line)
	}
}

// CreateTestPosition creates a position for testing purposes.
func CreateTestPosition(line, column int) types.Position {
	return types.Position{
		Line:   line,
		Column: column,
	}
}

// CreateTestLocation creates a source location for testing purposes.
func CreateTestLocation(startLine, startColumn, endLine, endColumn int) *types.SourceLocation {
	return &types.SourceLocation{
		Start: CreateTestPosition(startLine, startColumn),
		End:   CreateTestPosition(endLine, endColumn),
	}
}

// CreateSourceFile creates a TypeScript SourceFile from code for testing purposes.
func CreateSourceFile(code, filename string) *ast.SourceFile {
	factory := ast.NewNodeFactory()
	sourceFile := compiler.CreateSourceFile(
		factory,
		filename,
		code,
		core.ScriptTargetLatest,
		false,
		core.ScriptKindTS,
	)
	return sourceFile
}

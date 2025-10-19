package types_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/typescript-estree/types"
)

func TestBaseNode(t *testing.T) {
	t.Parallel()

	node := &types.BaseNode{
		NodeType: "Identifier",
		Location: &types.SourceLocation{
			Start: types.Position{Line: 1, Column: 0},
			End:   types.Position{Line: 1, Column: 3},
		},
		Span: [2]int{0, 3},
	}

	if node.Type() != "Identifier" {
		t.Errorf("Expected type 'Identifier', got '%s'", node.Type())
	}

	if node.Loc() == nil {
		t.Error("Expected non-nil location")
	}

	if node.Range()[0] != 0 || node.Range()[1] != 3 {
		t.Errorf("Expected range [0, 3], got %v", node.Range())
	}
}

func TestSourceLocation(t *testing.T) {
	t.Parallel()

	loc := &types.SourceLocation{
		Start: types.Position{Line: 1, Column: 0},
		End:   types.Position{Line: 2, Column: 5},
	}

	if loc.Start.Line != 1 {
		t.Errorf("Expected start line 1, got %d", loc.Start.Line)
	}

	if loc.End.Line != 2 {
		t.Errorf("Expected end line 2, got %d", loc.End.Line)
	}
}

func TestIdentifier(t *testing.T) {
	t.Parallel()

	id := &types.Identifier{
		BaseNode: types.BaseNode{
			NodeType: "Identifier",
			Location: &types.SourceLocation{
				Start: types.Position{Line: 1, Column: 0},
				End:   types.Position{Line: 1, Column: 3},
			},
			Span: [2]int{0, 3},
		},
		Name: "foo",
	}

	if id.Name != "foo" {
		t.Errorf("Expected name 'foo', got '%s'", id.Name)
	}

	if id.Type() != "Identifier" {
		t.Errorf("Expected type 'Identifier', got '%s'", id.Type())
	}
}

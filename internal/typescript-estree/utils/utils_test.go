package utils_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/typescript-estree/types"
	"github.com/web-infra-dev/rslint/internal/typescript-estree/utils"
)

func TestGetNodeType(t *testing.T) {
	t.Parallel()

	node := &types.BaseNode{NodeType: "TestNode"}
	if utils.GetNodeType(node) != "TestNode" {
		t.Errorf("Expected 'TestNode', got '%s'", utils.GetNodeType(node))
	}

	if utils.GetNodeType(nil) != "" {
		t.Errorf("Expected empty string for nil node, got '%s'", utils.GetNodeType(nil))
	}
}

func TestIsValidPosition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pos      types.Position
		expected bool
	}{
		{"valid position", types.Position{Line: 1, Column: 0}, true},
		{"invalid line", types.Position{Line: 0, Column: 0}, false},
		{"invalid column", types.Position{Line: 1, Column: -1}, false},
		{"both invalid", types.Position{Line: 0, Column: -1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := utils.IsValidPosition(tt.pos)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestComparePositions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		a        types.Position
		b        types.Position
		expected int
	}{
		{"a before b (line)", types.Position{Line: 1, Column: 0}, types.Position{Line: 2, Column: 0}, -1},
		{"a after b (line)", types.Position{Line: 2, Column: 0}, types.Position{Line: 1, Column: 0}, 1},
		{"a before b (column)", types.Position{Line: 1, Column: 0}, types.Position{Line: 1, Column: 5}, -1},
		{"a after b (column)", types.Position{Line: 1, Column: 5}, types.Position{Line: 1, Column: 0}, 1},
		{"equal", types.Position{Line: 1, Column: 0}, types.Position{Line: 1, Column: 0}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := utils.ComparePositions(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

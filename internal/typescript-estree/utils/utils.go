// Package utils provides utility functions for the typescript-estree module.
package utils

import (
	"github.com/web-infra-dev/rslint/internal/typescript-estree/types"
)

// GetNodeType returns the type string for a given node.
func GetNodeType(node types.Node) string {
	if node == nil {
		return ""
	}
	return node.Type()
}

// IsValidPosition checks if a position is valid.
func IsValidPosition(pos types.Position) bool {
	return pos.Line > 0 && pos.Column >= 0
}

// IsValidSourceLocation checks if a source location is valid.
func IsValidSourceLocation(loc *types.SourceLocation) bool {
	if loc == nil {
		return false
	}
	return IsValidPosition(loc.Start) && IsValidPosition(loc.End)
}

// ComparePositions compares two positions and returns:
// -1 if a comes before b
// 0 if a and b are equal
// 1 if a comes after b
func ComparePositions(a, b types.Position) int {
	if a.Line < b.Line {
		return -1
	}
	if a.Line > b.Line {
		return 1
	}
	if a.Column < b.Column {
		return -1
	}
	if a.Column > b.Column {
		return 1
	}
	return 0
}

// Package traverse provides AST traversal utilities for TypeScript ESTree nodes.
package traverse

import (
	"github.com/web-infra-dev/rslint/internal/typescript-estree/types"
)

// VisitorAction represents the action to take during traversal.
type VisitorAction int

const (
	// Continue indicates traversal should continue normally.
	Continue VisitorAction = iota
	// Skip indicates traversal should skip the children of the current node.
	Skip
	// Stop indicates traversal should stop entirely.
	Stop
)

// Visitor defines the interface for visiting AST nodes during traversal.
// Implementations can control traversal flow by returning different VisitorActions.
type Visitor interface {
	// Enter is called when entering a node (before visiting its children).
	// Returns a VisitorAction to control traversal flow.
	Enter(node types.Node, parent types.Node) VisitorAction

	// Leave is called when leaving a node (after visiting its children).
	// This is only called if Enter didn't return Stop.
	Leave(node types.Node, parent types.Node)
}

// SimpleVisitor provides a basic implementation of Visitor with optional callbacks.
type SimpleVisitor struct {
	// EnterFunc is called when entering a node.
	EnterFunc func(node types.Node, parent types.Node) VisitorAction

	// LeaveFunc is called when leaving a node.
	LeaveFunc func(node types.Node, parent types.Node)
}

// Enter implements the Visitor interface.
func (v *SimpleVisitor) Enter(node types.Node, parent types.Node) VisitorAction {
	if v.EnterFunc != nil {
		return v.EnterFunc(node, parent)
	}
	return Continue
}

// Leave implements the Visitor interface.
func (v *SimpleVisitor) Leave(node types.Node, parent types.Node) {
	if v.LeaveFunc != nil {
		v.LeaveFunc(node, parent)
	}
}

// TypedVisitor allows visiting specific node types with custom handlers.
// The map key is the node type string (e.g., "Identifier", "BinaryExpression").
type TypedVisitor struct {
	// Visitors maps node types to their specific visitor functions.
	Visitors map[string]func(node types.Node, parent types.Node) VisitorAction

	// DefaultEnter is called for nodes without a specific visitor.
	DefaultEnter func(node types.Node, parent types.Node) VisitorAction

	// DefaultLeave is called when leaving any node.
	DefaultLeave func(node types.Node, parent types.Node)
}

// Enter implements the Visitor interface with type-specific dispatch.
func (v *TypedVisitor) Enter(node types.Node, parent types.Node) VisitorAction {
	if node == nil {
		return Continue
	}

	nodeType := node.Type()
	if handler, ok := v.Visitors[nodeType]; ok {
		return handler(node, parent)
	}

	if v.DefaultEnter != nil {
		return v.DefaultEnter(node, parent)
	}

	return Continue
}

// Leave implements the Visitor interface.
func (v *TypedVisitor) Leave(node types.Node, parent types.Node) {
	if v.DefaultLeave != nil {
		v.DefaultLeave(node, parent)
	}
}

// VisitorFunc is a function type that can be used as a simple visitor.
type VisitorFunc func(node types.Node, parent types.Node) VisitorAction

// Visit creates a SimpleVisitor from a function.
func Visit(fn VisitorFunc) Visitor {
	return &SimpleVisitor{
		EnterFunc: fn,
	}
}

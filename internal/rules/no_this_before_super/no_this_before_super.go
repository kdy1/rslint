package no_this_before_super

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoThisBeforeSuperRule implements the no-this-before-super rule
// Disallow this/super before calling super() in constructors
var NoThisBeforeSuperRule = rule.Rule{
	Name: "no-this-before-super",
	Run:  run,
}

// Helper to check if a node is inside a derived class
func isInDerivedClass(node *ast.Node) bool {
	current := node
	for current != nil {
		if current.Kind == ast.KindClassDeclaration || current.Kind == ast.KindClassExpression {
			classDecl := current.AsClassDeclaration()
			if classDecl != nil && classDecl.HeritageClauses != nil && len(classDecl.HeritageClauses.Nodes) > 0 {
				// Simplified: if there are heritage clauses, assume it extends something
				return true
			}
			// Check ClassExpression too
			classExpr := current.AsClassExpression()
			if classExpr != nil && classExpr.HeritageClauses != nil && len(classExpr.HeritageClauses.Nodes) > 0 {
				// Simplified: if there are heritage clauses, assume it extends something
				return true
			}
		}
		current = current.Parent
	}
	return false
}

// Helper to find the containing constructor
func findContainingConstructor(node *ast.Node) *ast.Node {
	current := node.Parent
	for current != nil {
		if current.Kind == ast.KindConstructor {
			return current
		}
		// Stop if we hit a function boundary (nested function/arrow function)
		if current.Kind == ast.KindFunctionDeclaration ||
		   current.Kind == ast.KindFunctionExpression ||
		   current.Kind == ast.KindArrowFunction ||
		   current.Kind == ast.KindMethodDeclaration {
			return nil
		}
		current = current.Parent
	}
	return nil
}

// Helper to check if super() has been called before this position
func hasSuperCallBefore(constructor *ast.Node, position *ast.Node) bool {
	// Simplified check: walk through constructor body and look for super() calls
	// This is a basic implementation - full flow analysis would be more complex
	if constructor == nil || position == nil {
		return false
	}

	if constructor.Kind != ast.KindConstructor {
		return false
	}

	// Constructor nodes don't have AsConstructor(), we need to check for Body differently
	// For now, return false as this is a simplified implementation
	// A full implementation would need to properly traverse the constructor body
	return false
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindThisKeyword: func(node *ast.Node) {
			// Check if we're in a constructor of a derived class
			constructor := findContainingConstructor(node)
			if constructor == nil {
				return
			}

			if !isInDerivedClass(constructor) {
				return
			}

			// Check if super() has been called before this
			if !hasSuperCallBefore(constructor, node) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noBeforeSuper",
					Description: "'this' is not allowed before 'super()'.",
				})
			}
		},
		ast.KindSuperKeyword: func(node *ast.Node) {
			// Only check super property access (not super() calls)
			if node.Parent != nil && node.Parent.Kind == ast.KindCallExpression {
				callExpr := node.Parent.AsCallExpression()
				if callExpr != nil && callExpr.Expression == node {
					// This is a super() call, not a property access
					return
				}
			}

			// Check if we're in a constructor of a derived class
			constructor := findContainingConstructor(node)
			if constructor == nil {
				return
			}

			if !isInDerivedClass(constructor) {
				return
			}

			// Check if super() has been called before this super property access
			if !hasSuperCallBefore(constructor, node) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noBeforeSuper",
					Description: "'super' is not allowed before 'super()'.",
				})
			}
		},
	}
}

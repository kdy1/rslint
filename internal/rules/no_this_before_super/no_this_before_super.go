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
		if current.Kind == ast.KindClassDeclaration {
			classDecl := current.AsClassDeclaration()
			if classDecl != nil && classDecl.HeritageClauses != nil && len(classDecl.HeritageClauses.Nodes) > 0 {
				// Check if it extends something other than null
				for _, heritage := range classDecl.HeritageClauses.Nodes {
					if heritageClause := heritage.AsHeritageClause(); heritageClause != nil {
						for _, typeNode := range heritageClause.Types.Nodes {
							if exprWithType := typeNode.AsExpressionWithTypeArguments(); exprWithType != nil {
								if exprWithType.Expression.Kind != ast.KindNullKeyword {
									return true
								}
							}
						}
					}
				}
			}
		} else if current.Kind == ast.KindClassExpression {
			classExpr := current.AsClassExpression()
			if classExpr != nil && classExpr.HeritageClauses != nil && len(classExpr.HeritageClauses.Nodes) > 0 {
				// Check if it extends something other than null
				for _, heritage := range classExpr.HeritageClauses.Nodes {
					if heritageClause := heritage.AsHeritageClause(); heritageClause != nil {
						for _, typeNode := range heritageClause.Types.Nodes {
							if exprWithType := typeNode.AsExpressionWithTypeArguments(); exprWithType != nil {
								if exprWithType.Expression.Kind != ast.KindNullKeyword {
									return true
								}
							}
						}
					}
				}
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

	// Get the constructor declaration to access its body
	constructorDecl := constructor.AsConstructorDeclaration()
	if constructorDecl == nil || constructorDecl.Body == nil {
		return false
	}

	body := constructorDecl.Body.AsBlock()
	if body == nil || body.Statements == nil {
		return false
	}

	// Track if we've seen a super() call
	foundSuper := false
	positionStart := position.Pos()

	// Walk through the statements in the constructor body
	for _, stmt := range body.Statements.Nodes {
		// If this statement starts after the position we're checking, stop
		if stmt.Pos() >= positionStart {
			break
		}

		// Check if this statement (or its descendants) contains a super() call
		if containsSuperCall(stmt) {
			foundSuper = true
		}
	}

	return foundSuper
}

// Helper to check if a node or its descendants contain a super() call
func containsSuperCall(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check if this is a call expression with super as the callee
	if node.Kind == ast.KindCallExpression {
		callExpr := node.AsCallExpression()
		if callExpr != nil && callExpr.Expression != nil && callExpr.Expression.Kind == ast.KindSuperKeyword {
			return true
		}
	}

	// Recursively check children
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if child != nil && containsSuperCall(child) {
			found = true
		}
		return true // Continue iteration
	})

	return found
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

package no_await_in_loop

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildUnexpectedAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedAwait",
		Description: "Unexpected `await` inside a loop.",
	}
}

// isLoopNode checks if a node is a loop statement
func isLoopNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	kind := node.Kind
	return kind == ast.KindWhileStatement ||
		kind == ast.KindDoStatement ||
		kind == ast.KindForStatement ||
		kind == ast.KindForInStatement ||
		kind == ast.KindForOfStatement
}

// isForAwaitOfNode checks if a node is a for-await-of loop
func isForAwaitOfNode(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindForOfStatement {
		return false
	}
	// Check for await modifier on ForOfStatement
	stmt := node.AsForInOrOfStatement()
	if stmt == nil {
		return false
	}
	return stmt.AwaitModifier != nil
}

// isInLoop checks if we're currently inside a loop (excluding for-await-of)
func isInLoop(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		// If we hit a for-await-of, it's allowed
		if isForAwaitOfNode(current) {
			return false
		}

		// If we hit a regular loop, we're in a problematic context
		if isLoopNode(current) {
			return true
		}

		// Stop if we hit a function boundary (functions create new async contexts)
		kind := current.Kind
		if kind == ast.KindFunctionDeclaration ||
			kind == ast.KindFunctionExpression ||
			kind == ast.KindArrowFunction ||
			kind == ast.KindMethodDeclaration ||
			kind == ast.KindConstructor ||
			kind == ast.KindGetAccessor ||
			kind == ast.KindSetAccessor {
			return false
		}

		current = current.Parent
	}
	return false
}

// NoAwaitInLoopRule disallows await inside of loops
var NoAwaitInLoopRule = rule.CreateRule(rule.Rule{
	Name: "no-await-in-loop",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindAwaitExpression: func(node *ast.Node) {
				if node == nil {
					return
				}

				// Check if this await is inside a loop
				if isInLoop(node) {
					ctx.ReportNode(node, buildUnexpectedAwaitMessage())
				}
			},

			// Handle for-await-of in loops
			ast.KindForOfStatement: func(node *ast.Node) {
				if node == nil {
					return
				}

				// Check if this is a for-await-of
				if !isForAwaitOfNode(node) {
					return
				}

				// Check if the for-await-of is nested inside another loop
				if isInLoop(node) {
					// Report the await modifier
					stmt := node.AsForInOrOfStatement()
					if stmt != nil && stmt.AwaitModifier != nil {
						ctx.ReportNode(stmt.AwaitModifier, buildUnexpectedAwaitMessage())
					}
				}
			},
		}
	},
})

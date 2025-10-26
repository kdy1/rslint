package no_unsafe_finally

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnsafeFinallyRule implements the no-unsafe-finally rule
// Disallow control flow statements in `finally` blocks
var NoUnsafeFinallyRule = rule.Rule{
	Name: "no-unsafe-finally",
	Run:  run,
}

func buildUnsafeMessage(nodeKind ast.Kind) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeUsage",
		Description: fmt.Sprintf("Unsafe usage of %s.", nodeKind.String()),
	}
}

// isFinallyBlock checks if a node is a finally block
func isFinallyBlock(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBlock {
		return false
	}
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindTryStatement {
		return false
	}
	return parent.FinallyBlock() == node
}

// isFunction checks if a node is a function-like structure
func isFunction(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindArrowFunction, ast.KindMethodDeclaration,
		ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return true
	}
	return false
}

// isLoop checks if a node is a loop structure
func isLoop(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindForStatement, ast.KindForInStatement,
		ast.KindForOfStatement, ast.KindWhileStatement, ast.KindDoStatement:
		return true
	}
	return false
}

// isInFinallyBlock checks if a control flow statement is inside a finally block
func isInFinallyBlock(node *ast.Node, stmtKind ast.Kind) bool {
	current := node.Parent

	for current != nil {
		// Check if we're in a finally block
		if isFinallyBlock(current) {
			return true
		}

		// Stop at sentinel nodes based on statement type
		switch stmtKind {
		case ast.KindReturnStatement, ast.KindThrowStatement:
			// Stop at function boundaries
			if isFunction(current) {
				return false
			}
		case ast.KindBreakStatement:
			// Stop at loops and switch statements (in addition to functions)
			if isFunction(current) || isLoop(current) || current.Kind == ast.KindSwitchStatement {
				return false
			}
		case ast.KindContinueStatement:
			// Stop at loops (in addition to functions)
			if isFunction(current) || isLoop(current) {
				return false
			}
		}

		current = current.Parent
	}

	return false
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	check := func(node *ast.Node) {
		if node != nil && isInFinallyBlock(node, node.Kind) {
			ctx.ReportNode(node, buildUnsafeMessage(node.Kind))
		}
	}

	return rule.RuleListeners{
		ast.KindReturnStatement:   check,
		ast.KindThrowStatement:    check,
		ast.KindBreakStatement:    check,
		ast.KindContinueStatement: check,
	}
}

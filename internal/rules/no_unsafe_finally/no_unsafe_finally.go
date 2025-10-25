package no_unsafe_finally

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnsafeFinallyRule implements the no-unsafe-finally rule
// Disallow control flow statements in finally blocks
var NoUnsafeFinallyRule = rule.Rule{
	Name: "no-unsafe-finally",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindTryStatement: func(node *ast.Node) {
			tryStmt := node.AsTryStatement()
			if tryStmt == nil || tryStmt.FinallyBlock == nil {
				return
			}

			// Check the finally block for unsafe control flow statements
			checkFinallyBlock(ctx, tryStmt.FinallyBlock)
		},
	}
}

// checkFinallyBlock recursively checks a finally block for unsafe control flow statements
func checkFinallyBlock(ctx rule.RuleContext, finallyBlock *ast.Node) {
	if finallyBlock == nil {
		return
	}

	// We need to traverse the finally block and find any control flow statements
	// that are not nested inside functions or classes
	traverseForUnsafeStatements(ctx, finallyBlock, 0)
}

// traverseForUnsafeStatements traverses nodes looking for unsafe control flow statements
// depth tracks how deep we are in nested functions/classes (to avoid false positives)
func traverseForUnsafeStatements(ctx rule.RuleContext, node *ast.Node, nestedFunctionDepth int) {
	if node == nil {
		return
	}

	kind := node.Kind

	// If we're inside a function or class, control flow statements are safe
	if nestedFunctionDepth > 0 {
		// Still need to traverse into nested functions
		if isFunctionOrClass(kind) {
			traverseChildren(ctx, node, nestedFunctionDepth+1)
		} else {
			traverseChildren(ctx, node, nestedFunctionDepth)
		}
		return
	}

	// Check for unsafe control flow statements at the finally block level
	switch kind {
	case ast.KindReturnStatement:
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "unsafeUsage",
			Description: "Unsafe usage of ReturnStatement in finally block.",
		})

	case ast.KindThrowStatement:
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "unsafeUsage",
			Description: "Unsafe usage of ThrowStatement in finally block.",
		})

	case ast.KindBreakStatement:
		// Check if this break is for a switch/loop inside the finally block
		if !isBreakInsideLocalSwitch(node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "unsafeUsage",
				Description: "Unsafe usage of BreakStatement in finally block.",
			})
		}

	case ast.KindContinueStatement:
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "unsafeUsage",
			Description: "Unsafe usage of ContinueStatement in finally block.",
		})

	case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindArrowFunction, ast.KindClassDeclaration:
		// Traverse into functions/classes with increased depth
		traverseChildren(ctx, node, nestedFunctionDepth+1)
		return

	default:
		// Continue traversing
		traverseChildren(ctx, node, nestedFunctionDepth)
	}
}

// isFunctionOrClass checks if a node kind represents a function or class
func isFunctionOrClass(kind ast.Kind) bool {
	return kind == ast.KindFunctionDeclaration ||
		kind == ast.KindFunctionExpression ||
		kind == ast.KindArrowFunction ||
		kind == ast.KindClassDeclaration ||
		kind == ast.KindClassExpression ||
		kind == ast.KindMethodDeclaration ||
		kind == ast.KindConstructor ||
		kind == ast.KindGetAccessor ||
		kind == ast.KindSetAccessor
}

// isBreakInsideLocalSwitch checks if a break statement is inside a switch within the finally
// This is a simplified check - a full implementation would need to track scope properly
func isBreakInsideLocalSwitch(breakNode *ast.Node) bool {
	// For now, we'll report all breaks as potentially unsafe
	// A more sophisticated implementation would track the parent chain
	return false
}

// traverseChildren recursively traverses child nodes
func traverseChildren(ctx rule.RuleContext, node *ast.Node, nestedFunctionDepth int) {
	if node == nil {
		return
	}

	kind := node.Kind

	switch kind {
	case ast.KindBlock:
		block := node.AsBlock()
		if block != nil && block.Statements != nil {
			for _, stmt := range block.Statements {
				traverseForUnsafeStatements(ctx, &stmt, nestedFunctionDepth)
			}
		}

	case ast.KindIfStatement:
		ifStmt := node.AsIfStatement()
		if ifStmt != nil {
			traverseForUnsafeStatements(ctx, &ifStmt.ThenStatement, nestedFunctionDepth)
			if ifStmt.ElseStatement != nil {
				traverseForUnsafeStatements(ctx, ifStmt.ElseStatement, nestedFunctionDepth)
			}
		}

	case ast.KindSwitchStatement:
		switchStmt := node.AsSwitchStatement()
		if switchStmt != nil && switchStmt.CaseBlock != nil {
			caseBlock := switchStmt.CaseBlock.CaseBlock()
			if caseBlock != nil && caseBlock.Clauses != nil {
				for _, clause := range *caseBlock.Clauses {
					clauseNode := clause.CaseOrDefaultClause()
					if clauseNode != nil && clauseNode.Statements != nil {
						for _, stmt := range *clauseNode.Statements {
							traverseForUnsafeStatements(ctx, &stmt, nestedFunctionDepth)
						}
					}
				}
			}
		}

	case ast.KindWhileStatement:
		whileStmt := node.AsWhileStatement()
		if whileStmt != nil {
			traverseForUnsafeStatements(ctx, &whileStmt.Statement, nestedFunctionDepth)
		}

	case ast.KindDoStatement:
		doStmt := node.AsDoStatement()
		if doStmt != nil {
			traverseForUnsafeStatements(ctx, &doStmt.Statement, nestedFunctionDepth)
		}

	case ast.KindForStatement:
		forStmt := node.AsForStatement()
		if forStmt != nil {
			traverseForUnsafeStatements(ctx, &forStmt.Statement, nestedFunctionDepth)
		}

	case ast.KindForInStatement:
		forInStmt := node.AsForInStatement()
		if forInStmt != nil {
			traverseForUnsafeStatements(ctx, &forInStmt.Statement, nestedFunctionDepth)
		}

	case ast.KindForOfStatement:
		forOfStmt := node.AsForOfStatement()
		if forOfStmt != nil {
			traverseForUnsafeStatements(ctx, &forOfStmt.Statement, nestedFunctionDepth)
		}

	case ast.KindTryStatement:
		tryStmt := node.AsTryStatement()
		if tryStmt != nil {
			if tryStmt.TryBlock != nil {
				traverseForUnsafeStatements(ctx, tryStmt.TryBlock, nestedFunctionDepth)
			}
			if tryStmt.CatchClause != nil {
				catchClause := tryStmt.CatchClause.AsCatchClause()
				if catchClause != nil && catchClause.Block != nil {
					traverseForUnsafeStatements(ctx, catchClause.Block, nestedFunctionDepth)
				}
			}
			// Note: We don't recursively check nested finally blocks here
			// as they are checked by the main listener
		}

	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		// Don't traverse into function bodies when we're at depth 0
		// as we already handle this in traverseForUnsafeStatements
		if nestedFunctionDepth > 0 {
			funcNode := getFunctionBody(node)
			if funcNode != nil {
				traverseForUnsafeStatements(ctx, funcNode, nestedFunctionDepth)
			}
		}

	case ast.KindClassDeclaration, ast.KindClassExpression:
		// Don't traverse into class members when we're at depth 0
		// Class methods are inherently nested
		return
	}
}

// getFunctionBody extracts the body of a function node
func getFunctionBody(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl != nil {
			return funcDecl.Body
		}
	case ast.KindFunctionExpression:
		funcExpr := node.AsFunctionExpression()
		if funcExpr != nil {
			return funcExpr.Body
		}
	case ast.KindArrowFunction:
		arrowFunc := node.AsArrowFunction()
		if arrowFunc != nil {
			// Arrow functions might have expression bodies
			// For now, just return nil if not a block
			if arrowFunc.Body != nil && arrowFunc.Body.Kind == ast.KindBlock {
				return arrowFunc.Body
			}
		}
	}
	return nil
}

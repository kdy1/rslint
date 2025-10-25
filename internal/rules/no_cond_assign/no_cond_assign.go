package no_cond_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Options for no-cond-assign rule
type Options struct {
	Mode string `json:"mode"` // "except-parens" (default) or "always"
}

func parseOptions(options any) Options {
	opts := Options{
		Mode: "except-parens", // default
	}

	if options == nil {
		return opts
	}

	// Handle string option (ESLint compatibility)
	if mode, ok := options.(string); ok {
		opts.Mode = mode
		return opts
	}

	// Handle array format: ["except-parens"] or ["always"]
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		if mode, ok := optArray[0].(string); ok {
			opts.Mode = mode
			return opts
		}
		// Could also be [{mode: "except-parens"}]
		if optsMap, ok := optArray[0].(map[string]interface{}); ok {
			if mode, ok := optsMap["mode"].(string); ok {
				opts.Mode = mode
			}
		}
	}

	// Handle direct object format: {mode: "except-parens"}
	if optsMap, ok := options.(map[string]interface{}); ok {
		if mode, ok := optsMap["mode"].(string); ok {
			opts.Mode = mode
		}
	}

	return opts
}

func buildMissingMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missing",
		Description: "Expected a conditional expression and instead saw an assignment.",
	}
}

func buildUnexpectedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected assignment in conditional expression.",
	}
}

// NoCondAssignRule disallows assignment operators in conditional expressions
var NoCondAssignRule = rule.Rule{
	Name: "no-cond-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		listeners := rule.RuleListeners{}

		// Helper to check if a node is an assignment
		isAssignment := func(node *ast.Node) bool {
			if node == nil {
				return false
			}

			if node.Kind == ast.KindBinaryExpression {
				binary := node.AsBinaryExpression()
				if binary == nil {
					return false
				}

				op := binary.OperatorToken.Kind
				return op == ast.KindEqualsToken ||
					op == ast.KindPlusEqualsToken ||
					op == ast.KindMinusEqualsToken ||
					op == ast.KindAsteriskEqualsToken ||
					op == ast.KindSlashEqualsToken ||
					op == ast.KindPercentEqualsToken ||
					op == ast.KindAmpersandEqualsToken ||
					op == ast.KindBarEqualsToken ||
					op == ast.KindCaretEqualsToken ||
					op == ast.KindLessThanLessThanEqualsToken ||
					op == ast.KindGreaterThanGreaterThanEqualsToken ||
					op == ast.KindGreaterThanGreaterThanGreaterThanEqualsToken ||
					op == ast.KindAsteriskAsteriskEqualsToken
			}

			return false
		}

		// Helper to check if assignment is wrapped in extra parentheses
		isParenthesized := func(node *ast.Node) bool {
			if node == nil {
				return false
			}

			parent := node.Parent
			if parent == nil {
				return false
			}

			// Check if parent is a ParenthesizedExpression
			if parent.Kind == ast.KindParenthesizedExpression {
				// Check if the grandparent is also a ParenthesizedExpression (double parens)
				// or if we're directly in a conditional
				grandparent := parent.Parent
				if grandparent == nil {
					return false
				}

				// In except-parens mode, we need double parentheses: ((a = b))
				// Single parens: (a = b) - still an error
				// The parent being a paren means we have one level
				// We need to check if grandparent is also a paren
				if grandparent.Kind == ast.KindParenthesizedExpression {
					return true
				}
			}

			return false
		}

		// Helper to check test condition and report
		checkTestCondition := func(testNode *ast.Node, allowParens bool) {
			if testNode == nil {
				return
			}

			test := testNode

			// Recursively check for assignments
			var checkNode func(*ast.Node)
			checkNode = func(node *ast.Node) {
				if node == nil {
					return
				}

				if isAssignment(node) {
					// In except-parens mode, allow if wrapped in extra parentheses
					if allowParens && isParenthesized(node) {
						return
					}

					// Report the assignment
					if allowParens {
						ctx.ReportNode(node, buildMissingMessage())
					} else {
						ctx.ReportNode(node, buildUnexpectedMessage())
					}
					return
				}

				// Don't recurse into nested functions/arrows
				kind := node.Kind
				if kind == ast.KindFunctionExpression ||
					kind == ast.KindArrowFunction ||
					kind == ast.KindFunctionDeclaration {
					return
				}

				// For binary expressions, check both sides
				if kind == ast.KindBinaryExpression {
					binary := node.AsBinaryExpression()
					if binary != nil {
						checkNode(binary.Left)
						checkNode(binary.Right)
					}
				}

				// For parenthesized expressions, check the inner expression
				if kind == ast.KindParenthesizedExpression {
					paren := node.AsParenthesizedExpression()
					if paren != nil {
						checkNode(paren.Expression)
					}
				}

				// For logical expressions (&&, ||), check both sides
				// These are also BinaryExpression in TypeScript AST
			}

			checkNode(test)
		}

		allowParens := opts.Mode == "except-parens"

		// Listen to IfStatement
		listeners[ast.KindIfStatement] = func(node *ast.Node) {
			ifStmt := node.AsIfStatement()
			if ifStmt == nil || ifStmt.Expression == nil {
				return
			}

			checkTestCondition(ifStmt.Expression, allowParens)
		}

		// Listen to WhileStatement
		listeners[ast.KindWhileStatement] = func(node *ast.Node) {
			whileStmt := node.AsWhileStatement()
			if whileStmt == nil || whileStmt.Expression == nil {
				return
			}

			checkTestCondition(whileStmt.Expression, allowParens)
		}

		// Listen to DoStatement
		listeners[ast.KindDoStatement] = func(node *ast.Node) {
			doStmt := node.AsDoStatement()
			if doStmt == nil || doStmt.Expression == nil {
				return
			}

			checkTestCondition(doStmt.Expression, allowParens)
		}

		// Listen to ForStatement
		listeners[ast.KindForStatement] = func(node *ast.Node) {
			forStmt := node.AsForStatement()
			if forStmt == nil || forStmt.Condition == nil {
				return
			}

			checkTestCondition(forStmt.Condition, allowParens)
		}

		// Listen to ConditionalExpression (ternary)
		listeners[ast.KindConditionalExpression] = func(node *ast.Node) {
			condExpr := node.AsConditionalExpression()
			if condExpr == nil || condExpr.Condition == nil {
				return
			}

			checkTestCondition(condExpr.Condition, allowParens)
		}

		return listeners
	},
}

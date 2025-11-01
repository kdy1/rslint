package no_cond_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builders
func buildMissingMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missing",
		Description: "Expected a conditional expression and instead saw an assignment.",
	}
}

func buildUnexpectedMessage(nodeType string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected assignment within " + nodeType + ".",
	}
}

// isAssignmentExpression checks if a node is an assignment expression
func isAssignmentExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindBinaryExpression && isAssignmentOperator(node)
}

// isAssignmentOperator checks if a binary expression uses an assignment operator
func isAssignmentOperator(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}

	binary := node.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return false
	}

	// Check for all assignment operators
	switch binary.OperatorToken.Kind {
	case ast.KindEqualsToken, // =
		ast.KindPlusEqualsToken,              // +=
		ast.KindMinusEqualsToken,             // -=
		ast.KindAsteriskEqualsToken,          // *=
		ast.KindSlashEqualsToken,             // /=
		ast.KindPercentEqualsToken,           // %=
		ast.KindAsteriskAsteriskEqualsToken,  // **=
		ast.KindLessThanLessThanEqualsToken,  // <<=
		ast.KindGreaterThanGreaterThanEqualsToken,        // >>=
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken, // >>>=
		ast.KindAmpersandEqualsToken,         // &=
		ast.KindBarEqualsToken,               // |=
		ast.KindCaretEqualsToken:             // ^=
		return true
	}
	return false
}

// isConditionalTestExpression checks if a node is a test expression in a conditional statement
func isConditionalTestExpression(node, parent *ast.Node) bool {
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindIfStatement:
		ifStmt := parent.AsIfStatement()
		return ifStmt != nil && ifStmt.Expression != nil && ifStmt.Expression == node

	case ast.KindWhileStatement:
		whileStmt := parent.AsWhileStatement()
		return whileStmt != nil && whileStmt.Expression != nil && whileStmt.Expression == node

	case ast.KindDoStatement:
		doStmt := parent.AsDoStatement()
		return doStmt != nil && doStmt.Expression != nil && doStmt.Expression == node

	case ast.KindForStatement:
		forStmt := parent.AsForStatement()
		return forStmt != nil && forStmt.Condition != nil && forStmt.Condition == node

	case ast.KindConditionalExpression:
		condExpr := parent.AsConditionalExpression()
		return condExpr != nil && condExpr.Condition != nil && condExpr.Condition == node
	}

	return false
}

// getTestExpression returns the test expression for a conditional statement
func getTestExpression(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindIfStatement:
		ifStmt := node.AsIfStatement()
		if ifStmt != nil {
			return ifStmt.Expression
		}
	case ast.KindWhileStatement:
		whileStmt := node.AsWhileStatement()
		if whileStmt != nil {
			return whileStmt.Expression
		}
	case ast.KindDoStatement:
		doStmt := node.AsDoStatement()
		if doStmt != nil {
			return doStmt.Expression
		}
	case ast.KindForStatement:
		forStmt := node.AsForStatement()
		if forStmt != nil {
			return forStmt.Condition
		}
	case ast.KindConditionalExpression:
		// For ternary, return the entire expression
		return node
	}
	return nil
}

// getConditionalTypeName returns a human-readable name for the conditional statement type
func getConditionalTypeName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIfStatement:
		return "an 'if' statement"
	case ast.KindWhileStatement:
		return "a 'while' statement"
	case ast.KindDoStatement:
		return "a 'do...while' statement"
	case ast.KindForStatement:
		return "a 'for' statement"
	case ast.KindConditionalExpression:
		return "a conditional expression"
	}
	return ""
}

// isParenthesized checks if a node is wrapped in parentheses
// This is a simplified check - in a real implementation, we'd need to check the source tokens
func isParenthesized(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check if the node is wrapped in a ParenthesizedExpression
	return node.Kind == ast.KindParenthesizedExpression
}

// countParentheses counts the number of parentheses wrapping a node
func countParentheses(node, parent *ast.Node) int {
	count := 0
	current := parent

	// Walk up the tree counting ParenthesizedExpression nodes
	for current != nil && current.Kind == ast.KindParenthesizedExpression {
		count++
		// In a real implementation, we'd traverse up the AST
		// For now, we return the count we have
		break
	}

	return count
}

// NoCondAssignRule disallows assignment operators in conditional expressions
var NoCondAssignRule = rule.CreateRule(rule.Rule{
	Name: "no-cond-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Parse options - default is "except-parens"
		mode := "except-parens"
		if options != nil {
			if optMap, ok := options.(map[string]interface{}); ok {
				if modeStr, ok := optMap["mode"].(string); ok {
					mode = modeStr
				}
			} else if optStr, ok := options.(string); ok {
				mode = optStr
			}
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				// Check if this is an assignment expression
				if !isAssignmentExpression(node) {
					return
				}

				// Walk up to find if we're directly in a conditional test (not nested in other expressions)
				// Skip over ParenthesizedExpressions as they don't change the semantics
				var testExpressionRoot *ast.Node
				var conditionalAncestor *ast.Node
				current := node.Parent

				// Skip parentheses to find the real parent expression
				for current != nil && current.Kind == ast.KindParenthesizedExpression {
					current = current.Parent
				}

				// Now check if we're directly in a conditional or in a larger expression
				for current != nil {
					if current.Kind == ast.KindIfStatement ||
						current.Kind == ast.KindWhileStatement ||
						current.Kind == ast.KindDoStatement ||
						current.Kind == ast.KindForStatement ||
						current.Kind == ast.KindConditionalExpression {
						conditionalAncestor = current
						break
					}
					// Stop at function boundaries
					if current.Kind == ast.KindFunctionDeclaration ||
						current.Kind == ast.KindFunctionExpression ||
						current.Kind == ast.KindArrowFunction ||
						current.Kind == ast.KindMethodDeclaration {
						break
					}
					//  Remember this as a potential test expression root
					testExpressionRoot = current
					current = current.Parent
				}

				if conditionalAncestor == nil {
					return
				}

				// Get the actual test expression
				testExpr := getTestExpression(conditionalAncestor)
				if testExpr == nil {
					return
				}

				// Check if the assignment is in the test part of the conditional
				if !containsNode(testExpr, node) {
					return
				}

				// Apply the rule based on mode
				if mode == "always" {
					// Always report assignments in conditionals
					ctx.ReportNode(node, buildUnexpectedMessage(getConditionalTypeName(conditionalAncestor)))
				} else if mode == "except-parens" {
					// Check if the assignment is properly parenthesized
					// ESLint logic: assignments need to be "doubly parenthesized" except in for loops
					// where single parentheses suffice

					// Count consecutive ParenthesizedExpression nodes wrapping the assignment
					parenLevels := 0
					current := node.Parent

					// Count consecutive parenthesized expressions
					for current != nil && current.Kind == ast.KindParenthesizedExpression {
						parenLevels++
						current = current.Parent
					}

					var isProperlyParenthesized bool
					if conditionalAncestor.Kind == ast.KindForStatement {
						// For loops only require single parentheses
						isProperlyParenthesized = parenLevels >= 1
					} else {
						// Other statements (if, while, do-while, ternary) require double parentheses
						isProperlyParenthesized = parenLevels >= 2
					}

					if !isProperlyParenthesized {
						ctx.ReportNode(node, buildMissingMessage())
					}
				}
			},
		}
	},
})

// containsNode checks if a root node contains a target node in its subtree
func containsNode(root, target *ast.Node) bool {
	if root == nil || target == nil {
		return false
	}
	if root == target {
		return true
	}

	// Walk up from target to see if we reach root
	current := target.Parent
	for current != nil {
		if current == root {
			return true
		}
		current = current.Parent
	}

	return false
}

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

		// Track parent nodes to determine context
		parentStack := make([]*ast.Node, 0)

		return rule.RuleListeners{
			// Track parent nodes
			"*": func(node *ast.Node) {
				parentStack = append(parentStack, node)
			},
			"*:exit": func(node *ast.Node) {
				if len(parentStack) > 0 {
					parentStack = parentStack[:len(parentStack)-1]
				}
			},

			ast.KindBinaryExpression: func(node *ast.Node) {
				// Check if this is an assignment expression
				if !isAssignmentExpression(node) {
					return
				}

				// Get parent node
				var parent *ast.Node
				if len(parentStack) > 0 {
					parent = parentStack[len(parentStack)-1]
				}

				// Check if we're in a conditional test expression
				var conditionalAncestor *ast.Node
				for i := len(parentStack) - 1; i >= 0; i-- {
					p := parentStack[i]
					if p.Kind == ast.KindIfStatement ||
						p.Kind == ast.KindWhileStatement ||
						p.Kind == ast.KindDoStatement ||
						p.Kind == ast.KindForStatement ||
						p.Kind == ast.KindConditionalExpression {
						conditionalAncestor = p
						break
					}
					// Stop at function boundaries
					if p.Kind == ast.KindFunctionDeclaration ||
						p.Kind == ast.KindFunctionExpression ||
						p.Kind == ast.KindArrowFunction ||
						p.Kind == ast.KindMethodDeclaration {
						break
					}
				}

				if conditionalAncestor == nil {
					return
				}

				// Check if the assignment is in the test part of the conditional
				var inTestExpression bool
				switch conditionalAncestor.Kind {
				case ast.KindIfStatement:
					ifStmt := conditionalAncestor.AsIfStatement()
					inTestExpression = ifStmt != nil && containsNode(ifStmt.Expression, node)

				case ast.KindWhileStatement:
					whileStmt := conditionalAncestor.AsWhileStatement()
					inTestExpression = whileStmt != nil && containsNode(whileStmt.Expression, node)

				case ast.KindDoStatement:
					doStmt := conditionalAncestor.AsDoStatement()
					inTestExpression = doStmt != nil && containsNode(doStmt.Expression, node)

				case ast.KindForStatement:
					forStmt := conditionalAncestor.AsForStatement()
					inTestExpression = forStmt != nil && containsNode(forStmt.Condition, node)

				case ast.KindConditionalExpression:
					condExpr := conditionalAncestor.AsConditionalExpression()
					inTestExpression = condExpr != nil && containsNode(condExpr.Condition, node)
				}

				if !inTestExpression {
					return
				}

				// Apply the rule based on mode
				if mode == "always" {
					// Always report assignments in conditionals
					ctx.ReportNode(node, buildUnexpectedMessage(getConditionalTypeName(conditionalAncestor)))
				} else if mode == "except-parens" {
					// Check if the assignment is properly parenthesized
					isProperlyParenthesized := false

					// For ternary expressions, we need double parentheses
					if conditionalAncestor.Kind == ast.KindConditionalExpression {
						// Check if wrapped in at least 2 levels of parentheses
						parenCount := 0
						if parent != nil && parent.Kind == ast.KindParenthesizedExpression {
							parenCount++
							// Check grandparent
							if len(parentStack) >= 2 {
								grandparent := parentStack[len(parentStack)-2]
								if grandparent != nil && grandparent.Kind == ast.KindParenthesizedExpression {
									parenCount++
								}
							}
						}
						isProperlyParenthesized = parenCount >= 2
					} else {
						// For other statements, single parentheses suffice
						isProperlyParenthesized = parent != nil && parent.Kind == ast.KindParenthesizedExpression
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

	// Simple check - in a real implementation, we'd traverse the full AST
	// For now, we assume if target's position is within root's range, it's contained
	if root.Pos <= target.Pos && target.End <= root.End {
		return true
	}

	return false
}

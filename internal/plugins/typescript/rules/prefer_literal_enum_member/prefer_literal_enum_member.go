package prefer_literal_enum_member

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type PreferLiteralEnumMemberOptions struct {
	AllowBitwiseExpressions bool `json:"allowBitwiseExpressions"`
}

// isBitwiseOperator checks if the operator is a bitwise operator
func isBitwiseOperator(operator ast.SyntaxKind) bool {
	switch operator {
	case ast.KindBarToken,           // |
		ast.KindAmpersandToken,      // &
		ast.KindCaretToken,          // ^
		ast.KindLessThanLessThanToken,     // <<
		ast.KindGreaterThanGreaterThanToken,     // >>
		ast.KindGreaterThanGreaterThanGreaterThanToken: // >>>
		return true
	}
	return false
}

// isUnaryBitwiseOperator checks if the operator is a unary bitwise operator (~)
func isUnaryBitwiseOperator(operator ast.SyntaxKind) bool {
	return operator == ast.KindTildeToken
}

// isLiteralExpression checks if a node is a literal expression
func isLiteralExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindNumericLiteral,
		ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral:
		return true
	}
	return false
}

// isAllowedBitwiseExpression checks if a node is an allowed bitwise expression
func isAllowedBitwiseExpression(node *ast.Node, allowBitwiseExpressions bool) bool {
	if !allowBitwiseExpressions {
		return false
	}

	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindBinaryExpression:
		binaryExpr := node.AsBinaryExpression()
		if binaryExpr == nil {
			return false
		}

		// Check if operator is bitwise
		if !isBitwiseOperator(binaryExpr.OperatorToken.Kind) {
			return false
		}

		// Both operands must be literals, enum references, or bitwise expressions
		return isAllowedBitwiseOperand(binaryExpr.Left, allowBitwiseExpressions) &&
			isAllowedBitwiseOperand(binaryExpr.Right, allowBitwiseExpressions)

	case ast.KindPrefixUnaryExpression:
		unaryExpr := node.AsPrefixUnaryExpression()
		if unaryExpr == nil {
			return false
		}

		// Allow unary ~ and - for bitwise expressions
		if unaryExpr.Operator == ast.KindTildeToken {
			return isAllowedBitwiseOperand(unaryExpr.Operand, allowBitwiseExpressions)
		}

		// Allow unary - for negative numbers in bitwise expressions
		if unaryExpr.Operator == ast.KindMinusToken {
			return isAllowedBitwiseOperand(unaryExpr.Operand, allowBitwiseExpressions)
		}

	case ast.KindParenthesizedExpression:
		parenExpr := node.AsParenthesizedExpression()
		if parenExpr != nil {
			return isAllowedBitwiseExpression(parenExpr.Expression, allowBitwiseExpressions)
		}
	}

	return false
}

// isAllowedBitwiseOperand checks if a node can be an operand in a bitwise expression
func isAllowedBitwiseOperand(node *ast.Node, allowBitwiseExpressions bool) bool {
	if node == nil {
		return false
	}

	// Literals are always allowed
	if isLiteralExpression(node) {
		return true
	}

	// Unary minus on literals is allowed
	if node.Kind == ast.KindPrefixUnaryExpression {
		unaryExpr := node.AsPrefixUnaryExpression()
		if unaryExpr != nil && unaryExpr.Operator == ast.KindMinusToken {
			if isLiteralExpression(unaryExpr.Operand) {
				return true
			}
		}
	}

	// Unary plus on literals is allowed
	if node.Kind == ast.KindPrefixUnaryExpression {
		unaryExpr := node.AsPrefixUnaryExpression()
		if unaryExpr != nil && unaryExpr.Operator == ast.KindPlusToken {
			if isLiteralExpression(unaryExpr.Operand) {
				return true
			}
		}
	}

	// References to other enum members are allowed
	if node.Kind == ast.KindIdentifier {
		return true
	}

	// Property access to enum members (Foo.Bar or Foo['Bar'])
	if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
		return true
	}

	// Bitwise expressions are allowed recursively
	if isAllowedBitwiseExpression(node, allowBitwiseExpressions) {
		return true
	}

	// Parenthesized expressions
	if node.Kind == ast.KindParenthesizedExpression {
		parenExpr := node.AsParenthesizedExpression()
		if parenExpr != nil {
			return isAllowedBitwiseOperand(parenExpr.Expression, allowBitwiseExpressions)
		}
	}

	return false
}

// isValidEnumMemberInitializer checks if the enum member initializer is valid
func isValidEnumMemberInitializer(node *ast.Node, allowBitwiseExpressions bool) bool {
	if node == nil {
		return true // No initializer is valid (auto-increment)
	}

	// Literals are allowed
	if isLiteralExpression(node) {
		return true
	}

	// Unary + or - on literals is allowed
	if node.Kind == ast.KindPrefixUnaryExpression {
		unaryExpr := node.AsPrefixUnaryExpression()
		if unaryExpr != nil {
			if unaryExpr.Operator == ast.KindPlusToken || unaryExpr.Operator == ast.KindMinusToken {
				return isLiteralExpression(unaryExpr.Operand)
			}
		}
	}

	// Bitwise expressions if allowed
	if allowBitwiseExpressions && isAllowedBitwiseExpression(node, allowBitwiseExpressions) {
		return true
	}

	return false
}

var PreferLiteralEnumMemberRule = rule.CreateRule(rule.Rule{
	Name: "prefer-literal-enum-member",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := PreferLiteralEnumMemberOptions{
			AllowBitwiseExpressions: false,
		}

		// Parse options with dual-format support (handles both array and object formats)
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			// Handle array format: [{ option: value }]
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				// Handle direct object format: { option: value }
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if allowBitwiseExpressions, ok := optsMap["allowBitwiseExpressions"].(bool); ok {
					opts.AllowBitwiseExpressions = allowBitwiseExpressions
				}
			}
		}

		return rule.RuleListeners{
			ast.KindEnumDeclaration: func(node *ast.Node) {
				enumDecl := node.AsEnumDeclaration()
				if enumDecl == nil || enumDecl.Members == nil {
					return
				}

				for _, memberNode := range enumDecl.Members.Nodes {
					member := memberNode.AsEnumMember()
					if member == nil {
						continue
					}

					// No initializer is valid (auto-increment)
					if member.Initializer == nil {
						continue
					}

					// Check if the initializer is valid
					if !isValidEnumMemberInitializer(member.Initializer, opts.AllowBitwiseExpressions) {
						messageId := "notLiteral"
						description := "Enum member values should be literal values."

						// Use a more specific message if bitwise expressions are allowed
						if opts.AllowBitwiseExpressions {
							// Check if it's a bitwise expression with non-literal operands
							if member.Initializer.Kind == ast.KindBinaryExpression {
								binaryExpr := member.Initializer.AsBinaryExpression()
								if binaryExpr != nil && isBitwiseOperator(binaryExpr.OperatorToken.Kind) {
									messageId = "notLiteralOrBitwiseExpression"
									description = "Enum member values should be literal values or bitwise expressions with literal operands."
								}
							} else if member.Initializer.Kind == ast.KindPrefixUnaryExpression {
								unaryExpr := member.Initializer.AsPrefixUnaryExpression()
								if unaryExpr != nil && isUnaryBitwiseOperator(unaryExpr.Operator) {
									messageId = "notLiteralOrBitwiseExpression"
									description = "Enum member values should be literal values or bitwise expressions with literal operands."
								}
							}
						}

						ctx.ReportNode(memberNode, rule.RuleMessage{
							Id:          messageId,
							Description: description,
						})
					}
				}
			},
		}
	},
})

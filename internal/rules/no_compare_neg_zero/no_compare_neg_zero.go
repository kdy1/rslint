package no_compare_neg_zero

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoCompareNegZeroRule implements the no-compare-neg-zero rule
// Disallow comparing against `-0`
var NoCompareNegZeroRule = rule.Rule{
	Name: "no-compare-neg-zero",
	Run:  run,
}

// isNegativeZero checks if a node represents -0
func isNegativeZero(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check if it's a prefix unary expression with minus operator
	if node.Kind == ast.KindPrefixUnaryExpression {
		prefixUnary := node.AsPrefixUnaryExpression()
		if prefixUnary == nil || prefixUnary.Operator != ast.KindMinusToken {
			return false
		}

		// Check if the operand is the literal 0
		if prefixUnary.Operand == nil {
			return false
		}

		operand := prefixUnary.Operand
		if operand.Kind == ast.KindNumericLiteral {
			numLit := operand.AsNumericLiteral()
			if numLit != nil && numLit.Text != nil && *numLit.Text == "0" {
				return true
			}
		}
	}

	return false
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindBinaryExpression: func(node *ast.Node) {
			if node == nil {
				return
			}

			binExpr := node.AsBinaryExpression()
			if binExpr == nil || binExpr.OperatorToken == nil {
				return
			}

			operatorKind := binExpr.OperatorToken.Kind

			// Check if the operator is one of the comparison operators
			isComparisonOperator := operatorKind == ast.KindEqualsEqualsToken ||
				operatorKind == ast.KindEqualsEqualsEqualsToken ||
				operatorKind == ast.KindExclamationEqualsToken ||
				operatorKind == ast.KindExclamationEqualsEqualsToken ||
				operatorKind == ast.KindGreaterThanToken ||
				operatorKind == ast.KindGreaterThanEqualsToken ||
				operatorKind == ast.KindLessThanToken ||
				operatorKind == ast.KindLessThanEqualsToken

			if !isComparisonOperator {
				return
			}

			// Check if either side is -0
			leftIsNegZero := isNegativeZero(binExpr.Left)
			rightIsNegZero := isNegativeZero(binExpr.Right)

			if leftIsNegZero || rightIsNegZero {
				// Get operator text
				operatorText := utils.GetNodeText(ctx.SourceFile, binExpr.OperatorToken)

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpected",
					Description: "Do not use the '" + operatorText + "' operator to compare against -0.",
				})
			}
		},
	}
}

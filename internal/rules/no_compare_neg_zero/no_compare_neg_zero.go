package no_compare_neg_zero

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildUnexpectedMessage(operator string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Do not use the '" + operator + "' operator to compare against -0.",
	}
}

// NoCompareNegZeroRule disallows comparing against -0
var NoCompareNegZeroRule = rule.Rule{
	Name: "no-compare-neg-zero",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		listeners := rule.RuleListeners{}

		// Helper function to check if a node represents -0
		isNegZero := func(node *ast.Node) bool {
			if node == nil {
				return false
			}

			// Check for PrefixUnaryExpression with minus operator and 0 operand
			if node.GetKind() == ast.KindPrefixUnaryExpression {
				prefix := node.AsPrefixUnaryExpression()
				if prefix == nil {
					return false
				}

				// Check if operator is minus
				if prefix.Operator != ast.SyntaxKindMinusToken {
					return false
				}

				// Check if operand is numeric literal 0
				operand := ast.FromNode(prefix.Operand)
				if operand.GetKind() == ast.KindNumericLiteral {
					numLit := operand.AsNumericLiteral()
					if numLit != nil && numLit.Text() == "0" {
						return true
					}
				}
			}

			return false
		}

		// Listen to BinaryExpression nodes (comparisons)
		listeners[ast.KindBinaryExpression] = func(node *ast.Node) {
			binary := node.AsBinaryExpression()
			if binary == nil {
				return
			}

			// Check for comparison operators
			op := binary.OperatorToken.GetKind()
			isComparison := false
			operatorText := ""

			switch op {
			case ast.SyntaxKindEqualsEqualsToken:
				isComparison = true
				operatorText = "=="
			case ast.SyntaxKindEqualsEqualsEqualsToken:
				isComparison = true
				operatorText = "==="
			case ast.SyntaxKindLessThanToken:
				isComparison = true
				operatorText = "<"
			case ast.SyntaxKindLessThanEqualsToken:
				isComparison = true
				operatorText = "<="
			case ast.SyntaxKindGreaterThanToken:
				isComparison = true
				operatorText = ">"
			case ast.SyntaxKindGreaterThanEqualsToken:
				isComparison = true
				operatorText = ">="
			}

			if !isComparison {
				return
			}

			// Check if either side is -0
			left := ast.FromNode(binary.Left)
			right := ast.FromNode(binary.Right)

			if isNegZero(left) || isNegZero(right) {
				ctx.ReportNode(node, buildUnexpectedMessage(operatorText))
			}
		}

		return listeners
	},
}

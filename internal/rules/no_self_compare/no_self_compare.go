package no_self_compare

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildComparingToSelfMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "comparingToSelf",
		Description: "Comparing to itself is potentially pointless.",
	}
}

// nodesAreEqual checks if two nodes are structurally equal
func nodesAreEqual(srcFile *ast.SourceFile, left, right *ast.Node) bool {
	if left == nil || right == nil {
		return false
	}

	if left.Kind != right.Kind {
		return false
	}

	switch left.Kind {
	case ast.KindIdentifier:
		leftIdent := left.AsIdentifier()
		rightIdent := right.AsIdentifier()
		if leftIdent == nil || rightIdent == nil {
			return false
		}
		return leftIdent.Text == rightIdent.Text

	case ast.KindPropertyAccessExpression:
		leftProp := left.AsPropertyAccessExpression()
		rightProp := right.AsPropertyAccessExpression()
		if leftProp == nil || rightProp == nil {
			return false
		}

		// Check object part
		if !nodesAreEqual(srcFile, leftProp.Expression, rightProp.Expression) {
			return false
		}

		// Check property name
		if leftProp.Name() == nil || rightProp.Name() == nil {
			return false
		}
		return leftProp.Name().Text() == rightProp.Name().Text()

	case ast.KindElementAccessExpression:
		leftElem := left.AsElementAccessExpression()
		rightElem := right.AsElementAccessExpression()
		if leftElem == nil || rightElem == nil {
			return false
		}

		// Check object part
		if !nodesAreEqual(srcFile, leftElem.Expression, rightElem.Expression) {
			return false
		}

		// Check argument (index/key)
		return nodesAreEqual(srcFile, leftElem.ArgumentExpression, rightElem.ArgumentExpression)

	case ast.KindCallExpression:
		leftCall := left.AsCallExpression()
		rightCall := right.AsCallExpression()
		if leftCall == nil || rightCall == nil {
			return false
		}

		// Check function expression
		if !nodesAreEqual(srcFile, leftCall.Expression, rightCall.Expression) {
			return false
		}

		// Check arguments
		leftArgs := leftCall.Arguments
		rightArgs := rightCall.Arguments
		if leftArgs == nil || rightArgs == nil {
			return leftArgs == rightArgs
		}

		leftArgNodes := leftArgs.Nodes
		rightArgNodes := rightArgs.Nodes

		if len(leftArgNodes) != len(rightArgNodes) {
			return false
		}

		for i := range leftArgNodes {
			if !nodesAreEqual(srcFile, leftArgNodes[i], rightArgNodes[i]) {
				return false
			}
		}
		return true

	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral:
		// Compare literal values by text
		leftRange := utils.TrimNodeTextRange(srcFile, left)
		rightRange := utils.TrimNodeTextRange(srcFile, right)
		leftText := srcFile.Text()[leftRange.Pos():leftRange.End()]
		rightText := srcFile.Text()[rightRange.Pos():rightRange.End()]
		return leftText == rightText

	case ast.KindPrivateIdentifier:
		leftRange := utils.TrimNodeTextRange(srcFile, left)
		rightRange := utils.TrimNodeTextRange(srcFile, right)
		leftText := srcFile.Text()[leftRange.Pos():leftRange.End()]
		rightText := srcFile.Text()[rightRange.Pos():rightRange.End()]
		return leftText == rightText

	case ast.KindParenthesizedExpression:
		leftParen := left.AsParenthesizedExpression()
		rightParen := right.AsParenthesizedExpression()
		if leftParen == nil || rightParen == nil {
			return false
		}
		return nodesAreEqual(srcFile, leftParen.Expression, rightParen.Expression)

	case ast.KindThisKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		// Keywords are equal if they're the same kind
		return true
	}

	return false
}

// NoSelfCompareRule checks for comparisons where both sides are exactly the same
var NoSelfCompareRule = rule.CreateRule(rule.Rule{
	Name: "no-self-compare",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		listeners := rule.RuleListeners{}

		listeners[ast.KindBinaryExpression] = func(node *ast.Node) {
			binExpr := node.AsBinaryExpression()
			if binExpr == nil {
				return
			}

			op := binExpr.OperatorToken
			if op == nil {
				return
			}

			// Check if it's a comparison operator
			switch op.Kind {
			case ast.KindEqualsEqualsToken,           // ==
				ast.KindExclamationEqualsToken,        // !=
				ast.KindEqualsEqualsEqualsToken,       // ===
				ast.KindExclamationEqualsEqualsToken,  // !==
				ast.KindLessThanToken,                 // <
				ast.KindLessThanEqualsToken,           // <=
				ast.KindGreaterThanToken,              // >
				ast.KindGreaterThanEqualsToken:        // >=
				// This is a comparison operator, continue
			default:
				return
			}

			left := binExpr.Left
			right := binExpr.Right

			if left == nil || right == nil {
				return
			}

			// Check if both sides are equal
			if nodesAreEqual(ctx.SourceFile, left, right) {
				ctx.ReportNode(node, buildComparingToSelfMessage())
			}
		}

		return listeners
	},
})

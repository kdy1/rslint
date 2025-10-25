package no_self_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Options mirrors ESLint's no-self-assign options
type Options struct {
	Props bool `json:"props"`
}

func parseOptions(options any) Options {
	opts := Options{
		Props: true, // Default to true to check property assignments
	}

	if options == nil {
		return opts
	}

	// Parse options with dual-format support (handles both array and object formats)
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
		if v, ok := optsMap["props"].(bool); ok {
			opts.Props = v
		}
	}
	return opts
}

func buildSelfAssignmentMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "selfAssignment",
		Description: "'" + name + "' is assigned to itself.",
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

	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		// For property access like a.b or a['b'], check both object and property
		if left.Kind == ast.KindPropertyAccessExpression {
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
		}

		if left.Kind == ast.KindElementAccessExpression {
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
		}

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
	}

	return false
}

// getDisplayName returns a human-readable name for a node
func getDisplayName(srcFile *ast.SourceFile, node *ast.Node) string {
	if node == nil {
		return ""
	}

	nodeRange := utils.TrimNodeTextRange(srcFile, node)
	return srcFile.Text()[nodeRange.Pos():nodeRange.End()]
}

// checkAssignmentExpression checks if an assignment expression is self-assigning
func checkAssignmentExpression(ctx rule.RuleContext, node *ast.Node, opts Options) {
	binExpr := node.AsBinaryExpression()
	if binExpr == nil {
		return
	}

	left := binExpr.Left
	right := binExpr.Right

	if left == nil || right == nil {
		return
	}

	// Only check = assignments, not +=, -=, etc.
	// For logical assignment operators (&&=, ||=, ??=) we also check
	op := binExpr.OperatorToken
	if op == nil {
		return
	}

	// Check for simple assignment (=) or logical assignment operators (&&=, ||=, ??=)
	switch op.Kind {
	case ast.KindEqualsToken,
		ast.KindAmpersandAmpersandEqualsToken,
		ast.KindBarBarEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		// These are the operators we check
	default:
		return
	}

	// Simple identifier check (without props option)
	if left.Kind == ast.KindIdentifier && right.Kind == ast.KindIdentifier {
		if nodesAreEqual(ctx.SourceFile, left, right) {
			ctx.ReportNode(node, buildSelfAssignmentMessage(getDisplayName(ctx.SourceFile, left)))
			return
		}
	}

	// Property access check (only if opts.Props is true)
	if opts.Props {
		if nodesAreEqual(ctx.SourceFile, left, right) {
			ctx.ReportNode(node, buildSelfAssignmentMessage(getDisplayName(ctx.SourceFile, left)))
		}
	}
}

// checkArrayPattern checks array destructuring patterns
func checkArrayPattern(ctx rule.RuleContext, left *ast.Node, right *ast.Node, opts Options) {
	if left == nil || right == nil {
		return
	}

	if left.Kind != ast.KindArrayLiteralExpression || right.Kind != ast.KindArrayLiteralExpression {
		return
	}

	leftArray := left.AsArrayLiteralExpression()
	rightArray := right.AsArrayLiteralExpression()

	if leftArray == nil || rightArray == nil || leftArray.Elements == nil || rightArray.Elements == nil {
		return
	}

	leftElements := leftArray.Elements.Nodes
	rightElements := rightArray.Elements.Nodes

	// Check each element pair
	for i := 0; i < len(leftElements) && i < len(rightElements); i++ {
		leftElem := leftElements[i]
		rightElem := rightElements[i]

		if leftElem == nil || rightElem == nil {
			continue
		}

		// Skip spread elements and omitted expressions
		if leftElem.Kind == ast.KindSpreadElement || rightElem.Kind == ast.KindSpreadElement {
			continue
		}

		// Check if elements are equal
		if nodesAreEqual(ctx.SourceFile, leftElem, rightElem) {
			ctx.ReportNode(leftElem, buildSelfAssignmentMessage(getDisplayName(ctx.SourceFile, leftElem)))
		}
	}
}

// checkObjectPattern checks object destructuring patterns
func checkObjectPattern(ctx rule.RuleContext, left *ast.Node, right *ast.Node, opts Options) {
	if left == nil || right == nil {
		return
	}

	if left.Kind != ast.KindObjectLiteralExpression || right.Kind != ast.KindObjectLiteralExpression {
		return
	}

	leftObj := left.AsObjectLiteralExpression()
	rightObj := right.AsObjectLiteralExpression()

	if leftObj == nil || rightObj == nil || leftObj.Properties == nil || rightObj.Properties == nil {
		return
	}

	leftProps := leftObj.Properties.Nodes
	rightProps := rightObj.Properties.Nodes

	// Build a map of right-side properties by name
	rightPropMap := make(map[string]*ast.Node)
	for _, rightProp := range rightProps {
		if rightProp == nil || rightProp.Kind != ast.KindPropertyAssignment {
			continue
		}
		propAssign := rightProp.AsPropertyAssignment()
		if propAssign == nil || propAssign.Name() == nil {
			continue
		}
		name := propAssign.Name().Text()
		rightPropMap[name] = propAssign.Initializer
	}

	// Check left-side properties
	for _, leftProp := range leftProps {
		if leftProp == nil || leftProp.Kind != ast.KindPropertyAssignment {
			continue
		}
		propAssign := leftProp.AsPropertyAssignment()
		if propAssign == nil || propAssign.Name() == nil {
			continue
		}
		name := propAssign.Name().Text()

		// Find matching right property
		if rightValue, ok := rightPropMap[name]; ok {
			leftValue := propAssign.Initializer
			if leftValue != nil && nodesAreEqual(ctx.SourceFile, leftValue, rightValue) {
				ctx.ReportNode(leftValue, buildSelfAssignmentMessage(getDisplayName(ctx.SourceFile, leftValue)))
			}
		}
	}
}

// NoSelfAssignRule checks for assignments where both sides are exactly the same
var NoSelfAssignRule = rule.CreateRule(rule.Rule{
	Name: "no-self-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		listeners := rule.RuleListeners{}

		// Handle regular assignment expressions (a = a, a.b = a.b, etc.)
		listeners[ast.KindBinaryExpression] = func(node *ast.Node) {
			checkAssignmentExpression(ctx, node, opts)
		}

		return listeners
	},
})

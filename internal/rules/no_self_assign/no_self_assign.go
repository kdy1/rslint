package no_self_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoSelfAssignOptions defines the configuration options for this rule
type NoSelfAssignOptions struct {
	Props bool `json:"props"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) NoSelfAssignOptions {
	opts := NoSelfAssignOptions{
		Props: true, // Default is true
	}

	if options == nil {
		return opts
	}

	// Handle both array format [{ option: value }] and object format { option: value }
	var optsMap map[string]interface{}
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, _ = optArray[0].(map[string]interface{})
	} else {
		optsMap, _ = options.(map[string]interface{})
	}

	if optsMap != nil {
		if v, ok := optsMap["props"].(bool); ok {
			opts.Props = v
		}
	}

	return opts
}

// NoSelfAssignRule implements the no-self-assign rule
// Disallow assignments where both sides are exactly the same
var NoSelfAssignRule = rule.Rule{
	Name: "no-self-assign",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	return rule.RuleListeners{
		ast.KindBinaryExpression: func(node *ast.Node) {
			if node == nil {
				return
			}

			// Check for assignment operators
			op := node.OperatorToken()
			if op == nil {
				return
			}

			opKind := op.Kind

			// Check for regular assignment (=) and compound logical assignments (&&=, ||=, ??=)
			if opKind != ast.KindEqualsToken &&
				opKind != ast.KindAmpersandAmpersandEqualsToken &&
				opKind != ast.KindBarBarEqualsToken &&
				opKind != ast.KindQuestionQuestionEqualsToken {
				return
			}

			left := node.Left()
			right := node.Right()

			if left == nil || right == nil {
				return
			}

			// Check if left and right are the same
			if isSameExpression(left, right, opts) {
				reportSelfAssignment(ctx, node, left)
			}
		},

		// Handle destructuring assignments: [a] = [a]
		ast.KindArrayLiteralExpression: func(node *ast.Node) {
			parent := node.Parent()
			if parent == nil || parent.Kind != ast.KindBinaryExpression {
				return
			}

			// Check if this is the left side of an assignment
			if parent.Left() != node {
				return
			}

			// Check assignment operator
			op := parent.OperatorToken()
			if op == nil || op.Kind != ast.KindEqualsToken {
				return
			}

			right := parent.Right()
			if right == nil {
				return
			}

			checkArrayPatternAssignment(ctx, node, right, opts)
		},

		// Handle object destructuring: {a} = {a}
		ast.KindObjectLiteralExpression: func(node *ast.Node) {
			parent := node.Parent()
			if parent == nil || parent.Kind != ast.KindBinaryExpression {
				return
			}

			// Check if this is the left side of an assignment
			if parent.Left() != node {
				return
			}

			// Check assignment operator
			op := parent.OperatorToken()
			if op == nil || op.Kind != ast.KindEqualsToken {
				return
			}

			right := parent.Right()
			if right == nil {
				return
			}

			checkObjectPatternAssignment(ctx, node, right, opts)
		},
	}
}

// isSameExpression checks if two expressions are the same
func isSameExpression(left, right *ast.Node, opts NoSelfAssignOptions) bool {
	if left == nil || right == nil {
		return false
	}

	// Must be the same kind
	if left.Kind != right.Kind {
		return false
	}

	switch left.Kind {
	case ast.KindIdentifier:
		return left.Text() == right.Text()

	case ast.KindPropertyAccessExpression:
		if !opts.Props {
			return false
		}
		leftObj := left.Expression()
		rightObj := right.Expression()
		leftName := left.Name()
		rightName := right.Name()

		return isSameExpression(leftObj, rightObj, opts) &&
			isSameExpression(leftName, rightName, opts)

	case ast.KindElementAccessExpression:
		if !opts.Props {
			return false
		}
		leftObj := left.Expression()
		rightObj := right.Expression()
		leftArg := left.ArgumentExpression()
		rightArg := right.ArgumentExpression()

		return isSameExpression(leftObj, rightObj, opts) &&
			isSameExpression(leftArg, rightArg, opts)

	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return left.Text() == right.Text()

	case ast.KindThisKeyword:
		return true

	// For more complex expressions, we'll be conservative
	default:
		return false
	}
}

// checkArrayPatternAssignment checks for self-assignment in array destructuring
func checkArrayPatternAssignment(ctx rule.RuleContext, leftArray, rightExpr *ast.Node, opts NoSelfAssignOptions) {
	if rightExpr.Kind != ast.KindArrayLiteralExpression {
		return
	}

	leftElems := leftArray.Elements()
	rightElems := rightExpr.Elements()

	if len(leftElems) != len(rightElems) {
		return
	}

	for i := 0; i < len(leftElems); i++ {
		leftElem := leftElems[i]
		rightElem := rightElems[i]

		if leftElem == nil || rightElem == nil {
			continue
		}

		// Skip spread elements for now
		if leftElem.Kind == ast.KindSpreadElement || rightElem.Kind == ast.KindSpreadElement {
			continue
		}

		// Skip elements with defaults
		if leftElem.Kind == ast.KindBinaryExpression {
			continue
		}

		if isSameExpression(leftElem, rightElem, opts) {
			reportSelfAssignment(ctx, leftElem, leftElem)
		}
	}
}

// checkObjectPatternAssignment checks for self-assignment in object destructuring
func checkObjectPatternAssignment(ctx rule.RuleContext, leftObj, rightExpr *ast.Node, opts NoSelfAssignOptions) {
	if rightExpr.Kind != ast.KindObjectLiteralExpression {
		return
	}

	leftProps := leftObj.Properties()
	rightProps := rightExpr.Properties()

	// Create a map of right-side properties for lookup
	rightPropMap := make(map[string]*ast.Node)
	for _, prop := range rightProps {
		if prop == nil {
			continue
		}

		var propName string
		var propValue *ast.Node

		if prop.Kind == ast.KindPropertyAssignment {
			name := prop.Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				propName = name.Text()
			}
			propValue = prop.Initializer()
		} else if prop.Kind == ast.KindShorthandPropertyAssignment {
			name := prop.Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				propName = name.Text()
				propValue = name // For shorthand, the name is the value
			}
		}

		if propName != "" && propValue != nil {
			rightPropMap[propName] = propValue
		}
	}

	// Check left-side properties
	for _, prop := range leftProps {
		if prop == nil {
			continue
		}

		var leftPropName string
		var leftPropValue *ast.Node

		if prop.Kind == ast.KindPropertyAssignment {
			name := prop.Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				leftPropName = name.Text()
			}
			leftPropValue = prop.Initializer()
		} else if prop.Kind == ast.KindShorthandPropertyAssignment {
			name := prop.Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				leftPropName = name.Text()
				leftPropValue = name
			}
		}

		if leftPropName == "" || leftPropValue == nil {
			continue
		}

		// Check if there's a matching property on the right
		if rightPropValue, ok := rightPropMap[leftPropName]; ok {
			if isSameExpression(leftPropValue, rightPropValue, opts) {
				reportSelfAssignment(ctx, prop, leftPropValue)
			}
		}
	}
}

// reportSelfAssignment reports a self-assignment error
func reportSelfAssignment(ctx rule.RuleContext, node, target *ast.Node) {
	var targetName string

	if target.Kind == ast.KindIdentifier {
		targetName = target.Text()
	} else if target.Kind == ast.KindPropertyAccessExpression {
		targetName = getPropertyAccessName(target)
	} else if target.Kind == ast.KindElementAccessExpression {
		targetName = getElementAccessName(target)
	} else {
		targetName = "value"
	}

	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "selfAssignment",
		Description: "'" + targetName + "' is assigned to itself.",
	})
}

// getPropertyAccessName constructs a name for property access expressions
func getPropertyAccessName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindPropertyAccessExpression {
		return ""
	}

	obj := node.Expression()
	name := node.Name()

	var objName string
	if obj != nil {
		if obj.Kind == ast.KindIdentifier {
			objName = obj.Text()
		} else if obj.Kind == ast.KindPropertyAccessExpression {
			objName = getPropertyAccessName(obj)
		} else if obj.Kind == ast.KindThisKeyword {
			objName = "this"
		}
	}

	var propName string
	if name != nil && name.Kind == ast.KindIdentifier {
		propName = name.Text()
	}

	if objName != "" && propName != "" {
		return objName + "." + propName
	}

	return propName
}

// getElementAccessName constructs a name for element access expressions
func getElementAccessName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindElementAccessExpression {
		return ""
	}

	obj := node.Expression()
	arg := node.ArgumentExpression()

	var objName string
	if obj != nil {
		if obj.Kind == ast.KindIdentifier {
			objName = obj.Text()
		} else if obj.Kind == ast.KindPropertyAccessExpression {
			objName = getPropertyAccessName(obj)
		} else if obj.Kind == ast.KindThisKeyword {
			objName = "this"
		}
	}

	var indexName string
	if arg != nil {
		if arg.Kind == ast.KindStringLiteral {
			indexName = arg.Text()
		} else if arg.Kind == ast.KindIdentifier {
			indexName = arg.Text()
		}
	}

	if objName != "" && indexName != "" {
		return objName + "[" + indexName + "]"
	}

	return objName
}

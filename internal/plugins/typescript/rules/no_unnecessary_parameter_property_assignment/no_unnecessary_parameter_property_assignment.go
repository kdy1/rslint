package no_unnecessary_parameter_property_assignment

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildUnnecessaryAssignmentMessage(parameterName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessaryAssignment",
		Description: "This assignment is unnecessary since the parameter property `" + parameterName + "` automatically assigns to `this." + parameterName + "`.",
	}
}

// getParameterName extracts the parameter name from a parameter node
func getParameterName(param *ast.Node) string {
	if param == nil {
		return ""
	}

	paramDecl := param.AsParameterDeclaration()
	if paramDecl == nil || paramDecl.Name() == nil {
		return ""
	}

	name := paramDecl.Name()
	if name.Kind == ast.KindIdentifier {
		return name.AsIdentifier().Text
	}

	return ""
}

// isParameterProperty checks if a parameter has a modifier (public, private, protected, readonly)
func isParameterProperty(param *ast.Node) bool {
	if param == nil {
		return false
	}

	flags := ast.GetCombinedModifierFlags(param)

	// Parameter properties have public, private, protected, or readonly modifiers
	return (flags&ast.ModifierFlagsPublic != 0) ||
		(flags&ast.ModifierFlagsPrivate != 0) ||
		(flags&ast.ModifierFlagsProtected != 0) ||
		(flags&ast.ModifierFlagsReadonly != 0)
}

// isSimpleParameterReference checks if the right side of assignment is a simple parameter reference
// Returns the parameter name if it's a simple reference, empty string otherwise
func isSimpleParameterReference(node *ast.Node) string {
	if node == nil {
		return ""
	}

	// Skip parentheses
	node = ast.SkipParentheses(node)

	// Handle non-null assertions (foo!)
	if node.Kind == ast.KindNonNullExpression {
		nonNull := node.AsNonNullExpression()
		if nonNull != nil && nonNull.Expression != nil {
			node = ast.SkipParentheses(nonNull.Expression)
		}
	}

	// Handle type assertions (foo as any, <any>foo)
	if node.Kind == ast.KindAsExpression {
		asExpr := node.AsAsExpression()
		if asExpr != nil && asExpr.Expression != nil {
			node = ast.SkipParentheses(asExpr.Expression)
		}
	}

	if node.Kind == ast.KindTypeAssertionExpression {
		typeAssertion := node.AsTypeAssertion()
		if typeAssertion != nil && typeAssertion.Expression != nil {
			node = ast.SkipParentheses(typeAssertion.Expression)
		}
	}

	// Finally check if it's an identifier
	if node.Kind == ast.KindIdentifier {
		return node.AsIdentifier().Text
	}

	return ""
}

// getPropertyName extracts the property name from a this.property or this['property'] expression
func getPropertyName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	// Handle this.property (PropertyAccessExpression)
	if ast.IsPropertyAccessExpression(node) {
		propAccess := node.AsPropertyAccessExpression()
		if propAccess != nil && propAccess.Expression.Kind == ast.KindThisKeyword && propAccess.Name() != nil {
			if propAccess.Name().Kind == ast.KindIdentifier {
				return propAccess.Name().AsIdentifier().Text
			}
		}
	}

	// Handle this['property'] (ElementAccessExpression)
	if ast.IsElementAccessExpression(node) {
		elemAccess := node.AsElementAccessExpression()
		if elemAccess != nil && elemAccess.Expression.Kind == ast.KindThisKeyword && elemAccess.ArgumentExpression != nil {
			argExpr := elemAccess.ArgumentExpression
			if argExpr.Kind == ast.KindStringLiteral {
				// Extract the string value without quotes
				text := argExpr.AsStringLiteral().Text
				if len(text) >= 2 {
					return text[1 : len(text)-1]
				}
			}
		}
	}

	return ""
}

// isDirectlyInConstructor checks if a node is directly in the constructor body (not in a nested function)
func isDirectlyInConstructor(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Walk up the tree to find the containing function
	current := node.Parent
	for current != nil {
		// If we hit a constructor, we're good
		if ast.IsConstructorDeclaration(current) {
			return true
		}

		// If we hit any other function before a constructor, we're in a nested function
		if ast.IsFunctionDeclaration(current) ||
			ast.IsFunctionExpression(current) ||
			ast.IsArrowFunction(current) ||
			ast.IsMethodDeclaration(current) ||
			ast.IsGetAccessorDeclaration(current) ||
			ast.IsSetAccessorDeclaration(current) {
			return false
		}

		current = current.Parent
	}

	return false
}

var NoUnnecessaryParameterPropertyAssignmentRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-parameter-property-assignment",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Map to track parameter properties in the current constructor
		// Key: parameter name, Value: true if it's a parameter property
		var parameterProperties map[string]bool

		return rule.RuleListeners{
			// Track when entering a constructor to collect parameter properties
			ast.KindConstructorDeclaration: func(node *ast.Node) {
				parameterProperties = make(map[string]bool)

				constructor := node.AsConstructorDeclaration()
				if constructor == nil {
					return
				}

				// Collect all parameter properties
				for _, param := range constructor.Parameters.Nodes {
					if isParameterProperty(param) {
						paramName := getParameterName(param)
						if paramName != "" {
							parameterProperties[paramName] = true
						}
					}
				}
			},

			// Check assignments in the constructor
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil {
					return
				}

				// Only check simple assignments (=), not compound assignments (+=, -=, etc.)
				if binary.OperatorToken.Kind != ast.KindEqualsToken {
					return
				}

				// Check if we're directly in a constructor (not in a nested function)
				if !isDirectlyInConstructor(node) {
					return
				}

				// Check if left side is this.property or this['property']
				propertyName := getPropertyName(binary.Left)
				if propertyName == "" {
					return
				}

				// Check if this property corresponds to a parameter property
				if !parameterProperties[propertyName] {
					return
				}

				// Check if right side is a simple reference to the same parameter
				rightParamName := isSimpleParameterReference(binary.Right)
				if rightParamName == "" || rightParamName != propertyName {
					return
				}

				// Report the unnecessary assignment
				ctx.ReportNode(node, buildUnnecessaryAssignmentMessage(propertyName))
			},
		}
	},
})

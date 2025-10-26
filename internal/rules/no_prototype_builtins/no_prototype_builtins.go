package no_prototype_builtins

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoPrototypeBuiltinsRule implements the no-prototype-builtins rule
// Disallow calling some `Object.prototype` methods directly on objects
var NoPrototypeBuiltinsRule = rule.Rule{
	Name: "no-prototype-builtins",
	Run:  run,
}

// Object.prototype methods that should not be called directly
var prototypeBuiltins = map[string]bool{
	"hasOwnProperty":       true,
	"isPrototypeOf":        true,
	"propertyIsEnumerable": true,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindCallExpression: func(node *ast.Node) {
			checkCallExpression(ctx, node)
		},
	}
}

func checkCallExpression(ctx rule.RuleContext, node *ast.Node) {
	if node == nil {
		return
	}

	expr := node.Expression()
	if expr == nil {
		return
	}

	// Unwrap parenthesized expressions
	actualExpr := expr
	for actualExpr != nil && actualExpr.Kind == ast.KindParenthesizedExpression {
		actualExpr = actualExpr.Expression()
	}

	if actualExpr == nil {
		return
	}

	// Check for property access expressions: foo.hasOwnProperty('bar')
	if actualExpr.Kind == ast.KindPropertyAccessExpression {
		checkPropertyAccessCall(ctx, node, actualExpr)
		return
	}

	// Check for element access expressions: foo['hasOwnProperty']('bar')
	if actualExpr.Kind == ast.KindElementAccessExpression {
		checkElementAccessCall(ctx, node, actualExpr)
		return
	}
}

func checkPropertyAccessCall(ctx rule.RuleContext, callNode *ast.Node, propAccess *ast.Node) {
	if propAccess == nil {
		return
	}

	name := propAccess.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return
	}

	methodName := name.Text()

	// Check if this is one of the prototype builtins
	if !prototypeBuiltins[methodName] {
		return
	}

	obj := propAccess.Expression()
	if obj == nil {
		return
	}

	// Don't report if it's Object.prototype.hasOwnProperty.call/apply pattern
	if isObjectPrototypePattern(obj, methodName) {
		return
	}

	// Don't report if it's {}.hasOwnProperty.call/apply pattern
	if isObjectLiteralPattern(obj, methodName) {
		return
	}

	// Report the error
	ctx.ReportNode(callNode, rule.RuleMessage{
		Id:          "prototypeBuildIn",
		Description: "Do not access Object.prototype method '" + methodName + "' from target object.",
	})
}

func checkElementAccessCall(ctx rule.RuleContext, callNode *ast.Node, elemAccess *ast.Node) {
	if elemAccess == nil {
		return
	}

	arg := elemAccess.ArgumentExpression()
	if arg == nil {
		return
	}

	var methodName string

	// Check if it's a string literal: foo['hasOwnProperty']
	if arg.Kind == ast.KindStringLiteral {
		text := arg.Text()
		// Remove quotes
		if len(text) >= 2 {
			methodName = text[1 : len(text)-1]
		}
	}

	// Check if it's a template literal without substitutions: foo[`hasOwnProperty`]
	if arg.Kind == ast.KindNoSubstitutionTemplateLiteral {
		text := arg.Text()
		// Remove backticks
		if len(text) >= 2 {
			methodName = text[1 : len(text)-1]
		}
	}

	// If we couldn't extract a method name, return
	if methodName == "" {
		return
	}

	// Check if this is one of the prototype builtins
	if !prototypeBuiltins[methodName] {
		return
	}

	obj := elemAccess.Expression()
	if obj == nil {
		return
	}

	// Don't report if it's Object.prototype['hasOwnProperty'].call/apply pattern
	if isObjectPrototypePattern(obj, methodName) {
		return
	}

	// Don't report if it's {}['hasOwnProperty'].call/apply pattern
	if isObjectLiteralPattern(obj, methodName) {
		return
	}

	// Report the error
	ctx.ReportNode(callNode, rule.RuleMessage{
		Id:          "prototypeBuildIn",
		Description: "Do not access Object.prototype method '" + methodName + "' from target object.",
	})
}

// isObjectPrototypePattern checks if the object is Object.prototype
// Returns true for patterns like Object.prototype or Object['prototype']
func isObjectPrototypePattern(obj *ast.Node, methodName string) bool {
	if obj == nil {
		return false
	}

	// Check for Object.prototype pattern
	if obj.Kind == ast.KindPropertyAccessExpression {
		objExpr := obj.Expression()
		if objExpr != nil && objExpr.Kind == ast.KindIdentifier && objExpr.Text() == "Object" {
			propName := obj.Name()
			if propName != nil && propName.Kind == ast.KindIdentifier && propName.Text() == "prototype" {
				return true
			}
		}
	}

	// Check for Object['prototype'] pattern
	if obj.Kind == ast.KindElementAccessExpression {
		objExpr := obj.Expression()
		if objExpr != nil && objExpr.Kind == ast.KindIdentifier && objExpr.Text() == "Object" {
			arg := obj.ArgumentExpression()
			if arg != nil && arg.Kind == ast.KindStringLiteral {
				text := arg.Text()
				if len(text) >= 2 && text[1:len(text)-1] == "prototype" {
					return true
				}
			}
		}
	}

	return false
}

// isObjectLiteralPattern checks if the object is an empty object literal {}
func isObjectLiteralPattern(obj *ast.Node, methodName string) bool {
	if obj == nil {
		return false
	}

	// Unwrap parentheses: ({})
	actualObj := obj
	for actualObj != nil && actualObj.Kind == ast.KindParenthesizedExpression {
		actualObj = actualObj.Expression()
	}

	if actualObj == nil {
		return false
	}

	// Check if it's an object literal
	if actualObj.Kind == ast.KindObjectLiteralExpression {
		// Accept empty object literals or object literals without the method name
		props := actualObj.Properties()
		if props == nil || len(props) == 0 {
			return true
		}

		// Check that the object literal doesn't override the method
		for _, prop := range props {
			if prop == nil {
				continue
			}

			var propName string
			if prop.Kind == ast.KindPropertyAssignment || prop.Kind == ast.KindMethodDeclaration {
				name := prop.Name()
				if name != nil {
					if name.Kind == ast.KindIdentifier {
						propName = name.Text()
					} else if name.Kind == ast.KindStringLiteral {
						text := name.Text()
						if len(text) >= 2 {
							propName = text[1 : len(text)-1]
						}
					}
				}
			}

			// If the object literal has the method, it's not using the prototype
			if propName == methodName {
				return false
			}
		}

		return true
	}

	return false
}

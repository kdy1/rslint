package no_setter_return

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildReturnsValueMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "returnsValue",
		Description: "Setter cannot return a value.",
	}
}

// checkReturnStatements checks all return statements in a function body
func checkReturnStatements(ctx rule.RuleContext, body *ast.Node) {
	if body == nil {
		return
	}

	// For arrow functions with expression body, check if it's an implicit return
	if body.Kind != ast.KindBlock {
		// This is an implicit return in an arrow function
		// Report it as it returns a value
		ctx.ReportNode(body, buildReturnsValueMessage())
		return
	}

	// For block bodies, traverse and find return statements
	var traverse func(*ast.Node) bool
	traverse = func(node *ast.Node) bool {
		if node == nil {
			return false
		}

		// If we hit a function/arrow function, don't traverse into it (it's a nested function)
		if node.Kind == ast.KindFunctionExpression ||
			node.Kind == ast.KindArrowFunction ||
			node.Kind == ast.KindFunctionDeclaration {
			return false
		}

		if node.Kind == ast.KindReturnStatement {
			ret := node.AsReturnStatement()
			if ret != nil && ret.Expression != nil {
				// This return has a value
				ctx.ReportNode(node, buildReturnsValueMessage())
			}
		}

		// Traverse children
		node.ForEachChild(traverse)
		return false
	}

	traverse(body)
}

// findSetterInObjectLiteral looks for setter in object literal properties
func findSetterInObjectLiteral(objLit *ast.ObjectLiteralExpression) []*ast.Node {
	var setters []*ast.Node
	if objLit == nil || objLit.Properties == nil {
		return setters
	}

	for _, prop := range objLit.Properties.Nodes {
		if prop == nil {
			continue
		}

		if prop.Kind == ast.KindPropertyAssignment {
			propAssign := prop.AsPropertyAssignment()
			if propAssign == nil || propAssign.Name() == nil {
				continue
			}

			// Check if property name is "set"
			if propAssign.Name().Text() == "set" && propAssign.Initializer != nil {
				init := propAssign.Initializer
				if init.Kind == ast.KindFunctionExpression || init.Kind == ast.KindArrowFunction {
					setters = append(setters, init)
				}
			}
		}
	}

	return setters
}

// NoSetterReturnRule checks for return statements with values in setter methods
var NoSetterReturnRule = rule.CreateRule(rule.Rule{
	Name: "no-setter-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		listeners := rule.RuleListeners{}

		// Check class/object setters using KindSetAccessor
		listeners[ast.KindSetAccessor] = func(node *ast.Node) {
			accessor := node.AsSetAccessorDeclaration()
			if accessor == nil || accessor.Body == nil {
				return
			}
			checkReturnStatements(ctx, accessor.Body)
		}

		// Check Object.defineProperty, Object.defineProperties, Object.create patterns
		listeners[ast.KindCallExpression] = func(node *ast.Node) {
			call := node.AsCallExpression()
			if call == nil || call.Expression == nil || call.Arguments == nil {
				return
			}

			expr := call.Expression
			if expr.Kind != ast.KindPropertyAccessExpression {
				return
			}

			propAccess := expr.AsPropertyAccessExpression()
			if propAccess == nil || propAccess.Name() == nil || propAccess.Expression == nil {
				return
			}

			// Check if it's Object.defineProperty, Object.defineProperties, or Object.create
			methodName := propAccess.Name().Text()
			objExpr := propAccess.Expression

			var objName string
			if objExpr.Kind == ast.KindIdentifier {
				objName = objExpr.AsIdentifier().Text
			}

			if objName != "Object" && objName != "Reflect" {
				return
			}

			args := call.Arguments.Nodes
			if len(args) == 0 {
				return
			}

			switch methodName {
			case "defineProperty":
				// Object.defineProperty(obj, 'prop', { set(val) { return val; } })
				// Third argument is the descriptor
				if len(args) >= 3 && args[2] != nil && args[2].Kind == ast.KindObjectLiteralExpression {
					setters := findSetterInObjectLiteral(args[2].AsObjectLiteralExpression())
					for _, setter := range setters {
						if setter.Kind == ast.KindFunctionExpression {
							fnExpr := setter.AsFunctionExpression()
							if fnExpr != nil && fnExpr.Body != nil {
								checkReturnStatements(ctx, fnExpr.Body)
							}
						} else if setter.Kind == ast.KindArrowFunction {
							arrow := setter.AsArrowFunction()
							if arrow != nil && arrow.Body != nil {
								checkReturnStatements(ctx, arrow.Body)
							}
						}
					}
				}

			case "defineProperties", "create":
				// Object.defineProperties(obj, { prop: { set(val) { return val; } } })
				// Object.create(null, { prop: { set(val) { return val; } } })
				// Second argument contains property descriptors
				var descriptorsArg *ast.Node
				if methodName == "defineProperties" && len(args) >= 2 {
					descriptorsArg = args[1]
				} else if methodName == "create" && len(args) >= 2 {
					descriptorsArg = args[1]
				}

				if descriptorsArg != nil && descriptorsArg.Kind == ast.KindObjectLiteralExpression {
					objLit := descriptorsArg.AsObjectLiteralExpression()
					if objLit == nil || objLit.Properties == nil {
						return
					}

					// Each property might have a descriptor object
					for _, prop := range objLit.Properties.Nodes {
						if prop == nil || prop.Kind != ast.KindPropertyAssignment {
							continue
						}
						propAssign := prop.AsPropertyAssignment()
						if propAssign == nil || propAssign.Initializer == nil {
							continue
						}

						if propAssign.Initializer.Kind == ast.KindObjectLiteralExpression {
							descriptorObj := propAssign.Initializer.AsObjectLiteralExpression()
							setters := findSetterInObjectLiteral(descriptorObj)
							for _, setter := range setters {
								if setter.Kind == ast.KindFunctionExpression {
									fnExpr := setter.AsFunctionExpression()
									if fnExpr != nil && fnExpr.Body != nil {
										checkReturnStatements(ctx, fnExpr.Body)
									}
								} else if setter.Kind == ast.KindArrowFunction {
									arrow := setter.AsArrowFunction()
									if arrow != nil && arrow.Body != nil {
										checkReturnStatements(ctx, arrow.Body)
									}
								}
							}
						}
					}
				}
			}
		}

		return listeners
	},
})

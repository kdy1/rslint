package getter_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Options for getter-return rule
type Options struct {
	AllowImplicit bool `json:"allowImplicit"`
}

func parseOptions(options any) Options {
	opts := Options{
		AllowImplicit: false,
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
		if v, ok := optsMap["allowImplicit"].(bool); ok {
			opts.AllowImplicit = v
		}
	}
	return opts
}

func buildExpectedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expected",
		Description: "Expected to return a value in getter.",
	}
}

func buildExpectedAlwaysMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedAlways",
		Description: "Expected getter to always return a value.",
	}
}

// checkGetterReturn checks if a getter function has proper return statements
func checkGetterReturn(ctx rule.RuleContext, node *ast.Node, opts Options) {
	if node == nil {
		return
	}

	body := node.Body()
	if body == nil {
		return
	}

	// If allowImplicit is true, we don't check for return values
	if opts.AllowImplicit {
		return
	}

	// Simple check: look for any return statement in the body
	// For a full implementation, we'd need to traverse the AST more deeply
	// For now, we'll report if there's no body or if the body is empty

	if body.Kind == ast.KindBlock {
		statements := body.Statements()
		if statements == nil || len(statements) == 0 {
			ctx.ReportNode(node, buildExpectedMessage())
			return
		}

		// Check if there's at least one return with a value
		hasReturnWithValue := false
		for _, stmt := range statements {
			if stmt != nil && stmt.Kind == ast.KindReturnStatement {
				expr := stmt.Expression()
				if expr != nil {
					hasReturnWithValue = true
					break
				}
			}
		}

		if !hasReturnWithValue {
			ctx.ReportNode(node, buildExpectedMessage())
		}
	}
}

// GetterReturnRule enforces return statements in getters
var GetterReturnRule = rule.CreateRule(rule.Rule{
	Name: "getter-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindGetAccessor: func(node *ast.Node) {
				checkGetterReturn(ctx, node, opts)
			},

			// Handle Object.defineProperty, Reflect.defineProperty
			ast.KindCallExpression: func(node *ast.Node) {
				expr := node.Expression()
				if expr == nil {
					return
				}

				var objectName, methodName string

				// Check for Object.defineProperty, Reflect.defineProperty
				if expr.Kind == ast.KindPropertyAccessExpression {
					obj := expr.Expression()
					if obj != nil && obj.Kind == ast.KindIdentifier {
						objectName = obj.Text()
					}
					name := expr.Name()
					if name != nil && name.Kind == ast.KindIdentifier {
						methodName = name.Text()
					}
				}

				args := node.Arguments()
				if args == nil {
					return
				}

				var descriptorArg *ast.Node

				// Object.defineProperty(obj, 'prop', { get: function() {} })
				if (objectName == "Object" && methodName == "defineProperty") ||
					(objectName == "Reflect" && methodName == "defineProperty") {
					if len(args) >= 3 {
						descriptorArg = args[2]
					}
				}

				// Object.defineProperties(obj, { prop: { get: function() {} } })
				if objectName == "Object" && methodName == "defineProperties" {
					if len(args) >= 2 {
						propsArg := args[1]
						if propsArg != nil && propsArg.Kind == ast.KindObjectLiteralExpression {
							props := propsArg.Properties()
							for _, prop := range props {
								if prop != nil && (prop.Kind == ast.KindPropertyAssignment || prop.Kind == ast.KindShorthandPropertyAssignment) {
									init := prop.Initializer()
									if init != nil && init.Kind == ast.KindObjectLiteralExpression {
										checkDescriptorForGetter(ctx, init, opts)
									}
								}
							}
						}
					}
					return
				}

				// Object.create(proto, { prop: { get: function() {} } })
				if objectName == "Object" && methodName == "create" {
					if len(args) >= 2 {
						descriptorArg = args[1]
						if descriptorArg != nil && descriptorArg.Kind == ast.KindObjectLiteralExpression {
							props := descriptorArg.Properties()
							for _, prop := range props {
								if prop != nil && (prop.Kind == ast.KindPropertyAssignment || prop.Kind == ast.KindShorthandPropertyAssignment) {
									init := prop.Initializer()
									if init != nil && init.Kind == ast.KindObjectLiteralExpression {
										checkDescriptorForGetter(ctx, init, opts)
									}
								}
							}
						}
					}
					return
				}

				if descriptorArg != nil && descriptorArg.Kind == ast.KindObjectLiteralExpression {
					checkDescriptorForGetter(ctx, descriptorArg, opts)
				}
			},
		}
	},
})

// checkDescriptorForGetter checks property descriptors for get functions
func checkDescriptorForGetter(ctx rule.RuleContext, descriptor *ast.Node, opts Options) {
	if descriptor == nil || descriptor.Kind != ast.KindObjectLiteralExpression {
		return
	}

	props := descriptor.Properties()
	for _, prop := range props {
		if prop == nil {
			continue
		}

		// Look for 'get' property
		if prop.Kind == ast.KindPropertyAssignment || prop.Kind == ast.KindMethodDeclaration {
			var propName string
			if prop.Name() != nil {
				if prop.Name().Kind == ast.KindIdentifier {
					propName = prop.Name().Text()
				} else if prop.Name().Kind == ast.KindStringLiteral {
					propName = prop.Name().Text()
					// Remove quotes
					if len(propName) >= 2 {
						propName = propName[1 : len(propName)-1]
					}
				}
			}

			if propName == "get" {
				// Found a getter
				var getterFunc *ast.Node
				if prop.Kind == ast.KindPropertyAssignment {
					getterFunc = prop.Initializer()
				} else if prop.Kind == ast.KindMethodDeclaration {
					getterFunc = prop
				}

				if getterFunc != nil {
					if getterFunc.Kind == ast.KindFunctionExpression ||
						getterFunc.Kind == ast.KindArrowFunction ||
						getterFunc.Kind == ast.KindMethodDeclaration {
						checkGetterReturn(ctx, getterFunc, opts)
					}
				}
			}
		}
	}
}

package class_methods_use_this

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type ClassMethodsUseThisOptions struct {
	ExceptMethods          []string `json:"exceptMethods"`
	EnforceForClassFields  bool     `json:"enforceForClassFields"`
}

var ClassMethodsUseThisRule = rule.CreateRule(rule.Rule{
	Name: "class-methods-use-this",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := ClassMethodsUseThisOptions{
			ExceptMethods:         []string{},
			EnforceForClassFields: true,
		}

		// Parse options
		if options != nil {
			var optsMap map[string]interface{}
			if optsArray, ok := options.([]interface{}); ok && len(optsArray) > 0 {
				if opts, ok := optsArray[0].(map[string]interface{}); ok {
					optsMap = opts
				}
			} else if opts, ok := options.(map[string]interface{}); ok {
				optsMap = opts
			}

			if optsMap != nil {
				if exceptMethods, ok := optsMap["exceptMethods"].([]interface{}); ok {
					for _, method := range exceptMethods {
						if str, ok := method.(string); ok {
							opts.ExceptMethods = append(opts.ExceptMethods, str)
						}
					}
				}
				if enforceForClassFields, ok := optsMap["enforceForClassFields"].(bool); ok {
					opts.EnforceForClassFields = enforceForClassFields
				}
			}
		}

		// Helper to check if a method name is excepted
		isExceptedMethod := func(methodName string) bool {
			for _, name := range opts.ExceptMethods {
				if name == methodName {
					return true
				}
			}
			return false
		}

		// Check if a node uses 'this' or 'super'
		usesThisOrSuper := func(node *ast.Node) bool {
			found := false

			// Track function boundaries to not look into nested functions
			functionDepth := 0

			ast.ForEachChild(node, func(child *ast.Node) bool {
				// Don't descend into nested function expressions or arrow functions
				// unless they're arrow functions (which capture 'this')
				if child.Kind == ast.KindFunctionExpression || child.Kind == ast.KindFunctionDeclaration {
					functionDepth++
					if functionDepth > 0 {
						return false // Don't descend
					}
				}

				// Arrow functions capture 'this', so we should check them
				if child.Kind == ast.KindArrowFunction {
					// Check if the arrow function uses 'this'
					if containsThisInArrow(child) {
						found = true
						return false
					}
					return true // Continue checking
				}

				// Check for 'this' keyword
				if child.Kind == ast.KindThisKeyword {
					found = true
					return false
				}

				// Check for 'super' keyword
				if child.Kind == ast.KindSuperKeyword {
					found = true
					return false
				}

				return true // Continue traversal
			})

			return found
		}

		// Helper to check if an arrow function contains 'this'
		containsThisInArrow := func(arrowFunc *ast.Node) bool {
			found := false
			ast.ForEachChild(arrowFunc, func(child *ast.Node) bool {
				// Don't descend into nested regular functions
				if child.Kind == ast.KindFunctionExpression || child.Kind == ast.KindFunctionDeclaration {
					return false
				}

				if child.Kind == ast.KindThisKeyword || child.Kind == ast.KindSuperKeyword {
					found = true
					return false
				}

				return true
			})
			return found
		}

		// Get method name for display
		getMethodName := func(node *ast.Node) string {
			if node.Kind == ast.KindMethodDeclaration {
				method := node.AsMethodDeclaration()
				if method != nil && method.Name() != nil {
					name, isPrivate := utils.GetNameFromMember(ctx.SourceFile, method.Name())
					if isPrivate {
						if method.Kind == ast.KindGetAccessor {
							return "private getter " + name
						} else if method.Kind == ast.KindSetAccessor {
							return "private setter " + name
						} else if method.AsteriskToken != nil {
							return "private generator method " + name
						}
						return "private method " + name
					}

					if method.Kind == ast.KindGetAccessor {
						return "getter '" + name + "'"
					} else if method.Kind == ast.KindSetAccessor {
						return "setter '" + name + "'"
					} else if method.AsteriskToken != nil {
						return "generator method '" + name + "'"
					}
					return "method '" + name + "'"
				}
				return "method"
			} else if node.Kind == ast.KindGetAccessor {
				accessor := node.AsGetAccessorDeclaration()
				if accessor != nil && accessor.Name() != nil {
					name, isPrivate := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
					if isPrivate {
						return "private getter " + name
					}
					return "getter '" + name + "'"
				}
				return "getter"
			} else if node.Kind == ast.KindSetAccessor {
				accessor := node.AsSetAccessorDeclaration()
				if accessor != nil && accessor.Name() != nil {
					name, isPrivate := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
					if isPrivate {
						return "private setter " + name
					}
					if name == "" {
						return "setter"
					}
					return "setter '" + name + "'"
				}
				return "setter"
			} else if node.Kind == ast.KindPropertyDeclaration {
				prop := node.AsPropertyDeclaration()
				if prop != nil && prop.Name() != nil {
					name, isPrivate := utils.GetNameFromMember(ctx.SourceFile, prop.Name())
					if isPrivate {
						return "private method " + name
					}
					return "method '" + name + "'"
				}
				return "method"
			}
			return "method"
		}

		// Check method declarations
		checkMethod := func(node *ast.Node) {
			// Skip constructors
			if node.Kind == ast.KindConstructor {
				return
			}

			// Skip static methods
			if ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
				return
			}

			// Skip abstract methods
			if ast.HasSyntacticModifier(node, ast.ModifierFlagsAbstract) {
				return
			}

			// Check if method is in except list
			var methodName string
			if node.Kind == ast.KindMethodDeclaration {
				method := node.AsMethodDeclaration()
				if method != nil && method.Name() != nil {
					name, _ := utils.GetNameFromMember(ctx.SourceFile, method.Name())
					methodName = name
				}
			} else if node.Kind == ast.KindGetAccessor {
				accessor := node.AsGetAccessorDeclaration()
				if accessor != nil && accessor.Name() != nil {
					name, _ := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
					methodName = name
				}
			} else if node.Kind == ast.KindSetAccessor {
				accessor := node.AsSetAccessorDeclaration()
				if accessor != nil && accessor.Name() != nil {
					name, _ := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
					methodName = name
				}
			}

			if methodName != "" && isExceptedMethod(methodName) {
				return
			}

			// Get method body
			var body *ast.Node
			if node.Kind == ast.KindMethodDeclaration {
				method := node.AsMethodDeclaration()
				if method != nil {
					body = method.Body
				}
			} else if node.Kind == ast.KindGetAccessor {
				accessor := node.AsGetAccessorDeclaration()
				if accessor != nil {
					body = accessor.Body
				}
			} else if node.Kind == ast.KindSetAccessor {
				accessor := node.AsSetAccessorDeclaration()
				if accessor != nil {
					body = accessor.Body
				}
			}

			if body == nil {
				return
			}

			// Check if the method uses 'this' or 'super'
			if !usesThisOrSuper(body) {
				displayName := getMethodName(node)
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "missingThis",
					Description: "Expected 'this' to be used by class " + displayName + ".",
				})
			}
		}

		// Check property declarations with function values (class fields)
		checkPropertyDeclaration := func(node *ast.Node) {
			if node.Kind != ast.KindPropertyDeclaration {
				return
			}

			prop := node.AsPropertyDeclaration()
			if prop == nil {
				return
			}

			// Skip if enforceForClassFields is false
			if !opts.EnforceForClassFields {
				return
			}

			// Skip static properties
			if ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
				return
			}

			// Check if property name is in except list
			if prop.Name() != nil {
				name, _ := utils.GetNameFromMember(ctx.SourceFile, prop.Name())
				if name != "" && isExceptedMethod(name) {
					return
				}
			}

			// Check if initializer is a function or arrow function
			if prop.Initializer == nil {
				return
			}

			init := prop.Initializer
			if init.Kind != ast.KindFunctionExpression && init.Kind != ast.KindArrowFunction {
				return
			}

			// Get function body
			var body *ast.Node
			if init.Kind == ast.KindFunctionExpression {
				fn := init.AsFunctionExpression()
				if fn != nil {
					body = fn.Body
				}
			} else if init.Kind == ast.KindArrowFunction {
				fn := init.AsArrowFunction()
				if fn != nil && fn.Body != nil && fn.Body.Kind == ast.KindBlock {
					body = fn.Body
				}
			}

			if body == nil {
				return
			}

			// Check if the function uses 'this' or 'super'
			if !usesThisOrSuper(body) {
				displayName := getMethodName(node)
				ctx.ReportNode(init, rule.RuleMessage{
					Id:          "missingThis",
					Description: "Expected 'this' to be used by class " + displayName + ".",
				})
			}
		}

		return rule.RuleListeners{
			ast.KindMethodDeclaration:   checkMethod,
			ast.KindGetAccessor:         checkMethod,
			ast.KindSetAccessor:         checkMethod,
			ast.KindPropertyDeclaration: checkPropertyDeclaration,
		}
	},
})

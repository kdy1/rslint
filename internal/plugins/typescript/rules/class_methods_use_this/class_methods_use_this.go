package class_methods_use_this

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type ClassMethodsUseThisOptions struct {
	ExceptMethods          []string `json:"exceptMethods"`
	EnforceForClassFields  *bool    `json:"enforceForClassFields"`
}

var ClassMethodsUseThisRule = rule.CreateRule(rule.Rule{
	Name: "class-methods-use-this",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := ClassMethodsUseThisOptions{
			ExceptMethods:         []string{},
			EnforceForClassFields: nil, // defaults to true
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
					for _, m := range exceptMethods {
						if str, ok := m.(string); ok {
							opts.ExceptMethods = append(opts.ExceptMethods, str)
						}
					}
				}
				if enforceForClassFields, ok := optsMap["enforceForClassFields"].(bool); ok {
					opts.EnforceForClassFields = &enforceForClassFields
				}
			}
		}

		// Default enforceForClassFields to true if not set
		enforceForClassFields := true
		if opts.EnforceForClassFields != nil {
			enforceForClassFields = *opts.EnforceForClassFields
		}

		// Helper to check if method name is in exceptMethods
		isMethodExcepted := func(methodName string) bool {
			for _, name := range opts.ExceptMethods {
				if name == methodName {
					return true
				}
			}
			return false
		}

		// Helper to check if a subtree contains 'this' or 'super' at the current scope level
		containsThisOrSuper := func(node *ast.Node) bool {
			found := false

			var traverse func(*ast.Node, bool)
			traverse = func(n *ast.Node, inNestedFunction bool) {
				if n == nil || found {
					return
				}

				// If we're in a nested function/class, don't count 'this' or 'super' there
				// Arrow functions inherit 'this' from parent, so they don't create a new scope
				if inNestedFunction && n.Kind != ast.KindArrowFunction {
					// Check if this is a function/class that creates a new scope
					switch n.Kind {
					case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
						ast.KindMethodDeclaration, ast.KindConstructor,
						ast.KindGetAccessor, ast.KindSetAccessor,
						ast.KindClassDeclaration, ast.KindClassExpression:
						return // Don't traverse into nested scopes
					}
				}

				// Check for 'this' keyword
				if n.Kind == ast.KindThisKeyword {
					found = true
					return
				}

				// Check for 'super' keyword
				if n.Kind == ast.KindSuperKeyword {
					found = true
					return
				}

				// Mark when we enter a nested function (but not arrow function)
				shouldMarkNested := false
				switch n.Kind {
				case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
					ast.KindMethodDeclaration, ast.KindConstructor,
					ast.KindGetAccessor, ast.KindSetAccessor,
					ast.KindClassDeclaration, ast.KindClassExpression:
					shouldMarkNested = true
				}

				// Traverse children
				ast.ForEachChild(n, func(child *ast.Node) bool {
					if shouldMarkNested {
						traverse(child, true)
					} else {
						traverse(child, inNestedFunction)
					}
					return !found // Continue if not found
				})
			}

			traverse(node, false)
			return found
		}

		// Helper to get method name for error reporting
		getMethodName := func(node *ast.Node) (string, string) {
			var nameNode *ast.Node
			var isPrivate bool
			var isGetter bool
			var isSetter bool
			var isGenerator bool

			switch node.Kind {
			case ast.KindMethodDeclaration:
				method := node.AsMethodDeclaration()
				if method != nil {
					nameNode = method.Name()
					isGenerator = method.AsteriskToken != nil
				}
			case ast.KindGetAccessor:
				accessor := node.AsGetAccessorDeclaration()
				if accessor != nil {
					nameNode = accessor.Name()
					isGetter = true
				}
			case ast.KindSetAccessor:
				accessor := node.AsSetAccessorDeclaration()
				if accessor != nil {
					nameNode = accessor.Name()
					isSetter = true
				}
			case ast.KindPropertyDeclaration:
				prop := node.AsPropertyDeclaration()
				if prop != nil {
					nameNode = prop.Name()
				}
			}

			if nameNode == nil {
				return "", "method"
			}

			// Check if it's a private identifier
			if nameNode.Kind == ast.KindPrivateIdentifier {
				privateIdent := nameNode.AsPrivateIdentifier()
				if privateIdent != nil {
					isPrivate = true
					name := privateIdent.Text
					if isGetter {
						return name, "private getter " + name
					}
					if isSetter {
						return name, "private setter " + name
					}
					return name, "private method " + name
				}
			}

			// Get the name
			name, _ := utils.GetNameFromMember(ctx.SourceFile, nameNode)
			if name == "" {
				if isGetter {
					return "", "getter"
				}
				if isSetter {
					return "", "setter"
				}
				return "", "method"
			}

			// Build the display name
			if isGetter {
				return name, "getter '" + name + "'"
			}
			if isSetter {
				return name, "setter '" + name + "'"
			}
			if isGenerator {
				return name, "generator method '" + name + "'"
			}
			return name, "method '" + name + "'"
		}

		// Check method declaration
		checkMethodDeclaration := func(node *ast.Node) {
			method := node.AsMethodDeclaration()
			if method == nil {
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

			// Skip if no body
			if method.Body == nil {
				return
			}

			// Get method name
			methodName, displayName := getMethodName(node)

			// Check if method is excepted
			if isMethodExcepted(methodName) {
				return
			}

			// Check if method uses 'this' or 'super'
			if containsThisOrSuper(method.Body) {
				return
			}

			// Report the error
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "missingThis",
				Description: "Expected 'this' to be used by class " + displayName + ".",
				Data: map[string]interface{}{
					"name": displayName,
				},
			})
		}

		// Check getter/setter
		checkAccessor := func(node *ast.Node) {
			var body *ast.Node
			var methodName, displayName string

			switch node.Kind {
			case ast.KindGetAccessor:
				accessor := node.AsGetAccessorDeclaration()
				if accessor == nil {
					return
				}
				body = accessor.Body
				methodName, displayName = getMethodName(node)
			case ast.KindSetAccessor:
				accessor := node.AsSetAccessorDeclaration()
				if accessor == nil {
					return
				}
				body = accessor.Body
				methodName, displayName = getMethodName(node)
			default:
				return
			}

			// Skip static accessors
			if ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
				return
			}

			// Skip abstract accessors
			if ast.HasSyntacticModifier(node, ast.ModifierFlagsAbstract) {
				return
			}

			// Skip if no body
			if body == nil {
				return
			}

			// Check if method is excepted
			if isMethodExcepted(methodName) {
				return
			}

			// Check if accessor uses 'this' or 'super'
			if containsThisOrSuper(body) {
				return
			}

			// Report the error
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "missingThis",
				Description: "Expected 'this' to be used by class " + displayName + ".",
				Data: map[string]interface{}{
					"name": displayName,
				},
			})
		}

		// Check property with function value (class fields)
		checkPropertyDeclaration := func(node *ast.Node) {
			if !enforceForClassFields {
				return
			}

			prop := node.AsPropertyDeclaration()
			if prop == nil {
				return
			}

			// Skip static properties
			if ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
				return
			}

			// Skip if no initializer
			if prop.Initializer == nil {
				return
			}

			// Check if initializer is a function or arrow function
			var funcBody *ast.Node
			switch prop.Initializer.Kind {
			case ast.KindFunctionExpression:
				fn := prop.Initializer.AsFunctionExpression()
				if fn != nil {
					funcBody = fn.Body
				}
			case ast.KindArrowFunction:
				fn := prop.Initializer.AsArrowFunction()
				if fn != nil && fn.Body != nil && fn.Body.Kind == ast.KindBlock {
					funcBody = fn.Body
				}
			default:
				return
			}

			if funcBody == nil {
				return
			}

			// Get method name
			methodName, displayName := getMethodName(node)

			// Check if method is excepted
			if isMethodExcepted(methodName) {
				return
			}

			// Check if function uses 'this' or 'super'
			if containsThisOrSuper(funcBody) {
				return
			}

			// Report the error on the initializer
			ctx.ReportNode(prop.Initializer, rule.RuleMessage{
				Id:          "missingThis",
				Description: "Expected 'this' to be used by class " + displayName + ".",
				Data: map[string]interface{}{
					"name": displayName,
				},
			})
		}

		return rule.RuleListeners{
			ast.KindMethodDeclaration:   checkMethodDeclaration,
			ast.KindGetAccessor:          checkAccessor,
			ast.KindSetAccessor:          checkAccessor,
			ast.KindPropertyDeclaration:  checkPropertyDeclaration,
		}
	},
})

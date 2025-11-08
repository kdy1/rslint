package max_params

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type MaxParamsOptions struct {
	Max           int  `json:"max"`
	Maximum       int  `json:"maximum"`
	CountVoidThis bool `json:"countVoidThis"`
}

var MaxParamsRule = rule.CreateRule(rule.Rule{
	Name: "max-params",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Default options: max = 3, countVoidThis = false
		opts := MaxParamsOptions{
			Max:           3,
			Maximum:       0,
			CountVoidThis: false,
		}

		// Parse options with dual-format support (handles both array and object formats)
		if options != nil {
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
				if max, ok := optsMap["max"].(float64); ok {
					opts.Max = int(max)
				}
				if maximum, ok := optsMap["maximum"].(float64); ok {
					opts.Maximum = int(maximum)
				}
				if countVoidThis, ok := optsMap["countVoidThis"].(bool); ok {
					opts.CountVoidThis = countVoidThis
				}
			}
		}

		// Use 'maximum' if specified, otherwise use 'max'
		maxParams := opts.Max
		if opts.Maximum > 0 {
			maxParams = opts.Maximum
		}

		checkFunction := func(node *ast.Node, params *ast.NodeArray) {
			if params == nil {
				return
			}

			paramCount := 0
			for _, param := range params.Nodes {
				paramNode := param.AsParameterDeclaration()
				if paramNode == nil {
					paramCount++
					continue
				}

				// Check if this is a 'this' parameter
				isThisParam := false
				if paramNode.Name != nil && paramNode.Name.Kind == ast.KindIdentifier {
					identifier := paramNode.Name.AsIdentifier()
					if identifier != nil && identifier.EscapedText == "this" {
						isThisParam = true
					}
				}

				// Handle TypeScript 'this' parameter
				if isThisParam {
					// Check if it's a void this parameter
					isVoidThis := false
					if paramNode.Type != nil {
						typeNode := paramNode.Type
						if typeNode.Kind == ast.KindVoidKeyword {
							isVoidThis = true
						}
					}

					// Count void this only if countVoidThis is true
					if isVoidThis {
						if opts.CountVoidThis {
							paramCount++
						}
						// Otherwise, skip counting this parameter
					} else {
						// Non-void this parameters are always counted
						paramCount++
					}
				} else {
					// Regular parameter - always count it
					paramCount++
				}
			}

			if paramCount > maxParams {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "exceed",
					Description: "Function has too many parameters.",
				})
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				funcDecl := node.AsFunctionDeclaration()
				if funcDecl == nil {
					return
				}
				checkFunction(node, funcDecl.Parameters)
			},
			ast.KindFunctionExpression: func(node *ast.Node) {
				funcExpr := node.AsFunctionExpression()
				if funcExpr == nil {
					return
				}
				checkFunction(node, funcExpr.Parameters)
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				arrowFunc := node.AsArrowFunction()
				if arrowFunc == nil {
					return
				}
				checkFunction(node, arrowFunc.Parameters)
			},
			ast.KindMethodDeclaration: func(node *ast.Node) {
				methodDecl := node.AsMethodDeclaration()
				if methodDecl == nil {
					return
				}
				checkFunction(node, methodDecl.Parameters)
			},
			ast.KindConstructor: func(node *ast.Node) {
				constructor := node.AsConstructorDeclaration()
				if constructor == nil {
					return
				}
				checkFunction(node, constructor.Parameters)
			},
		}
	},
})

package no_dupe_args

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoDupeArgsRule implements the no-dupe-args rule
// Disallow duplicate arguments in `function` definitions
var NoDupeArgsRule = rule.Rule{
	Name: "no-dupe-args",
	Run:  run,
}

// checkDuplicateParameters checks for duplicate parameter names
func checkDuplicateParameters(ctx rule.RuleContext, node *ast.Node) {
	if node == nil {
		return
	}

	params := node.Parameters()
	if params == nil || len(params) == 0 {
		return
	}

	// Track parameter names
	paramNames := make(map[string]*ast.Node)

	for _, param := range params {
		if param == nil {
			continue
		}

		// Skip destructuring patterns and rest parameters
		// Only check simple identifiers
		if param.Kind == ast.KindIdentifier {
			name := param.Text()
			if name == "" {
				continue
			}

			if firstOccurrence, exists := paramNames[name]; exists {
				// Found a duplicate
				ctx.ReportNode(param, rule.RuleMessage{
					Id:          "unexpected",
					Description: "Duplicate param '" + name + "'.",
					Data: map[string]interface{}{
						"name": name,
					},
				})
				// Mark the first occurrence as well
				_ = firstOccurrence
			} else {
				paramNames[name] = param
			}
		} else if param.Kind == ast.KindParameter {
			// Handle Parameter nodes that have a name
			name := param.Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				paramName := name.Text()
				if paramName == "" {
					continue
				}

				if firstOccurrence, exists := paramNames[paramName]; exists {
					// Found a duplicate
					ctx.ReportNode(name, rule.RuleMessage{
						Id:          "unexpected",
						Description: "Duplicate param '" + paramName + "'.",
						Data: map[string]interface{}{
							"name": paramName,
						},
					})
					_ = firstOccurrence
				} else {
					paramNames[paramName] = name
				}
			}
		}
	}
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			checkDuplicateParameters(ctx, node)
		},
		ast.KindFunctionExpression: func(node *ast.Node) {
			checkDuplicateParameters(ctx, node)
		},
	}
}

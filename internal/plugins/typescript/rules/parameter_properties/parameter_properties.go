package parameter_properties

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// ParameterPropertiesOptions represents the configuration options
type ParameterPropertiesOptions struct {
	Prefer string `json:"prefer"` // "class-property" or "parameter-property"
}

var ParameterPropertiesRule = rule.CreateRule(rule.Rule{
	Name: "parameter-properties",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := ParameterPropertiesOptions{
			Prefer: "class-property",
		}

		// Parse options
		if options != nil {
			var optsMap map[string]interface{}
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, _ = optArray[0].(map[string]interface{})
			} else {
				optsMap, _ = options.(map[string]interface{})
			}

			if optsMap != nil {
				if prefer, ok := optsMap["prefer"].(string); ok {
					opts.Prefer = prefer
				}
			}
		}

		// TODO: Implement parameter properties checking
		// This rule enforces a consistent style for parameter properties
		// Either prefer "parameter-property" (constructor(public name: string))
		// Or prefer "class-property" (separate property declaration + assignment)

		return rule.RuleListeners{
			ast.KindConstructorDeclaration: func(node *ast.Node) {
				// TODO: Check constructor parameters and class properties
			},
		}
	},
})

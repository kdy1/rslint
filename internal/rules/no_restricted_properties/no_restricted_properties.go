package no_restricted_properties

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoRestrictedPropertiesOptions defines the configuration options for this rule
type NoRestrictedPropertiesOptions struct {
	// TODO: Add option fields here
	// Example: AllowSomePattern bool `json:"allowSomePattern"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) NoRestrictedPropertiesOptions {
	opts := NoRestrictedPropertiesOptions{
		// Set default values here
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
		// TODO: Parse option values from optsMap
		// Example:
		// if v, ok := optsMap["allowSomePattern"].(bool); ok {
		//     opts.AllowSomePattern = v
		// }
	}

	return opts
}

// NoRestrictedPropertiesRule implements the no-restricted-properties rule
// Disallow certain properties on certain objects
var NoRestrictedPropertiesRule = rule.Rule{
	Name: "no-restricted-properties",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)
	_ = opts // Use opts in your rule logic

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			// TODO: Implement rule logic for FunctionDeclaration

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "restrictedObjectProperty",
			//         Description: "TODO: Add error message",
			//     })
			// }
		},
	}
}

package no_unused_expressions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnusedExpressionsOptions represents the configuration options
type NoUnusedExpressionsOptions struct {
	AllowShortCircuit bool `json:"allowShortCircuit"`
	AllowTernary      bool `json:"allowTernary"`
	AllowTaggedTemplates bool `json:"allowTaggedTemplates"`
	EnforceForJSX     bool `json:"enforceForJSX"`
}

var NoUnusedExpressionsRule = rule.CreateRule(rule.Rule{
	Name: "no-unused-expressions",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoUnusedExpressionsOptions{
			AllowShortCircuit:    false,
			AllowTernary:         false,
			AllowTaggedTemplates: false,
			EnforceForJSX:        false,
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
				if allowShortCircuit, ok := optsMap["allowShortCircuit"].(bool); ok {
					opts.AllowShortCircuit = allowShortCircuit
				}
				if allowTernary, ok := optsMap["allowTernary"].(bool); ok {
					opts.AllowTernary = allowTernary
				}
				if allowTaggedTemplates, ok := optsMap["allowTaggedTemplates"].(bool); ok {
					opts.AllowTaggedTemplates = allowTaggedTemplates
				}
				if enforceForJSX, ok := optsMap["enforceForJSX"].(bool); ok {
					opts.EnforceForJSX = enforceForJSX
				}
			}
		}

		// TODO: Implement unused expression detection
		// This rule needs to:
		// 1. Identify expression statements that don't have side effects
		// 2. Handle short-circuit and ternary operators based on options
		// 3. Check for TypeScript-specific unused expressions (type assertions, etc.)
		// 4. Report expressions that should be removed or used

		return rule.RuleListeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				// TODO: Implement unused expression checking
			},
		}
	},
})

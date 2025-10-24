package no_shadow

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoShadowOptions represents the configuration options
type NoShadowOptions struct {
	Builtins             bool     `json:"builtinGlobals"`
	Hoist                string   `json:"hoist"` // "all", "functions", "never"
	Allow                []string `json:"allow"`
	IgnoreTypeValueShadow bool    `json:"ignoreTypeValueShadow"`
	IgnoreOnInitialization bool   `json:"ignoreOnInitialization"`
}

var NoShadowRule = rule.CreateRule(rule.Rule{
	Name: "no-shadow",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoShadowOptions{
			Builtins:              false,
			Hoist:                 "functions",
			Allow:                 []string{},
			IgnoreTypeValueShadow: true,
			IgnoreOnInitialization: false,
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
				if builtins, ok := optsMap["builtinGlobals"].(bool); ok {
					opts.Builtins = builtins
				}
				if hoist, ok := optsMap["hoist"].(string); ok {
					opts.Hoist = hoist
				}
				if allowVal, ok := optsMap["allow"].([]interface{}); ok {
					for _, pattern := range allowVal {
						if str, ok := pattern.(string); ok {
							opts.Allow = append(opts.Allow, str)
						}
					}
				}
				if ignoreTypeValueShadow, ok := optsMap["ignoreTypeValueShadow"].(bool); ok {
					opts.IgnoreTypeValueShadow = ignoreTypeValueShadow
				}
				if ignoreOnInitialization, ok := optsMap["ignoreOnInitialization"].(bool); ok {
					opts.IgnoreOnInitialization = ignoreOnInitialization
				}
			}
		}

		// TODO: Implement shadow variable detection
		// This rule needs to:
		// 1. Track variable declarations in each scope
		// 2. Detect when a variable shadows another from an outer scope
		// 3. Handle TypeScript-specific cases (type/value shadowing)
		// 4. Report shadowing violations

		return rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				// TODO: Implement shadow checking
			},
		}
	},
})

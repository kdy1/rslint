package no_type_alias

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoTypeAliasOptions represents the configuration options
type NoTypeAliasOptions struct {
	AllowAliases             string   `json:"allowAliases"` // "always", "never", "in-unions", "in-intersections", "in-unions-and-intersections"
	AllowCallbacks           string   `json:"allowCallbacks"`
	AllowConditionalTypes    string   `json:"allowConditionalTypes"`
	AllowConstructors        string   `json:"allowConstructors"`
	AllowLiterals            string   `json:"allowLiterals"`
	AllowMappedTypes         string   `json:"allowMappedTypes"`
	AllowTupleTypes          string   `json:"allowTupleTypes"`
	AllowGenerics            string   `json:"allowGenerics"`
}

var NoTypeAliasRule = rule.CreateRule(rule.Rule{
	Name: "no-type-alias",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoTypeAliasOptions{
			AllowAliases:          "never",
			AllowCallbacks:        "never",
			AllowConditionalTypes: "never",
			AllowConstructors:     "never",
			AllowLiterals:         "never",
			AllowMappedTypes:      "never",
			AllowTupleTypes:       "never",
			AllowGenerics:         "never",
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
				if allowAliases, ok := optsMap["allowAliases"].(string); ok {
					opts.AllowAliases = allowAliases
				}
				if allowCallbacks, ok := optsMap["allowCallbacks"].(string); ok {
					opts.AllowCallbacks = allowCallbacks
				}
				if allowConditionalTypes, ok := optsMap["allowConditionalTypes"].(string); ok {
					opts.AllowConditionalTypes = allowConditionalTypes
				}
				if allowConstructors, ok := optsMap["allowConstructors"].(string); ok {
					opts.AllowConstructors = allowConstructors
				}
				if allowLiterals, ok := optsMap["allowLiterals"].(string); ok {
					opts.AllowLiterals = allowLiterals
				}
				if allowMappedTypes, ok := optsMap["allowMappedTypes"].(string); ok {
					opts.AllowMappedTypes = allowMappedTypes
				}
				if allowTupleTypes, ok := optsMap["allowTupleTypes"].(string); ok {
					opts.AllowTupleTypes = allowTupleTypes
				}
				if allowGenerics, ok := optsMap["allowGenerics"].(string); ok {
					opts.AllowGenerics = allowGenerics
				}
			}
		}

		// TODO: Implement type alias restrictions
		// This rule disallows the use of type aliases with various exceptions
		// 1. Check type alias declarations
		// 2. Analyze the type being aliased
		// 3. Determine if it matches allowed patterns
		// 4. Report violations suggesting interface or other alternatives

		return rule.RuleListeners{
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				// TODO: Check if type alias should be disallowed
			},
		}
	},
})

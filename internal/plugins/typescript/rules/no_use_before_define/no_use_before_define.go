package no_use_before_define

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUseBeforeDefineOptions represents the configuration options
type NoUseBeforeDefineOptions struct {
	Functions         string   `json:"functions"`         // "nofunc" to ignore functions
	Classes           bool     `json:"classes"`
	Variables         bool     `json:"variables"`
	Enums             bool     `json:"enums"`
	TypeAliases       bool     `json:"typeAliases"`
	IgnoreTypeReferences bool  `json:"ignoreTypeReferences"`
}

var NoUseBeforeDefineRule = rule.CreateRule(rule.Rule{
	Name: "no-use-before-define",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoUseBeforeDefineOptions{
			Functions:            "",
			Classes:              true,
			Variables:            true,
			Enums:                true,
			TypeAliases:          true,
			IgnoreTypeReferences: false,
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
				if functions, ok := optsMap["functions"].(string); ok {
					opts.Functions = functions
				}
				if classes, ok := optsMap["classes"].(bool); ok {
					opts.Classes = classes
				}
				if variables, ok := optsMap["variables"].(bool); ok {
					opts.Variables = variables
				}
				if enums, ok := optsMap["enums"].(bool); ok {
					opts.Enums = enums
				}
				if typeAliases, ok := optsMap["typeAliases"].(bool); ok {
					opts.TypeAliases = typeAliases
				}
				if ignoreTypeReferences, ok := optsMap["ignoreTypeReferences"].(bool); ok {
					opts.IgnoreTypeReferences = ignoreTypeReferences
				}
			}
		}

		// TODO: Implement tracking of declarations and references
		// This rule needs to:
		// 1. Track all declarations in the current scope with their positions
		// 2. Check each reference to see if it appears before its declaration
		// 3. Handle TypeScript-specific cases (enums, type aliases, etc.)
		// 4. Report uses that occur before definitions

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				// TODO: Implement use-before-define checking
			},
		}
	},
})

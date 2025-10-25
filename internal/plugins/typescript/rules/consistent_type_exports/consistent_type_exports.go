package consistent_type_exports

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// ConsistentTypeExportsOptions defines the configuration options for this rule
type ConsistentTypeExportsOptions struct {
	FixMixedExportsWithInlineTypeSpecifier bool `json:"fixMixedExportsWithInlineTypeSpecifier"` // default: false
}

// parseOptions parses and validates the rule options
func parseOptions(options any) ConsistentTypeExportsOptions {
	opts := ConsistentTypeExportsOptions{
		FixMixedExportsWithInlineTypeSpecifier: false,
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
		if v, ok := optsMap["fixMixedExportsWithInlineTypeSpecifier"].(bool); ok {
			opts.FixMixedExportsWithInlineTypeSpecifier = v
		}
	}

	return opts
}

// ConsistentTypeExportsRule implements the consistent-type-exports rule
// Enforce consistent usage of type exports
var ConsistentTypeExportsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-type-exports",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)
	_ = opts

	return rule.RuleListeners{
		ast.KindExportDeclaration: func(node *ast.Node) {
			// TODO: Implement full logic
			// This is a placeholder implementation for the draft PR
			// Full implementation would:
			// 1. Check if export declaration exports types that should use "export type"
			// 2. Use TypeChecker to determine if exported symbols are types or values
			// 3. Handle mixed exports (both types and values in same export)
			// 4. Based on opts.FixMixedExportsWithInlineTypeSpecifier:
			//    - true: export { type Button, ButtonProps } (inline type specifiers)
			//    - false: export { Button }; export type { ButtonProps } (separate)
			// 5. Handle star exports: export * from '...' vs export type * from '...'
			// 6. Handle namespace exports: export * as ns from '...'
			// 7. Skip exports from unknown modules
			// 8. Preserve comments and formatting
			//
			// See TypeScript-ESLint implementation:
			// https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/src/rules/consistent-type-exports.ts
		},
	}
}

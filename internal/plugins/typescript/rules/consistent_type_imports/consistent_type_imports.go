package consistent_type_imports

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// ConsistentTypeImportsOptions defines the configuration options for this rule
type ConsistentTypeImportsOptions struct {
	Prefer                  string `json:"prefer"`                  // "type-imports" or "no-type-imports"
	FixStyle                string `json:"fixStyle"`                // "separate-type-imports" or "inline-type-imports"
	DisallowTypeAnnotations bool   `json:"disallowTypeAnnotations"` // default: true
}

// parseOptions parses and validates the rule options
func parseOptions(options any) ConsistentTypeImportsOptions {
	opts := ConsistentTypeImportsOptions{
		Prefer:                  "type-imports",
		FixStyle:                "separate-type-imports",
		DisallowTypeAnnotations: true,
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
		if v, ok := optsMap["prefer"].(string); ok {
			opts.Prefer = v
		}
		if v, ok := optsMap["fixStyle"].(string); ok {
			opts.FixStyle = v
		}
		if v, ok := optsMap["disallowTypeAnnotations"].(bool); ok {
			opts.DisallowTypeAnnotations = v
		}
	}

	return opts
}

// ConsistentTypeImportsRule implements the consistent-type-imports rule
// Enforce consistent usage of type imports
var ConsistentTypeImportsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-type-imports",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)
	_ = opts

	return rule.RuleListeners{
		ast.KindImportDeclaration: func(node *ast.Node) {
			// TODO: Implement full logic
			// This is a placeholder implementation for the draft PR
			// Full implementation would:
			// 1. Analyze if imports are only used in type positions using TypeChecker
			// 2. Check if import has "type" keyword (ImportClause.IsTypeOnly)
			// 3. Based on opts.Prefer, determine if we need to add/remove "type" keyword
			// 4. Generate appropriate fixes based on opts.FixStyle:
			//    - separate-type-imports: import type { Foo } from '...'
			//    - inline-type-imports: import { type Foo } from '...'
			// 5. Handle mixed imports (some types, some values)
			// 6. Handle default imports and namespace imports
			// 7. Preserve comments and formatting
			//
			// See TypeScript-ESLint implementation:
			// https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/src/rules/consistent-type-imports.ts
		},
	}
}

package prefer_nullish_coalescing

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// PreferNullishCoalescingRule implements the prefer-nullish-coalescing rule
// Enforce nullish coalescing operator over logical or
var PreferNullishCoalescingRule = rule.CreateRule(rule.Rule{
	Name: "prefer-nullish-coalescing",
	Run:  run,
})

type Options struct {
	IgnoreTernaryTests            bool                 `json:"ignoreTernaryTests"`
	IgnorePrimitives              *PrimitivesOptions   `json:"ignorePrimitives"`
	IgnoreMixedLogicalExpressions bool                 `json:"ignoreMixedLogicalExpressions"`
}

type PrimitivesOptions struct {
	String  bool `json:"string"`
	Number  bool `json:"number"`
	Boolean bool `json:"boolean"`
	Bigint  bool `json:"bigint"`
}

func parseOptions(options any) Options {
	opts := Options{
		IgnoreTernaryTests:            false,
		IgnoreMixedLogicalExpressions: false,
	}

	if options == nil {
		return opts
	}

	var optsMap map[string]interface{}
	var ok bool

	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, ok = optArray[0].(map[string]interface{})
	} else {
		optsMap, ok = options.(map[string]interface{})
	}

	if ok {
		if v, ok := optsMap["ignoreTernaryTests"].(bool); ok {
			opts.IgnoreTernaryTests = v
		}
		if v, ok := optsMap["ignoreMixedLogicalExpressions"].(bool); ok {
			opts.IgnoreMixedLogicalExpressions = v
		}
		if primMap, ok := optsMap["ignorePrimitives"].(map[string]interface{}); ok {
			prim := &PrimitivesOptions{}
			if v, ok := primMap["string"].(bool); ok {
				prim.String = v
			}
			if v, ok := primMap["number"].(bool); ok {
				prim.Number = v
			}
			if v, ok := primMap["boolean"].(bool); ok {
				prim.Boolean = v
			}
			if v, ok := primMap["bigint"].(bool); ok {
				prim.Bigint = v
			}
			opts.IgnorePrimitives = prim
		}
	}
	return opts
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	if ctx.TypeChecker == nil {
		// Silently return - rule requires type checking
		return rule.RuleListeners{}
	}

	opts := parseOptions(options)
	_ = opts // TODO: Use options in implementation

	return rule.RuleListeners{
		ast.KindBinaryExpression: func(node *ast.Node) {
			// TODO: Implement full logic
			// This is a simplified placeholder implementation
			// Full implementation would:
			// 1. Check for || operator
			// 2. Check if left side has nullish type using TypeChecker
			// 3. Report with fix to use ?? instead

			// Example pattern to detect:
			// const x: string | undefined = ...
			// x || 'default'  // Should be: x ?? 'default'
		},
		ast.KindConditionalExpression: func(node *ast.Node) {
			if opts.IgnoreTernaryTests {
				return
			}

			// TODO: Implement ternary pattern detection
			// Pattern: x ? x : y or !x ? y : x
			// Should suggest: x ?? y
		},
	}
}

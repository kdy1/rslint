package prefer_optional_chain

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// PreferOptionalChainRule implements the prefer-optional-chain rule
// Enforce optional chain expressions over chained ternaries
var PreferOptionalChainRule = rule.CreateRule(rule.Rule{
	Name: "prefer-optional-chain",
	Run:  run,
})

type Options struct {
	RequireNullish                                            bool `json:"requireNullish"`
	AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing bool `json:"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing"`
}

func parseOptions(options any) Options {
	opts := Options{
		RequireNullish: false,
		AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing: false,
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
		if v, ok := optsMap["requireNullish"].(bool); ok {
			opts.RequireNullish = v
		}
		if v, ok := optsMap["allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing"].(bool); ok {
			opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing = v
		}
	}
	return opts
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	if ctx.TypeChecker == nil {
		return rule.RuleListeners{}
	}

	opts := parseOptions(options)
	_ = opts // TODO: Use options in implementation

	return rule.RuleListeners{
		ast.KindBinaryExpression: func(node *ast.Node) {
			// TODO: Implement full logic
			// This is a simplified placeholder implementation
			// Full implementation would:
			// 1. Check for && chains: foo && foo.a && foo.a.b
			// 2. Check for || with empty object: (foo || {}).a
			// 3. Suggest optional chain: foo?.a?.b

			// Example patterns to detect:
			// foo && foo.bar && foo.bar.baz  // Should be: foo?.bar?.baz
			// (foo || {}).bar                 // Should be: foo?.bar
		},
	}
}

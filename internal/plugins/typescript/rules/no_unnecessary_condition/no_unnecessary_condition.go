package no_unnecessary_condition

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoUnnecessaryConditionOptions defines the configuration options for this rule
type NoUnnecessaryConditionOptions struct {
	AllowConstantLoopConditions bool `json:"allowConstantLoopConditions"`
	AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing bool `json:"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) NoUnnecessaryConditionOptions {
	opts := NoUnnecessaryConditionOptions{
		AllowConstantLoopConditions: false,
		AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: false,
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
		if v, ok := optsMap["allowConstantLoopConditions"].(bool); ok {
			opts.AllowConstantLoopConditions = v
		}
		if v, ok := optsMap["allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"].(bool); ok {
			opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = v
		}
	}

	return opts
}

// NoUnnecessaryConditionRule implements the no-unnecessary-condition rule
// Disallows conditionals where the type is always truthy or always falsy
var NoUnnecessaryConditionRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-condition",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)
	_ = opts // Will be used in implementation

	return rule.RuleListeners{
		ast.KindIfStatement: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: Implement type-based analysis for if statement conditions
			// Check if condition is always truthy or always falsy using type information
		},
		ast.KindConditionalExpression: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: Implement type-based analysis for ternary expressions
		},
		ast.KindWhileStatement: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: Implement type-based analysis for while loop conditions
			// Unless allowConstantLoopConditions is true
		},
		ast.KindDoStatement: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: Implement type-based analysis for do-while loop conditions
		},
		ast.KindForStatement: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: Implement type-based analysis for for loop conditions
		},
		ast.KindBinaryExpression: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: Implement type-based analysis for logical operators (&&, ||, ??)
		},
	}
}

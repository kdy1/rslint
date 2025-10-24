package strict_boolean_expressions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// StrictBooleanExpressionsOptions defines the configuration options for this rule
type StrictBooleanExpressionsOptions struct {
	AllowString            bool `json:"allowString"`
	AllowNumber            bool `json:"allowNumber"`
	AllowNullableObject    bool `json:"allowNullableObject"`
	AllowNullableBoolean   bool `json:"allowNullableBoolean"`
	AllowNullableString    bool `json:"allowNullableString"`
	AllowNullableNumber    bool `json:"allowNullableNumber"`
	AllowNullableEnum      bool `json:"allowNullableEnum"`
	AllowAny               bool `json:"allowAny"`
	AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing bool `json:"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) StrictBooleanExpressionsOptions {
	opts := StrictBooleanExpressionsOptions{
		AllowString:          false,
		AllowNumber:          false,
		AllowNullableObject:  false,
		AllowNullableBoolean: false,
		AllowNullableString:  false,
		AllowNullableNumber:  false,
		AllowNullableEnum:    false,
		AllowAny:             false,
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
		if v, ok := optsMap["allowString"].(bool); ok {
			opts.AllowString = v
		}
		if v, ok := optsMap["allowNumber"].(bool); ok {
			opts.AllowNumber = v
		}
		if v, ok := optsMap["allowNullableObject"].(bool); ok {
			opts.AllowNullableObject = v
		}
		if v, ok := optsMap["allowNullableBoolean"].(bool); ok {
			opts.AllowNullableBoolean = v
		}
		if v, ok := optsMap["allowNullableString"].(bool); ok {
			opts.AllowNullableString = v
		}
		if v, ok := optsMap["allowNullableNumber"].(bool); ok {
			opts.AllowNullableNumber = v
		}
		if v, ok := optsMap["allowNullableEnum"].(bool); ok {
			opts.AllowNullableEnum = v
		}
		if v, ok := optsMap["allowAny"].(bool); ok {
			opts.AllowAny = v
		}
		if v, ok := optsMap["allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"].(bool); ok {
			opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = v
		}
	}

	return opts
}

// StrictBooleanExpressionsRule implements the strict-boolean-expressions rule
// Disallows non-boolean types in boolean contexts
var StrictBooleanExpressionsRule = rule.CreateRule(rule.Rule{
	Name: "strict-boolean-expressions",
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

			// TODO: Check if condition is strictly boolean
			// Report if condition type is not boolean (unless allowed by options)
		},
		ast.KindConditionalExpression: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: Check if condition is strictly boolean
		},
		ast.KindWhileStatement: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: Check if condition is strictly boolean
		},
		ast.KindDoStatement: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: Check if condition is strictly boolean
		},
		ast.KindForStatement: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: Check if condition is strictly boolean
		},
		ast.KindBinaryExpression: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: For logical operators (&&, ||), check operands are strictly boolean
		},
		ast.KindPrefixUnaryExpression: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: For negation operator (!), check operand is strictly boolean
		},
	}
}

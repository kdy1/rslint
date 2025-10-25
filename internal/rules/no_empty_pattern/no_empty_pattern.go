package no_empty_pattern

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoEmptyPatternOptions struct {
	AllowObjectPatternsAsParameters bool `json:"allowObjectPatternsAsParameters"`
}

// NoEmptyPatternRule implements the no-empty-pattern rule
// Disallow empty destructuring patterns
var NoEmptyPatternRule = rule.Rule{
	Name: "no-empty-pattern",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := NoEmptyPatternOptions{
		AllowObjectPatternsAsParameters: false,
	}

	// Parse options with dual-format support (handles both array and object formats)
	if options != nil {
		var optsMap map[string]interface{}
		var ok bool

		// Handle array format: [{ option: value }]
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			optsMap, ok = optArray[0].(map[string]interface{})
		} else {
			// Handle direct object format: { option: value }
			optsMap, ok = options.(map[string]interface{})
		}

		if ok {
			if allowObjectPatternsAsParameters, ok := optsMap["allowObjectPatternsAsParameters"].(bool); ok {
				opts.AllowObjectPatternsAsParameters = allowObjectPatternsAsParameters
			}
		}
	}

	return rule.RuleListeners{
		ast.KindObjectBindingPattern: func(node *ast.Node) {
			objectPattern := node.AsObjectBindingPattern()
			if objectPattern == nil {
				return
			}

			// Check if pattern has elements
			if objectPattern.Elements != nil && len(objectPattern.Elements.Nodes) > 0 {
				return
			}

			// If option is enabled, allow empty object patterns as function parameters
			if opts.AllowObjectPatternsAsParameters && isParameter(node) {
				return
			}

			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "unexpected",
				Description: "Unexpected empty object pattern.",
			})
		},
		ast.KindArrayBindingPattern: func(node *ast.Node) {
			arrayPattern := node.AsArrayBindingPattern()
			if arrayPattern == nil {
				return
			}

			// Check if pattern has elements
			if arrayPattern.Elements != nil && len(arrayPattern.Elements.Nodes) > 0 {
				return
			}

			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "unexpected",
				Description: "Unexpected empty array pattern.",
			})
		},
	}
}

// isParameter checks if a binding pattern is a function parameter
func isParameter(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// Check if parent is a parameter
	if parent.Kind == ast.KindParameter {
		return true
	}

	// Check if parent is a binding element, then check its parent
	if parent.Kind == ast.KindBindingElement {
		return isParameter(parent)
	}

	return false
}

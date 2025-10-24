package member_ordering

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// MemberOrderingOptions represents the configuration options
type MemberOrderingOptions struct {
	Default []string `json:"default"`
	Classes []string `json:"classes"`
	ClassExpressions []string `json:"classExpressions"`
	Interfaces []string `json:"interfaces"`
	TypeLiterals []string `json:"typeLiterals"`
}

var MemberOrderingRule = rule.CreateRule(rule.Rule{
	Name: "member-ordering",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := MemberOrderingOptions{
			Default: []string{},
			Classes: []string{},
			ClassExpressions: []string{},
			Interfaces: []string{},
			TypeLiterals: []string{},
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
				if defaultVal, ok := optsMap["default"].([]interface{}); ok {
					for _, item := range defaultVal {
						if str, ok := item.(string); ok {
							opts.Default = append(opts.Default, str)
						}
					}
				}
				// Parse other options similarly
			}
		}

		// TODO: Implement member ordering checking
		// This rule enforces a consistent order for class members:
		// 1. Parse the desired ordering from options
		// 2. Track actual order of members in classes/interfaces
		// 3. Compare against expected order
		// 4. Report violations with suggestions for reordering

		return rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				// TODO: Check member ordering in classes
			},
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				// TODO: Check member ordering in interfaces
			},
		}
	},
})

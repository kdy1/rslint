package no_invalid_void_type

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoInvalidVoidTypeOptions struct {
	AllowInGenericTypeArguments interface{} `json:"allowInGenericTypeArguments"`
	AllowAsThisParameter        bool        `json:"allowAsThisParameter"`
}

var NoInvalidVoidTypeRule = rule.CreateRule(rule.Rule{
	Name: "no-invalid-void-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoInvalidVoidTypeOptions{
			AllowInGenericTypeArguments: true,
			AllowAsThisParameter:        false,
		}

		// Parse options with dual-format support
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
				if allowInGenericTypeArguments, ok := optsMap["allowInGenericTypeArguments"]; ok {
					opts.AllowInGenericTypeArguments = allowInGenericTypeArguments
				}
				if allowAsThisParameter, ok := optsMap["allowAsThisParameter"].(bool); ok {
					opts.AllowAsThisParameter = allowAsThisParameter
				}
			}
		}

		// Helper to check if void is allowed in generic context
		isAllowedInGeneric := func() bool {
			// If allowInGenericTypeArguments is false, never allow
			if allow, ok := opts.AllowInGenericTypeArguments.(bool); ok && !allow {
				return false
			}

			// If it's true (default), allow all generics
			if allow, ok := opts.AllowInGenericTypeArguments.(bool); ok && allow {
				return true
			}

			// If it's an array/whitelist, this would need more complex checking
			// For now, simplified implementation
			return true
		}

		// Helper to check if node is in valid context
		isValidVoidContext := func(node *ast.Node) bool {
			parent := node.Parent
			if parent == nil {
				return false
			}

			// Allow in return types
			if parent.Kind == ast.KindFunctionDeclaration ||
				parent.Kind == ast.KindMethodDeclaration ||
				parent.Kind == ast.KindArrowFunction ||
				parent.Kind == ast.KindFunctionExpression {
				return true
			}

			// Allow in union with never
			if parent.Kind == ast.KindUnionType {
				// Simplified check - allow void in union types for now
				return true
			}

			// Allow in generic type arguments (Promise<void>, etc.)
			if parent.Kind == ast.KindTypeReference {
				return isAllowedInGeneric()
			}

			return false
		}

		return rule.RuleListeners{
			ast.KindVoidKeyword: func(node *ast.Node) {
				// Check if in valid context
				if isValidVoidContext(node) {
					return
				}

				// Report invalid void usage
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "invalidVoidNotReturn",
					Description: "`void` is only valid as a return type or generic type argument.",
				})
			},
		}
	},
})

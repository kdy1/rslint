package require_array_sort_compare

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildRequireCompareMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "requireCompare",
		Description: "Require 'compare' argument.",
	}
}

type RequireArraySortCompareOptions struct {
	IgnoreStringArrays bool
}

var RequireArraySortCompareRule = rule.CreateRule(rule.Rule{
	Name: "require-array-sort-compare",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := RequireArraySortCompareOptions{
			IgnoreStringArrays: true, // Default value is true
		}

		// Parse options from map format
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
				if val, exists := optsMap["ignoreStringArrays"]; exists {
					if boolVal, isBool := val.(bool); isBool {
						opts.IgnoreStringArrays = boolVal
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				expr := node.AsCallExpression()
				if len(expr.Arguments.Nodes) != 0 {
					return
				}
				callee := expr.Expression

				if !ast.IsAccessExpression(callee) {
					return
				}

				if propertyName, found := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee); !found || (propertyName != "sort" && propertyName != "toSorted") {
					return
				}

				calleeObjType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, callee.Expression())

				if opts.IgnoreStringArrays && checker.Checker_isArrayOrTupleType(ctx.TypeChecker, calleeObjType) {
					if utils.Every(checker.Checker_getTypeArguments(ctx.TypeChecker, calleeObjType), func(t *checker.Type) bool {
						return utils.IsTypeFlagSet(t, checker.TypeFlagsString)
					}) {
						return
					}
				}

				if utils.Every(utils.UnionTypeParts(calleeObjType), func(t *checker.Type) bool {
					return checker.Checker_isArrayOrTupleType(ctx.TypeChecker, t)
				}) {
					ctx.ReportNode(node, buildRequireCompareMessage())
				}
			},
		}
	},
})

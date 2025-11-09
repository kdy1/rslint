package no_meaningless_void_operator

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type NoMeaninglessVoidOperatorOptions struct {
	CheckNever bool `json:"checkNever"`
}

var NoMeaninglessVoidOperatorRule = rule.CreateRule(rule.Rule{
	Name: "no-meaningless-void-operator",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoMeaninglessVoidOperatorOptions{
			CheckNever: false,
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
				if checkNever, ok := optsMap["checkNever"].(bool); ok {
					opts.CheckNever = checkNever
				}
			}
		}

		return rule.RuleListeners{
			ast.KindVoidExpression: func(node *ast.Node) {
				voidExpr := node.AsVoidExpression()
				if voidExpr == nil {
					return
				}

				// Get the type of the expression being voided
				exprType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, voidExpr.Expression)
				if exprType == nil {
					return
				}

				// Check if the expression is already void
				isVoid := utils.IsTypeFlagSet(exprType, checker.TypeFlagsVoid)
				isUndefined := utils.IsTypeFlagSet(exprType, checker.TypeFlagsUndefined)
				isNever := utils.IsTypeFlagSet(exprType, checker.TypeFlagsNever)

				// If checkNever is enabled and type is never, report with suggestion only
				if opts.CheckNever && isNever {
					message := rule.RuleMessage{
						Id:          "meaninglessVoidOperator",
						Description: "void operator is meaningless on a value that is never.",
					}

					suggestion := rule.RuleSuggestion{
						Desc: rule.RuleMessage{
							Id:          "removeVoid",
							Description: "Remove void operator.",
						},
						Fix: rule.RuleFixReplace(ctx.SourceFile, node, ctx.SourceFile.Text()[voidExpr.Expression.Pos():voidExpr.Expression.End()]),
					}

					ctx.ReportNodeWithSuggestions(node, message, suggestion)
					return
				}

				// If type is void or undefined (but not never when checkNever is false), report with auto-fix
				if (isVoid || isUndefined) && !isNever {
					message := rule.RuleMessage{
						Id:          "meaninglessVoidOperator",
						Description: "void operator is meaningless on a value that is already void or undefined.",
					}

					fix := rule.RuleFixReplace(ctx.SourceFile, node, ctx.SourceFile.Text()[voidExpr.Expression.Pos():voidExpr.Expression.End()])

					ctx.ReportNodeWithFixes(node, message, fix)
					return
				}
			},
		}
	},
})

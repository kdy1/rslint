package no_prototype_builtins

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoPrototypeBuiltinsRule implements the no-prototype-builtins rule
// Disallow calling Object.prototype methods directly
var NoPrototypeBuiltinsRule = rule.Rule{
	Name: "no-prototype-builtins",
	Run:  run,
}

// prototypeMethods lists the Object.prototype methods we're checking
var prototypeMethods = map[string]bool{
	"hasOwnProperty":       true,
	"isPrototypeOf":        true,
	"propertyIsEnumerable": true,
}

// getMethodName extracts the method name from a property access or element access expression
func getMethodName(callExpr *ast.CallExpression) (string, bool) {
	if callExpr == nil || callExpr.Expression == nil {
		return "", false
	}

	expr := callExpr.Expression

	// Check for property access: obj.hasOwnProperty(...)
	if expr.Kind == ast.KindPropertyAccessExpression {
		pae := expr.AsPropertyAccessExpression()
		if pae != nil && pae.Name() != nil {
			methodName := pae.Name().Text()
			return methodName, prototypeMethods[methodName]
		}
	}

	// Check for element access: obj['hasOwnProperty'](...)
	if expr.Kind == ast.KindElementAccessExpression {
		eae := expr.AsElementAccessExpression()
		if eae != nil && eae.ArgumentExpression != nil {
			arg := eae.ArgumentExpression
			// Check for string literal
			if arg.Kind == ast.KindStringLiteral {
				if strLit := arg.AsStringLiteral(); strLit != nil {
					methodName := strLit.Text
					return methodName, prototypeMethods[methodName]
				}
			}
			// Check for no-substitution template literal
			if arg.Kind == ast.KindNoSubstitutionTemplateLiteral {
				if tmpl := arg.AsNoSubstitutionTemplateLiteral(); tmpl != nil {
					methodName := tmpl.Text
					return methodName, prototypeMethods[methodName]
				}
			}
		}
	}

	return "", false
}

// hasOptionalChaining checks if the call expression or its parents use optional chaining
func hasOptionalChaining(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check if this is an optional call
	if node.Kind == ast.KindCallExpression {
		if callExpr := node.AsCallExpression(); callExpr != nil {
			// Check the QuestionDotToken field if available
			// For now, we'll check the expression
			if callExpr.Expression != nil {
				expr := callExpr.Expression
				if expr.Kind == ast.KindPropertyAccessExpression {
					if pae := expr.AsPropertyAccessExpression(); pae != nil {
						// Check if expression chain contains optional chaining
						if pae.Expression != nil && hasOptionalChaining(pae.Expression) {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindCallExpression: func(node *ast.Node) {
			callExpr := node.AsCallExpression()
			if callExpr == nil || callExpr.Expression == nil {
				return
			}

			methodName, isPrototypeMethod := getMethodName(callExpr)
			if !isPrototypeMethod {
				return
			}

			// Check for optional chaining - we can't easily fix those
			if hasOptionalChaining(node) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "prototypeBuildIn",
					Description: "Do not access Object.prototype method '" + methodName + "' from target object.",
				})
				return
			}

			// Get the object being called on
			var objectExpr *ast.Node
			expr := callExpr.Expression

			if expr.Kind == ast.KindPropertyAccessExpression {
				if pae := expr.AsPropertyAccessExpression(); pae != nil {
					objectExpr = pae.Expression
				}
			} else if expr.Kind == ast.KindElementAccessExpression {
				if eae := expr.AsElementAccessExpression(); eae != nil {
					objectExpr = eae.Expression
				}
			}

			if objectExpr == nil {
				return
			}

			// Build the fix
			text := ctx.SourceFile.Text()
			objectRange := utils.TrimNodeTextRange(ctx.SourceFile, objectExpr)
			objectText := text[objectRange.Pos():objectRange.End()]

			// Check if object needs parentheses (e.g., for comma expressions)
			needsParens := false
			if objectExpr.Kind == ast.KindBinaryExpression ||
				objectExpr.Kind == ast.KindCommaListExpression {
				needsParens = true
			}

			if needsParens {
				objectText = "(" + objectText + ")"
			}

			// Get the arguments
			var argsText string
			if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
				firstArg := callExpr.Arguments.Nodes[0]
				lastArg := callExpr.Arguments.Nodes[len(callExpr.Arguments.Nodes)-1]
				argsRange := core.NewTextRange(
					utils.TrimNodeTextRange(ctx.SourceFile, firstArg).Pos(),
					utils.TrimNodeTextRange(ctx.SourceFile, lastArg).End(),
				)
				argsText = text[argsRange.Pos():argsRange.End()]
			}

			// Build the replacement
			var replacement string
			if argsText != "" {
				replacement = "Object.prototype." + methodName + ".call(" + objectText + ", " + argsText + ")"
			} else {
				replacement = "Object.prototype." + methodName + ".call(" + objectText + ")"
			}

			// Find the complete call expression including any chained property access after it
			// e.g., foo.hasOwnProperty('bar').baz should preserve .baz
			nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)

			// Check if there's anything after the call expression
			afterCall := ""
			if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression {
				parentRange := utils.TrimNodeTextRange(ctx.SourceFile, node.Parent)
				afterCall = text[nodeRange.End():parentRange.End()]
			}

			if afterCall != "" {
				replacement = replacement + afterCall
				// Report on parent to include the full expression
				ctx.ReportNodeWithFixes(node.Parent, rule.RuleMessage{
					Id:          "prototypeBuildIn",
					Description: "Do not access Object.prototype method '" + methodName + "' from target object.",
				}, rule.RuleFixReplace(ctx.SourceFile, node.Parent, replacement))
			} else {
				ctx.ReportNodeWithFixes(node, rule.RuleMessage{
					Id:          "prototypeBuildIn",
					Description: "Do not access Object.prototype method '" + methodName + "' from target object.",
				}, rule.RuleFixReplace(ctx.SourceFile, node, replacement))
			}
		},
	}
}

// isStringLiteralOrTemplate checks if a node is a string literal or template
func isStringLiteralOrTemplate(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindStringLiteral ||
		node.Kind == ast.KindNoSubstitutionTemplateLiteral
}

// getStringValue extracts the string value from a string literal or template
func getStringValue(node *ast.Node, sourceText string) string {
	if node == nil {
		return ""
	}

	rng := utils.TrimNodeTextRange(nil, node)
	text := sourceText[rng.Pos():rng.End()]

	// Remove quotes or backticks
	if len(text) >= 2 {
		if (strings.HasPrefix(text, "'") && strings.HasSuffix(text, "'")) ||
			(strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"")) ||
			(strings.HasPrefix(text, "`") && strings.HasSuffix(text, "`")) {
			return text[1 : len(text)-1]
		}
	}
	return text
}

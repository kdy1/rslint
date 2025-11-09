package no_inferrable_types

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoInferrableTypesOptions struct {
	IgnoreParameters bool `json:"ignoreParameters"`
	IgnoreProperties bool `json:"ignoreProperties"`
}

var NoInferrableTypesRule = rule.CreateRule(rule.Rule{
	Name: "no-inferrable-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoInferrableTypesOptions{
			IgnoreParameters: false,
			IgnoreProperties: false,
		}

		// Parse options
		if options != nil {
			var optsMap map[string]interface{}
			if optsArray, ok := options.([]interface{}); ok && len(optsArray) > 0 {
				if opts, ok := optsArray[0].(map[string]interface{}); ok {
					optsMap = opts
				}
			} else if opts, ok := options.(map[string]interface{}); ok {
				optsMap = opts
			}

			if optsMap != nil {
				if ignoreParams, ok := optsMap["ignoreParameters"].(bool); ok {
					opts.IgnoreParameters = ignoreParams
				}
				if ignoreProps, ok := optsMap["ignoreProperties"].(bool); ok {
					opts.IgnoreProperties = ignoreProps
				}
			}
		}

		// Helper function to check if a type annotation matches an inferrable type
		isInferrableType := func(typeNode *ast.Node, initializer *ast.Node) (bool, string) {
			if typeNode == nil || initializer == nil {
				return false, ""
			}

			// Get the type reference text
			var typeText string
			if typeNode.Kind == ast.KindTypeReference {
				typeRef := typeNode.AsTypeReferenceNode()
				if typeRef != nil && typeRef.TypeName != nil {
					if typeRef.TypeName.Kind == ast.KindIdentifier {
						ident := typeRef.TypeName.AsIdentifier()
						if ident != nil {
							typeText = ident.Text
						}
					}
				}
			} else if typeNode.Kind == ast.KindBigIntKeyword {
				typeText = "bigint"
			} else if typeNode.Kind == ast.KindBooleanKeyword {
				typeText = "boolean"
			} else if typeNode.Kind == ast.KindNumberKeyword {
				typeText = "number"
			} else if typeNode.Kind == ast.KindNullKeyword {
				typeText = "null"
			} else if typeNode.Kind == ast.KindStringKeyword {
				typeText = "string"
			} else if typeNode.Kind == ast.KindSymbolKeyword {
				typeText = "symbol"
			} else if typeNode.Kind == ast.KindUndefinedKeyword {
				typeText = "undefined"
			}

			if typeText == "" {
				return false, ""
			}

			// Handle unary expressions (e.g., -10, +10, !0)
			initExpr := initializer
			if initializer.Kind == ast.KindPrefixUnaryExpression {
				unary := initializer.AsPrefixUnaryExpression()
				if unary != nil && unary.Operand != nil {
					initExpr = unary.Operand
				}
			}

			// Check if the initializer matches the type
			switch typeText {
			case "bigint":
				if initExpr.Kind == ast.KindBigIntLiteral {
					return true, "bigint"
				}
				// Check for BigInt() or BigInt?.() calls
				if initExpr.Kind == ast.KindCallExpression {
					call := initExpr.AsCallExpression()
					if call != nil && call.Expression != nil {
						if call.Expression.Kind == ast.KindIdentifier {
							ident := call.Expression.AsIdentifier()
							if ident != nil && ident.Text == "BigInt" {
								return true, "bigint"
							}
						}
					}
				}

			case "boolean":
				if initExpr.Kind == ast.KindTrueKeyword || initExpr.Kind == ast.KindFalseKeyword {
					return true, "boolean"
				}
				// Check for Boolean() or Boolean?.() calls
				if initExpr.Kind == ast.KindCallExpression {
					call := initExpr.AsCallExpression()
					if call != nil && call.Expression != nil {
						if call.Expression.Kind == ast.KindIdentifier {
							ident := call.Expression.AsIdentifier()
							if ident != nil && ident.Text == "Boolean" {
								return true, "boolean"
							}
						}
					}
				}

			case "number":
				if initExpr.Kind == ast.KindNumericLiteral {
					return true, "number"
				}
				// Check for Infinity or NaN
				if initExpr.Kind == ast.KindIdentifier {
					ident := initExpr.AsIdentifier()
					if ident != nil && (ident.Text == "Infinity" || ident.Text == "NaN") {
						return true, "number"
					}
				}
				// Check for Number() or Number?.() calls
				if initExpr.Kind == ast.KindCallExpression {
					call := initExpr.AsCallExpression()
					if call != nil && call.Expression != nil {
						if call.Expression.Kind == ast.KindIdentifier {
							ident := call.Expression.AsIdentifier()
							if ident != nil && ident.Text == "Number" {
								return true, "number"
							}
						}
					}
				}

			case "null":
				if initExpr.Kind == ast.KindNullKeyword {
					return true, "null"
				}

			case "RegExp":
				if initExpr.Kind == ast.KindRegularExpressionLiteral {
					return true, "RegExp"
				}
				// Check for RegExp() or new RegExp() or RegExp?.() calls
				if initExpr.Kind == ast.KindCallExpression || initExpr.Kind == ast.KindNewExpression {
					var expr *ast.Node
					if initExpr.Kind == ast.KindCallExpression {
						call := initExpr.AsCallExpression()
						if call != nil {
							expr = call.Expression
						}
					} else if initExpr.Kind == ast.KindNewExpression {
						newExpr := initExpr.AsNewExpression()
						if newExpr != nil {
							expr = newExpr.Expression
						}
					}

					if expr != nil && expr.Kind == ast.KindIdentifier {
						ident := expr.AsIdentifier()
						if ident != nil && ident.Text == "RegExp" {
							return true, "RegExp"
						}
					}
				}

			case "string":
				if initExpr.Kind == ast.KindStringLiteral || initExpr.Kind == ast.KindNoSubstitutionTemplateLiteral {
					return true, "string"
				}
				// Check for String() or String?.() calls
				if initExpr.Kind == ast.KindCallExpression {
					call := initExpr.AsCallExpression()
					if call != nil && call.Expression != nil {
						if call.Expression.Kind == ast.KindIdentifier {
							ident := call.Expression.AsIdentifier()
							if ident != nil && ident.Text == "String" {
								return true, "string"
							}
						}
					}
				}

			case "symbol":
				// Check for Symbol() or Symbol?.() calls
				if initExpr.Kind == ast.KindCallExpression {
					call := initExpr.AsCallExpression()
					if call != nil && call.Expression != nil {
						if call.Expression.Kind == ast.KindIdentifier {
							ident := call.Expression.AsIdentifier()
							if ident != nil && ident.Text == "Symbol" {
								return true, "symbol"
							}
						}
					}
				}

			case "undefined":
				if initExpr.Kind == ast.KindUndefinedKeyword {
					return true, "undefined"
				}
				// Check for void expressions
				if initializer.Kind == ast.KindVoidExpression {
					return true, "undefined"
				}
			}

			return false, ""
		}

		// Check variable declaration
		checkVariableDeclaration := func(node *ast.Node) {
			varDecl := node.AsVariableDeclaration()
			if varDecl == nil {
				return
			}

			// Skip if no type annotation or no initializer
			if varDecl.Type == nil || varDecl.Initializer == nil {
				return
			}

			// Check if the type is "any" - this is allowed
			if varDecl.Type.Kind == ast.KindAnyKeyword {
				return
			}

			isInferrable, typeName := isInferrableType(varDecl.Type, varDecl.Initializer)
			if isInferrable {
				// Report the error on the identifier
				reportNode := varDecl.Name()
				if reportNode == nil {
					reportNode = node
				}

				// Create the fix
				// Find the position of the type annotation (after the identifier)
				typeStart := varDecl.Type.Pos()
				typeEnd := varDecl.Type.End()

				// We need to remove ": type" including the colon and spaces
				// Find the colon before the type
				nameEnd := varDecl.Name().End()
				colonPos := nameEnd

				// Find the colon between name and type
				sourceText := ctx.SourceFile.Text()
				for i := nameEnd; i < typeStart && i < len(sourceText); i++ {
					if sourceText[i] == ':' {
						colonPos = i
						break
					}
				}

				// Check if there's a definite assignment assertion (!)
				definiteToken := ""
				for i := nameEnd; i < colonPos && i < len(sourceText); i++ {
					if sourceText[i] == '!' {
						definiteToken = "!"
						break
					}
				}

				// Remove from after the name (or definite token) to the end of the type
				startRemove := nameEnd
				if definiteToken != "" {
					// Keep the name, remove the "!" and type annotation
					for i := nameEnd; i < colonPos && i < len(sourceText); i++ {
						if sourceText[i] == '!' {
							startRemove = i
							break
						}
					}
				} else {
					// Just remove from after the name
					startRemove = colonPos
				}

				fix := rule.RuleFixRemoveRange(core.NewTextRange(startRemove, typeEnd))

				ctx.ReportNodeWithFixes(reportNode, rule.RuleMessage{
					Id:          "noInferrableType",
					Description: "Type " + typeName + " trivially inferred from a " + typeName + " literal, remove type annotation.",
				}, fix)
			}
		}

		// Check function parameters
		checkParameter := func(node *ast.Node) {
			if opts.IgnoreParameters {
				return
			}

			param := node.AsParameterDeclaration()
			if param == nil {
				return
			}

			// Skip if no type annotation or no initializer
			if param.Type == nil || param.Initializer == nil {
				return
			}

			// Check if the type is "any" - this is allowed
			if param.Type.Kind == ast.KindAnyKeyword {
				return
			}

			isInferrable, typeName := isInferrableType(param.Type, param.Initializer)
			if isInferrable {
				reportNode := param.Name()
				if reportNode == nil {
					reportNode = node
				}

				// Create the fix
				// Remove the type annotation from parameter
				nameEnd := param.Name().End()
				typeStart := param.Type.Pos()
				typeEnd := param.Type.End()

				sourceText := ctx.SourceFile.Text()
				colonPos := nameEnd

				// Find the colon between name and type
				for i := nameEnd; i < typeStart && i < len(sourceText); i++ {
					if sourceText[i] == ':' {
						colonPos = i
						break
					}
				}

				// Check for optional token (?)
				hasOptional := false
				for i := nameEnd; i < colonPos && i < len(sourceText); i++ {
					if sourceText[i] == '?' {
						hasOptional = true
						break
					}
				}

				startRemove := colonPos
				if hasOptional {
					// Remove from the optional token
					for i := nameEnd; i < colonPos && i < len(sourceText); i++ {
						if sourceText[i] == '?' {
							startRemove = i
							break
						}
					}
				}

				fix := rule.RuleFixRemoveRange(core.NewTextRange(startRemove, typeEnd))

				ctx.ReportNodeWithFixes(reportNode, rule.RuleMessage{
					Id:          "noInferrableType",
					Description: "Type " + typeName + " trivially inferred from a " + typeName + " literal, remove type annotation.",
				}, fix)
			}
		}

		// Check class properties
		checkPropertyDeclaration := func(node *ast.Node) {
			if opts.IgnoreProperties {
				return
			}

			propDecl := node.AsPropertyDeclaration()
			if propDecl == nil {
				return
			}

			// Skip if no type annotation or no initializer
			if propDecl.Type == nil || propDecl.Initializer == nil {
				return
			}

			// Skip optional properties (they're allowed to have type annotations)
			if propDecl.PostfixToken != nil && propDecl.PostfixToken.Kind == ast.KindQuestionToken {
				return
			}

			// Skip readonly properties (they're allowed to have type annotations)
			if ast.HasSyntacticModifier(node, ast.ModifierFlagsReadonly) {
				return
			}

			// Check if the type is "any" - this is allowed
			if propDecl.Type.Kind == ast.KindAnyKeyword {
				return
			}

			isInferrable, typeName := isInferrableType(propDecl.Type, propDecl.Initializer)
			if isInferrable {
				reportNode := propDecl.Name()
				if reportNode == nil {
					reportNode = node
				}

				// Create the fix
				// Remove the type annotation from property
				nameEnd := propDecl.Name().End()
				typeStart := propDecl.Type.Pos()
				typeEnd := propDecl.Type.End()

				sourceText := ctx.SourceFile.Text()
				colonPos := nameEnd

				// Find the colon between name and type
				for i := nameEnd; i < typeStart && i < len(sourceText); i++ {
					if sourceText[i] == ':' {
						colonPos = i
						break
					}
				}

				// Check for definite assignment assertion (!)
				startRemove := colonPos
				for i := nameEnd; i < colonPos && i < len(sourceText); i++ {
					if sourceText[i] == '!' {
						startRemove = i
						break
					}
				}

				fix := rule.RuleFixRemoveRange(core.NewTextRange(startRemove, typeEnd))

				ctx.ReportNodeWithFixes(reportNode, rule.RuleMessage{
					Id:          "noInferrableType",
					Description: "Type " + typeName + " trivially inferred from a " + typeName + " literal, remove type annotation.",
				}, fix)
			}
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration: checkVariableDeclaration,
			ast.KindParameter:            checkParameter,
			ast.KindPropertyDeclaration:  checkPropertyDeclaration,
		}
	},
})

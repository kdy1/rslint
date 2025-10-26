package no_inferrable_types

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type NoInferrableTypesOptions struct {
	IgnoreParameters bool `json:"ignoreParameters"`
	IgnoreProperties bool `json:"ignoreProperties"`
}

func parseOptions(options any) NoInferrableTypesOptions {
	opts := NoInferrableTypesOptions{
		IgnoreParameters: false,
		IgnoreProperties: false,
	}

	if options == nil {
		return opts
	}

	// Handle array format: ["error", { ignoreParameters: true }]
	if arr, ok := options.([]any); ok && len(arr) > 0 {
		if objMap, ok := arr[0].(map[string]any); ok {
			if val, ok := objMap["ignoreParameters"].(bool); ok {
				opts.IgnoreParameters = val
			}
			if val, ok := objMap["ignoreProperties"].(bool); ok {
				opts.IgnoreProperties = val
			}
		}
		return opts
	}

	// Handle object format: { ignoreParameters: true }
	if objMap, ok := options.(map[string]any); ok {
		if val, ok := objMap["ignoreParameters"].(bool); ok {
			opts.IgnoreParameters = val
		}
		if val, ok := objMap["ignoreProperties"].(bool); ok {
			opts.IgnoreProperties = val
		}
	}

	return opts
}

// NoInferrableTypesRule implements the no-inferrable-types rule
// Disallow explicit type declarations for variables or parameters initialized to a number, string, or boolean
var NoInferrableTypesRule = rule.CreateRule(rule.Rule{
	Name: "no-inferrable-types",
	Run:  run,
})

func buildNoInferrableTypeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noInferrableType",
		Description: "Type can be trivially inferred from its value.",
	}
}

// Check if an expression is inferrable for a given type annotation
func isInferrableExpression(typeNode *ast.Node, init *ast.Node) bool {
	if typeNode == nil || init == nil {
		return false
	}

	typeKind := typeNode.Kind

	// Handle unary expressions (e.g., -10, +10, !true)
	actualInit := init
	if init.Kind == ast.KindPrefixUnaryExpression {
		unary := init.AsPrefixUnaryExpression()
		actualInit = unary.Operand
	}

	switch typeKind {
	case ast.KindBigIntKeyword:
		return isBigIntLiteral(actualInit)
	case ast.KindBooleanKeyword:
		return isBooleanLiteral(actualInit)
	case ast.KindNumberKeyword:
		return isNumberLiteral(actualInit)
	case ast.KindNullKeyword:
		return actualInit.Kind == ast.KindNullKeyword
	case ast.KindStringKeyword:
		return isStringLiteral(actualInit)
	case ast.KindSymbolKeyword:
		return isSymbolCall(actualInit)
	case ast.KindUndefinedKeyword:
		return isUndefinedLiteral(actualInit)
	case ast.KindTypeReference:
		// Check for RegExp
		typeRef := typeNode.AsTypeReferenceNode()
		if typeRef.TypeName != nil && typeRef.TypeName.Kind == ast.KindIdentifier {
			typeName := typeRef.TypeName.AsIdentifier().Text()
			if typeName == "RegExp" {
				return isRegExpLiteral(actualInit)
			}
		}
	}

	return false
}

func isBigIntLiteral(node *ast.Node) bool {
	if node.Kind == ast.KindBigIntLiteral {
		return true
	}
	// BigInt(10) or BigInt?.(10)
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		expr := call.Expression
		if expr.Kind == ast.KindIdentifier && expr.AsIdentifier().Text() == "BigInt" {
			return true
		}
	}
	return false
}

func isBooleanLiteral(node *ast.Node) bool {
	if node.Kind == ast.KindTrueKeyword || node.Kind == ast.KindFalseKeyword {
		return true
	}
	// Boolean(null) or Boolean?.(null)
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		expr := call.Expression
		if expr.Kind == ast.KindIdentifier && expr.AsIdentifier().Text() == "Boolean" {
			return true
		}
	}
	// !0 or !!value
	if node.Kind == ast.KindPrefixUnaryExpression {
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindExclamationToken {
			return true
		}
	}
	return false
}

func isNumberLiteral(node *ast.Node) bool {
	if node.Kind == ast.KindNumericLiteral {
		return true
	}
	// Check for Infinity or NaN
	if node.Kind == ast.KindIdentifier {
		id := node.AsIdentifier().Text()
		if id == "Infinity" || id == "NaN" {
			return true
		}
	}
	// Number('1') or Number?.('1')
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		expr := call.Expression
		if expr.Kind == ast.KindIdentifier && expr.AsIdentifier().Text() == "Number" {
			return true
		}
	}
	return false
}

func isStringLiteral(node *ast.Node) bool {
	return node.Kind == ast.KindStringLiteral ||
		node.Kind == ast.KindNoSubstitutionTemplateLiteral ||
		node.Kind == ast.KindTemplateExpression ||
		isStringCall(node)
}

func isStringCall(node *ast.Node) bool {
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		expr := call.Expression
		if expr.Kind == ast.KindIdentifier && expr.AsIdentifier().Text() == "String" {
			return true
		}
	}
	return false
}

func isSymbolCall(node *ast.Node) bool {
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		expr := call.Expression
		if expr.Kind == ast.KindIdentifier && expr.AsIdentifier().Text() == "Symbol" {
			return true
		}
	}
	return false
}

func isUndefinedLiteral(node *ast.Node) bool {
	if node.Kind == ast.KindIdentifier && node.AsIdentifier().Text() == "undefined" {
		return true
	}
	// void someValue
	if node.Kind == ast.KindVoidExpression {
		return true
	}
	return false
}

func isRegExpLiteral(node *ast.Node) bool {
	if node.Kind == ast.KindRegularExpressionLiteral {
		return true
	}
	// RegExp('a') or RegExp?.('a') or new RegExp('a')
	if node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression {
		var expr *ast.Node
		if node.Kind == ast.KindCallExpression {
			expr = node.AsCallExpression().Expression
		} else {
			expr = node.AsNewExpression().Expression
		}
		if expr != nil && expr.Kind == ast.KindIdentifier && expr.AsIdentifier().Text() == "RegExp" {
			return true
		}
	}
	return false
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	return rule.RuleListeners{
		// Check variable declarations
		ast.KindVariableDeclaration: func(node *ast.Node) {
			varDecl := node.AsVariableDeclaration()

			// Skip if no type annotation or no initializer
			if varDecl.Type == nil || varDecl.Initializer == nil {
				return
			}

			// Check if the type is inferrable from the initializer
			if isInferrableExpression(varDecl.Type, varDecl.Initializer) {
				message := buildNoInferrableTypeMessage()

				// Create a fix that removes the type annotation
				// Pattern: "const a: number = 5" -> "const a = 5"
				// Get the source text for the variable name and initializer
				nameRange := utils.TrimNodeTextRange(ctx.SourceFile, varDecl.Name())
				nameText := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

				initRange := utils.TrimNodeTextRange(ctx.SourceFile, varDecl.Initializer)
				initText := ctx.SourceFile.Text()[initRange.Pos():initRange.End()]

				// Build the replacement without the type annotation
				replacement := nameText + " = " + initText

				// Replace from name start to initializer end
				replaceRange := nameRange.WithEnd(initRange.End())
				fix := rule.RuleFix{
					Text:  replacement,
					Range: replaceRange,
				}

				ctx.ReportNodeWithFixes(varDecl.Type, message, fix)
			}
		},

		// Check parameter declarations
		ast.KindParameter: func(node *ast.Node) {
			if opts.IgnoreParameters {
				return
			}

			param := node.AsParameterDeclaration()

			// Skip if no type annotation or no initializer
			if param.Type == nil || param.Initializer == nil {
				return
			}

			// Skip optional parameters (they need explicit types)
			if param.QuestionToken != nil {
				return
			}

			// Check if the type is inferrable from the initializer
			if isInferrableExpression(param.Type, param.Initializer) {
				message := buildNoInferrableTypeMessage()

				// Create a fix that removes the type annotation
				nameRange := utils.TrimNodeTextRange(ctx.SourceFile, param.Name())
				nameText := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

				initRange := utils.TrimNodeTextRange(ctx.SourceFile, param.Initializer)
				initText := ctx.SourceFile.Text()[initRange.Pos():initRange.End()]

				replacement := nameText + " = " + initText

				replaceRange := nameRange.WithEnd(initRange.End())
				fix := rule.RuleFix{
					Text:  replacement,
					Range: replaceRange,
				}

				ctx.ReportNodeWithFixes(param.Type, message, fix)
			}
		},

		// Check property declarations
		ast.KindPropertyDeclaration: func(node *ast.Node) {
			if opts.IgnoreProperties {
				return
			}

			propDecl := node.AsPropertyDeclaration()

			// Skip if no type annotation or no initializer
			if propDecl.Type == nil || propDecl.Initializer == nil {
				return
			}

			// Skip optional properties
			// QuestionToken field not available, skipping optional property check for now
				return
			}

			// Check for readonly modifier
			if propDecl.Modifiers() != nil {
				for _, mod := range propDecl.Modifiers().Nodes {
					if mod.Kind == ast.KindReadonlyKeyword {
						return
					}
				}
			}

			// Check if the type is inferrable from the initializer
			if isInferrableExpression(propDecl.Type, propDecl.Initializer) {
				message := buildNoInferrableTypeMessage()

				// Create a fix that removes the type annotation
				nameRange := utils.TrimNodeTextRange(ctx.SourceFile, propDecl.Name())
				nameText := ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]

				initRange := utils.TrimNodeTextRange(ctx.SourceFile, propDecl.Initializer)
				initText := ctx.SourceFile.Text()[initRange.Pos():initRange.End()]

				replacement := nameText + " = " + initText

				replaceRange := nameRange.WithEnd(initRange.End())
				fix := rule.RuleFix{
					Text:  replacement,
					Range: replaceRange,
				}

				ctx.ReportNodeWithFixes(propDecl.Type, message, fix)
			}
		},
	}
}

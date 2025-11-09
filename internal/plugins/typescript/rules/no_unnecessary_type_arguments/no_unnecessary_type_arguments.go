// Package no_unnecessary_type_arguments implements the @typescript-eslint/no-unnecessary-type-arguments rule.
// This rule disallows type arguments that are equal to the default, helping to reduce code verbosity
// by removing unnecessary type parameter specifications when they match their defaults.
package no_unnecessary_type_arguments

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoUnnecessaryTypeArgumentsRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-type-arguments",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		if ctx.TypeChecker == nil {
			return rule.RuleListeners{}
		}

		checkTypeArguments := func(node *ast.Node, typeArguments []*ast.Node, getSignatureFunc func() *checker.Signature) {
			if typeArguments == nil || len(typeArguments) == 0 {
				return
			}

			signature := getSignatureFunc()
			if signature == nil {
				return
			}

			typeParameters := checker.Signature_typeParameters(signature)
			if typeParameters == nil || len(typeParameters) == 0 {
				return
			}

			// Find the first unnecessary type argument from the end
			// We check from the end because trailing defaults can be omitted
			//
			// NOTE: This implementation is currently incomplete.
			// TODO: Add proper default type parameter checking once shim supports accessing
			// the TypeParameterDeclaration.Default field through an accessor function.
			// For now, the rule is registered but doesn't report any errors.
			unnecessaryIndex := -1
			_ = typeParameters // Prevent unused variable warning

			if unnecessaryIndex >= 0 {
				// Report the first unnecessary type argument
				unnecessaryArg := typeArguments[unnecessaryIndex]

				// Build the fix by removing unnecessary type arguments from the found index
				var newTypeArgs string
				if unnecessaryIndex == 0 {
					// Remove all type arguments
					newTypeArgs = ""
				} else {
					// Keep only the necessary type arguments
					newTypeArgs = "<"
					for i := 0; i < unnecessaryIndex; i++ {
						if i > 0 {
							newTypeArgs += ", "
						}
						typeArgRange := utils.TrimNodeTextRange(ctx.SourceFile, typeArguments[i])
						newTypeArgs += ctx.SourceFile.Text()[typeArgRange.Pos():typeArgRange.End()]
					}
					newTypeArgs += ">"
				}

				// Calculate the range to replace (the entire type arguments section)
				// Find the opening < and closing >
				firstTypeArg := typeArguments[0]
				lastTypeArg := typeArguments[len(typeArguments)-1]
				firstRange := utils.TrimNodeTextRange(ctx.SourceFile, firstTypeArg)
				lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastTypeArg)

				// The type arguments range includes the angle brackets
				typeArgsStart := firstRange.Pos() - 1 // Include opening <
				typeArgsEnd := lastRange.End() + 1     // Include closing >

				ctx.ReportNodeWithFixes(
					unnecessaryArg,
					rule.RuleMessage{
						Id:          "unnecessaryTypeParameter",
						Description: "This is the default value for this type parameter, so it can be omitted.",
					},
					rule.RuleFix{
						Text:  newTypeArgs,
						Range: core.NewTextRange(typeArgsStart, typeArgsEnd),
					},
				)
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				if callExpr == nil || callExpr.TypeArguments == nil || callExpr.TypeArguments.Nodes == nil {
					return
				}

				getSignature := func() *checker.Signature {
					// Get the signature of the called function
					return checker.Checker_getResolvedSignature(ctx.TypeChecker, node, nil, checker.CheckModeNormal)
				}

				checkTypeArguments(node, callExpr.TypeArguments.Nodes, getSignature)
			},

			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				if newExpr == nil || newExpr.TypeArguments == nil || newExpr.TypeArguments.Nodes == nil {
					return
				}

				getSignature := func() *checker.Signature {
					return checker.Checker_getResolvedSignature(ctx.TypeChecker, node, nil, checker.CheckModeNormal)
				}

				checkTypeArguments(node, newExpr.TypeArguments.Nodes, getSignature)
			},

			// Note: TypeReference and ExpressionWithTypeArguments are commented out for now
			// because they require handling type parameter defaults which needs more
			// complex shim support. These can be added in a future enhancement.
			//
			// ast.KindTypeReference: func(node *ast.Node) { ... },
			// ast.KindExpressionWithTypeArguments: func(node *ast.Node) { ... },

			ast.KindJsxOpeningElement: func(node *ast.Node) {
				jsxOpening := node.AsJsxOpeningElement()
				if jsxOpening == nil || jsxOpening.TypeArguments == nil || jsxOpening.TypeArguments.Nodes == nil {
					return
				}

				getSignature := func() *checker.Signature {
					// Get the type of the JSX tag
					tagType := ctx.TypeChecker.GetTypeAtLocation(jsxOpening.TagName)
					if tagType == nil {
						return nil
					}

					// Get call signatures (JSX elements are essentially function calls)
					signatures := utils.GetCallSignatures(ctx.TypeChecker, tagType)
					if len(signatures) > 0 {
						return signatures[0]
					}

					return nil
				}

				checkTypeArguments(node, jsxOpening.TypeArguments.Nodes, getSignature)
			},

			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				jsxSelfClosing := node.AsJsxSelfClosingElement()
				if jsxSelfClosing == nil || jsxSelfClosing.TypeArguments == nil || jsxSelfClosing.TypeArguments.Nodes == nil {
					return
				}

				getSignature := func() *checker.Signature {
					// Get the type of the JSX tag
					tagType := ctx.TypeChecker.GetTypeAtLocation(jsxSelfClosing.TagName)
					if tagType == nil {
						return nil
					}

					// Get call signatures
					signatures := utils.GetCallSignatures(ctx.TypeChecker, tagType)
					if len(signatures) > 0 {
						return signatures[0]
					}

					return nil
				}

				checkTypeArguments(node, jsxSelfClosing.TypeArguments.Nodes, getSignature)
			},

			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				taggedTemplate := node.AsTaggedTemplateExpression()
				if taggedTemplate == nil || taggedTemplate.TypeArguments == nil || taggedTemplate.TypeArguments.Nodes == nil {
					return
				}

				getSignature := func() *checker.Signature {
					return checker.Checker_getResolvedSignature(ctx.TypeChecker, node, nil, checker.CheckModeNormal)
				}

				checkTypeArguments(node, taggedTemplate.TypeArguments.Nodes, getSignature)
			},
		}
	},
})

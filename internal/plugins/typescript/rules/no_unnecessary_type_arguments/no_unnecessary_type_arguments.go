// Package no_unnecessary_type_arguments implements the @typescript-eslint/no-unnecessary-type-arguments rule.
// This rule disallows type arguments that are equal to the default, helping to reduce code verbosity
// by removing unnecessary type parameter specifications when they match their defaults.
package no_unnecessary_type_arguments

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoUnnecessaryTypeArgumentsRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-type-arguments",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		if ctx.TypeChecker == nil {
			return rule.RuleListeners{}
		}

		checkTypeArguments := func(node *ast.Node, typeArguments *ast.NodeArray, getSignatureFunc func() checker.Signature) {
			if typeArguments == nil || typeArguments.Nodes == nil || len(typeArguments.Nodes) == 0 {
				return
			}

			signature := getSignatureFunc()
			if signature == nil {
				return
			}

			typeParameters := signature.TypeParameters()
			if typeParameters == nil || len(typeParameters) == 0 {
				return
			}

			// Find the first unnecessary type argument from the end
			// We check from the end because trailing defaults can be omitted
			unnecessaryIndex := -1
			for i := len(typeArguments.Nodes) - 1; i >= 0; i-- {
				if i >= len(typeParameters) {
					// More type arguments than parameters - this is already an error elsewhere
					break
				}

				typeParam := typeParameters[i]
				if typeParam == nil {
					break
				}

				defaultType := typeParam.Default()
				if defaultType == nil {
					// No default - all previous type arguments are necessary
					break
				}

				typeArg := typeArguments.Nodes[i]
				argType := ctx.TypeChecker.GetTypeFromTypeNode(typeArg)

				if argType != nil && defaultType != nil {
					// Check if the provided type is the same as the default
					if ctx.TypeChecker.IsTypeIdenticalTo(argType, defaultType) {
						unnecessaryIndex = i
					} else {
						// Different from default - all previous type arguments are necessary
						break
					}
				} else {
					break
				}
			}

			if unnecessaryIndex >= 0 {
				// Report the first unnecessary type argument
				unnecessaryArg := typeArguments.Nodes[unnecessaryIndex]

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
						typeArgRange := utils.TrimNodeTextRange(ctx.SourceFile, typeArguments.Nodes[i])
						newTypeArgs += ctx.SourceFile.Text()[typeArgRange.Pos():typeArgRange.End()]
					}
					newTypeArgs += ">"
				}

				// Calculate the range to replace (the entire type arguments section)
				// Find the opening < and closing >
				firstTypeArg := typeArguments.Nodes[0]
				lastTypeArg := typeArguments.Nodes[len(typeArguments.Nodes)-1]
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
						Message: "Remove unnecessary type arguments",
						Edits: []rule.RuleFixEdit{
							{
								Range: utils.TextRange{Start: typeArgsStart, End: typeArgsEnd},
								Text:  newTypeArgs,
							},
						},
					},
				)
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				if callExpr == nil || callExpr.TypeArguments == nil {
					return
				}

				getSignature := func() checker.Signature {
					// Get the signature of the called function
					return ctx.TypeChecker.GetResolvedSignature(node)
				}

				checkTypeArguments(node, callExpr.TypeArguments, getSignature)
			},

			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				if newExpr == nil || newExpr.TypeArguments == nil {
					return
				}

				getSignature := func() checker.Signature {
					return ctx.TypeChecker.GetResolvedSignature(node)
				}

				checkTypeArguments(node, newExpr.TypeArguments, getSignature)
			},

			ast.KindTypeReference: func(node *ast.Node) {
				typeRef := node.AsTypeReference()
				if typeRef == nil || typeRef.TypeArguments == nil {
					return
				}

				getSignature := func() checker.Signature {
					// For type references, we need to get the type and its symbol
					typeRefType := ctx.TypeChecker.GetTypeFromTypeNode(node)
					if typeRefType == nil {
						return nil
					}

					symbol := typeRefType.Symbol()
					if symbol == nil {
						return nil
					}

					// Get the declarations to find type parameters
					declarations := symbol.Declarations
					if len(declarations) == 0 {
						return nil
					}

					// Look for type parameters in the declaration
					for _, decl := range declarations {
						switch decl.Kind {
						case ast.KindTypeAliasDeclaration:
							typeAlias := decl.AsTypeAliasDeclaration()
							if typeAlias != nil && typeAlias.TypeParameters != nil {
								// Create a pseudo-signature to check type parameters
								return createPseudoSignature(ctx, typeAlias.TypeParameters)
							}
						case ast.KindInterfaceDeclaration:
							interfaceDecl := decl.AsInterfaceDeclaration()
							if interfaceDecl != nil && interfaceDecl.TypeParameters != nil {
								return createPseudoSignature(ctx, interfaceDecl.TypeParameters)
							}
						case ast.KindClassDeclaration:
							classDecl := decl.AsClassDeclaration()
							if classDecl != nil && classDecl.TypeParameters != nil {
								return createPseudoSignature(ctx, classDecl.TypeParameters)
							}
						}
					}

					return nil
				}

				checkTypeArguments(node, typeRef.TypeArguments, getSignature)
			},

			ast.KindExpressionWithTypeArguments: func(node *ast.Node) {
				exprWithTypeArgs := node.AsExpressionWithTypeArguments()
				if exprWithTypeArgs == nil || exprWithTypeArgs.TypeArguments == nil {
					return
				}

				getSignature := func() checker.Signature {
					// Get the type of the expression
					exprType := ctx.TypeChecker.GetTypeAtLocation(exprWithTypeArgs.Expression)
					if exprType == nil {
						return nil
					}

					symbol := exprType.Symbol()
					if symbol == nil {
						return nil
					}

					declarations := symbol.Declarations
					if len(declarations) == 0 {
						return nil
					}

					// Look for type parameters in the declaration
					for _, decl := range declarations {
						switch decl.Kind {
						case ast.KindClassDeclaration:
							classDecl := decl.AsClassDeclaration()
							if classDecl != nil && classDecl.TypeParameters != nil {
								return createPseudoSignature(ctx, classDecl.TypeParameters)
							}
						case ast.KindInterfaceDeclaration:
							interfaceDecl := decl.AsInterfaceDeclaration()
							if interfaceDecl != nil && interfaceDecl.TypeParameters != nil {
								return createPseudoSignature(ctx, interfaceDecl.TypeParameters)
							}
						}
					}

					return nil
				}

				checkTypeArguments(node, exprWithTypeArgs.TypeArguments, getSignature)
			},

			ast.KindJsxOpeningElement: func(node *ast.Node) {
				jsxOpening := node.AsJsxOpeningElement()
				if jsxOpening == nil || jsxOpening.TypeArguments == nil {
					return
				}

				getSignature := func() checker.Signature {
					// Get the type of the JSX tag
					tagType := ctx.TypeChecker.GetTypeAtLocation(jsxOpening.TagName)
					if tagType == nil {
						return nil
					}

					// Get call signatures (JSX elements are essentially function calls)
					signatures := tagType.GetCallSignatures()
					if len(signatures) > 0 {
						return signatures[0]
					}

					return nil
				}

				checkTypeArguments(node, jsxOpening.TypeArguments, getSignature)
			},

			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				jsxSelfClosing := node.AsJsxSelfClosingElement()
				if jsxSelfClosing == nil || jsxSelfClosing.TypeArguments == nil {
					return
				}

				getSignature := func() checker.Signature {
					// Get the type of the JSX tag
					tagType := ctx.TypeChecker.GetTypeAtLocation(jsxSelfClosing.TagName)
					if tagType == nil {
						return nil
					}

					// Get call signatures
					signatures := tagType.GetCallSignatures()
					if len(signatures) > 0 {
						return signatures[0]
					}

					return nil
				}

				checkTypeArguments(node, jsxSelfClosing.TypeArguments, getSignature)
			},

			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				taggedTemplate := node.AsTaggedTemplateExpression()
				if taggedTemplate == nil || taggedTemplate.TypeArguments == nil {
					return
				}

				getSignature := func() checker.Signature {
					return ctx.TypeChecker.GetResolvedSignature(node)
				}

				checkTypeArguments(node, taggedTemplate.TypeArguments, getSignature)
			},
		}
	},
})

// createPseudoSignature creates a pseudo-signature for checking type parameters
// This is a helper to work with type declarations that have type parameters but aren't functions
func createPseudoSignature(ctx rule.RuleContext, typeParameters *ast.NodeArray) checker.Signature {
	if typeParameters == nil || typeParameters.Nodes == nil || len(typeParameters.Nodes) == 0 {
		return nil
	}

	// We need to extract type parameters from the declaration and create a structure
	// that can be queried for defaults. Since TypeScript's checker API provides
	// TypeParameter objects through signatures, we attempt to get them through
	// the type system.

	var typeParams []checker.TypeParameter
	for _, typeParamNode := range typeParameters.Nodes {
		typeParam := typeParamNode.AsTypeParameterDeclaration()
		if typeParam == nil {
			continue
		}

		// Get the type parameter from the type system
		symbol := ctx.TypeChecker.GetSymbolAtLocation(typeParam.Name())
		if symbol != nil {
			// Get the type parameter from the symbol
			typeParamType := ctx.TypeChecker.GetDeclaredTypeOfSymbol(symbol)
			if typeParamType != nil {
				// Check if this is a TypeParameter
				if tp, ok := typeParamType.(checker.TypeParameter); ok {
					typeParams = append(typeParams, tp)
				}
			}
		}
	}

	if len(typeParams) == 0 {
		return nil
	}

	// Create a pseudo-signature wrapper
	return &pseudoSignature{
		typeParams: typeParams,
	}
}

// pseudoSignature is a minimal implementation of checker.Signature
// that only provides TypeParameters() for our use case
type pseudoSignature struct {
	typeParams []checker.TypeParameter
}

func (ps *pseudoSignature) TypeParameters() []checker.TypeParameter {
	return ps.typeParams
}

func (ps *pseudoSignature) GetDeclaration() *ast.Node                  { return nil }
func (ps *pseudoSignature) Parameters() []checker.Symbol               { return nil }
func (ps *pseudoSignature) ThisParameter() checker.Symbol              { return nil }
func (ps *pseudoSignature) GetReturnType() checker.Type                { return nil }
func (ps *pseudoSignature) GetDocumentationComment() []checker.SymbolDisplayPart { return nil }
func (ps *pseudoSignature) GetJsDocTags() []checker.JSDocTagInfo       { return nil }

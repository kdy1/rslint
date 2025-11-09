// Package no_unnecessary_type_arguments implements the @typescript-eslint/no-unnecessary-type-arguments rule.
// This rule disallows type arguments that are equal to the default, helping to reduce code verbosity
// by removing unnecessary type parameter specifications when they match their defaults.
package no_unnecessary_type_arguments

import (
	"unsafe"

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
			unnecessaryIndex := -1

			for i := len(typeArguments) - 1; i >= 0; i-- {
				if i >= len(typeParameters) {
					break
				}

				typeArg := typeArguments[i]
				typeParam := typeParameters[i]

				// Get the default type for this type parameter
				// We need to get the type parameter declaration to access its default
				var defaultType *checker.Type

				// Get the symbol from the type parameter
				symbol := checker.Type_symbol(typeParam)
				if symbol == nil {
					break
				}

				// Get the declarations for this symbol
				declarations := symbol.Declarations
				if len(declarations) == 0 {
					break
				}

				// Find the type parameter declaration
				var typeParamDecl *ast.Node
				for _, decl := range declarations {
					if ast.IsTypeParameterDeclaration(decl) {
						typeParamDecl = decl
						break
					}
				}

				if typeParamDecl == nil {
					break
				}

				// First try getting the default from the checker
				defaultType = checker.Checker_getDefaultFromTypeParameter(ctx.TypeChecker, typeParam)

				// If that didn't work, try getting it from the AST node
				if defaultType == nil && ast.IsTypeParameterDeclaration(typeParamDecl) {
					// Cast to TypeParameterDeclaration using unsafe pointer
					tpDecl := (*ast.TypeParameterDeclaration)(unsafe.Pointer(typeParamDecl))
					if tpDecl.DefaultType != nil {
						defaultType = ctx.TypeChecker.GetTypeFromTypeNode(tpDecl.DefaultType)
					}
				}

				if defaultType == nil {
					// No default, so we can stop checking
					break
				}

				// Get the type of the argument
				argType := ctx.TypeChecker.GetTypeFromTypeNode(typeArg)
				if argType == nil {
					break
				}

				// Check if the argument type is identical to the default type
				// We use multiple methods to check equality
				isEqual := false

				// Method 1: Direct type identity check
				if checker.Checker_isTypeIdenticalTo(ctx.TypeChecker, argType, defaultType) {
					isEqual = true
				}

				// Method 2: Compare type strings as fallback
				if !isEqual {
					argTypeStr := checker.Checker_typeToString(ctx.TypeChecker, argType, nil)
					defaultTypeStr := checker.Checker_typeToString(ctx.TypeChecker, defaultType, nil)
					if argTypeStr == defaultTypeStr {
						isEqual = true
					}
				}

				if isEqual {
					unnecessaryIndex = i
				} else {
					// Not identical, so we can stop checking
					break
				}
			}

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

		checkTypeReference := func(node *ast.Node, typeArguments []*ast.Node, typeNode *ast.Node, isValueContext bool) {
			if typeArguments == nil || len(typeArguments) == 0 {
				return
			}

			// Get the type from the type reference
			refType := ctx.TypeChecker.GetTypeAtLocation(typeNode)
			if refType == nil {
				return
			}

			// Get the symbol from the type
			symbol := checker.Type_symbol(refType)
			if symbol == nil {
				return
			}

			// Get the declarations for this symbol
			declarations := symbol.Declarations
			if len(declarations) == 0 {
				return
			}

			// Determine which declaration to use based on context
			// When there are multiple declarations (e.g., interface and class with same name),
			// we need to pick the right one based on how it's being used
			var relevantDecl *ast.Node

			if isValueContext {
				// In value context (extends/implements), determine the appropriate declaration
				// Check if this is extends vs implements
				isExtendsClause := false
				if node.Parent != nil && node.Parent.Kind == ast.KindHeritageClause {
					heritageClause := node.Parent.AsHeritageClause()
					if heritageClause != nil {
						isExtendsClause = (heritageClause.Token == ast.KindExtendsKeyword)
					}
				}

				if isExtendsClause {
					// For extends, prefer class declarations, then constructor signatures
					for _, decl := range declarations {
						if ast.IsClassDeclaration(decl) {
							relevantDecl = decl
							break
						}
					}
					if relevantDecl == nil {
						for _, decl := range declarations {
							if decl.Kind == ast.KindConstructSignature {
								relevantDecl = decl
								break
							}
						}
					}
				} else {
					// For implements, prefer interface declarations
					for _, decl := range declarations {
						if ast.IsInterfaceDeclaration(decl) {
							relevantDecl = decl
							break
						}
					}
				}
			}

			// If we haven't found a relevant declaration, use the first one with type parameters
			// Prioritize type declarations (interfaces, type aliases) over value declarations (classes)
			if relevantDecl == nil {
				// First try type declarations
				for _, decl := range declarations {
					if ast.IsInterfaceDeclaration(decl) || ast.IsTypeAliasDeclaration(decl) {
						relevantDecl = decl
						break
					}
				}
				// Then try value declarations
				if relevantDecl == nil {
					for _, decl := range declarations {
						if ast.IsClassDeclaration(decl) || decl.Kind == ast.KindConstructSignature {
							relevantDecl = decl
							break
						}
					}
				}
			}

			// Use the relevant declaration, or fall back to the first one
			if relevantDecl == nil && len(declarations) > 0 {
				relevantDecl = declarations[0]
			}

			if relevantDecl == nil {
				return
			}

			// Get type parameters from the relevant declaration
			sortedDecls := []*ast.Node{relevantDecl}

			// Find a declaration that has type parameters
			var typeParameters []*checker.Type
			for _, decl := range sortedDecls {
				var typeParametersNodes []*ast.Node

				// Check different kinds of declarations that can have type parameters
				if ast.IsClassDeclaration(decl) {
					classDecl := decl.AsClassDeclaration()
					if classDecl.TypeParameters != nil && classDecl.TypeParameters.Nodes != nil {
						typeParametersNodes = classDecl.TypeParameters.Nodes
					}
				} else if ast.IsInterfaceDeclaration(decl) {
					interfaceDecl := decl.AsInterfaceDeclaration()
					if interfaceDecl.TypeParameters != nil && interfaceDecl.TypeParameters.Nodes != nil {
						typeParametersNodes = interfaceDecl.TypeParameters.Nodes
					}
				} else if ast.IsTypeAliasDeclaration(decl) {
					typeAliasDecl := decl.AsTypeAliasDeclaration()
					if typeAliasDecl.TypeParameters != nil && typeAliasDecl.TypeParameters.Nodes != nil {
						typeParametersNodes = typeAliasDecl.TypeParameters.Nodes
					}
				} else if decl.Kind == ast.KindConstructSignature {
					constructSig := decl.AsConstructSignatureDeclaration()
					if constructSig != nil && constructSig.TypeParameters != nil && constructSig.TypeParameters.Nodes != nil {
						typeParametersNodes = constructSig.TypeParameters.Nodes
					}
				}

				if len(typeParametersNodes) > 0 {
					// Convert AST type parameters to checker types
					typeParameters = make([]*checker.Type, 0, len(typeParametersNodes))
					for _, typeParamNode := range typeParametersNodes {
						if ast.IsTypeParameterDeclaration(typeParamNode) {
							typeParamType := ctx.TypeChecker.GetTypeAtLocation(typeParamNode)
							if typeParamType != nil {
								typeParameters = append(typeParameters, typeParamType)
							}
						}
					}
					if len(typeParameters) > 0 {
						break
					}
				}
			}

			// Also store the AST nodes for type parameters so we can check defaults correctly
			var typeParameterNodes []*ast.Node
			for _, decl := range sortedDecls {
				// Check different kinds of declarations that can have type parameters
				if ast.IsClassDeclaration(decl) {
					classDecl := decl.AsClassDeclaration()
					if classDecl.TypeParameters != nil && classDecl.TypeParameters.Nodes != nil {
						typeParameterNodes = classDecl.TypeParameters.Nodes
					}
				} else if ast.IsInterfaceDeclaration(decl) {
					interfaceDecl := decl.AsInterfaceDeclaration()
					if interfaceDecl.TypeParameters != nil && interfaceDecl.TypeParameters.Nodes != nil {
						typeParameterNodes = interfaceDecl.TypeParameters.Nodes
					}
				} else if ast.IsTypeAliasDeclaration(decl) {
					typeAliasDecl := decl.AsTypeAliasDeclaration()
					if typeAliasDecl.TypeParameters != nil && typeAliasDecl.TypeParameters.Nodes != nil {
						typeParameterNodes = typeAliasDecl.TypeParameters.Nodes
					}
				} else if decl.Kind == ast.KindConstructSignature {
					constructSig := decl.AsConstructSignatureDeclaration()
					if constructSig != nil && constructSig.TypeParameters != nil && constructSig.TypeParameters.Nodes != nil {
						typeParameterNodes = constructSig.TypeParameters.Nodes
					}
				}

				if len(typeParameterNodes) > 0 {
					break
				}
			}

			if len(typeParameters) > 0 && len(typeParameterNodes) > 0 {
				// Since we can't create a signature easily, we'll check inline here
				// This is similar to checkTypeArguments but adapted for type references
				unnecessaryIndex := -1

				for i := len(typeArguments) - 1; i >= 0; i-- {
					if i >= len(typeParameters) || i >= len(typeParameterNodes) {
						break
					}

					typeArg := typeArguments[i]
					typeParamNode := typeParameterNodes[i]

					// Get the default type for this type parameter directly from the AST node
					var defaultType *checker.Type

					if ast.IsTypeParameterDeclaration(typeParamNode) {
						// Cast to TypeParameterDeclaration using unsafe pointer
						tpDecl := (*ast.TypeParameterDeclaration)(unsafe.Pointer(typeParamNode))
						if tpDecl.DefaultType != nil {
							// Get the type from the default type node
							defaultType = ctx.TypeChecker.GetTypeFromTypeNode(tpDecl.DefaultType)
						}
					}

					if defaultType == nil {
						// No default, so we can stop checking
						break
					}

					// Get the type of the argument
					argType := ctx.TypeChecker.GetTypeFromTypeNode(typeArg)
					if argType == nil {
						break
					}

					// Check if the argument type is identical to the default type
					// We use multiple methods to check equality
					isEqual := false

					// Method 1: Direct type identity check
					if checker.Checker_isTypeIdenticalTo(ctx.TypeChecker, argType, defaultType) {
						isEqual = true
					}

					// Method 2: Compare type strings as fallback
					if !isEqual {
						argTypeStr := checker.Checker_typeToString(ctx.TypeChecker, argType, nil)
						defaultTypeStr := checker.Checker_typeToString(ctx.TypeChecker, defaultType, nil)
						if argTypeStr == defaultTypeStr {
							isEqual = true
						}
					}

					if isEqual {
						unnecessaryIndex = i
					} else {
						// Not identical, so we can stop checking
						break
					}
				}

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

			ast.KindTypeReference: func(node *ast.Node) {
				typeRef := node.AsTypeReference()
				if typeRef == nil || typeRef.TypeArguments == nil || typeRef.TypeArguments.Nodes == nil {
					return
				}

				checkTypeReference(node, typeRef.TypeArguments.Nodes, typeRef.TypeName, false)
			},

			ast.KindExpressionWithTypeArguments: func(node *ast.Node) {
				exprWithTypeArgs := node.AsExpressionWithTypeArguments()
				if exprWithTypeArgs == nil || exprWithTypeArgs.TypeArguments == nil || exprWithTypeArgs.TypeArguments.Nodes == nil {
					return
				}

				// Determine if this is in a value context (extends/implements)
				// In a value context (class extends/implements), prioritize value declarations
				isValueContext := false
				if node.Parent != nil {
					if node.Parent.Kind == ast.KindHeritageClause {
						isValueContext = true
					}
				}

				checkTypeReference(node, exprWithTypeArgs.TypeArguments.Nodes, exprWithTypeArgs.Expression, isValueContext)
			},

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

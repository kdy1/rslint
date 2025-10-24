package consistent_generic_constructors

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ConsistentGenericConstructorsOptions defines the configuration
type ConsistentGenericConstructorsOptions struct {
	Style string `json:"style"` // "constructor" or "type-annotation"
}

func parseOptions(options interface{}) ConsistentGenericConstructorsOptions {
	opts := ConsistentGenericConstructorsOptions{
		Style: "constructor", // Default
	}

	if options == nil {
		return opts
	}

	switch v := options.(type) {
	case string:
		if v == "type-annotation" || v == "constructor" {
			opts.Style = v
		}
	case map[string]interface{}:
		if style, ok := v["style"].(string); ok {
			if style == "type-annotation" || style == "constructor" {
				opts.Style = style
			}
		}
	}

	return opts
}

func buildPreferConstructorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferConstructor",
		Description: "The generic type arguments should be specified on the constructor type.",
	}
}

func buildPreferTypeAnnotationMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferTypeAnnotation",
		Description: "The generic type arguments should be specified on the type annotation.",
	}
}

// Check if a type node has generic type arguments
func hasTypeArguments(typeNode *ast.Node) bool {
	if typeNode == nil {
		return false
	}

	if typeNode.Kind == ast.KindTypeReference {
		typeRef := typeNode.AsTypeReference()
		if typeRef == nil {
			return false
		}
		return typeRef.TypeArguments != nil && len(typeRef.TypeArguments.Nodes) > 0
	}

	return false
}

// Check if a new expression has generic type arguments
func hasNewExpressionTypeArguments(newExpr *ast.NewExpression) bool {
	if newExpr == nil {
		return false
	}
	return newExpr.TypeArguments != nil && len(newExpr.TypeArguments.Nodes) > 0
}

// Get the type name from a type node
func getTypeName(typeNode *ast.Node, sourceFile *ast.SourceFile) string {
	if typeNode == nil {
		return ""
	}

	if typeNode.Kind == ast.KindTypeReference {
		typeRef := typeNode.AsTypeReference()
		if typeRef == nil {
			return ""
		}

		if typeRef.TypeName != nil {
			typeRange := utils.TrimNodeTextRange(sourceFile, typeRef.TypeName)
			return sourceFile.Text()[typeRange.Pos():typeRange.End()]
		}
	}

	return ""
}

// Get the constructor name from a new expression
func getConstructorName(newExpr *ast.NewExpression, sourceFile *ast.SourceFile) string {
	if newExpr == nil || newExpr.Expression == nil {
		return ""
	}

	exprRange := utils.TrimNodeTextRange(sourceFile, newExpr.Expression)
	return sourceFile.Text()[exprRange.Pos():exprRange.End()]
}

var ConsistentGenericConstructorsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-generic-constructors",
	Run: func(ctx rule.RuleContext, options interface{}) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				if node.Kind != ast.KindVariableDeclaration {
					return
				}

				varDecl := node.AsVariableDeclaration()
				if varDecl == nil {
					return
				}

				// Check if initializer is a new expression
				if varDecl.Initializer == nil || varDecl.Initializer.Kind != ast.KindNewExpression {
					return
				}

				newExpr := varDecl.Initializer.AsNewExpression()
				if newExpr == nil {
					return
				}

				// Check type annotation
				typeAnnotation := varDecl.Type
				hasTypeArgs := hasTypeArguments(typeAnnotation)
				hasConstructorArgs := hasNewExpressionTypeArguments(newExpr)

				// Only report if generics appear on one side only (not both or neither)
				if hasTypeArgs && !hasConstructorArgs {
					// Type arguments only on annotation
					if opts.Style == "constructor" {
						// Get type arguments text
						typeRef := typeAnnotation.AsTypeReference()
						if typeRef != nil && typeRef.TypeArguments != nil {
							// Get the text of the type arguments including brackets
							typeArgsNodes := typeRef.TypeArguments.Nodes
							if len(typeArgsNodes) > 0 {
								// Get the full text including the angle brackets
								// Position of '<' is one character before TypeArguments.Pos()
								// Position after '>' is from the end of the '>' token
								typeArgsStartPos := typeRef.TypeArguments.Pos() - 1
								closingBracketRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, typeRef.TypeArguments.End())
								typeArgsEndPos := closingBracketRange.End()
								typeArgsText := ctx.SourceFile.Text()[typeArgsStartPos:typeArgsEndPos]

								// Create fix: move type args to constructor
								ctx.ReportNodeWithFixes(
									typeAnnotation,
									buildPreferConstructorMessage(),
									// Remove type args from annotation
									rule.RuleFixReplaceRange(
										core.NewTextRange(typeArgsStartPos, typeArgsEndPos),
										"",
									),
									// Add type args to constructor
									rule.RuleFixInsertAfter(newExpr.Expression, typeArgsText),
								)
							}
						}
					}
				} else if !hasTypeArgs && hasConstructorArgs {
					// Type arguments only on constructor
					if opts.Style == "type-annotation" {
						// Get constructor type arguments text
						if newExpr.TypeArguments != nil {
							typeArgsNodes := newExpr.TypeArguments.Nodes
							if len(typeArgsNodes) > 0 {
								// Get the full text including the angle brackets
								typeArgsStartPos := newExpr.TypeArguments.Pos() - 1
								closingBracketRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, newExpr.TypeArguments.End())
								typeArgsEndPos := closingBracketRange.End()
								typeArgsText := ctx.SourceFile.Text()[typeArgsStartPos:typeArgsEndPos]

								// Build the constructor name (what type annotation should be)
								constructorName := getConstructorName(newExpr, ctx.SourceFile)

								// If there's a type annotation, verify it matches constructor name
								if typeAnnotation != nil {
									typeName := getTypeName(typeAnnotation, ctx.SourceFile)

									// Only provide fix if names match
									if typeName == constructorName {
										ctx.ReportNodeWithFixes(
											node,
											buildPreferTypeAnnotationMessage(),
											// Remove type args from constructor
											rule.RuleFixReplaceRange(
												core.NewTextRange(typeArgsStartPos, typeArgsEndPos),
												"",
											),
											// Add type args to annotation
											rule.RuleFixInsertAfter(typeAnnotation.AsTypeReference().TypeName, typeArgsText),
										)
									} else {
										// Names don't match, report without fix
										ctx.ReportNode(node, buildPreferTypeAnnotationMessage())
									}
								} else {
									// No type annotation exists - need to add one
									// Get variable name
									varName := ""
									if varDecl.Name() != nil {
										nameRange := utils.TrimNodeTextRange(ctx.SourceFile, varDecl.Name())
										varName = ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
									}

									if varName != "" {
										// Create the full type annotation including generic args
										fullTypeAnnotation := ": " + constructorName + typeArgsText

										ctx.ReportNodeWithFixes(
											node,
											buildPreferTypeAnnotationMessage(),
											// Add type annotation after variable name
											rule.RuleFixInsertAfter(varDecl.Name(), fullTypeAnnotation),
											// Remove type args from constructor
											rule.RuleFixReplaceRange(
												core.NewTextRange(typeArgsStartPos, typeArgsEndPos),
												"",
											),
										)
									}
								}
							}
						}
					}
				}
			},
		}
	},
})

package no_wrapper_object_types

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildErrorMessage(typeName string, primitiveType string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "bannedClassType",
		Description: "Don't use `" + typeName + "` as a type. Use `" + primitiveType + "` instead.",
	}
}

// Map of wrapper types to their primitive equivalents
var wrapperTypes = map[string]string{
	"BigInt":  "bigint",
	"Boolean": "boolean",
	"Number":  "number",
	"Object":  "object",
	"String":  "string",
	"Symbol":  "symbol",
}

var NoWrapperObjectTypesRule = rule.CreateRule(rule.Rule{
	Name: "no-wrapper-object-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		checkTypeReference := func(node *ast.Node) {
			typeRef := node.AsTypeReferenceNode()
			if typeRef == nil || typeRef.TypeName == nil {
				return
			}

			// Get the type name
			typeName := getTypeNameText(ctx, typeRef.TypeName)
			if typeName == "" {
				return
			}

			// Check if this is a wrapper type
			primitiveType, isWrapper := wrapperTypes[typeName]
			if !isWrapper {
				return
			}

			// Report with a fix to replace with primitive type
			ctx.ReportNodeWithFixes(
				node,
				buildErrorMessage(typeName, primitiveType),
				rule.RuleFixReplace(ctx.SourceFile, node, primitiveType),
			)
		}

		return rule.RuleListeners{
			ast.KindTypeReference: checkTypeReference,
		}
	},
})

func getTypeNameText(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		ident := node.AsIdentifier()
		if ident != nil {
			return ident.Text
		}
	case ast.KindQualifiedName:
		// For qualified names like A.B.C, we only care about the rightmost identifier
		qual := node.AsQualifiedName()
		if qual != nil && qual.Right != nil {
			return getTypeNameText(ctx, qual.Right)
		}
	}

	// Fallback: get text from source
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	text := ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
	return strings.TrimSpace(text)
}

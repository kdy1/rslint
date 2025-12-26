package no_wrapper_object_types

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Map of wrapper object types to their primitive equivalents
var wrapperTypeToPrimitive = map[string]string{
	"BigInt":  "bigint",
	"Boolean": "boolean",
	"Number":  "number",
	"Object":  "object",
	"String":  "string",
	"Symbol":  "symbol",
}

func isWrapperType(name string) bool {
	_, exists := wrapperTypeToPrimitive[name]
	return exists
}

func getPrimitiveType(wrapperType string) string {
	return wrapperTypeToPrimitive[wrapperType]
}

func isInExtendsOrImplementsClause(node *ast.Node) bool {
	// Check if the type reference is in an extends or implements clause
	// We need to walk up the parent chain looking for a HeritageClause
	parent := node
	for parent != nil {
		if parent.Kind == ast.KindHeritageClause {
			return true
		}
		// Stop searching if we reach a class or interface declaration without finding a heritage clause
		// (but keep going if we're still inside the type reference itself)
		if parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindInterfaceDeclaration {
			return false
		}
		parent = parent.Parent
	}
	return false
}

func hasLocalTypeDeclaration(symbol *ast.Symbol) bool {
	// Check if the symbol has a local type declaration (type alias, type parameter)
	// These are always user-defined, never built-in wrapper types
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}

	for _, decl := range symbol.Declarations {
		if decl.Kind == ast.KindTypeAliasDeclaration || decl.Kind == ast.KindTypeParameter {
			return true
		}
	}
	return false
}

func isShadowedByLocalDeclaration(ctx rule.RuleContext, node *ast.Node, typeName string) bool {
	// Check if there's a local declaration that shadows the global wrapper type
	// The main case is a type alias or type parameter with the same name

	// Get the symbol at this location
	if ctx.TypeChecker == nil {
		return false
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return false
	}

	// If there's a type alias or type parameter, it's definitely a local shadow
	if hasLocalTypeDeclaration(symbol) {
		return true
	}

	return false
}

var NoWrapperObjectTypesRule = rule.CreateRule(rule.Rule{
	Name: "no-wrapper-object-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindTypeReference: func(node *ast.Node) {
				typeRef := node.AsTypeReference()
				if typeRef == nil {
					return
				}

				// Check if the type name is an identifier
				if !ast.IsIdentifier(typeRef.TypeName) {
					return
				}

				identifier := typeRef.TypeName.AsIdentifier()
				if identifier == nil {
					return
				}

				typeName := identifier.Text

				// Check if it's a wrapper type
				if !isWrapperType(typeName) {
					return
				}

				// Check if this is in an extends or implements clause
				// In these contexts, we can't auto-fix because the semantics are different
				if isInExtendsOrImplementsClause(node) {
					// Report without fix
					primitiveType := getPrimitiveType(typeName)
					ctx.ReportNode(typeRef.TypeName, rule.RuleMessage{
						Id:          "bannedClassType",
						Description: "Prefer using the primitive '" + primitiveType + "' type instead of the wrapper '" + typeName + "' type.",
					})
					return
				}

				// Check if this wrapper type is shadowed by a local declaration
				if isShadowedByLocalDeclaration(ctx, typeRef.TypeName, typeName) {
					// This is a local type, not the global wrapper type, so don't report
					return
				}

				// Get the primitive type
				primitiveType := getPrimitiveType(typeName)

				// Report with auto-fix
				ctx.ReportNodeWithFixes(
					typeRef.TypeName,
					rule.RuleMessage{
						Id:          "bannedClassType",
						Description: "Prefer using the primitive '" + primitiveType + "' type instead of the wrapper '" + typeName + "' type.",
					},
					rule.RuleFixReplace(ctx.SourceFile, typeRef.TypeName, primitiveType),
				)
			},
		}
	},
})

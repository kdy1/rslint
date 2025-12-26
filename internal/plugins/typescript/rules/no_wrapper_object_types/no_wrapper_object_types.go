package no_wrapper_object_types

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var wrapperObjectTypes = map[string]string{
	"BigInt":  "bigint",
	"Boolean": "boolean",
	"Number":  "number",
	"Object":  "object",
	"String":  "string",
	"Symbol":  "symbol",
}

func buildNoWrapperObjectTypesMessage(typeName, preferred string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "bannedClassType",
		Description: "Use `" + preferred + "` instead of `" + typeName + "`.",
	}
}

var NoWrapperObjectTypesRule = rule.CreateRule(rule.Rule{
	Name: "no-wrapper-object-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		checkTypeReference := func(node *ast.Node) {
			typeRef := node.AsTypeReferenceNode()
			if typeRef == nil || typeRef.TypeName == nil {
				return
			}

			// Get the type name - must be an identifier
			var typeName string
			if ast.IsIdentifier(typeRef.TypeName) {
				ident := typeRef.TypeName.AsIdentifier()
				if ident != nil {
					typeName = ident.Text
				}
			} else {
				// For qualified names or other complex type names, skip
				return
			}

			// Check if this is a wrapper object type
			preferred, isWrapper := wrapperObjectTypes[typeName]
			if !isWrapper {
				return
			}

			// Check if this identifier is referencing a type parameter, type alias, or interface
			// by checking if there's a local declaration in scope
			// We need to skip if the type name is actually a user-defined type

			// Simple heuristic: check if we're in a scope where this name is defined as a type
			// Walk up the parent chain to look for declarations
			isLocalType := false
			current := node.Parent

			for current != nil {
				// Check for type alias declaration with the same name
				if ast.IsTypeAliasDeclaration(current) {
					typeAlias := current.AsTypeAliasDeclaration()
					if typeAlias != nil && typeAlias.Name() != nil {
						if nameNode := typeAlias.Name(); nameNode.Kind == ast.KindIdentifier {
							nameIdent := nameNode.AsIdentifier()
							if nameIdent != nil && nameIdent.Text == typeName {
								// We found a type alias with this name in the parent scope
								isLocalType = true
								break
							}
						}
					}
				}

				// Check for interface declaration with the same name
				if ast.IsInterfaceDeclaration(current) {
					interfaceDecl := current.AsInterfaceDeclaration()
					if interfaceDecl != nil && interfaceDecl.Name() != nil {
						if nameNode := interfaceDecl.Name(); nameNode.Kind == ast.KindIdentifier {
							nameIdent := nameNode.AsIdentifier()
							if nameIdent != nil && nameIdent.Text == typeName {
								isLocalType = true
								break
							}
						}
					}
				}

				current = current.Parent
			}

			if isLocalType {
				return
			}

			// Check if this is in an extends/implements clause
			// If so, we cannot auto-fix AND we should not report at all in this initial implementation
			// because it's a complex case that requires renaming the entire class/interface
			parent := node.Parent
			if parent != nil {
				// Check if parent is ExpressionWithTypeArguments in a heritage clause
				if ast.IsExpressionWithTypeArguments(parent) {
					// Check if grandparent is HeritageClause
					grandparent := parent.Parent
					if grandparent != nil && ast.IsHeritageClause(grandparent) {
						// For extends/implements, don't report in this version
						return
					}
				}
			}

			shouldAutoFix := true

			// Report the wrapper object type
			if shouldAutoFix {
				ctx.ReportNodeWithFixes(
					node,
					buildNoWrapperObjectTypesMessage(typeName, preferred),
					rule.RuleFixReplace(ctx.SourceFile, node, preferred),
				)
			} else {
				// Report without autofix for extends/implements
				ctx.ReportNode(
					node,
					buildNoWrapperObjectTypesMessage(typeName, preferred),
				)
			}
		}

		return rule.RuleListeners{
			ast.KindTypeReference: checkTypeReference,
		}
	},
})

package no_unnecessary_qualifier

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildUnnecessaryQualifierMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessaryQualifier",
		Description: "Qualifier is unnecessary since it is already in scope.",
	}
}

// getLeftmostIdentifier gets the leftmost identifier from a qualified name or property access chain
func getLeftmostIdentifier(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node
	case ast.KindQualifiedName:
		qualifiedName := node.AsQualifiedName()
		if qualifiedName != nil {
			return getLeftmostIdentifier(qualifiedName.Left)
		}
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		if propAccess != nil {
			return getLeftmostIdentifier(propAccess.Expression)
		}
	}

	return nil
}

// getRightmostName gets the rightmost name from a qualified name or property access chain
func getRightmostName(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node.Text()
	case ast.KindQualifiedName:
		qualifiedName := node.AsQualifiedName()
		if qualifiedName != nil && qualifiedName.Right != nil {
			return qualifiedName.Right.Text()
		}
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		if propAccess != nil && propAccess.Name != nil {
			// Name is a function, we need to call it
			name := propAccess.Name()
			if ast.IsIdentifier(name) {
				return name.Text()
			}
		}
	}

	return ""
}

// isNamespaceOrEnumDeclaration checks if a node is a namespace or enum declaration
func isNamespaceOrEnumDeclaration(node *ast.Node) bool {
	if node == nil {
		return false
	}

	return ast.IsModuleDeclaration(node) || ast.IsEnumDeclaration(node)
}

// getContainingNamespaceOrEnum finds the containing namespace or enum declaration for a node
func getContainingNamespaceOrEnum(node *ast.Node) *ast.Node {
	current := node
	for current != nil {
		if isNamespaceOrEnumDeclaration(current) {
			return current
		}
		current = current.Parent
	}
	return nil
}

// checkIfQualifierIsUnnecessary checks if a qualifier is unnecessary
func checkIfQualifierIsUnnecessary(ctx rule.RuleContext, node *ast.Node, qualifier *ast.Node) bool {
	// Get the symbol of the leftmost identifier in the qualifier
	leftmostId := getLeftmostIdentifier(qualifier)
	if leftmostId == nil {
		return false
	}

	qualifierSymbol := ctx.TypeChecker.GetSymbolAtLocation(leftmostId)
	if qualifierSymbol == nil {
		return false
	}

	// Get the containing namespace or enum
	containingScope := getContainingNamespaceOrEnum(node)
	if containingScope == nil {
		return false
	}

	// Check if we're referencing the containing scope
	containingScopeName := containingScope.Name()
	if containingScopeName == nil {
		return false
	}

	containingScopeSymbol := ctx.TypeChecker.GetSymbolAtLocation(containingScopeName)
	if containingScopeSymbol == nil {
		return false
	}

	// If the qualifier's symbol matches the containing scope's symbol, it's unnecessary
	if qualifierSymbol == containingScopeSymbol {
		return true
	}

	// Check parent scopes for nested namespaces
	currentScope := containingScope
	for currentScope != nil {
		scopeName := currentScope.Name()
		if scopeName != nil {
			scopeSymbol := ctx.TypeChecker.GetSymbolAtLocation(scopeName)
			if scopeSymbol == qualifierSymbol {
				return true
			}
		}
		currentScope = getContainingNamespaceOrEnum(currentScope.Parent)
	}

	return false
}

var NoUnnecessaryQualifierRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-qualifier",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		checkQualifiedName := func(node *ast.Node) {
			if !ast.IsQualifiedName(node) {
				return
			}

			qualifiedName := node.AsQualifiedName()
			if qualifiedName == nil || qualifiedName.Left == nil {
				return
			}

			if checkIfQualifierIsUnnecessary(ctx, node, qualifiedName.Left) {
				replacement := getRightmostName(ctx, node)
				if replacement == "" {
					ctx.ReportNode(node, buildUnnecessaryQualifierMessage())
					return
				}

				ctx.ReportNodeWithFixes(
					node,
					buildUnnecessaryQualifierMessage(),
					rule.RuleFixReplace(ctx.SourceFile, node, replacement),
				)
			}
		}

		checkPropertyAccess := func(node *ast.Node) {
			if !ast.IsPropertyAccessExpression(node) {
				return
			}

			propAccess := node.AsPropertyAccessExpression()
			if propAccess == nil || propAccess.Expression == nil {
				return
			}

			if checkIfQualifierIsUnnecessary(ctx, node, propAccess.Expression) {
				replacement := getRightmostName(ctx, node)
				if replacement == "" {
					ctx.ReportNode(node, buildUnnecessaryQualifierMessage())
					return
				}

				ctx.ReportNodeWithFixes(
					node,
					buildUnnecessaryQualifierMessage(),
					rule.RuleFixReplace(ctx.SourceFile, node, replacement),
				)
			}
		}

		checkTypeReference := func(node *ast.Node) {
			if !ast.IsTypeReferenceNode(node) {
				return
			}

			typeRef := node.AsTypeReferenceNode()
			if typeRef == nil || typeRef.TypeName == nil {
				return
			}

			typeName := typeRef.TypeName

			// Only check qualified names
			if ast.IsQualifiedName(typeName) {
				checkQualifiedName(typeName)
			}
		}

		return rule.RuleListeners{
			ast.KindQualifiedName: checkQualifiedName,
			ast.KindPropertyAccessExpression: checkPropertyAccess,
			ast.KindTypeReference: checkTypeReference,
		}
	},
})

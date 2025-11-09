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

// getRightmostIdentifier gets the rightmost identifier from a qualified name or property access chain
func getRightmostIdentifier(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node
	case ast.KindQualifiedName:
		qualifiedName := node.AsQualifiedName()
		if qualifiedName != nil && qualifiedName.Right != nil {
			return qualifiedName.Right
		}
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		if propAccess != nil {
			name := propAccess.Name()
			if ast.IsIdentifier(name) {
				return name
			}
		}
	}

	return nil
}

// getRightmostName gets the rightmost name from a qualified name or property access chain
func getRightmostName(ctx rule.RuleContext, node *ast.Node) string {
	identifier := getRightmostIdentifier(node)
	if identifier != nil {
		return identifier.Text()
	}
	return ""
}

// isNamespaceOrEnumDeclaration checks if a node is a namespace, module or enum declaration
func isNamespaceOrEnumDeclaration(node *ast.Node) bool {
	if node == nil {
		return false
	}

	return ast.IsModuleDeclaration(node) || ast.IsEnumDeclaration(node)
}

// getContainingNamespacesAndEnums finds all containing namespace or enum declarations for a node
func getContainingNamespacesAndEnums(node *ast.Node) []*ast.Node {
	result := []*ast.Node{}
	current := node.Parent
	for current != nil {
		if isNamespaceOrEnumDeclaration(current) {
			result = append([]*ast.Node{current}, result...) // Prepend to maintain order from outer to inner
		}
		current = current.Parent
	}
	return result
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

	// Get all containing namespaces/enums
	namespacesInScope := getContainingNamespacesAndEnums(node)

	// Check if the qualifier symbol is a namespace in scope
	if !symbolIsNamespaceInScope(ctx, qualifierSymbol, namespacesInScope) {
		return false
	}

	// Get the full symbol being accessed (the whole qualified name)
	accessedSymbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if accessedSymbol == nil {
		return false
	}

	// Special case: if the accessed symbol is an enum member, check if we're inside the same enum
	// If we're accessing an enum member from outside the enum, the qualifier is always necessary
	// But if we're inside the enum, we can access members without the qualifier
	if accessedSymbol.Flags&ast.SymbolFlagsEnumMember != 0 {
		// Check if we're inside an enum declaration
		insideEnum := false
		for _, scope := range namespacesInScope {
			if ast.IsEnumDeclaration(scope) {
				// Check if this enum contains the accessed enum member
				enumSymbol := ctx.TypeChecker.GetSymbolAtLocation(scope.Name())
				if enumSymbol != nil && enumSymbol.Exports != nil {
					// Check if the accessed symbol is a member of this enum
					for _, member := range enumSymbol.Exports {
						if member == accessedSymbol {
							insideEnum = true
							break
						}
					}
				}
				if insideEnum {
					break
				}
			}
		}

		// If we're not inside the enum that contains this member, the qualifier is necessary
		if !insideEnum {
			return false
		}
	}

	// Get the rightmost identifier (the actual member being accessed)
	rightmostId := getRightmostIdentifier(node)
	if rightmostId == nil {
		return false
	}

	// The key insight: if the rightmost identifier can be resolved to the same symbol
	// without the qualifier, then the qualifier is unnecessary.
	// We use the TypeChecker to resolve the symbol at the rightmost identifier location.
	// This gives us what the identifier would resolve to in the current scope.
	unqualifiedSymbol := ctx.TypeChecker.GetSymbolAtLocation(rightmostId)

	// If the rightmost identifier doesn't resolve to anything, the qualifier is necessary
	if unqualifiedSymbol == nil {
		return false
	}

	// Check if the symbols are equal (refer to the same export)
	// If they're not equal, it means the unqualified name resolves to something different
	// (e.g., a shadowing symbol), so the qualifier IS necessary
	return symbolsAreEqual(ctx, accessedSymbol, unqualifiedSymbol)
}

// symbolIsNamespaceInScope checks if a symbol is a namespace/module/enum declaration in the current scope
func symbolIsNamespaceInScope(ctx rule.RuleContext, symbol *ast.Symbol, namespacesInScope []*ast.Node) bool {
	if symbol == nil {
		return false
	}

	declarations := symbol.Declarations
	if declarations == nil {
		return false
	}

	for _, decl := range declarations {
		if decl != nil {
			for _, ns := range namespacesInScope {
				if decl == ns {
					return true
				}
			}
		}
	}

	// Check if this is an aliased symbol - only call GetAliasedSymbol if this is actually an alias
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		if alias := ctx.TypeChecker.GetAliasedSymbol(symbol); alias != nil && alias != symbol {
			return symbolIsNamespaceInScope(ctx, alias, namespacesInScope)
		}
	}

	return false
}

// symbolsAreEqual checks if two symbols refer to the same export
func symbolsAreEqual(ctx rule.RuleContext, accessed *ast.Symbol, inScope *ast.Symbol) bool {
	if accessed == nil || inScope == nil {
		return false
	}

	// Direct equality
	if accessed == inScope {
		return true
	}

	// Check if the export symbol of inScope equals accessed
	exportSymbol := ctx.TypeChecker.GetExportSymbolOfSymbol(inScope)
	if exportSymbol != nil && exportSymbol == accessed {
		return true
	}

	// Check if both have the same export symbol
	accessedExport := ctx.TypeChecker.GetExportSymbolOfSymbol(accessed)
	if accessedExport != nil && exportSymbol != nil && accessedExport == exportSymbol {
		return true
	}

	return false
}

// isPartOfLargerQualifiedNameOrPropertyAccess checks if this node is part of a larger qualified name chain
// For example, in "A.B.C", the node "A.B" is part of a larger chain
func isPartOfLargerQualifiedNameOrPropertyAccess(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	// Check if parent is a QualifiedName and this node is its left side
	if ast.IsQualifiedName(parent) {
		qn := parent.AsQualifiedName()
		if qn != nil && qn.Left == node {
			return true
		}
	}

	// Check if parent is a PropertyAccessExpression and this node is its expression
	if ast.IsPropertyAccessExpression(parent) {
		pa := parent.AsPropertyAccessExpression()
		if pa != nil && pa.Expression == node {
			return true
		}
	}

	return false
}

var NoUnnecessaryQualifierRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-qualifier",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		checkNode := func(node *ast.Node, qualifier *ast.Node) {
			// Skip if this node is part of a larger qualified name/property access chain
			// We only want to check the outermost/complete qualified name
			if isPartOfLargerQualifiedNameOrPropertyAccess(node) {
				return
			}

			if checkIfQualifierIsUnnecessary(ctx, node, qualifier) {
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

		checkQualifiedName := func(node *ast.Node) {
			if !ast.IsQualifiedName(node) {
				return
			}

			qualifiedName := node.AsQualifiedName()
			if qualifiedName == nil || qualifiedName.Left == nil {
				return
			}

			checkNode(node, qualifiedName.Left)
		}

		checkPropertyAccess := func(node *ast.Node) {
			if !ast.IsPropertyAccessExpression(node) {
				return
			}

			propAccess := node.AsPropertyAccessExpression()
			if propAccess == nil || propAccess.Expression == nil {
				return
			}

			checkNode(node, propAccess.Expression)
		}

		return rule.RuleListeners{
			ast.KindQualifiedName:            checkQualifiedName,
			ast.KindPropertyAccessExpression: checkPropertyAccess,
		}
	},
})

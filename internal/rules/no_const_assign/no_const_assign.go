package no_const_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builder
func buildConstMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "const",
		Description: "'" + name + "' is constant.",
	}
}

// isConstBinding checks if a variable declaration is a const binding
func isConstBinding(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindVariableDeclarationList {
		return false
	}

	varDeclList := node.AsVariableDeclarationList()
	if varDeclList == nil {
		return false
	}

	// Check if the declaration is const (or using/await using in the future)
	// In TypeScript AST, const declarations have flags
	return (varDeclList.Flags & ast.NodeFlagsConst) != 0
}

// getIdentifierName gets the name of an identifier node
func getIdentifierName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindIdentifier {
		return ""
	}

	identifier := node.AsIdentifier()
	if identifier == nil {
		return ""
	}

	return identifier.EscapedText
}

// isWriteReference checks if a reference is a write operation (assignment, increment, decrement, etc.)
func isWriteReference(node *ast.Node) bool {
	if node == nil {
		return false
	}

	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindBinaryExpression:
		// Check if this is an assignment operation
		binary := parent.AsBinaryExpression()
		if binary == nil {
			return false
		}

		// Check if the node is on the left side of an assignment
		if binary.Left != node {
			return false
		}

		// Check for all assignment operators
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken, // =
			ast.KindPlusEqualsToken,              // +=
			ast.KindMinusEqualsToken,             // -=
			ast.KindAsteriskEqualsToken,          // *=
			ast.KindSlashEqualsToken,             // /=
			ast.KindPercentEqualsToken,           // %=
			ast.KindAsteriskAsteriskEqualsToken,  // **=
			ast.KindLessThanLessThanEqualsToken,  // <<=
			ast.KindGreaterThanGreaterThanEqualsToken,        // >>=
			ast.KindGreaterThanGreaterThanGreaterThanEqualsToken, // >>>=
			ast.KindAmpersandEqualsToken,         // &=
			ast.KindBarEqualsToken,               // |=
			ast.KindCaretEqualsToken,             // ^=
			ast.KindQuestionQuestionEqualsToken,  // ??=
			ast.KindAmpersandAmpersandEqualsToken, // &&=
			ast.KindBarBarEqualsToken:            // ||=
			return true
		}

	case ast.KindPrefixUnaryExpression:
		// Check for ++ and -- prefix operators
		prefix := parent.AsPrefixUnaryExpression()
		if prefix == nil {
			return false
		}

		switch prefix.Operator {
		case ast.KindPlusPlusToken,   // ++
			ast.KindMinusMinusToken: // --
			return true
		}

	case ast.KindPostfixUnaryExpression:
		// Check for ++ and -- postfix operators
		postfix := parent.AsPostfixUnaryExpression()
		if postfix == nil {
			return false
		}

		switch postfix.Operator {
		case ast.KindPlusPlusToken,   // ++
			ast.KindMinusMinusToken: // --
			return true
		}
	}

	return false
}

// isInInitializer checks if a node is in the initializer of its declaration
func isInInitializer(identifierNode *ast.Node, declNode *ast.Node) bool {
	if identifierNode == nil || declNode == nil {
		return false
	}

	// Walk up from the identifier to see if we're in the initializer
	current := identifierNode.Parent
	for current != nil && current != declNode {
		// If we hit a VariableDeclaration, check if we're in its initializer
		if current.Kind == ast.KindVariableDeclaration {
			varDecl := current.AsVariableDeclaration()
			if varDecl != nil && varDecl.Initializer != nil {
				// Check if the identifier is within the initializer
				if containsNode(varDecl.Initializer, identifierNode) {
					return true
				}
			}
			break
		}
		current = current.Parent
	}

	return false
}

// containsNode checks if a root node contains a target node in its subtree
func containsNode(root, target *ast.Node) bool {
	if root == nil || target == nil {
		return false
	}
	if root == target {
		return true
	}

	// Walk up from target to see if we reach root
	current := target.Parent
	for current != nil {
		if current == root {
			return true
		}
		current = current.Parent
	}

	return false
}

// findVariableDeclaration finds the declaration for a given identifier
func findVariableDeclaration(identifier *ast.Node, variableDeclarationList *ast.Node) *ast.Node {
	if identifier == nil || variableDeclarationList == nil {
		return nil
	}

	identName := getIdentifierName(identifier)
	if identName == "" {
		return nil
	}

	varDeclList := variableDeclarationList.AsVariableDeclarationList()
	if varDeclList == nil || varDeclList.Declarations == nil {
		return nil
	}

	// Search through all declarations
	for _, decl := range varDeclList.Declarations.Slice() {
		if decl.Kind != ast.KindVariableDeclaration {
			continue
		}

		varDecl := decl.AsVariableDeclaration()
		if varDecl == nil || varDecl.Name == nil {
			continue
		}

		// Check if this declaration matches our identifier
		if matchesIdentifier(varDecl.Name, identName) {
			return decl
		}
	}

	return nil
}

// matchesIdentifier checks if a binding name matches an identifier
func matchesIdentifier(bindingName *ast.Node, identName string) bool {
	if bindingName == nil {
		return false
	}

	switch bindingName.Kind {
	case ast.KindIdentifier:
		return getIdentifierName(bindingName) == identName

	case ast.KindObjectBindingPattern:
		// Check if any element in the object binding matches
		objBinding := bindingName.AsObjectBindingPattern()
		if objBinding == nil || objBinding.Elements == nil {
			return false
		}

		for _, elem := range objBinding.Elements.Slice() {
			if elem.Kind == ast.KindBindingElement {
				bindingElem := elem.AsBindingElement()
				if bindingElem != nil && bindingElem.Name != nil {
					if matchesIdentifier(bindingElem.Name, identName) {
						return true
					}
				}
			}
		}

	case ast.KindArrayBindingPattern:
		// Check if any element in the array binding matches
		arrBinding := bindingName.AsArrayBindingPattern()
		if arrBinding == nil || arrBinding.Elements == nil {
			return false
		}

		for _, elem := range arrBinding.Elements.Slice() {
			if elem.Kind == ast.KindBindingElement {
				bindingElem := elem.AsBindingElement()
				if bindingElem != nil && bindingElem.Name != nil {
					if matchesIdentifier(bindingElem.Name, identName) {
						return true
					}
				}
			}
		}
	}

	return false
}

// NoConstAssignRule disallows reassigning const variables
var NoConstAssignRule = rule.CreateRule(rule.Rule{
	Name: "no-const-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Track const declarations and their identifiers
		constDeclarations := make(map[string]*ast.Node) // maps identifier name to declaration list node

		return rule.RuleListeners{
			// Track const variable declarations
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if !isConstBinding(node) {
					return
				}

				varDeclList := node.AsVariableDeclarationList()
				if varDeclList == nil || varDeclList.Declarations == nil {
					return
				}

				// Track all identifiers declared as const
				for _, decl := range varDeclList.Declarations.Slice() {
					if decl.Kind != ast.KindVariableDeclaration {
						continue
					}

					varDecl := decl.AsVariableDeclaration()
					if varDecl == nil || varDecl.Name == nil {
						continue
					}

					// Collect all identifiers from the binding name
					collectIdentifiers(varDecl.Name, node, constDeclarations)
				}
			},

			// Check for reassignments to const variables
			ast.KindIdentifier: func(node *ast.Node) {
				identName := getIdentifierName(node)
				if identName == "" {
					return
				}

				// Check if this identifier refers to a const variable
				declListNode, isConst := constDeclarations[identName]
				if !isConst {
					return
				}

				// Check if this is a write reference (assignment, increment, etc.)
				if !isWriteReference(node) {
					return
				}

				// Find the specific variable declaration for this identifier
				varDecl := findVariableDeclaration(node, declListNode)
				if varDecl == nil {
					return
				}

				// Don't report if this is part of the initializer
				if isInInitializer(node, varDecl) {
					return
				}

				// Report the violation
				ctx.ReportNode(node, buildConstMessage(identName))
			},
		}
	},
})

// collectIdentifiers recursively collects all identifiers from a binding pattern
func collectIdentifiers(bindingName *ast.Node, declListNode *ast.Node, constDeclarations map[string]*ast.Node) {
	if bindingName == nil {
		return
	}

	switch bindingName.Kind {
	case ast.KindIdentifier:
		name := getIdentifierName(bindingName)
		if name != "" {
			constDeclarations[name] = declListNode
		}

	case ast.KindObjectBindingPattern:
		objBinding := bindingName.AsObjectBindingPattern()
		if objBinding == nil || objBinding.Elements == nil {
			return
		}

		for _, elem := range objBinding.Elements.Slice() {
			if elem.Kind == ast.KindBindingElement {
				bindingElem := elem.AsBindingElement()
				if bindingElem != nil && bindingElem.Name != nil {
					collectIdentifiers(bindingElem.Name, declListNode, constDeclarations)
				}
			}
		}

	case ast.KindArrayBindingPattern:
		arrBinding := bindingName.AsArrayBindingPattern()
		if arrBinding == nil || arrBinding.Elements == nil {
			return
		}

		for _, elem := range arrBinding.Elements.Slice() {
			if elem.Kind == ast.KindBindingElement {
				bindingElem := elem.AsBindingElement()
				if bindingElem != nil && bindingElem.Name != nil {
					collectIdentifiers(bindingElem.Name, declListNode, constDeclarations)
				}
			}
		}
	}
}

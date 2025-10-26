package no_import_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoImportAssignRule implements the no-import-assign rule
// Disallow assigning to imported bindings
var NoImportAssignRule = rule.Rule{
	Name: "no-import-assign",
	Run:  run,
}

func buildReadonlyMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "readonly",
		Description: "'" + name + "' is read-only.",
	}
}

func buildReadonlyMemberMessage(name string, member string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "readonlyMember",
		Description: "The members of '" + name + "' are read-only.",
	}
}

// isCompoundAssignmentOperator checks if a token is a compound assignment operator
func isCompoundAssignmentOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindPlusEqualsToken,
		ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken,
		ast.KindAsteriskAsteriskEqualsToken,
		ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken,
		ast.KindBarBarEqualsToken,
		ast.KindAmpersandAmpersandEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// isAssignmentLeft checks if a node is on the left side of an assignment
func isAssignmentLeft(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	// Direct assignment: x = value
	if parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsToken || isCompoundAssignmentOperator(binary.OperatorToken.Kind) {
			return binary.Left == node
		}
	}

	// Destructuring patterns: [x] = arr, {prop: x} = obj
	// Check if we're in a destructuring pattern by looking up the tree
	current := node
	for current != nil {
		if current.Parent == nil {
			break
		}

		switch current.Parent.Kind {
		case ast.KindArrayLiteralExpression:
			// Continue checking if this array is in a destructuring position
			current = current.Parent
			continue

		case ast.KindPropertyAssignment:
			// In object destructuring, we care about the initializer
			prop := current.Parent.AsPropertyAssignment()
			if prop.Initializer == current {
				current = current.Parent
				continue
			}

		case ast.KindShorthandPropertyAssignment:
			current = current.Parent
			continue

		case ast.KindObjectLiteralExpression:
			current = current.Parent
			continue

		case ast.KindSpreadElement:
			current = current.Parent
			continue

		case ast.KindBinaryExpression:
			binary := current.Parent.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindEqualsToken {
				// Found the assignment, check if we're on the left
				return binary.Left == current
			}
			return false

		case ast.KindVariableDeclaration:
			// In a variable declaration with initializer
			return true

		default:
			return false
		}
	}

	return false
}

// isOperandOfMutationUnaryOperator checks if a node is operand of ++, --, or delete
func isOperandOfMutationUnaryOperator(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	switch parent.Kind {
	case ast.KindPrefixUnaryExpression:
		prefix := parent.AsPrefixUnaryExpression()
		if prefix.Operand == node {
			switch prefix.Operator {
			case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
				return true
			}
		}
	case ast.KindPostfixUnaryExpression:
		postfix := parent.AsPostfixUnaryExpression()
		if postfix.Operand == node {
			switch postfix.Operator {
			case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
				return true
			}
		}
	case ast.KindDeleteExpression:
		deleteExpr := parent.AsDeleteExpression()
		return deleteExpr.Expression == node
	}

	return false
}

// isIterationVariable checks if a node is a for-in or for-of loop variable
func isIterationVariable(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	// Check if parent is a for-in or for-of statement
	parent := node.Parent
	if parent.Kind == ast.KindForInStatement || parent.Kind == ast.KindForOfStatement {
		stmt := parent.AsForInOrOfStatement()
		return stmt.Initializer == node
	}

	// Check if we're inside a variable declaration in a for-in/for-of
	if parent.Kind == ast.KindVariableDeclaration {
		if parent.Parent != nil {
			grandparent := parent.Parent
			if grandparent.Kind == ast.KindVariableDeclarationList {
				if grandparent.Parent != nil {
					greatGrandparent := grandparent.Parent
					if greatGrandparent.Kind == ast.KindForInStatement || greatGrandparent.Kind == ast.KindForOfStatement {
						return true
					}
				}
			}
		}
	}

	return false
}

// isArgumentOfWellKnownMutationFunction checks if node is first argument to Object.assign, etc.
func isArgumentOfWellKnownMutationFunction(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent
	if parent.Kind != ast.KindCallExpression {
		return false
	}

	call := parent.AsCallExpression()

	// Check if node is the first argument
	if len(call.Arguments.Nodes) == 0 || call.Arguments.Nodes[0] != node {
		return false
	}

	// Get the callee
	callee := call.Expression

	// Handle optional chaining: Object?.defineProperty
	if callee.Kind == ast.KindNonNullExpression {
		callee = callee.AsNonNullExpression().Expression
	}

	// Handle parenthesized expression: (Object.defineProperty)
	for callee.Kind == ast.KindParenthesizedExpression {
		callee = callee.AsParenthesizedExpression().Expression
	}

	// Check for property access: Object.method
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}

	propAccess := callee.AsPropertyAccessExpression()
	obj := propAccess.Expression
	method := propAccess.Name()

	if obj == nil || method == nil {
		return false
	}

	// Get object name
	var objectName string
	if obj.Kind == ast.KindIdentifier {
		objectName = obj.AsIdentifier().Text
	} else {
		return false
	}

	// Get method name
	var methodName string
	if method.Kind == ast.KindIdentifier {
		methodName = method.AsIdentifier().Text
	} else {
		return false
	}

	// Check against well-known mutation functions
	if objectName == "Object" {
		switch methodName {
		case "assign", "defineProperty", "defineProperties", "freeze", "setPrototypeOf":
			return true
		}
	} else if objectName == "Reflect" {
		switch methodName {
		case "defineProperty", "deleteProperty", "set", "setPrototypeOf":
			return true
		}
	}

	return false
}

// isMemberWrite checks if a member access is being written to
func isMemberWrite(node *ast.Node) bool {
	if node == nil {
		return false
	}

	return isAssignmentLeft(node) ||
		isOperandOfMutationUnaryOperator(node) ||
		isIterationVariable(node) ||
		isArgumentOfWellKnownMutationFunction(node)
}

// getIdentifierName safely gets the text of an identifier
func getIdentifierName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindIdentifier {
		return ""
	}
	return node.AsIdentifier().Text
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// Track all imported identifiers
	importedNames := make(map[string]bool)    // Regular imports
	namespaceImports := make(map[string]bool) // Namespace imports (import * as x)

	return rule.RuleListeners{
		ast.KindImportDeclaration: func(node *ast.Node) {
			importDecl := node.AsImportDeclaration()
			if importDecl.ImportClause == nil {
				return
			}

			clause := importDecl.ImportClause

			// Handle default import: import x from 'mod'
			if clause.Name != nil {
				name := getIdentifierName(clause.Name)
				if name != "" {
					importedNames[name] = true
				}
			}

			// Handle named bindings
			if clause.NamedBindings != nil {
				bindings := clause.NamedBindings

				// Handle namespace import: import * as x from 'mod'
				if bindings.Kind == ast.KindNamespaceImport {
					nsImport := bindings.AsNamespaceImport()
					if nsImport.Name != nil {
						name := getIdentifierName(nsImport.Name)
						if name != "" {
							namespaceImports[name] = true
						}
					}
				}

				// Handle named imports: import { x, y as z } from 'mod'
				if bindings.Kind == ast.KindNamedImports {
					namedImports := bindings.AsNamedImports()
					for _, element := range namedImports.Elements.Nodes {
						if element.Kind == ast.KindImportSpecifier {
							spec := element.AsImportSpecifier()
							// Use the local name (or the imported name if no alias)
							var name string
							if spec.Name != nil {
								name = getIdentifierName(spec.Name)
							} else if spec.PropertyName != nil {
								name = getIdentifierName(spec.PropertyName)
							}
							if name != "" {
								importedNames[name] = true
							}
						}
					}
				}
			}
		},

		// Check all identifier references
		ast.KindIdentifier: func(node *ast.Node) {
			name := getIdentifierName(node)
			if name == "" {
				return
			}

			// Skip if this identifier is part of an import declaration itself
			if node.Parent != nil {
				parent := node.Parent
				// Skip identifiers that are part of the import declaration
				if parent.Kind == ast.KindImportSpecifier ||
					parent.Kind == ast.KindNamespaceImport ||
					parent.Kind == ast.KindImportClause {
					return
				}
			}

			// Check if it's a regular imported name being assigned
			if importedNames[name] {
				if isAssignmentLeft(node) || isOperandOfMutationUnaryOperator(node) || isIterationVariable(node) {
					ctx.ReportNode(node, buildReadonlyMessage(name))
				}
			}
		},

		// Check property access expressions for namespace imports
		ast.KindPropertyAccessExpression: func(node *ast.Node) {
			propAccess := node.AsPropertyAccessExpression()
			if propAccess.Expression == nil || propAccess.Name() == nil {
				return
			}

			// Check if the object is a namespace import
			if propAccess.Expression.Kind == ast.KindIdentifier {
				objName := getIdentifierName(propAccess.Expression)

				if namespaceImports[objName] {
					// Check if this member is being written to
					if isMemberWrite(node) {
						ctx.ReportNode(node, buildReadonlyMemberMessage(objName, getIdentifierName(propAccess.Name())))
					}
				}
			}
		},

		// Check delete expressions on optional chaining: delete mod?.prop
		ast.KindDeleteExpression: func(node *ast.Node) {
			deleteExpr := node.AsDeleteExpression()
			expr := deleteExpr.Expression

			// Handle delete on property access with optional chaining
			// The property access would have already been checked, but we need to ensure
			// optional chaining delete is caught
			if expr != nil && expr.Kind == ast.KindPropertyAccessExpression {
				propAccess := expr.AsPropertyAccessExpression()
				if propAccess.Expression != nil && propAccess.Expression.Kind == ast.KindIdentifier {
					objName := getIdentifierName(propAccess.Expression)
					if namespaceImports[objName] {
						ctx.ReportNode(expr, buildReadonlyMemberMessage(objName, getIdentifierName(propAccess.Name())))
					}
				}
			}
		},
	}
}

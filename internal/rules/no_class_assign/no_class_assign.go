package no_class_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builder
func buildClassReassignmentMessage(className string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "classReassignment",
		Description: "'" + className + "' is a class.",
		Data: map[string]string{
			"name": className,
		},
	}
}

// getIdentifierName extracts the name from an identifier node
func getIdentifierName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindIdentifier {
		return ""
	}
	return node.Text()
}

// isWriteReference checks if a node is a write reference (assignment target)
func isWriteReference(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}

		// Check if it's an assignment operator and node is on the left side
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken,
			ast.KindPlusEqualsToken,
			ast.KindMinusEqualsToken,
			ast.KindAsteriskAsteriskEqualsToken,
			ast.KindAsteriskEqualsToken,
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
			return binary.Left == node
		}

	case ast.KindPostfixUnaryExpression:
		postfix := parent.AsPostfixUnaryExpression()
		if postfix == nil {
			return false
		}
		// ++ and -- are write operations
		switch postfix.Operator {
		case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			return postfix.Operand == node
		}

	case ast.KindPrefixUnaryExpression:
		prefix := parent.AsPrefixUnaryExpression()
		if prefix == nil {
			return false
		}
		// ++ and -- are write operations
		switch prefix.Operator {
		case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			return prefix.Operand == node
		}

	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		// This is a destructuring pattern - check if it's part of an assignment
		return isWriteReference(parent)

	case ast.KindBindingElement:
		// Check if the binding element is part of a write context
		return isWriteReference(parent)

	case ast.KindShorthandPropertyAssignment:
		// In destructuring like {A} = obj, A is a write reference
		shorthand := parent.AsShorthandPropertyAssignment()
		if shorthand != nil && shorthand.Name() == node {
			return isInDestructuringAssignment(parent)
		}

	case ast.KindPropertyAssignment:
		// In destructuring like {b: A} = obj, A is a write reference
		propAssignment := parent.AsPropertyAssignment()
		if propAssignment != nil && propAssignment.Initializer == node {
			return isInDestructuringAssignment(parent)
		}
	}

	return false
}

// isInDestructuringAssignment checks if a node is part of a destructuring assignment pattern
func isInDestructuringAssignment(node *ast.Node) bool {
	current := node
	for current != nil {
		if current.Kind == ast.KindObjectLiteralExpression {
			// Check if this object literal is the left side of an assignment
			if current.Parent != nil && current.Parent.Kind == ast.KindBinaryExpression {
				binary := current.Parent.AsBinaryExpression()
				if binary != nil && binary.Left == current && binary.OperatorToken != nil {
					return binary.OperatorToken.Kind == ast.KindEqualsToken
				}
			}
			return false
		}
		if current.Kind == ast.KindArrayLiteralExpression {
			// Check if this array literal is the left side of an assignment
			if current.Parent != nil && current.Parent.Kind == ast.KindBinaryExpression {
				binary := current.Parent.AsBinaryExpression()
				if binary != nil && binary.Left == current && binary.OperatorToken != nil {
					return binary.OperatorToken.Kind == ast.KindEqualsToken
				}
			}
			return false
		}
		current = current.Parent
	}
	return false
}

// isNameShadowed checks if an identifier references a different variable due to shadowing
func isNameShadowed(node *ast.Node, className string, classNode *ast.Node, ctx *rule.RuleContext) bool {
	if node == nil || ctx.TypeChecker == nil {
		return false
	}

	// Get the symbol at the identifier location
	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return false
	}

	// Get the symbol of the class declaration
	var classSymbol *checker.Symbol
	if classNode.Kind == ast.KindClassDeclaration {
		classDecl := classNode.AsClassDeclaration()
		if classDecl != nil && classDecl.Name() != nil {
			classSymbol = ctx.TypeChecker.GetSymbolAtLocation(classDecl.Name())
		}
	} else if classNode.Kind == ast.KindClassExpression {
		classExpr := classNode.AsClassExpression()
		if classExpr != nil && classExpr.Name() != nil {
			classSymbol = ctx.TypeChecker.GetSymbolAtLocation(classExpr.Name())
		}
	}

	// If symbols are different, the name is shadowed
	if classSymbol != nil {
		return symbol != classSymbol
	}

	// Fallback: check if the identifier is within a scope that shadows the class name
	return isInShadowingScope(node, className, classNode)
}

// isInShadowingScope checks if a node is within a scope that shadows the class name
func isInShadowingScope(node *ast.Node, className string, classNode *ast.Node) bool {
	current := node.Parent
	for current != nil && current != classNode {
		switch current.Kind {
		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindConstructor,
			ast.KindGetAccessor,
			ast.KindSetAccessor:
			// Check if there's a parameter with the same name
			if hasShadowingParameter(current, className) {
				return true
			}

		case ast.KindBlock:
			// Check if there's a variable declaration with the same name
			if hasShadowingVariable(current, className) {
				return true
			}

		case ast.KindCatchClause:
			// Check if the catch variable has the same name
			catchClause := current.AsCatchClause()
			if catchClause != nil && catchClause.VariableDeclaration != nil {
				varDecl := catchClause.VariableDeclaration.AsVariableDeclaration()
				if varDecl != nil && varDecl.Name() != nil {
					if getIdentifierName(varDecl.Name()) == className {
						return true
					}
				}
			}
		}
		current = current.Parent
	}
	return false
}

// hasShadowingParameter checks if a function has a parameter with the given name
func hasShadowingParameter(node *ast.Node, name string) bool {
	var params []*ast.ParameterDeclaration

	switch node.Kind {
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl != nil {
			params = funcDecl.Parameters
		}
	case ast.KindFunctionExpression:
		funcExpr := node.AsFunctionExpression()
		if funcExpr != nil {
			params = funcExpr.Parameters
		}
	case ast.KindArrowFunction:
		arrowFunc := node.AsArrowFunction()
		if arrowFunc != nil {
			params = arrowFunc.Parameters
		}
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		if method != nil {
			params = method.Parameters
		}
	case ast.KindConstructor:
		constructor := node.AsConstructorDeclaration()
		if constructor != nil {
			params = constructor.Parameters
		}
	case ast.KindGetAccessor:
		getter := node.AsGetAccessorDeclaration()
		if getter != nil {
			params = getter.Parameters
		}
	case ast.KindSetAccessor:
		setter := node.AsSetAccessorDeclaration()
		if setter != nil {
			params = setter.Parameters
		}
	}

	for _, param := range params {
		if param.Name() != nil && getIdentifierName(param.Name()) == name {
			return true
		}
	}

	return false
}

// hasShadowingVariable checks if a block contains a variable declaration with the given name
func hasShadowingVariable(node *ast.Node, name string) bool {
	if node.Kind != ast.KindBlock {
		return false
	}

	block := node.AsBlock()
	if block == nil {
		return false
	}

	for _, stmt := range block.Statements {
		if stmt.Kind == ast.KindVariableStatement {
			varStmt := stmt.AsVariableStatement()
			if varStmt != nil && varStmt.DeclarationList != nil {
				declList := varStmt.DeclarationList.AsVariableDeclarationList()
				if declList != nil {
					for _, decl := range declList.Declarations {
						varDecl := decl.AsVariableDeclaration()
						if varDecl != nil && varDecl.Name() != nil {
							if getIdentifierName(varDecl.Name()) == name {
								return true
							}
						}
					}
				}
			}
		}
	}

	return false
}

// checkClassReassignments finds all reassignments to the class name
func checkClassReassignments(classNode *ast.Node, className string, ctx *rule.RuleContext) {
	if className == "" {
		return
	}

	// Walk the tree to find all identifiers with the class name
	ast.ForEachChild(classNode, func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier && getIdentifierName(node) == className {
			// Skip if this is the class name declaration itself
			if node.Parent == classNode {
				return
			}

			// Check if this is a write reference
			if isWriteReference(node) {
				// Check if the name is shadowed by a local variable
				if !isNameShadowed(node, className, classNode, ctx) {
					ctx.ReportNode(node, buildClassReassignmentMessage(className))
				}
			}
		}

		// Recursively check children
		ast.ForEachChild(node, func(child *ast.Node) {})
	})
}

// NoClassAssignRule disallows reassigning class declarations
var NoClassAssignRule = rule.CreateRule(rule.Rule{
	Name: "no-class-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			// Check class declarations
			ast.KindClassDeclaration: func(node *ast.Node) {
				classDecl := node.AsClassDeclaration()
				if classDecl == nil || classDecl.Name() == nil {
					return
				}

				className := getIdentifierName(classDecl.Name())
				checkClassReassignments(node, className, &ctx)
			},

			// Check named class expressions
			ast.KindClassExpression: func(node *ast.Node) {
				classExpr := node.AsClassExpression()
				if classExpr == nil || classExpr.Name() == nil {
					return
				}

				// Only check named class expressions
				// For `let A = class A { ... }`, we need to check reassignments inside the class
				className := getIdentifierName(classExpr.Name())
				checkClassReassignments(node, className, &ctx)
			},
		}
	},
})

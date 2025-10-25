package no_class_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildClassMessage(className string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "class",
		Description: "'" + className + "' is a class.",
	}
}

// NoClassAssignRule disallows reassigning class members
var NoClassAssignRule = rule.Rule{
	Name: "no-class-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		listeners := rule.RuleListeners{}

		// Track class declarations by name
		classNames := make(map[string]*ast.Node)

		// Helper function to check if we're in a class method or nested scope
		isInClassMethodScope := func(node *ast.Node, className string) bool {
			if node == nil {
				return false
			}

			// Walk up the tree to see if we're inside a method that shadows the class name
			parent := ast.FromNode(ctx.SourceFile.GetParent(node.InternalNode))
			for parent != nil {
				kind := parent.GetKind()

				// Check if we're in a function/method parameter that shadows the class
				if kind == ast.KindParameter {
					param := parent.AsParameter()
					if param != nil && param.Name != nil {
						paramName := ast.FromNode(param.Name)
						if paramName.GetKind() == ast.KindIdentifier {
							ident := paramName.AsIdentifier()
							if ident != nil && ident.EscapedText() == className {
								return true
							}
						}
					}
				}

				// Check if we're in a variable declaration that shadows the class
				if kind == ast.KindVariableDeclaration {
					varDecl := parent.AsVariableDeclaration()
					if varDecl != nil && varDecl.Name != nil {
						varName := ast.FromNode(varDecl.Name)
						if varName.GetKind() == ast.KindIdentifier {
							ident := varName.AsIdentifier()
							if ident != nil && ident.EscapedText() == className {
								return true
							}
						}
					}
				}

				parent = ast.FromNode(ctx.SourceFile.GetParent(parent.InternalNode))
			}

			return false
		}

		// Listen to ClassDeclaration nodes
		listeners[ast.KindClassDeclaration] = func(node *ast.Node) {
			classDecl := node.AsClassDeclaration()
			if classDecl == nil || classDecl.Name == nil {
				return
			}

			className := classDecl.Name.EscapedText()
			classNames[className] = node
		}

		// Listen to ClassExpression nodes (named class expressions)
		listeners[ast.KindClassExpression] = func(node *ast.Node) {
			classExpr := node.AsClassExpression()
			if classExpr == nil || classExpr.Name == nil {
				return
			}

			className := classExpr.Name.EscapedText()
			// Named class expressions create immutable bindings
			classNames[className] = node
		}

		// Listen to BinaryExpression for assignments
		listeners[ast.KindBinaryExpression] = func(node *ast.Node) {
			binary := node.AsBinaryExpression()
			if binary == nil {
				return
			}

			// Check for assignment operators
			op := binary.OperatorToken.GetKind()
			isAssignment := op == ast.SyntaxKindEqualsToken ||
				op == ast.SyntaxKindPlusEqualsToken ||
				op == ast.SyntaxKindMinusEqualsToken ||
				op == ast.SyntaxKindAsteriskEqualsToken ||
				op == ast.SyntaxKindSlashEqualsToken ||
				op == ast.SyntaxKindPercentEqualsToken ||
				op == ast.SyntaxKindAmpersandEqualsToken ||
				op == ast.SyntaxKindBarEqualsToken ||
				op == ast.SyntaxKindCaretEqualsToken ||
				op == ast.SyntaxKindLessThanLessThanEqualsToken ||
				op == ast.SyntaxKindGreaterThanGreaterThanEqualsToken ||
				op == ast.SyntaxKindGreaterThanGreaterThanGreaterThanEqualsToken ||
				op == ast.SyntaxKindAsteriskAsteriskEqualsToken

			if !isAssignment {
				return
			}

			// Check if left side is an identifier
			left := ast.FromNode(binary.Left)
			if left.GetKind() != ast.KindIdentifier {
				return
			}

			ident := left.AsIdentifier()
			if ident == nil {
				return
			}

			identName := ident.EscapedText()

			// Check if this identifier is a class name
			if _, isClass := classNames[identName]; isClass {
				// Check if we're in a scope that shadows the class name
				if !isInClassMethodScope(node, identName) {
					ctx.ReportNode(left, buildClassMessage(identName))
				}
			}
		}

		// Listen to PostfixUnaryExpression (++, --)
		listeners[ast.KindPostfixUnaryExpression] = func(node *ast.Node) {
			postfix := node.AsPostfixUnaryExpression()
			if postfix == nil {
				return
			}

			operand := ast.FromNode(postfix.Operand)
			if operand.GetKind() != ast.KindIdentifier {
				return
			}

			ident := operand.AsIdentifier()
			if ident == nil {
				return
			}

			identName := ident.EscapedText()

			if _, isClass := classNames[identName]; isClass {
				if !isInClassMethodScope(node, identName) {
					ctx.ReportNode(operand, buildClassMessage(identName))
				}
			}
		}

		// Listen to PrefixUnaryExpression (++, --)
		listeners[ast.KindPrefixUnaryExpression] = func(node *ast.Node) {
			prefix := node.AsPrefixUnaryExpression()
			if prefix == nil {
				return
			}

			// Only care about ++ and --
			if prefix.Operator != ast.SyntaxKindPlusPlusToken &&
			   prefix.Operator != ast.SyntaxKindMinusMinusToken {
				return
			}

			operand := ast.FromNode(prefix.Operand)
			if operand.GetKind() != ast.KindIdentifier {
				return
			}

			ident := operand.AsIdentifier()
			if ident == nil {
				return
			}

			identName := ident.EscapedText()

			if _, isClass := classNames[identName]; isClass {
				if !isInClassMethodScope(node, identName) {
					ctx.ReportNode(operand, buildClassMessage(identName))
				}
			}
		}

		return listeners
	},
}

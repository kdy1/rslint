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
			parent := node.Parent
			for parent != nil {
				kind := parent.Kind

				// Check if we're in a function/method parameter that shadows the class
				if kind == ast.KindParameter {
					param := parent.AsParameterDeclaration()
					if param != nil {
						paramName := param.Name()
						if paramName != nil && paramName.Kind == ast.KindIdentifier {
							ident := paramName.AsIdentifier()
							if ident != nil && ident.Text == className {
								return true
							}
						}
					}
				}

				// Check if we're in a variable declaration that shadows the class
				if kind == ast.KindVariableDeclaration {
					varDecl := parent.AsVariableDeclaration()
					if varDecl != nil {
						varName := varDecl.Name()
						if varName != nil && varName.Kind == ast.KindIdentifier {
							ident := varName.AsIdentifier()
							if ident != nil && ident.Text == className {
								return true
							}
						}
					}
				}

				parent = parent.Parent
			}

			return false
		}

		// Listen to ClassDeclaration nodes
		listeners[ast.KindClassDeclaration] = func(node *ast.Node) {
			classDecl := node.AsClassDeclaration()
			nameNode := classDecl.Name()
			if classDecl == nil || nameNode == nil {
				return
			}

			className := nameNode.Text()
			classNames[className] = node
		}

		// Listen to ClassExpression nodes (named class expressions)
		listeners[ast.KindClassExpression] = func(node *ast.Node) {
			classExpr := node.AsClassExpression()
			nameNode := classExpr.Name()
			if classExpr == nil || nameNode == nil {
				return
			}

			className := nameNode.Text()
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
			op := binary.OperatorToken.Kind
			isAssignment := op == ast.KindEqualsToken ||
				op == ast.KindPlusEqualsToken ||
				op == ast.KindMinusEqualsToken ||
				op == ast.KindAsteriskEqualsToken ||
				op == ast.KindSlashEqualsToken ||
				op == ast.KindPercentEqualsToken ||
				op == ast.KindAmpersandEqualsToken ||
				op == ast.KindBarEqualsToken ||
				op == ast.KindCaretEqualsToken ||
				op == ast.KindLessThanLessThanEqualsToken ||
				op == ast.KindGreaterThanGreaterThanEqualsToken ||
				op == ast.KindGreaterThanGreaterThanGreaterThanEqualsToken ||
				op == ast.KindAsteriskAsteriskEqualsToken

			if !isAssignment {
				return
			}

			// Check if left side is an identifier
			left := binary.Left
			if left.Kind != ast.KindIdentifier {
				return
			}

			ident := left.AsIdentifier()
			if ident == nil {
				return
			}

			identName := ident.Text

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

			operand := postfix.Operand
			if operand.Kind != ast.KindIdentifier {
				return
			}

			ident := operand.AsIdentifier()
			if ident == nil {
				return
			}

			identName := ident.Text

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
			if prefix.Operator != ast.KindPlusPlusToken &&
			   prefix.Operator != ast.KindMinusMinusToken {
				return
			}

			operand := prefix.Operand
			if operand.Kind != ast.KindIdentifier {
				return
			}

			ident := operand.AsIdentifier()
			if ident == nil {
				return
			}

			identName := ident.Text

			if _, isClass := classNames[identName]; isClass {
				if !isInClassMethodScope(node, identName) {
					ctx.ReportNode(operand, buildClassMessage(identName))
				}
			}
		}

		return listeners
	},
}

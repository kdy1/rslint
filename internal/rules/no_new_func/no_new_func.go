package no_new_func

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildNoFunctionConstructorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noFunctionConstructor",
		Description: "The Function constructor is eval.",
	}
}

// isFunctionIdentifier checks if a node is the global Function identifier
func isFunctionIdentifier(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}

	ident := node.AsIdentifier()
	if ident == nil {
		return false
	}

	return ident.Text == "Function"
}

// isNodeInsideNode checks if nodeToCheck is a descendant of possibleAncestor
func isNodeInsideNode(nodeToCheck *ast.Node, possibleAncestor *ast.Node) bool {
	if nodeToCheck == nil || possibleAncestor == nil {
		return false
	}

	current := nodeToCheck
	for current != nil {
		if current == possibleAncestor {
			return true
		}
		current = current.Parent
	}
	return false
}

// isShadowedFunction checks if the Function identifier is shadowed by a local declaration
func isShadowedFunction(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Walk up to find any scope that declares Function
	current := node.Parent
	for current != nil {
		// Check for function expression with name Function - named function expressions
		// create a binding only within their own body
		if current.Kind == ast.KindFunctionExpression {
			fnExpr := current.AsFunctionExpression()
			if fnExpr != nil && fnExpr.Name() != nil && fnExpr.Name().Kind == ast.KindIdentifier {
				nameIdent := fnExpr.Name().AsIdentifier()
				if nameIdent != nil && nameIdent.Text == "Function" {
					// The name is only accessible within the function body
					return true
				}
			}
		}

		// Check in block scopes for declarations - simplified to only check source file scope
		if current.Kind == ast.KindSourceFile {
			sf := current.AsSourceFile()
			if sf != nil && sf.Statements != nil {
				for _, child := range sf.Statements.Nodes {
					// Check for function declarations
					if child.Kind == ast.KindFunctionDeclaration {
						fnDecl := child.AsFunctionDeclaration()
						if fnDecl != nil && fnDecl.Name() != nil && fnDecl.Name().Kind == ast.KindIdentifier {
							nameIdent := fnDecl.Name().AsIdentifier()
							if nameIdent != nil && nameIdent.Text == "Function" {
								return true
							}
						}
					}

					// Check for class declarations
					if child.Kind == ast.KindClassDeclaration {
						classDecl := child.AsClassDeclaration()
						if classDecl != nil && classDecl.Name() != nil && classDecl.Name().Kind == ast.KindIdentifier {
							nameIdent := classDecl.Name().AsIdentifier()
							if nameIdent != nil && nameIdent.Text == "Function" {
								return true
							}
						}
					}
				}
			}
		} else if current.Kind == ast.KindBlock {
			block := current.AsBlock()
			if block != nil && block.Statements != nil {
				for _, child := range block.Statements.Nodes {
					// Check for function declarations
					if child.Kind == ast.KindFunctionDeclaration {
						fnDecl := child.AsFunctionDeclaration()
						if fnDecl != nil && fnDecl.Name() != nil && fnDecl.Name().Kind == ast.KindIdentifier {
							nameIdent := fnDecl.Name().AsIdentifier()
							if nameIdent != nil && nameIdent.Text == "Function" {
								return true
							}
						}
					}

					// Check for class declarations
					if child.Kind == ast.KindClassDeclaration {
						classDecl := child.AsClassDeclaration()
						if classDecl != nil && classDecl.Name() != nil && classDecl.Name().Kind == ast.KindIdentifier {
							nameIdent := classDecl.Name().AsIdentifier()
							if nameIdent != nil && nameIdent.Text == "Function" {
								return true
							}
						}
					}
				}
			}
		}

		current = current.Parent
	}

	return false
}

// NoNewFuncRule disallows new Function(...)
var NoNewFuncRule = rule.CreateRule(rule.Rule{
	Name: "no-new-func",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		listeners := rule.RuleListeners{}

		// Check for new Function() and Function() calls
		listeners[ast.KindNewExpression] = func(node *ast.Node) {
			newExpr := node.AsNewExpression()
			if newExpr == nil || newExpr.Expression == nil {
				return
			}

			if isFunctionIdentifier(newExpr.Expression) && !isShadowedFunction(newExpr.Expression) {
				ctx.ReportNode(newExpr.Expression, buildNoFunctionConstructorMessage())
			}
		}

		listeners[ast.KindCallExpression] = func(node *ast.Node) {
			callExpr := node.AsCallExpression()
			if callExpr == nil || callExpr.Expression == nil {
				return
			}

			// Direct Function() call
			if isFunctionIdentifier(callExpr.Expression) && !isShadowedFunction(callExpr.Expression) {
				ctx.ReportNode(callExpr.Expression, buildNoFunctionConstructorMessage())
				return
			}

			// Function.call(), Function.apply(), Function.bind()
			if callExpr.Expression.Kind == ast.KindPropertyAccessExpression {
				propAccess := callExpr.Expression.AsPropertyAccessExpression()
				if propAccess != nil && propAccess.Expression != nil && propAccess.Name() != nil {
					if isFunctionIdentifier(propAccess.Expression) && !isShadowedFunction(propAccess.Expression) {
						methodName := propAccess.Name().Text()
						if methodName == "call" || methodName == "apply" || methodName == "bind" {
							ctx.ReportNode(propAccess.Expression, buildNoFunctionConstructorMessage())
						}
					}
				}
			}

			// Function?.call() with optional chaining
			if callExpr.Expression.Kind == ast.KindPropertyAccessExpression {
				propAccess := callExpr.Expression.AsPropertyAccessExpression()
				if propAccess != nil && propAccess.QuestionDotToken != nil && propAccess.Expression != nil {
					if isFunctionIdentifier(propAccess.Expression) && !isShadowedFunction(propAccess.Expression) {
						ctx.ReportNode(propAccess.Expression, buildNoFunctionConstructorMessage())
					}
				}
			}

			// Check for parenthesized expressions like (Function?.call)
			if callExpr.Expression.Kind == ast.KindParenthesizedExpression {
				parenExpr := callExpr.Expression.AsParenthesizedExpression()
				if parenExpr != nil && parenExpr.Expression != nil {
					if parenExpr.Expression.Kind == ast.KindPropertyAccessExpression {
						propAccess := parenExpr.Expression.AsPropertyAccessExpression()
						if propAccess != nil && propAccess.Expression != nil {
							if isFunctionIdentifier(propAccess.Expression) && !isShadowedFunction(propAccess.Expression) {
								methodName := ""
								if propAccess.Name() != nil {
									methodName = propAccess.Name().Text()
								}
								if methodName == "call" || methodName == "apply" || methodName == "bind" {
									ctx.ReportNode(propAccess.Expression, buildNoFunctionConstructorMessage())
								}
							}
						}
					}
				}
			}
		}

		// Check for Function["call"] and similar bracket notation
		listeners[ast.KindElementAccessExpression] = func(node *ast.Node) {
			elemAccess := node.AsElementAccessExpression()
			if elemAccess == nil || elemAccess.Expression == nil || elemAccess.ArgumentExpression == nil {
				return
			}

			// Check if object is Function identifier
			if !isFunctionIdentifier(elemAccess.Expression) {
				return
			}

			// Check if it's shadowed
			if isShadowedFunction(elemAccess.Expression) {
				return
			}

			// Check if we're accessing call, apply, or bind
			if elemAccess.ArgumentExpression.Kind == ast.KindStringLiteral {
				strLit := elemAccess.ArgumentExpression.AsStringLiteral()
				if strLit != nil {
					text := strLit.Text
					// Strip quotes
					if len(text) >= 2 && (text[0] == '"' || text[0] == '\'') {
						text = text[1 : len(text)-1]
					}
					if text == "call" || text == "apply" || text == "bind" {
						// This is Function["call"] etc., check if it's being called
						parent := node.Parent
						if parent != nil && parent.Kind == ast.KindCallExpression {
							ctx.ReportNode(elemAccess.Expression, buildNoFunctionConstructorMessage())
						}
					}
				}
			}
		}

		return listeners
	},
})

package no_ex_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoExAssignRule implements the no-ex-assign rule
// Disallow reassigning exceptions in catch clauses
var NoExAssignRule = rule.Rule{
	Name: "no-ex-assign",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindCatchClause: func(node *ast.Node) {
			catchClause := node.AsCatchClause()
			if catchClause == nil {
				return
			}

			// Get the catch parameter
			if catchClause.VariableDeclaration == nil {
				return
			}

			variableDecl := catchClause.VariableDeclaration.AsVariableDeclaration()
			if variableDecl == nil || variableDecl.Name() == nil {
				return
			}

			// Collect all parameter names (for destructured parameters)
			paramNames := collectBindingNames(variableDecl.Name())
			if len(paramNames) == 0 {
				return
			}

			// Check the catch block for assignments to the parameter
			if catchClause.Block != nil {
				checkForAssignments(ctx, catchClause.Block, paramNames)
			}
		},
	}
}

// collectBindingNames extracts all binding names from a binding pattern
func collectBindingNames(node *ast.Node) []string {
	var names []string

	switch node.Kind {
	case ast.KindIdentifier:
		if ast.IsIdentifier(node) {
			names = append(names, node.AsIdentifier().Text)
		}

	case ast.KindObjectBindingPattern:
		if ast.IsObjectBindingPattern(node) {
			bindingPattern := node.AsBindingPattern()
			if bindingPattern != nil && bindingPattern.Elements != nil {
				for _, elem := range bindingPattern.Elements.Nodes {
					bindingElem := elem.AsBindingElement()
					if bindingElem != nil && bindingElem.Name() != nil {
						names = append(names, collectBindingNames(bindingElem.Name())...)
					}
				}
			}
		}

	case ast.KindArrayBindingPattern:
		if ast.IsArrayBindingPattern(node) {
			bindingPattern := node.AsBindingPattern()
			if bindingPattern != nil && bindingPattern.Elements != nil {
				for _, elem := range bindingPattern.Elements.Nodes {
					bindingElem := elem.AsBindingElement()
					if bindingElem != nil && bindingElem.Name() != nil {
						names = append(names, collectBindingNames(bindingElem.Name())...)
					}
				}
			}
		}
	}

	return names
}

// checkForAssignments walks the AST and checks for assignments to the catch parameter
func checkForAssignments(ctx rule.RuleContext, node *ast.Node, paramNames []string) {
	if node == nil {
		return
	}

	// Check if this node is an assignment to one of the parameter names
	switch node.Kind {
	case ast.KindBinaryExpression:
		binExpr := node.AsBinaryExpression()
		if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindEqualsToken {
			// Check if left side is one of our parameter names
			if binExpr.Left != nil && binExpr.Left.Kind == ast.KindIdentifier {
				if ast.IsIdentifier(binExpr.Left) {
					leftIdent := binExpr.Left.AsIdentifier()
					if leftIdent != nil {
						for _, paramName := range paramNames {
							if leftIdent.Text == paramName {
								ctx.ReportNode(binExpr.Left, rule.RuleMessage{
									Id:          "unexpected",
									Description: "Do not assign to the exception parameter.",
								})
							}
						}
					}
				}
			}
		}

	case ast.KindArrayLiteralExpression:
		// Check for destructuring assignments like [e] = []
		arrLit := node.AsArrayLiteralExpression()
		if arrLit != nil && node.Parent != nil && node.Parent.Kind == ast.KindBinaryExpression {
			binExpr := node.Parent.AsBinaryExpression()
			if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindEqualsToken && binExpr.Left == node {
				// This is a destructuring assignment
				if arrLit.Elements != nil {
					for _, elem := range arrLit.Elements.Nodes {
						if elem.Kind == ast.KindIdentifier && ast.IsIdentifier(elem) {
							elemIdent := elem.AsIdentifier()
							if elemIdent != nil {
								for _, paramName := range paramNames {
									if elemIdent.Text == paramName {
										ctx.ReportNode(elem, rule.RuleMessage{
											Id:          "unexpected",
											Description: "Do not assign to the exception parameter.",
										})
									}
								}
							}
						}
					}
				}
			}
		}

	case ast.KindObjectLiteralExpression:
		// Check for destructuring assignments like {x: e = 0} = {}
		objLit := node.AsObjectLiteralExpression()
		if objLit != nil && node.Parent != nil && node.Parent.Kind == ast.KindBinaryExpression {
			binExpr := node.Parent.AsBinaryExpression()
			if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindEqualsToken && binExpr.Left == node {
				// This is a destructuring assignment
				checkObjectLiteralForParamNames(ctx, objLit, paramNames)
				return // Don't recursively check children as we've already processed this
			}
		}
	}

	// Recursively check children
	node.ForEachChild(func(child *ast.Node) bool {
		checkForAssignments(ctx, child, paramNames)
		return false
	})
}

// checkObjectLiteralForParamNames checks object literal for parameter names in destructuring
func checkObjectLiteralForParamNames(ctx rule.RuleContext, objLit *ast.ObjectLiteralExpression, paramNames []string) {
	if objLit.Properties == nil {
		return
	}

	for _, prop := range objLit.Properties.Nodes {
		if prop.Kind == ast.KindPropertyAssignment {
			propAssign := prop.AsPropertyAssignment()
			if propAssign != nil && propAssign.Initializer != nil {
				// Check if initializer contains parameter names
				checkInitializerForParamNames(ctx, propAssign.Initializer, paramNames)
			}
		} else if prop.Kind == ast.KindShorthandPropertyAssignment {
			shorthand := prop.AsShorthandPropertyAssignment()
			if shorthand != nil && shorthand.Name() != nil && shorthand.Name().Kind == ast.KindIdentifier {
				if ast.IsIdentifier(shorthand.Name()) {
					nameIdent := shorthand.Name().AsIdentifier()
					if nameIdent != nil {
						for _, paramName := range paramNames {
							if nameIdent.Text == paramName {
								ctx.ReportNode(shorthand.Name(), rule.RuleMessage{
									Id:          "unexpected",
									Description: "Do not assign to the exception parameter.",
								})
							}
						}
					}
				}
			}
		}
	}
}

// checkInitializerForParamNames checks if an initializer contains parameter names
func checkInitializerForParamNames(ctx rule.RuleContext, node *ast.Node, paramNames []string) {
	if node == nil {
		return
	}

	if node.Kind == ast.KindIdentifier && ast.IsIdentifier(node) {
		ident := node.AsIdentifier()
		if ident != nil {
			for _, paramName := range paramNames {
				if ident.Text == paramName {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpected",
						Description: "Do not assign to the exception parameter.",
					})
				}
			}
		}
	} else if node.Kind == ast.KindBinaryExpression {
		// Handle cases like {x: e = 0}
		binExpr := node.AsBinaryExpression()
		if binExpr != nil && binExpr.Left != nil {
			checkInitializerForParamNames(ctx, binExpr.Left, paramNames)
		}
	}
}

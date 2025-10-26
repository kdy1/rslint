package prefer_destructuring

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// PreferDestructuringOptions defines the configuration options for this rule
type PreferDestructuringOptions struct {
	EnforceForDeclarationWithTypeAnnotation bool
	// Simplified: we'll use basic object/array boolean flags
	VariableDeclaratorObject bool
	VariableDeclaratorArray  bool
	AssignmentExpressionObject bool
	AssignmentExpressionArray  bool
}

// parseOptions parses and validates the rule options
func parseOptions(options any) PreferDestructuringOptions {
	opts := PreferDestructuringOptions{
		EnforceForDeclarationWithTypeAnnotation: false,
		VariableDeclaratorObject:               true,
		VariableDeclaratorArray:                true,
		AssignmentExpressionObject:             false,
		AssignmentExpressionArray:              false,
	}

	if options == nil {
		return opts
	}

	// Handle both array format and object format
	var optsMap map[string]interface{}
	var secondOpt map[string]interface{}

	if optArray, isArray := options.([]interface{}); isArray {
		if len(optArray) > 0 {
			optsMap, _ = optArray[0].(map[string]interface{})
		}
		if len(optArray) > 1 {
			secondOpt, _ = optArray[1].(map[string]interface{})
		}
	} else {
		optsMap, _ = options.(map[string]interface{})
	}

	if optsMap != nil {
		// Check for object/array pattern
		if obj, ok := optsMap["object"].(bool); ok {
			opts.VariableDeclaratorObject = obj
			opts.AssignmentExpressionObject = obj
		}
		if arr, ok := optsMap["array"].(bool); ok {
			opts.VariableDeclaratorArray = arr
			opts.AssignmentExpressionArray = arr
		}

		// Check for specific patterns
		if vd, ok := optsMap["VariableDeclarator"].(map[string]interface{}); ok {
			if obj, ok := vd["object"].(bool); ok {
				opts.VariableDeclaratorObject = obj
			}
			if arr, ok := vd["array"].(bool); ok {
				opts.VariableDeclaratorArray = arr
			}
		}

		if ae, ok := optsMap["AssignmentExpression"].(map[string]interface{}); ok {
			if obj, ok := ae["object"].(bool); ok {
				opts.AssignmentExpressionObject = obj
			}
			if arr, ok := ae["array"].(bool); ok {
				opts.AssignmentExpressionArray = arr
			}
		}
	}

	if secondOpt != nil {
		if v, ok := secondOpt["enforceForDeclarationWithTypeAnnotation"].(bool); ok {
			opts.EnforceForDeclarationWithTypeAnnotation = v
		}
	}

	return opts
}

// PreferDestructuringRule implements the prefer-destructuring rule
// Whether to enforce destructuring on variable declarations with type annotations.
var PreferDestructuringRule = rule.CreateRule(rule.Rule{
	Name: "prefer-destructuring",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	return rule.RuleListeners{
		ast.KindVariableStatement: func(node *ast.Node) {
			varStmt := node.AsVariableStatement()
			if varStmt == nil || varStmt.DeclarationList == nil {
				return
			}

			declList := varStmt.DeclarationList.AsVariableDeclarationList()
			if declList == nil {
				return
			}

			for _, decl := range declList.Declarations.Nodes {
				if decl.Kind != ast.KindVariableDeclaration {
					continue
				}

				varDeclNode := decl.AsVariableDeclaration()
				if varDeclNode == nil || varDeclNode.Initializer == nil {
					continue
				}

				// Skip if has type annotation and option is false
				if varDeclNode.Type != nil && !opts.EnforceForDeclarationWithTypeAnnotation {
					continue
				}

				// Check if initializer is a member expression
				if varDeclNode.Initializer.Kind != ast.KindPropertyAccessExpression &&
					varDeclNode.Initializer.Kind != ast.KindElementAccessExpression {
					continue
				}

				// Check for property access (object.property)
				if varDeclNode.Initializer.Kind == ast.KindPropertyAccessExpression {
					if !opts.VariableDeclaratorObject {
						continue
					}

					propAccess := varDeclNode.Initializer.AsPropertyAccessExpression()
					if propAccess == nil {
						continue
					}

					// Get variable name
					varName := getVariableName(varDeclNode)
					if varName == "" {
						continue
					}

					// Get property name
					propName := ""
					if propAccess.Name() != nil && propAccess.Name().Kind == ast.KindIdentifier {
						propName = propAccess.Name().AsIdentifier().Text
					}

					// Only report if names match (avoid renaming cases)
					if varName == propName {
						reportObjectDestructuring(ctx, decl, varName)
					}
				}

				// Check for element access (array[0])
				if varDeclNode.Initializer.Kind == ast.KindElementAccessExpression {
					if !opts.VariableDeclaratorArray {
						continue
					}

					elemAccess := varDeclNode.Initializer.AsElementAccessExpression()
					if elemAccess == nil || elemAccess.ArgumentExpression == nil {
						continue
					}

					// Check if accessing with numeric literal
					if elemAccess.ArgumentExpression.Kind == ast.KindNumericLiteral {
						numLit := elemAccess.ArgumentExpression.AsNumericLiteral()
						if numLit != nil {
							reportArrayDestructuring(ctx, decl)
						}
					}
				}
			}
		},

		ast.KindBinaryExpression: func(node *ast.Node) {
			binExpr := node.AsBinaryExpression()
			if binExpr == nil || binExpr.OperatorToken.Kind != ast.KindEqualsToken {
				return
			}

			// Skip if right side is not a member expression
			if binExpr.Right == nil {
				return
			}

			// Check for property access (object.property)
			if binExpr.Right.Kind == ast.KindPropertyAccessExpression {
				if !opts.AssignmentExpressionObject {
					return
				}

				propAccess := binExpr.Right.AsPropertyAccessExpression()
				if propAccess == nil {
					return
				}

				// Get left side name
				leftName := ""
				if binExpr.Left != nil && binExpr.Left.Kind == ast.KindIdentifier {
					leftName = binExpr.Left.AsIdentifier().Text
				}

				// Get property name
				propName := ""
				if propAccess.Name() != nil && propAccess.Name().Kind == ast.KindIdentifier {
					propName = propAccess.Name().AsIdentifier().Text
				}

				// Only report if names match
				if leftName != "" && leftName == propName {
					reportObjectDestructuring(ctx, node, leftName)
				}
			}

			// Check for element access (array[0])
			if binExpr.Right.Kind == ast.KindElementAccessExpression {
				if !opts.AssignmentExpressionArray {
					return
				}

				elemAccess := binExpr.Right.AsElementAccessExpression()
				if elemAccess == nil || elemAccess.ArgumentExpression == nil {
					return
				}

				// Check if accessing with numeric literal
				if elemAccess.ArgumentExpression.Kind == ast.KindNumericLiteral {
					reportArrayDestructuring(ctx, node)
				}
			}
		},
	}
}

func getVariableName(varDecl *ast.VariableDeclaration) string {
	if varDecl.Name() == nil {
		return ""
	}

	if varDecl.Name().Kind == ast.KindIdentifier {
		ident := varDecl.Name().AsIdentifier()
		if ident != nil {
			return ident.Text
		}
	}

	return ""
}

func reportObjectDestructuring(ctx rule.RuleContext, node *ast.Node, propName string) {
	_ = propName // Use propName if needed for better messages
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "preferDestructuring",
		Description: "Use object destructuring.",
	})
}

func reportArrayDestructuring(ctx rule.RuleContext, node *ast.Node) {
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "preferDestructuring",
		Description: "Use array destructuring.",
	})
}

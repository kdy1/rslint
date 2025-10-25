package no_eval

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// Options mirrors ESLint no-eval options
type Options struct {
	AllowIndirect bool `json:"allowIndirect"`
}

func parseOptions(options any) Options {
	opts := Options{
		AllowIndirect: false,
	}

	if options == nil {
		return opts
	}

	// Parse options with dual-format support (handles both array and object formats)
	var optsMap map[string]interface{}
	var ok bool

	// Handle array format: [{ option: value }]
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, ok = optArray[0].(map[string]interface{})
	} else {
		// Handle direct object format: { option: value }
		optsMap, ok = options.(map[string]interface{})
	}

	if ok {
		if v, ok := optsMap["allowIndirect"].(bool); ok {
			opts.AllowIndirect = v
		}
	}
	return opts
}

func buildUnexpectedEvalMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "eval can be harmful.",
	}
}

// NoEvalRule disallows the use of eval()
var NoEvalRule = rule.CreateRule(rule.Rule{
	Name: "no-eval",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		listeners := rule.RuleListeners{}

		// Check for direct eval() calls
		listeners[ast.KindCallExpression] = func(node *ast.Node) {
			callExpr := node.AsCallExpression()
			if callExpr == nil || callExpr.Expression == nil {
				return
			}

			// Check if this is a direct eval call
			if callExpr.Expression.Kind == ast.KindIdentifier {
				ident := callExpr.Expression.AsIdentifier()
				if ident != nil && ident.Text == "eval" {
					// Direct eval() call
					ctx.ReportNode(callExpr.Expression, buildUnexpectedEvalMessage())
					return
				}
			}

			// Check for property access eval (e.g., window.eval, this.eval, globalThis.eval)
			if callExpr.Expression.Kind == ast.KindPropertyAccessExpression {
				propAccess := callExpr.Expression.AsPropertyAccessExpression()
				if propAccess != nil && propAccess.Name() != nil && propAccess.Name().Text() == "eval" {
					// Always allow this.eval() - in methods it's likely a custom method
					// At global scope in non-strict mode it would be eval, but we can't reliably detect this
					if propAccess.Expression != nil && propAccess.Expression.Kind == ast.KindThisKeyword {
						return
					}

					// Check if allowIndirect is enabled for other property accesses
					if opts.AllowIndirect {
						// Allow indirect eval from window, global, globalThis
						if propAccess.Expression != nil && propAccess.Expression.Kind == ast.KindIdentifier {
							objIdent := propAccess.Expression.AsIdentifier()
							if objIdent != nil {
								objName := objIdent.Text
								// Allow window.eval, global.eval, globalThis.eval
								if objName == "window" || objName == "global" || objName == "globalThis" {
									return
								}
							}
						}
					}
					// Report property access eval
					ctx.ReportNode(propAccess.Name(), buildUnexpectedEvalMessage())
					return
				}
			}

			// Check for optional chaining eval (e.g., window?.eval)
			if callExpr.Expression.Kind == ast.KindPropertyAccessExpression {
				propAccess := callExpr.Expression.AsPropertyAccessExpression()
				if propAccess != nil && propAccess.QuestionDotToken != nil && propAccess.Name() != nil && propAccess.Name().Text() == "eval" {
					ctx.ReportNode(propAccess.Name(), buildUnexpectedEvalMessage())
					return
				}
			}

			// Check for sequence expression eval: (0, eval)('code')
			if !opts.AllowIndirect && callExpr.Expression.Kind == ast.KindParenthesizedExpression {
				parenExpr := callExpr.Expression.AsParenthesizedExpression()
				if parenExpr != nil && parenExpr.Expression != nil {
					if parenExpr.Expression.Kind == ast.KindBinaryExpression {
						binExpr := parenExpr.Expression.AsBinaryExpression()
						if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindCommaToken {
							// Check if right side is eval identifier
							if binExpr.Right != nil && binExpr.Right.Kind == ast.KindIdentifier {
								rightIdent := binExpr.Right.AsIdentifier()
								if rightIdent != nil && rightIdent.Text == "eval" {
									ctx.ReportNode(binExpr.Right, buildUnexpectedEvalMessage())
									return
								}
							}
						}
					}
				}
			}
		}

		// Check for identifier references to eval (like var x = eval; x('code'))
		listeners[ast.KindIdentifier] = func(node *ast.Node) {
			ident := node.AsIdentifier()
			if ident == nil || ident.Text != "eval" {
				return
			}

			// Skip if this identifier is part of a call expression we already handled
			parent := node.Parent
			if parent != nil && parent.Kind == ast.KindCallExpression {
				callExpr := parent.AsCallExpression()
				if callExpr != nil && callExpr.Expression == node {
					// Already handled in CallExpression listener
					return
				}
			}

			// Skip if this is part of a property access expression
			if parent != nil && parent.Kind == ast.KindPropertyAccessExpression {
				propAccess := parent.AsPropertyAccessExpression()
				if propAccess != nil && propAccess.Name() == node {
					// Already handled in CallExpression listener
					return
				}
			}

			// Check if this is an assignment or variable initialization with eval
			// var x = eval; or EVAL = eval;
			if !opts.AllowIndirect {
				// Walk up to see if this is on the right side of an assignment
				if parent != nil {
					switch parent.Kind {
					case ast.KindVariableDeclaration:
						varDecl := parent.AsVariableDeclaration()
						if varDecl != nil && varDecl.Initializer == node {
							ctx.ReportNode(node, buildUnexpectedEvalMessage())
						}
					case ast.KindBinaryExpression:
						binExpr := parent.AsBinaryExpression()
						if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindEqualsToken && binExpr.Right == node {
							ctx.ReportNode(node, buildUnexpectedEvalMessage())
						}
					case ast.KindPropertyAccessExpression:
						// Check if this is window.eval, global.eval, globalThis.eval reference
						propAccess := parent.AsPropertyAccessExpression()
						if propAccess != nil && propAccess.Name() == node {
							// Check the object
							if propAccess.Expression != nil && propAccess.Expression.Kind == ast.KindIdentifier {
								objIdent := propAccess.Expression.AsIdentifier()
								if objIdent != nil {
									objName := objIdent.Text
									if objName == "window" || objName == "global" || objName == "globalThis" {
										// Check if this property access is being used in an assignment
										grandParent := parent.Parent
										if grandParent != nil {
											switch grandParent.Kind {
											case ast.KindVariableDeclaration:
												varDecl := grandParent.AsVariableDeclaration()
												if varDecl != nil && varDecl.Initializer == parent {
													ctx.ReportNode(node, buildUnexpectedEvalMessage())
												}
											case ast.KindBinaryExpression:
												binExpr := grandParent.AsBinaryExpression()
												if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindEqualsToken && binExpr.Right == parent {
													ctx.ReportNode(node, buildUnexpectedEvalMessage())
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}

		return listeners
	},
})

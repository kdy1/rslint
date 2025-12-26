package no_unused_expressions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoUnusedExpressionsOptions struct {
	AllowShortCircuit      bool `json:"allowShortCircuit"`
	AllowTernary           bool `json:"allowTernary"`
	AllowTaggedTemplates   bool `json:"allowTaggedTemplates"`
	EnforceForJSX          bool `json:"enforceForJSX"`
	IgnoreDirectives       bool `json:"ignoreDirectives"`
}

var NoUnusedExpressionsRule = rule.CreateRule(rule.Rule{
	Name: "no-unused-expressions",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoUnusedExpressionsOptions{
			AllowShortCircuit:    false,
			AllowTernary:         false,
			AllowTaggedTemplates: false,
			EnforceForJSX:        false,
			IgnoreDirectives:     false,
		}

		// Parse options
		if options != nil {
			var optsMap map[string]interface{}
			if optsArray, ok := options.([]interface{}); ok && len(optsArray) > 0 {
				if m, ok := optsArray[0].(map[string]interface{}); ok {
					optsMap = m
				}
			} else if m, ok := options.(map[string]interface{}); ok {
				optsMap = m
			}

			if optsMap != nil {
				if v, ok := optsMap["allowShortCircuit"].(bool); ok {
					opts.AllowShortCircuit = v
				}
				if v, ok := optsMap["allowTernary"].(bool); ok {
					opts.AllowTernary = v
				}
				if v, ok := optsMap["allowTaggedTemplates"].(bool); ok {
					opts.AllowTaggedTemplates = v
				}
				if v, ok := optsMap["enforceForJSX"].(bool); ok {
					opts.EnforceForJSX = v
				}
				if v, ok := optsMap["ignoreDirectives"].(bool); ok {
					opts.IgnoreDirectives = v
				}
			}
		}

		// Check if an expression is a directive (e.g., "use strict")
		isDirective := func(node *ast.Node, parent *ast.Node) bool {
			if node.Kind != ast.KindExpressionStatement {
				return false
			}

			exprStmt := node.AsExpressionStatement()
			if exprStmt == nil || exprStmt.Expression == nil {
				return false
			}

			expr := exprStmt.Expression

			// Must be a string literal (or potentially wrapped in parentheses)
			unwrapped := ast.SkipParentheses(expr)
			if unwrapped.Kind != ast.KindStringLiteral && unwrapped.Kind != ast.KindNoSubstitutionTemplateLiteral {
				return false
			}

			// Check if parent is a body that allows directives
			if parent == nil {
				return false
			}

			var body []*ast.Node
			switch parent.Kind {
			case ast.KindSourceFile:
				sf := parent.AsSourceFile()
				if sf != nil && sf.Statements != nil {
					body = sf.Statements.Nodes
				}
			case ast.KindBlock:
				block := parent.AsBlock()
				if block != nil && block.Statements != nil {
					body = block.Statements.Nodes
				}
			case ast.KindModuleBlock:
				mb := parent.AsModuleBlock()
				if mb != nil && mb.Statements != nil {
					body = mb.Statements.Nodes
				}
			default:
				return false
			}

			// Directives must be at the beginning (before any non-directive statements)
			for _, stmt := range body {
				if stmt == node {
					return true
				}
				// If we hit a non-directive statement before finding our node, it's not a directive
				if stmt.Kind != ast.KindExpressionStatement {
					return false
				}
				exprStmt := stmt.AsExpressionStatement()
				if exprStmt != nil && exprStmt.Expression != nil {
					unwrapped := ast.SkipParentheses(exprStmt.Expression)
					if unwrapped.Kind != ast.KindStringLiteral && unwrapped.Kind != ast.KindNoSubstitutionTemplateLiteral {
						return false
					}
				}
			}

			return false
		}

		// Unwrap TypeScript-specific expression wrappers
		unwrapTypescriptNode := func(node *ast.Node) *ast.Node {
			current := node
			for {
				switch current.Kind {
				case ast.KindAsExpression, ast.KindTypeAssertion:
					// Type assertion: expr as Type or <Type>expr
					if current.Kind == ast.KindAsExpression {
						asExpr := current.AsAsExpression()
						if asExpr != nil {
							current = asExpr.Expression
							continue
						}
					} else {
						typeAssertion := current.AsTypeAssertion()
						if typeAssertion != nil {
							current = typeAssertion.Expression
							continue
						}
					}
				case ast.KindNonNullExpression:
					// Non-null assertion: expr!
					nnExpr := current.AsNonNullExpression()
					if nnExpr != nil {
						current = nnExpr.Expression
						continue
					}
				case ast.KindParenthesizedExpression:
					// Parenthesized expression
					parenExpr := current.AsParenthesizedExpression()
					if parenExpr != nil {
						current = parenExpr.Expression
						continue
					}
				case ast.KindExpressionWithTypeArguments:
					// Instantiation expression without new: Foo<string>
					// This is an unused expression
					return current
				}
				break
			}
			return current
		}

		// Check if expression has side effects
		hasSideEffects := func(node *ast.Node) bool {
			unwrapped := unwrapTypescriptNode(node)

			switch unwrapped.Kind {
			// Always allowed - expressions with side effects
			case ast.KindCallExpression, ast.KindNewExpression:
				return true

			// Dynamic import has side effects
			case ast.KindImportKeyword:
				return true

			// Assignment expressions have side effects
			case ast.KindBinaryExpression:
				binExpr := unwrapped.AsBinaryExpression()
				if binExpr != nil {
					// Check for assignment operators
					switch binExpr.OperatorToken.Kind {
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
						return true
					}
				}

			// Update expressions have side effects (++, --)
			case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
				unaryExpr := unwrapped.AsPrefixUnaryExpression()
				if unaryExpr == nil {
					unaryExpr = unwrapped.AsPostfixUnaryExpression()
				}
				if unaryExpr != nil {
					switch unaryExpr.Operator {
					case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
						return true
					}
				}

			// Await expressions may have side effects
			case ast.KindAwaitExpression:
				return true

			// Yield expressions have side effects
			case ast.KindYieldExpression:
				return true

			// Delete expressions have side effects
			case ast.KindDeleteExpression:
				return true

			// Tagged template literals
			case ast.KindTaggedTemplateExpression:
				return opts.AllowTaggedTemplates

			// JSX elements
			case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
				return !opts.EnforceForJSX

			// Short-circuit evaluations (&&, ||, ??)
			case ast.KindBinaryExpression:
				if !opts.AllowShortCircuit {
					return false
				}
				binExpr := unwrapped.AsBinaryExpression()
				if binExpr != nil {
					switch binExpr.OperatorToken.Kind {
					case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
						return true
					}
				}

			// Conditional (ternary) expressions
			case ast.KindConditionalExpression:
				return opts.AllowTernary

			// Comma/sequence expressions - check last element
			case ast.KindCommaListExpression, ast.KindBinaryExpression:
				if unwrapped.Kind == ast.KindBinaryExpression {
					binExpr := unwrapped.AsBinaryExpression()
					if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindCommaToken {
						return hasSideEffects(binExpr.Right)
					}
				} else {
					commaList := unwrapped.AsCommaListExpression()
					if commaList != nil && commaList.Elements != nil && len(commaList.Elements.Nodes) > 0 {
						return hasSideEffects(commaList.Elements.Nodes[len(commaList.Elements.Nodes)-1])
					}
				}
			}

			return false
		}

		return rule.RuleListeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				if node.Kind != ast.KindExpressionStatement {
					return
				}

				exprStmt := node.AsExpressionStatement()
				if exprStmt == nil || exprStmt.Expression == nil {
					return
				}

				// Check if it's a directive
				if !opts.IgnoreDirectives && isDirective(node, node.Parent) {
					return
				}

				// Check if expression has side effects
				if hasSideEffects(exprStmt.Expression) {
					return
				}

				// Report unused expression
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unusedExpression",
					Description: "Expected an assignment or function call and instead saw an expression.",
				})
			},
		}
	},
})

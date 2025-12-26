package no_unused_expressions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Options mirrors @typescript-eslint/no-unused-expressions options
type Options struct {
	AllowShortCircuit    bool `json:"allowShortCircuit"`
	AllowTernary         bool `json:"allowTernary"`
	AllowTaggedTemplates bool `json:"allowTaggedTemplates"`
	EnforceForJSX        bool `json:"enforceForJSX"`
}

func parseOptions(options any) Options {
	opts := Options{
		AllowShortCircuit:    false,
		AllowTernary:         false,
		AllowTaggedTemplates: false,
		EnforceForJSX:        false,
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
	}
	return opts
}

func buildUnusedExpressionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unusedExpression",
		Description: "Expected an assignment or function call and instead saw an expression.",
	}
}

// isDirective checks if an expression statement is a directive (e.g., "use strict")
// Directives are string literals in expression statements
// For simplicity, we allow string literals since directives are valid
func isDirective(exprStmt *ast.ExpressionStatement) bool {
	if exprStmt == nil || exprStmt.Expression == nil {
		return false
	}

	// The expression must be a string literal
	return exprStmt.Expression.Kind == ast.KindStringLiteral
}

// isValidExpression checks if an expression has side effects or is valid in statement position
func isValidExpression(node *ast.Node, opts Options) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	// Assignment expressions are always valid
	case ast.KindBinaryExpression:
		bin := node.AsBinaryExpression()
		if bin != nil && ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
			return true
		}
		// Comma expressions - check the last expression
		if bin != nil && bin.OperatorToken.Kind == ast.KindCommaToken {
			return isValidExpression(bin.Right, opts)
		}
		// Logical expressions with allowShortCircuit
		if opts.AllowShortCircuit && bin != nil {
			if bin.OperatorToken.Kind == ast.KindAmpersandAmpersandToken ||
				bin.OperatorToken.Kind == ast.KindBarBarToken ||
				bin.OperatorToken.Kind == ast.KindQuestionQuestionToken {
				// Check if the right side has effects
				return isValidExpression(bin.Right, opts)
			}
		}
		return false

	// Function calls, new expressions, delete, void, await, yield are valid
	case ast.KindCallExpression, ast.KindNewExpression:
		return true

	case ast.KindDeleteExpression, ast.KindVoidExpression,
		ast.KindAwaitExpression, ast.KindYieldExpression:
		return true

	// Tagged templates are valid if allowed
	case ast.KindTaggedTemplateExpression:
		return opts.AllowTaggedTemplates

	// Import calls are valid
	case ast.KindImportKeyword:
		return true

	// Conditional expressions with allowTernary
	case ast.KindConditionalExpression:
		if opts.AllowTernary {
			cond := node.AsConditionalExpression()
			if cond != nil {
				// Both branches must have effects
				return isValidExpression(cond.WhenTrue, opts) &&
					isValidExpression(cond.WhenFalse, opts)
			}
		}
		return false

	// Prefix/Postfix increment/decrement are valid (modify state)
	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix != nil {
			if prefix.Operator == ast.KindPlusPlusToken || prefix.Operator == ast.KindMinusMinusToken {
				return true
			}
		}
		return false

	case ast.KindPostfixUnaryExpression:
		postfix := node.AsPostfixUnaryExpression()
		if postfix != nil {
			if postfix.Operator == ast.KindPlusPlusToken || postfix.Operator == ast.KindMinusMinusToken {
				return true
			}
		}
		return false

	// TypeScript-specific: unwrap type assertions and non-null assertions
	case ast.KindAsExpression:
		asExpr := node.AsAsExpression()
		if asExpr != nil && asExpr.Expression != nil {
			return isValidExpression(asExpr.Expression, opts)
		}
		return false

	case ast.KindTypeAssertionExpression:
		typeAssert := node.AsTypeAssertion()
		if typeAssert != nil && typeAssert.Expression != nil {
			return isValidExpression(typeAssert.Expression, opts)
		}
		return false

	case ast.KindNonNullExpression:
		nonNull := node.AsNonNullExpression()
		if nonNull != nil && nonNull.Expression != nil {
			return isValidExpression(nonNull.Expression, opts)
		}
		return false

	// TypeScript-specific: instantiation expressions (e.g., Foo<string>)
	case ast.KindExpressionWithTypeArguments:
		// Type instantiation expressions don't have runtime effects but are allowed for type testing
		return false

	// Parenthesized expressions - unwrap
	case ast.KindParenthesizedExpression:
		paren := node.AsParenthesizedExpression()
		if paren != nil && paren.Expression != nil {
			return isValidExpression(paren.Expression, opts)
		}
		return false

	// JSX elements
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
		return !opts.EnforceForJSX

	// Sequence expressions - check the last expression
	case ast.KindCommaListExpression:
		// This is for comma-separated expressions
		return false

	default:
		return false
	}
}

var NoUnusedExpressionsRule = rule.CreateRule(rule.Rule{
	Name: "no-unused-expressions",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				exprStmt := node.AsExpressionStatement()
				if exprStmt == nil || exprStmt.Expression == nil {
					return
				}

				// Check if this is a directive - directives are allowed
				if isDirective(exprStmt) {
					return
				}

				expr := exprStmt.Expression

				// Check if the expression is valid
				if !isValidExpression(expr, opts) {
					ctx.ReportNode(node, buildUnusedExpressionMessage())
				}
			},
		}
	},
})

package no_async_promise_executor

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoAsyncPromiseExecutorRule implements the no-async-promise-executor rule
// Disallow using an async function as a Promise executor
var NoAsyncPromiseExecutorRule = rule.Rule{
	Name: "no-async-promise-executor",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {

	return rule.RuleListeners{
		ast.KindNewExpression: func(node *ast.Node) {
			if node == nil {
				return
			}

			newExpr := node.AsNewExpression()
			if newExpr == nil {
				return
			}

			// Check if the callee is "Promise"
			if newExpr.Expression == nil {
				return
			}

			callee := newExpr.Expression
			var calleeName string

			// Handle direct identifier (Promise)
			if callee.Kind == ast.KindIdentifier {
				ident := callee.AsIdentifier()
				if ident != nil && ident.EscapedText != nil {
					calleeName = *ident.EscapedText
				}
			}

			// Only check if the constructor is "Promise"
			if calleeName != "Promise" {
				return
			}

			// Check if there are arguments and the first one is async
			if newExpr.Arguments == nil || len(newExpr.Arguments.Elements) == 0 {
				return
			}

			firstArg := newExpr.Arguments.Elements[0]
			if firstArg == nil {
				return
			}

			// Check if the first argument is an async function
			isAsync := false
			var asyncNode *ast.Node = firstArg

			switch firstArg.Kind {
			case ast.KindArrowFunction:
				arrowFunc := firstArg.AsArrowFunction()
				if arrowFunc != nil && arrowFunc.Modifiers != nil {
					for _, mod := range arrowFunc.Modifiers.Elements {
						if mod != nil && mod.Kind == ast.KindAsyncKeyword {
							isAsync = true
							asyncNode = mod
							break
						}
					}
				}
			case ast.KindFunctionExpression:
				funcExpr := firstArg.AsFunctionExpression()
				if funcExpr != nil && funcExpr.Modifiers != nil {
					for _, mod := range funcExpr.Modifiers.Elements {
						if mod != nil && mod.Kind == ast.KindAsyncKeyword {
							isAsync = true
							asyncNode = mod
							break
						}
					}
				}
			case ast.KindParenthesizedExpression:
				// Unwrap parenthesized expressions
				current := firstArg
				for current != nil && current.Kind == ast.KindParenthesizedExpression {
					parenExpr := current.AsParenthesizedExpression()
					if parenExpr == nil || parenExpr.Expression == nil {
						break
					}
					current = parenExpr.Expression

					// Check if the unwrapped expression is an async function
					if current.Kind == ast.KindArrowFunction {
						arrowFunc := current.AsArrowFunction()
						if arrowFunc != nil && arrowFunc.Modifiers != nil {
							for _, mod := range arrowFunc.Modifiers.Elements {
								if mod != nil && mod.Kind == ast.KindAsyncKeyword {
									isAsync = true
									asyncNode = mod
									break
								}
							}
						}
					} else if current.Kind == ast.KindFunctionExpression {
						funcExpr := current.AsFunctionExpression()
						if funcExpr != nil && funcExpr.Modifiers != nil {
							for _, mod := range funcExpr.Modifiers.Elements {
								if mod != nil && mod.Kind == ast.KindAsyncKeyword {
									isAsync = true
									asyncNode = mod
									break
								}
							}
						}
					}
				}
			}

			// Report if async
			if isAsync {
				ctx.ReportNode(asyncNode, rule.RuleMessage{
					Id:          "async",
					Description: "Promise executor functions should not be async.",
				})
			}
		},
	}
}

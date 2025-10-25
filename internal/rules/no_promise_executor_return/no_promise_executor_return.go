package no_promise_executor_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoPromiseExecutorReturnOptions defines the configuration options for this rule
type NoPromiseExecutorReturnOptions struct {
	AllowVoid bool `json:"allowVoid"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) NoPromiseExecutorReturnOptions {
	opts := NoPromiseExecutorReturnOptions{
		AllowVoid: false, // Default to false
	}

	if options == nil {
		return opts
	}

	// Handle both array format [{ option: value }] and object format { option: value }
	var optsMap map[string]interface{}
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, _ = optArray[0].(map[string]interface{})
	} else {
		optsMap, _ = options.(map[string]interface{})
	}

	if optsMap != nil {
		if v, ok := optsMap["allowVoid"].(bool); ok {
			opts.AllowVoid = v
		}
	}

	return opts
}

// NoPromiseExecutorReturnRule implements the no-promise-executor-return rule
// Disallow returning values from Promise executors
var NoPromiseExecutorReturnRule = rule.Rule{
	Name: "no-promise-executor-return",
	Run:  run,
}

// isPromiseConstructor checks if the expression is the Promise constructor
func isPromiseConstructor(expr *ast.Node) bool {
	if expr == nil {
		return false
	}
	if expr.Kind == ast.KindIdentifier {
		if ident := expr.AsIdentifier(); ident != nil {
			return ident.Text() == "Promise"
		}
	}
	return false
}

// isVoidExpression checks if an expression is a void expression
func isVoidExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindVoidExpression {
		return true
	}
	return false
}

// checkReturnStatement checks if a return statement in the executor is problematic
func checkReturnStatement(returnNode *ast.Node, allowVoid bool) bool {
	if returnNode == nil || returnNode.Kind != ast.KindReturnStatement {
		return false
	}

	returnStmt := returnNode.AsReturnStatement()
	if returnStmt == nil {
		return false
	}

	// Empty return is OK (return; for control flow)
	if returnStmt.Expression == nil {
		return false
	}

	// If allowVoid is true, void expressions are allowed
	if allowVoid && isVoidExpression(returnStmt.Expression) {
		return false
	}

	// Any other return with a value is problematic
	return true
}

// isDirectExecutorFunction checks if we're in the direct executor function, not a nested function
func isDirectExecutorFunction(node *ast.Node, executorNode *ast.Node) bool {
	// Walk up the tree to find the enclosing function
	current := node
	for current != nil {
		if current == executorNode {
			return true
		}
		// Stop if we hit another function (nested function)
		if current.Kind == ast.KindFunctionDeclaration ||
			current.Kind == ast.KindFunctionExpression ||
			current.Kind == ast.KindArrowFunction {
			return current == executorNode
		}
		current = current.Parent
	}
	return false
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	return rule.RuleListeners{
		ast.KindNewExpression: func(node *ast.Node) {
			newExpr := node.AsNewExpression()
			if newExpr == nil {
				return
			}

			// Check if this is new Promise(...)
			if !isPromiseConstructor(newExpr.Expression) {
				return
			}

			// Get the executor function (first argument)
			if newExpr.Arguments == nil || len(newExpr.Arguments.Nodes) == 0 {
				return
			}

			executor := newExpr.Arguments.Nodes[0]
			if executor == nil {
				return
			}

			// Check if executor is a function (arrow function or function expression)
			var executorBody *ast.Node

			switch executor.Kind {
			case ast.KindArrowFunction:
				arrowFunc := executor.AsArrowFunction()
				if arrowFunc == nil {
					return
				}

				// For arrow functions with expression body, check if it's returning a value
				if arrowFunc.Body != nil {
					if arrowFunc.Body.Kind != ast.KindBlock {
						// Expression body - this is an implicit return
						if !opts.AllowVoid || !isVoidExpression(arrowFunc.Body) {
							ctx.ReportNode(arrowFunc.Body, rule.RuleMessage{
								Id:          "returnsValue",
								Description: "Return values from promise executor functions cannot be read.",
							})
						}
						return
					}
					executorBody = arrowFunc.Body
				}

			case ast.KindFunctionExpression:
				funcExpr := executor.AsFunctionExpression()
				if funcExpr != nil && funcExpr.Body != nil {
					executorBody = funcExpr.Body
				}

			default:
				return
			}

			// Walk through the executor body looking for return statements
			if executorBody != nil {
				walkForReturns(ctx, executorBody, executor, opts.AllowVoid)
			}
		},
	}
}

// walkForReturns walks the AST to find return statements in the executor function
func walkForReturns(ctx rule.RuleContext, node *ast.Node, executorNode *ast.Node, allowVoid bool) {
	if node == nil {
		return
	}

	// Check if this is a return statement
	if node.Kind == ast.KindReturnStatement {
		if checkReturnStatement(node, allowVoid) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "returnsValue",
				Description: "Return values from promise executor functions cannot be read.",
			})
		}
		return
	}

	// Don't descend into nested functions
	if node != executorNode {
		if node.Kind == ast.KindFunctionDeclaration ||
			node.Kind == ast.KindFunctionExpression ||
			node.Kind == ast.KindArrowFunction {
			return
		}
	}

	// Walk children
	for _, child := range node.Children() {
		walkForReturns(ctx, child, executorNode, allowVoid)
	}
}

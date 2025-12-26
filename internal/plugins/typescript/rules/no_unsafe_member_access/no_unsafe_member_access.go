package no_unsafe_member_access

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type NoUnsafeMemberAccessOptions struct {
	AllowOptionalChaining bool `json:"allowOptionalChaining"`
}

func buildUnsafeComputedMemberAccessMessage(property, t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeComputedMemberAccess",
		Description: fmt.Sprintf("Computed name %v resolves to an %v value.", property, t),
	}
}
func buildUnsafeMemberExpressionMessage(property, t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeMemberExpression",
		Description: fmt.Sprintf("Unsafe member access %v on an %v value.", property, t),
	}
}
func buildUnsafeThisMemberExpressionMessage(property string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "unsafeThisMemberExpression",
		Description: fmt.Sprintf("Unsafe member access %v on an `any` value. `this` is typed as `any`.", property) +
			"You can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.",
	}
}

type state uint8

const (
	stateUnsafe state = iota
	stateSafe   state = iota
)

func createDataType(t *checker.Type) string {
	if utils.IsIntrinsicErrorType(t) {
		return "`error` typed"
	}
	return "`any`"
}

func parseOptions(options any) NoUnsafeMemberAccessOptions {
	opts := NoUnsafeMemberAccessOptions{}
	if options == nil {
		return opts
	}
	// Handle array format: [{ option: value }]
	if arr, ok := options.([]interface{}); ok {
		if len(arr) > 0 {
			if m, ok := arr[0].(map[string]interface{}); ok {
				if v, ok := m["allowOptionalChaining"].(bool); ok {
					opts.AllowOptionalChaining = v
				}
			}
		}
		return opts
	}
	// Handle direct object format
	if m, ok := options.(map[string]interface{}); ok {
		if v, ok := m["allowOptionalChaining"].(bool); ok {
			opts.AllowOptionalChaining = v
		}
	}
	return opts
}

// Check if the current member expression uses optional chaining directly
func isDirectlyOptionalChain(node *ast.Node) bool {
	if ast.IsPropertyAccessExpression(node) && node.AsPropertyAccessExpression().QuestionDotToken != nil {
		return true
	}
	if ast.IsElementAccessExpression(node) && node.AsElementAccessExpression().QuestionDotToken != nil {
		return true
	}
	return false
}

// shouldAllowWithOptionalChaining checks if a node should be allowed when allowOptionalChaining is true
// The rule is: if the parent expression of the current node uses optional chaining (i.e., has ?.)
// then we allow the current node
func shouldAllowWithOptionalChaining(node *ast.Node) bool {
	// Get the expression that this member access is operating on
	expression := node.Expression()
	if expression == nil {
		return false
	}

	// Check if the expression itself is an optional chain
	// e.g., in `value?.middle.inner`, when we're checking `.inner`,
	// the expression is `value?.middle` which has optional chaining
	if ast.IsAccessExpression(expression) && isDirectlyOptionalChain(expression) {
		return true
	}

	return false
}

var NoUnsafeMemberAccessRule = rule.CreateRule(rule.Rule{
	Name: "no-unsafe-member-access",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		compilerOptions := ctx.Program.Options()
		isNoImplicitThis := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.NoImplicitThis,
		)

		stateCache := map[*ast.Node]state{}

		var checkMemberExpression func(node *ast.Node) state
		checkMemberExpression = func(node *ast.Node) state {
			cachedState, ok := stateCache[node]
			if ok {
				return cachedState
			}

			// TODO(port-perf): couldn't it be done at visitchildren phase?
			parent := node.Parent
			for !ast.IsSourceFile(parent) {
				// ignore MemberExpressions with ancestors of type `TSClassImplements` or `TSInterfaceHeritage`
				if ast.IsHeritageClause(parent) {
					return stateSafe
				}
				parent = parent.Parent
			}

			expression := node.Expression()
			if ast.IsAccessExpression(expression) {
				objectState := checkMemberExpression(expression)
				if objectState == stateUnsafe {
					// if the object is unsafe, we know this will be unsafe as well
					// we don't need to report, as we have already reported on the inner member expr
					stateCache[node] = objectState
					return objectState
				}
			}

			t := ctx.TypeChecker.GetTypeAtLocation(expression)
			state := stateSafe
			if utils.IsTypeAnyType(t) {
				state = stateUnsafe
			}
			stateCache[node] = state

			if state == stateUnsafe {
				// If allowOptionalChaining is enabled and the parent expression uses optional chaining
				// then we should allow it (skip reporting)
				if opts.AllowOptionalChaining && shouldAllowWithOptionalChaining(node) {
					return state
				}

				var property *ast.Node
				var propertyName string
				if ast.IsPropertyAccessExpression(node) {
					property = node.Name()
					loc := utils.TrimNodeTextRange(ctx.SourceFile, property)
					propertyName = "." + ctx.SourceFile.Text()[loc.Pos():loc.End()]
				} else if ast.IsElementAccessExpression(node) {
					property = node.AsElementAccessExpression().ArgumentExpression
					loc := utils.TrimNodeTextRange(ctx.SourceFile, property)
					propertyName = "[" + ctx.SourceFile.Text()[loc.Pos():loc.End()] + "]"
				}

				// let messageId: 'unsafeMemberExpression' | 'unsafeThisMemberExpression' =
				//   'unsafeMemberExpression';

				if !isNoImplicitThis {
					// `this.foo` or `this.foo[bar]`
					thisExpression := utils.GetThisExpression(node)

					if thisExpression != nil && utils.IsTypeAnyType(
						utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, thisExpression)) {
						ctx.ReportNode(property, buildUnsafeThisMemberExpressionMessage(propertyName))
						return state
					}
				}

				ctx.ReportNode(property, buildUnsafeMemberExpressionMessage(propertyName, createDataType(t)))
			}

			return state
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				checkMemberExpression(node)
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				checkMemberExpression(node)

				arg := node.AsElementAccessExpression().ArgumentExpression
				// x[1]
				if ast.IsLiteralExpression(arg) {
					// perf optimizations - literals can obviously never be `any`
					return
				}

				// x[1++] x[++x] etc
				// FUN FACT - **all** update expressions return type number, regardless of the argument's type,
				// because JS engines return NaN if there the argument is not a number.
				unaryOperatorKind := ast.KindUnknown
				if ast.IsPrefixUnaryExpression(arg) {
					unaryOperatorKind = arg.AsPrefixUnaryExpression().Operator
				} else if arg.Kind == ast.KindPostfixUnaryExpression {
					unaryOperatorKind = arg.AsPostfixUnaryExpression().Operator
				}
				if unaryOperatorKind == ast.KindPlusPlusToken || unaryOperatorKind == ast.KindMinusMinusToken {
					return
				}

				t := ctx.TypeChecker.GetTypeAtLocation(arg)

				if utils.IsTypeAnyType(t) {
					loc := utils.TrimNodeTextRange(ctx.SourceFile, arg)
					propertyName := "[" + ctx.SourceFile.Text()[loc.Pos():loc.End()] + "]"
					ctx.ReportNode(arg, buildUnsafeComputedMemberAccessMessage(propertyName, createDataType(t)))
				}
			},
		}
	},
})

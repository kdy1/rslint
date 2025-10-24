package prefer_regexp_exec

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// PreferRegexpExecRule implements the prefer-regexp-exec rule
// Enforces using RegExp#exec over String#match when no global flag is present
var PreferRegexpExecRule = rule.CreateRule(rule.Rule{
	Name: "prefer-regexp-exec",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindCallExpression: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// TODO: Implement logic to detect String#match calls
			// 1. Check if this is a method call on a string type
			// 2. Check if the method name is "match"
			// 3. Check if the argument is a RegExp literal or RegExp type
			// 4. If RegExp has no global flag, suggest using RegExp#exec instead
			//
			// Example to detect:
			//   "foo".match(/bar/)  -> Should use /bar/.exec("foo")
			//   str.match(regex)    -> Should use regex.exec(str) if regex has no 'g' flag
			//
			// Valid cases (should not report):
			//   "foo".match(/bar/g) -> Global flag present, match is appropriate
			//   /bar/.exec("foo")   -> Already using exec
		},
	}
}

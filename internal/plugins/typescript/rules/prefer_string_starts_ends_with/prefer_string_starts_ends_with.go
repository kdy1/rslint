package prefer_string_starts_ends_with

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// PreferStringStartsEndsWithRule implements the prefer-string-starts-ends-with rule
// Enforce startsWith/endsWith over indexOf/regex
var PreferStringStartsEndsWithRule = rule.CreateRule(rule.Rule{
	Name: "prefer-string-starts-ends-with",
	Run:  run,
})

type Options struct {
	AllowSingleElementEquality string `json:"allowSingleElementEquality"`
}

func parseOptions(options any) Options {
	opts := Options{
		AllowSingleElementEquality: "never",
	}

	if options == nil {
		return opts
	}

	var optsMap map[string]interface{}
	var ok bool

	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, ok = optArray[0].(map[string]interface{})
	} else {
		optsMap, ok = options.(map[string]interface{})
	}

	if ok {
		if v, ok := optsMap["allowSingleElementEquality"].(string); ok {
			opts.AllowSingleElementEquality = v
		}
	}
	return opts
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	if ctx.TypeChecker == nil {
		return rule.RuleListeners{}
	}

	opts := parseOptions(options)
	_ = opts // TODO: Use options in implementation

	return rule.RuleListeners{
		ast.KindBinaryExpression: func(node *ast.Node) {
			// TODO: Implement full logic
			// This is a simplified placeholder implementation
			// Full implementation would detect patterns like:
			// - s.indexOf('foo') === 0        -> s.startsWith('foo')
			// - s.charAt(0) === 'a'           -> s.startsWith('a')
			// - s[0] === 'a'                  -> s.startsWith('a')
			// - s.slice(0, 3) === 'foo'       -> s.startsWith('foo')
			// - s.match(/^foo/) !== null      -> s.startsWith('foo')
			// - s.lastIndexOf('foo') === ...  -> s.endsWith('foo')

			// Each pattern needs:
			// 1. Type checking to ensure s is a string
			// 2. Pattern detection
			// 3. Auto-fix generation
		},
		ast.KindCallExpression: func(node *ast.Node) {
			// TODO: Check for regex patterns
			// - /^foo/.test(s)  -> s.startsWith('foo')
			// - /foo$/.test(s)  -> s.endsWith('foo')
		},
	}
}

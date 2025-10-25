package no_console

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Options for the no-console rule
type Options struct {
	Allow []string `json:"allow"` // List of allowed console methods
}

func parseOptions(options any) Options {
	opts := Options{
		Allow: []string{},
	}

	if options == nil {
		return opts
	}

	// Handle array format: [{ "allow": ["warn", "error"] }]
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		if optsMap, ok := optArray[0].(map[string]interface{}); ok {
			if allowArray, ok := optsMap["allow"].([]interface{}); ok {
				for _, item := range allowArray {
					if str, ok := item.(string); ok {
						opts.Allow = append(opts.Allow, str)
					}
				}
			}
		}
	} else if optsMap, ok := options.(map[string]interface{}); ok {
		// Handle direct object format
		if allowArray, ok := optsMap["allow"].([]interface{}); ok {
			for _, item := range allowArray {
				if str, ok := item.(string); ok {
					opts.Allow = append(opts.Allow, str)
				}
			}
		}
	}

	return opts
}

func buildUnexpectedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected console statement.",
	}
}

func buildUnexpectedLimitedMessage(allow []string) rule.RuleMessage {
	methods := ""
	for i, method := range allow {
		if i > 0 {
			methods += ", "
		}
		methods += method
	}
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Unexpected console statement. Allowed methods: " + methods,
	}
}

// Check if a method is in the allow list
func isAllowed(method string, allow []string) bool {
	for _, m := range allow {
		if m == method {
			return true
		}
	}
	return false
}

// Try to get the property name from different node types
func getPropertyName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		id := node.AsIdentifier()
		if id != nil {
			return id.Text
		}
	case ast.KindStringLiteral:
		lit := node.AsStringLiteral()
		if lit != nil {
			return lit.Text
		}
	case ast.KindNumericLiteral:
		// Handle console[0]() type calls
		return ""
	}

	return ""
}

var NoConsoleRule = rule.CreateRule(rule.Rule{
	Name: "no-console",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		listeners := rule.RuleListeners{}

		// Handle PropertyAccessExpression: console.log(), console.warn(), etc.
		listeners[ast.KindPropertyAccessExpression] = func(node *ast.Node) {
			pae := node.AsPropertyAccessExpression()
			if pae == nil || pae.Name() == nil {
				return
			}

			// Get method name - Name() returns a DeclarationName/Node
			methodName := ""
			nameNode := pae.Name()
			if nameNode != nil && ast.IsIdentifier(nameNode) {
				methodName = nameNode.AsIdentifier().Text
			}

			// Must be part of a call expression
			// We need to check if the parent is a CallExpression
			// For now, we'll check all property accesses to console
			expr := pae.Expression
			if expr == nil || expr.Kind != ast.KindIdentifier {
				return
			}

			id := expr.AsIdentifier()
			if id == nil {
				return
			}

			objName := id.Text
			if objName != "console" {
				return
			}

			// Check if 'console' is a local variable (shadowing the global)
			// If the identifier has a symbol with declarations, it's locally defined
			if ctx.TypeChecker != nil {
				symbol := ctx.TypeChecker.GetSymbolAtLocation(expr)
				if symbol != nil {
					declarations := symbol.Declarations()
					if len(declarations) > 0 {
						// console is locally defined, skip this reference
						return
					}
				}
			}

			// Check if method is allowed
			if len(opts.Allow) > 0 && methodName != "" && isAllowed(methodName, opts.Allow) {
				return
			}

			// Report the error
			msg := buildUnexpectedMessage()
			if len(opts.Allow) > 0 {
				msg = buildUnexpectedLimitedMessage(opts.Allow)
			}

			ctx.ReportNode(node, msg)
		}

		// Handle ElementAccessExpression: console['log'](), console[method]()
		listeners[ast.KindElementAccessExpression] = func(node *ast.Node) {
			eae := node.AsElementAccessExpression()
			if eae == nil {
				return
			}

			expr := eae.Expression
			if expr == nil || expr.Kind != ast.KindIdentifier {
				return
			}

			id := expr.AsIdentifier()
			if id == nil {
				return
			}

			objName := id.Text
			if objName != "console" {
				return
			}

			// Check if 'console' is a local variable (shadowing the global)
			// If the identifier has a symbol with declarations, it's locally defined
			if ctx.TypeChecker != nil {
				symbol := ctx.TypeChecker.GetSymbolAtLocation(expr)
				if symbol != nil {
					declarations := symbol.Declarations()
					if len(declarations) > 0 {
						// console is locally defined, skip this reference
						return
					}
				}
			}

			// Try to get the method name
			methodName := ""
			if eae.ArgumentExpression != nil {
				methodName = getPropertyName(eae.ArgumentExpression)
			}

			// Check if method is allowed
			if len(opts.Allow) > 0 && methodName != "" && isAllowed(methodName, opts.Allow) {
				return
			}

			// Report the error
			msg := buildUnexpectedMessage()
			if len(opts.Allow) > 0 {
				msg = buildUnexpectedLimitedMessage(opts.Allow)
			}

			ctx.ReportNode(node, msg)
		}

		return listeners
	},
})

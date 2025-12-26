package no_var_requires

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoVarRequiresOptions struct {
	Allow []string `json:"allow"`
}

func buildNoVarRequiresMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noVarReqs",
		Description: "Require statement not part of import statement.",
	}
}

// Helper function to check if a string argument matches any of the allowed patterns
func isAllowed(arg string, allowPatterns []string) bool {
	if len(allowPatterns) == 0 {
		return false
	}

	for _, pattern := range allowPatterns {
		matched, err := regexp.MatchString(pattern, arg)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// Helper function to extract string literal value from a node
func getStringLiteralValue(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}

	if ast.IsStringLiteral(node) {
		return node.AsStringLiteral().Text(), true
	}

	if ast.IsNoSubstitutionTemplateLiteral(node) {
		return node.AsNoSubstitutionTemplateLiteral().Text(), true
	}

	return "", false
}

// Helper function to check if a call expression is a require() call
func isRequireCall(node *ast.Node) bool {
	if node == nil || !ast.IsCallExpression(node) {
		return false
	}

	callExpr := node.AsCallExpression()
	expr := callExpr.Expression

	// Check for direct require identifier
	if ast.IsIdentifier(expr) && expr.AsIdentifier().EscapedText() == "require" {
		return true
	}

	return false
}

// Helper function to check if a node is within a variable declaration
func findRequireInExpression(node *ast.Node, allowPatterns []string) *ast.Node {
	if node == nil {
		return nil
	}

	// Check if this node itself is a require call
	if isRequireCall(node) {
		callExpr := node.AsCallExpression()
		// Check if it's allowed
		if callExpr.Arguments != nil && len(callExpr.Arguments.Nodes) > 0 {
			firstArg := callExpr.Arguments.Nodes[0]
			if value, ok := getStringLiteralValue(firstArg); ok {
				if isAllowed(value, allowPatterns) {
					return nil
				}
			}
		}
		return node
	}

	// Recursively check common expression types
	if ast.IsCallExpression(node) {
		callExpr := node.AsCallExpression()
		// Check the callee
		if found := findRequireInExpression(callExpr.Expression, allowPatterns); found != nil {
			return found
		}
		// Check arguments
		if callExpr.Arguments != nil {
			for _, arg := range callExpr.Arguments.Nodes {
				if found := findRequireInExpression(arg, allowPatterns); found != nil {
					return found
				}
			}
		}
	}

	if ast.IsAsExpression(node) {
		return findRequireInExpression(node.AsAsExpression().Expression, allowPatterns)
	}

	if ast.IsTypeAssertionExpression(node) {
		return findRequireInExpression(node.AsTypeAssertionExpression().Expression, allowPatterns)
	}

	if ast.IsPropertyAccessExpression(node) {
		return findRequireInExpression(node.AsPropertyAccessExpression().Expression, allowPatterns)
	}

	if ast.IsElementAccessExpression(node) {
		elemAccess := node.AsElementAccessExpression()
		if found := findRequireInExpression(elemAccess.Expression, allowPatterns); found != nil {
			return found
		}
		return findRequireInExpression(elemAccess.ArgumentExpression, allowPatterns)
	}

	if ast.IsParenthesizedExpression(node) {
		return findRequireInExpression(node.AsParenthesizedExpression().Expression, allowPatterns)
	}

	return nil
}

var NoVarRequiresRule = rule.CreateRule(rule.Rule{
	Name: "no-var-requires",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoVarRequiresOptions{
			Allow: []string{},
		}

		// Parse options
		if options != nil {
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
				if allow, ok := optsMap["allow"].([]interface{}); ok {
					opts.Allow = make([]string, 0, len(allow))
					for _, pattern := range allow {
						if str, ok := pattern.(string); ok {
							opts.Allow = append(opts.Allow, str)
						}
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				if node.Kind != ast.KindVariableDeclaration {
					return
				}

				varDecl := node.AsVariableDeclaration()
				if varDecl.Initializer == nil {
					return
				}

				// Find require() call in the initializer
				requireNode := findRequireInExpression(varDecl.Initializer, opts.Allow)
				if requireNode != nil {
					ctx.ReportNode(requireNode, buildNoVarRequiresMessage())
				}
			},
		}
	},
})

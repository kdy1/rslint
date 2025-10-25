package prefer_arrow_callback

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// PreferArrowCallbackRule implements the prefer-arrow-callback rule
// Require arrow functions as callbacks
var PreferArrowCallbackRule = rule.Rule{
	Name: "prefer-arrow-callback",
	Run:  run,
}

// Options for prefer-arrow-callback rule
type Options struct {
	AllowNamedFunctions bool `json:"allowNamedFunctions"`
	AllowUnboundThis    bool `json:"allowUnboundThis"`
}

func parseOptions(options any) Options {
	opts := Options{
		AllowNamedFunctions: false,
		AllowUnboundThis:    true,
	}

	if options == nil {
		return opts
	}

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
		if v, ok := optsMap["allowNamedFunctions"].(bool); ok {
			opts.AllowNamedFunctions = v
		}
		if v, ok := optsMap["allowUnboundThis"].(bool); ok {
			opts.AllowUnboundThis = v
		}
	}

	return opts
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)

	return rule.RuleListeners{
		ast.KindCallExpression: func(node *ast.Node) {
			callExpr := node.AsCallExpression()
			if callExpr == nil || callExpr.Arguments == nil {
				return
			}

			// Check each argument
			for _, arg := range callExpr.Arguments.Nodes {
				if arg == nil {
					continue
				}

				// Check if argument is a function expression
				funcExpr := arg.AsFunctionExpression()
				if funcExpr == nil {
					continue
				}

				// Skip generator functions
				if funcExpr.AsteriskToken != nil {
					continue
				}

				// Skip if allowNamedFunctions is true and function has a name
				if opts.AllowNamedFunctions && funcExpr.Name() != nil {
					continue
				}

				// Check if function uses 'this'
				usesThis := checkUsesThis(funcExpr.Body)

				// Skip if function uses 'this' and doesn't have .bind(this)
				if usesThis && opts.AllowUnboundThis {
					// Check if this is followed by .bind(this)
					if !hasBindThis(ctx, node, arg) {
						continue
					}
				}

				// Check if function references itself (named function used recursively)
				if funcExpr.Name() != nil && checkSelfReferential(funcExpr.Body, funcExpr.Name().Text()) {
					continue
				}

				// Should suggest arrow function
				msg := rule.RuleMessage{
					Id:          "preferArrow",
					Description: "Unexpected function expression. Use an arrow function instead.",
				}

				// Create fix - convert function to arrow
				fix := createArrowFix(ctx, arg, funcExpr)
				if fix != nil {
					ctx.ReportNodeWithFixes(arg, msg, *fix)
				} else {
					ctx.ReportNode(arg, msg)
				}
			}
		},
	}
}

// checkUsesThis checks if a node or its descendants use 'this'
func checkUsesThis(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check current node
	if node.Kind == ast.KindThisKeyword {
		return true
	}

	// Don't descend into nested functions
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		return false
	}

	// Check children
	for _, child := range ast.GetChildren(node) {
		if checkUsesThis(child) {
			return true
		}
	}

	return false
}

// checkSelfReferential checks if function references itself by name
func checkSelfReferential(node *ast.Node, name string) bool {
	if node == nil || name == "" {
		return false
	}

	// Check if current node is an identifier matching the function name
	if node.Kind == ast.KindIdentifier {
		if ident := node.AsIdentifier(); ident != nil && ident.Text() == name {
			return true
		}
	}

	// Don't descend into nested functions
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		return false
	}

	// Check children
	for _, child := range ast.GetChildren(node) {
		if checkSelfReferential(child, name) {
			return true
		}
	}

	return false
}

// hasBindThis checks if the call expression has .bind(this) after the function
func hasBindThis(ctx rule.RuleContext, callNode *ast.Node, funcArg *ast.Node) bool {
	// Check if funcArg is wrapped in a .bind(this) call
	// This is a simplified check - full implementation would need more analysis
	text := ctx.SourceFile.Text()
	argRange := utils.TrimNodeTextRange(ctx.SourceFile, funcArg)

	// Check if there's .bind(this) after the function
	endPos := argRange.End()
	remaining := text[endPos:]

	// Simple pattern matching for .bind(this)
	if strings.HasPrefix(strings.TrimSpace(remaining), ".bind(this)") {
		return true
	}

	return false
}

// createArrowFix creates a fix to convert function expression to arrow function
func createArrowFix(ctx rule.RuleContext, node *ast.Node, funcExpr *ast.FunctionExpression) *rule.RuleFix {
	text := ctx.SourceFile.Text()
	nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	fullText := text[nodeRange.Pos():nodeRange.End()]

	// Check for duplicate parameters (can't fix)
	if hasDuplicateParams(funcExpr) {
		return nil
	}

	// Build arrow function
	var replacement strings.Builder

	// Parameters
	if funcExpr.Parameters != nil && len(funcExpr.Parameters.Nodes) > 0 {
		// Get parameter list text
		paramsStart := funcExpr.Parameters.Pos() - nodeRange.Pos()
		paramsEnd := funcExpr.Parameters.End() - nodeRange.Pos()

		if paramsStart >= 0 && paramsEnd <= len(fullText) {
			params := fullText[paramsStart:paramsEnd]

			// Single parameter without type annotation can omit parens
			if len(funcExpr.Parameters.Nodes) == 1 && !strings.Contains(params, ":") {
				// Check if it's a simple identifier
				if param := funcExpr.Parameters.Nodes[0]; param != nil {
					if binding := param.AsParameterDeclaration(); binding != nil && binding.Name != nil {
						if ident := binding.Name.AsIdentifier(); ident != nil {
							replacement.WriteString(ident.Text())
						} else {
							replacement.WriteString(params)
						}
					}
				}
			} else {
				replacement.WriteString(params)
			}
		}
	} else {
		replacement.WriteString("()")
	}

	// Arrow
	replacement.WriteString(" => ")

	// Body
	if funcExpr.Body != nil {
		bodyStart := funcExpr.Body.Pos() - nodeRange.Pos()
		bodyEnd := funcExpr.Body.End() - nodeRange.Pos()

		if bodyStart >= 0 && bodyEnd <= len(fullText) {
			replacement.WriteString(fullText[bodyStart:bodyEnd])
		}
	}

	return &rule.RuleFix{
		Range:       nodeRange,
		Replacement: replacement.String(),
	}
}

// hasDuplicateParams checks if function has duplicate parameter names
func hasDuplicateParams(funcExpr *ast.FunctionExpression) bool {
	if funcExpr.Parameters == nil {
		return false
	}

	seen := make(map[string]bool)
	for _, param := range funcExpr.Parameters.Nodes {
		if param == nil {
			continue
		}

		paramDecl := param.AsParameterDeclaration()
		if paramDecl == nil || paramDecl.Name == nil {
			continue
		}

		if ident := paramDecl.Name.AsIdentifier(); ident != nil {
			name := ident.Text()
			if seen[name] {
				return true
			}
			seen[name] = true
		}
	}

	return false
}

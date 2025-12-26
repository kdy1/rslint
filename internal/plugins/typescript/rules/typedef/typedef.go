// Package typedef implements the @typescript-eslint/typedef rule.
// This rule enforces type annotations in various code contexts to ensure
// explicit type declarations, improving code documentation and type safety.
package typedef

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type TypedefOptions struct {
	ArrayDestructuring             bool `json:"arrayDestructuring"`
	ArrowParameter                 bool `json:"arrowParameter"`
	MemberVariableDeclaration      bool `json:"memberVariableDeclaration"`
	ObjectDestructuring            bool `json:"objectDestructuring"`
	Parameter                      bool `json:"parameter"`
	PropertyDeclaration            bool `json:"propertyDeclaration"`
	VariableDeclaration            bool `json:"variableDeclaration"`
	VariableDeclarationIgnoreFunction bool `json:"variableDeclarationIgnoreFunction"`
}

func parseOptions(options any) TypedefOptions {
	opts := TypedefOptions{
		ArrayDestructuring:             false,
		ArrowParameter:                 false,
		MemberVariableDeclaration:      false,
		ObjectDestructuring:            false,
		Parameter:                      false,
		PropertyDeclaration:            false,
		VariableDeclaration:            false,
		VariableDeclarationIgnoreFunction: false,
	}

	if options == nil {
		return opts
	}

	var optsMap map[string]interface{}
	if optsArray, ok := options.([]interface{}); ok && len(optsArray) > 0 {
		if m, ok := optsArray[0].(map[string]interface{}); ok {
			optsMap = m
		}
	} else if m, ok := options.(map[string]interface{}); ok {
		optsMap = m
	}

	if optsMap != nil {
		if v, ok := optsMap["arrayDestructuring"].(bool); ok {
			opts.ArrayDestructuring = v
		}
		if v, ok := optsMap["arrowParameter"].(bool); ok {
			opts.ArrowParameter = v
		}
		if v, ok := optsMap["memberVariableDeclaration"].(bool); ok {
			opts.MemberVariableDeclaration = v
		}
		if v, ok := optsMap["objectDestructuring"].(bool); ok {
			opts.ObjectDestructuring = v
		}
		if v, ok := optsMap["parameter"].(bool); ok {
			opts.Parameter = v
		}
		if v, ok := optsMap["propertyDeclaration"].(bool); ok {
			opts.PropertyDeclaration = v
		}
		if v, ok := optsMap["variableDeclaration"].(bool); ok {
			opts.VariableDeclaration = v
		}
		if v, ok := optsMap["variableDeclarationIgnoreFunction"].(bool); ok {
			opts.VariableDeclarationIgnoreFunction = v
		}
	}

	return opts
}

func buildExpectedTypedefMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedTypedef",
		Description: "Expected a type annotation.",
	}
}

func buildExpectedTypedefNamedMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedTypedefNamed",
		Description: "Expected " + name + " to have a type annotation.",
	}
}

// Check if a node has a type annotation
func hasTypeAnnotation(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindParameter:
		param := node.AsParameterDeclaration()
		return param != nil && param.Type != nil
	case ast.KindVariableDeclaration:
		varDecl := node.AsVariableDeclaration()
		return varDecl != nil && varDecl.Type != nil
	case ast.KindPropertyDeclaration:
		prop := node.AsPropertyDeclaration()
		return prop != nil && prop.Type != nil
	}
	return false
}

// Get the name of an identifier
func getIdentifierName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	if ast.IsIdentifier(node) {
		ident := node.AsIdentifier()
		if ident != nil {
			return ident.Text
		}
	}

	return ""
}

// Check if variable declaration initializer is a function
func isVariableDeclarationFunction(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindVariableDeclaration {
		return false
	}

	varDecl := node.AsVariableDeclaration()
	if varDecl == nil || varDecl.Initializer == nil {
		return false
	}

	kind := varDecl.Initializer.Kind
	return kind == ast.KindFunctionExpression || kind == ast.KindArrowFunction
}

// Check if we're in a for-of or for-in loop
func isInForLoop(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		if current.Kind == ast.KindForOfStatement || current.Kind == ast.KindForInStatement {
			return true
		}
		// Stop at function boundaries
		if current.Kind == ast.KindFunctionDeclaration ||
			current.Kind == ast.KindFunctionExpression ||
			current.Kind == ast.KindArrowFunction {
			break
		}
		current = current.Parent
	}
	return false
}

// Check if parameter is a catch clause parameter
func isInCatchClause(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		if current.Kind == ast.KindCatchClause {
			return true
		}
		// Stop at function boundaries
		if current.Kind == ast.KindFunctionDeclaration ||
			current.Kind == ast.KindFunctionExpression ||
			current.Kind == ast.KindArrowFunction {
			break
		}
		current = current.Parent
	}
	return false
}

// Check if node is in an assignment pattern context (not a declaration)
func isInAssignmentPattern(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		// If we find a variable declaration statement, it's a declaration
		if current.Kind == ast.KindVariableDeclarationList || current.Kind == ast.KindVariableStatement {
			return false
		}
		// If we find an assignment pattern in an expression context, it's an assignment
		if current.Kind == ast.KindBinaryExpression {
			return true
		}
		// Stop at certain boundaries
		if current.Kind == ast.KindFunctionDeclaration ||
			current.Kind == ast.KindFunctionExpression ||
			current.Kind == ast.KindArrowFunction ||
			current.Kind == ast.KindBlock {
			break
		}
		current = current.Parent
	}
	return false
}

var TypedefRule = rule.CreateRule(rule.Rule{
	Name: "@typescript-eslint/typedef",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			// Handle arrow function parameters
			ast.KindArrowFunction: func(node *ast.Node) {
				if !opts.ArrowParameter {
					return
				}

				arrowFn := node.AsArrowFunction()
				if arrowFn == nil || arrowFn.Parameters == nil {
					return
				}

				for _, param := range arrowFn.Parameters.Nodes {
					if param == nil || param.Kind != ast.KindParameter {
						continue
					}

					if hasTypeAnnotation(param) {
						continue
					}

					paramDecl := param.AsParameterDeclaration()
					if paramDecl == nil || paramDecl.Name() == nil {
						continue
					}

					name := getIdentifierName(paramDecl.Name().AsNode())
					if name != "" {
						ctx.ReportNode(param, buildExpectedTypedefNamedMessage(name))
					}
				}
			},

			// Handle class member variables
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				if !opts.MemberVariableDeclaration {
					return
				}

				if hasTypeAnnotation(node) {
					return
				}

				prop := node.AsPropertyDeclaration()
				if prop == nil || prop.Name() == nil {
					return
				}

				// Check if it's a function assignment and should be ignored
				if opts.VariableDeclarationIgnoreFunction && prop.Initializer != nil {
					kind := prop.Initializer.Kind
					if kind == ast.KindFunctionExpression || kind == ast.KindArrowFunction {
						return
					}
				}

				name := getIdentifierName(prop.Name().AsNode())
				if name != "" {
					ctx.ReportNode(node, buildExpectedTypedefNamedMessage(name))
				} else {
					// For computed properties like ['state']
					ctx.ReportNode(node, buildExpectedTypedefMessage())
				}
			},

			// Handle function and method parameters
			ast.KindParameter: func(node *ast.Node) {
				if !opts.Parameter {
					return
				}

				if hasTypeAnnotation(node) {
					return
				}

				// Skip catch clause parameters (TypeScript doesn't allow type annotations)
				if isInCatchClause(node) {
					return
				}

				param := node.AsParameterDeclaration()
				if param == nil || param.Name() == nil {
					return
				}

				// Skip if parent is arrow function (handled separately)
				if node.Parent != nil && node.Parent.Kind == ast.KindArrowFunction {
					return
				}

				name := getIdentifierName(param.Name().AsNode())
				if name != "" {
					// Check if parameter has a default value
					if param.Initializer != nil {
						ctx.ReportNode(node, buildExpectedTypedefMessage())
					} else {
						ctx.ReportNode(node, buildExpectedTypedefNamedMessage(name))
					}
				} else {
					// For destructuring parameters
					ctx.ReportNode(node, buildExpectedTypedefMessage())
				}
			},

			// Handle variable declarations
			ast.KindVariableDeclaration: func(node *ast.Node) {
				if !opts.VariableDeclaration {
					return
				}

				varDecl := node.AsVariableDeclaration()
				if varDecl == nil || varDecl.Name() == nil {
					return
				}

				// Skip if already has type annotation
				if varDecl.Type != nil {
					return
				}

				// Skip for-in and for-of loops
				if isInForLoop(node) {
					return
				}

				// Skip catch clause variables
				if isInCatchClause(node) {
					return
				}

				// Check if we should ignore function declarations
				if opts.VariableDeclarationIgnoreFunction && isVariableDeclarationFunction(node) {
					return
				}

				name := getIdentifierName(varDecl.Name().AsNode())
				if name != "" {
					ctx.ReportNode(node, buildExpectedTypedefNamedMessage(name))
				}
			},
		}
	},
})

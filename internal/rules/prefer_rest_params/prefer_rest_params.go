package prefer_rest_params

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// PreferRestParamsRule implements the prefer-rest-params rule
// Require rest parameters instead of arguments
var PreferRestParamsRule = rule.Rule{
	Name: "prefer-rest-params",
	Run:  run,
}

func buildMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferRestParams",
		Description: "Use the rest parameters instead of 'arguments'.",
	}
}

// isInFunctionScope checks if we're inside a function that has its own arguments object
func isInFunctionScope(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		kind := current.Kind
		// Arrow functions don't have their own arguments
		if kind == ast.KindArrowFunction {
			return false
		}
		// Regular functions have arguments
		if kind == ast.KindFunctionDeclaration ||
			kind == ast.KindFunctionExpression ||
			kind == ast.KindMethodDeclaration ||
			kind == ast.KindConstructor ||
			kind == ast.KindGetAccessor ||
			kind == ast.KindSetAccessor {
			return true
		}
		current = current.Parent
	}
	return false
}

// hasVariableNamed searches a Block/SourceFile for a var declaration with the given name
// This is a simple iterative search, not a deep recursive one
func hasVariableNamed(block *ast.Node, name string) bool {
	if block == nil {
		return false
	}

	var statements []*ast.Node
	if block.Kind == ast.KindBlock {
		if b := block.AsBlock(); b != nil && b.Statements != nil {
			statements = b.Statements.Nodes
		}
	}

	for _, stmt := range statements {
		// Check if this is a VariableStatement
		if stmt.Kind == ast.KindVariableStatement {
			if varStmt := stmt.AsVariableStatement(); varStmt != nil && varStmt.DeclarationList != nil {
				if declList := varStmt.DeclarationList.AsVariableDeclarationList(); declList != nil && declList.Declarations != nil {
					for _, decl := range declList.Declarations.Nodes {
						if decl.Kind == ast.KindVariableDeclaration {
							if varDecl := decl.AsVariableDeclaration(); varDecl != nil && varDecl.Name() != nil {
								if ident := varDecl.Name().AsIdentifier(); ident != nil && ident.Text == name {
									return true
								}
							}
						}
					}
				}
			}
		}
	}

	return false
}

// isShadowedArguments checks if 'arguments' is shadowed by a parameter or local variable
func isShadowedArguments(node *ast.Node) bool {
	// Walk up to find the enclosing function
	current := node.Parent
	for current != nil {
		kind := current.Kind

		// When we hit a function, check if it has a parameter named 'arguments'
		// or a variable declaration named 'arguments'
		if kind == ast.KindFunctionDeclaration ||
			kind == ast.KindFunctionExpression ||
			kind == ast.KindArrowFunction ||
			kind == ast.KindMethodDeclaration ||
			kind == ast.KindConstructor {

			// Check function parameters
			var params []*ast.Node
			var body *ast.Node
			switch kind {
			case ast.KindFunctionDeclaration:
				if fn := current.AsFunctionDeclaration(); fn != nil {
					if fn.Parameters != nil {
						params = fn.Parameters.Nodes
					}
					body = fn.Body
				}
			case ast.KindFunctionExpression:
				if fn := current.AsFunctionExpression(); fn != nil {
					if fn.Parameters != nil {
						params = fn.Parameters.Nodes
					}
					body = fn.Body
				}
			case ast.KindArrowFunction:
				if fn := current.AsArrowFunction(); fn != nil {
					if fn.Parameters != nil {
						params = fn.Parameters.Nodes
					}
					body = fn.Body
				}
			case ast.KindMethodDeclaration:
				if fn := current.AsMethodDeclaration(); fn != nil {
					if fn.Parameters != nil {
						params = fn.Parameters.Nodes
					}
					body = fn.Body
				}
			case ast.KindConstructor:
				if fn := current.AsConstructorDeclaration(); fn != nil {
					if fn.Parameters != nil {
						params = fn.Parameters.Nodes
					}
					body = fn.Body
				}
			}

			// Check if any parameter is named 'arguments'
			for _, param := range params {
				if param.Kind == ast.KindParameter {
					if p := param.AsParameterDeclaration(); p != nil && p.Name() != nil {
						if ident := p.Name().AsIdentifier(); ident != nil && ident.Text == "arguments" {
							return true
						}
					}
				}
			}

			// Check if the function body has a variable declaration named 'arguments'
			if hasVariableNamed(body, "arguments") {
				return true
			}

			// Stop searching at function boundary
			break
		}

		current = current.Parent
	}
	return false
}

// isArgumentsPropertyAccess checks if this is accessing a safe property of arguments
func isArgumentsPropertyAccess(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}

	parent := node.Parent
	// Check if parent is a property access expression
	if parent.Kind == ast.KindPropertyAccessExpression {
		if propAccess := parent.AsPropertyAccessExpression(); propAccess != nil {
			if propAccess.Expression == node && propAccess.Name() != nil {
				propName := propAccess.Name().Text()
				// Allow .length and .callee
				return propName == "length" || propName == "callee"
			}
		}
	}

	return false
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindIdentifier: func(node *ast.Node) {
			ident := node.AsIdentifier()
			if ident == nil || ident.Text != "arguments" {
				return
			}

			// Don't flag if this identifier is the name of a variable declaration
			// (i.e., we're declaring a variable named 'arguments', not referencing it)
			if node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclaration {
				if varDecl := node.Parent.AsVariableDeclaration(); varDecl != nil {
					if varDecl.Name() == node {
						// This is the declaration itself, not a reference
						return
					}
				}
			}

			// Don't flag if this identifier is a parameter name
			if node.Parent != nil && node.Parent.Kind == ast.KindParameter {
				if param := node.Parent.AsParameterDeclaration(); param != nil {
					if param.Name() == node {
						// This is the parameter name itself, not a reference
						return
					}
				}
			}

			// Don't flag if shadowed by parameter or variable
			if isShadowedArguments(node) {
				return
			}

			// Don't flag if not in a function scope (or in arrow function)
			if !isInFunctionScope(node) {
				return
			}

			// Don't flag if accessing .length or .callee properties
			if isArgumentsPropertyAccess(node) {
				return
			}

			// Report the violation
			// Note: Auto-fix is complex (requires modifying function signature),
			// so we only report without providing a fix
			ctx.ReportNode(node, buildMessage())
		},
	}
}

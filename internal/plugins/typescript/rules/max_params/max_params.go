package max_params

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// MaxParamsRule enforces a maximum number of parameters in function definitions
var MaxParamsRule = rule.CreateRule(rule.Rule{
	Name: "max-params",
	Run:  run,
})

// Options represents the configuration for the max-params rule
type Options struct {
	Max           *int  `json:"max"`
	Maximum       *int  `json:"maximum"`
	CountVoidThis *bool `json:"countVoidThis"`
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// Parse options
	opts := Options{}
	if options != nil {
		if err := rule.ParseOptions(options, &opts); err == nil {
			// Use parsed options
		}
	}

	// Determine the maximum number of parameters
	// Default is 3 according to ESLint's max-params rule
	maxParams := 3
	if opts.Max != nil {
		maxParams = *opts.Max
	} else if opts.Maximum != nil {
		maxParams = *opts.Maximum
	}

	// Determine if we should count "this: void" parameters
	countVoidThis := false
	if opts.CountVoidThis != nil {
		countVoidThis = *opts.CountVoidThis
	}

	// Helper function to check if a parameter is a "this" parameter
	isThisParam := func(node *ast.Node) bool {
		if node == nil || node.Kind != ast.KindParameter {
			return false
		}

		param := node.AsParameterDeclaration()
		if param == nil || param.Name() == nil {
			return false
		}

		// Check if parameter name is "this"
		name := param.Name()
		if name.Kind == ast.KindIdentifier {
			ident := name.AsIdentifier()
			if ident != nil && ident.Text() == "this" {
				return true
			}
		}

		return false
	}

	// Helper function to check if a parameter is a "this: void" parameter
	isVoidThisParam := func(node *ast.Node) bool {
		if !isThisParam(node) {
			return false
		}

		param := node.AsParameterDeclaration()
		if param == nil || param.Type == nil {
			return false
		}

		// Check if the type is void
		typeNode := param.Type
		if typeNode.Kind == ast.KindVoidKeyword {
			return true
		}

		return false
	}

	// Check function for parameter count
	checkMaxParams := func(node *ast.Node) {
		var params []*ast.Node

		// Get parameters based on node type
		switch node.Kind {
		case ast.KindFunctionDeclaration:
			funcDecl := node.AsFunctionDeclaration()
			if funcDecl != nil && funcDecl.Parameters != nil {
				params = funcDecl.Parameters.Nodes
			}
		case ast.KindFunctionExpression:
			funcExpr := node.AsFunctionExpression()
			if funcExpr != nil && funcExpr.Parameters != nil {
				params = funcExpr.Parameters.Nodes
			}
		case ast.KindArrowFunction:
			arrowFunc := node.AsArrowFunction()
			if arrowFunc != nil && arrowFunc.Parameters != nil {
				params = arrowFunc.Parameters.Nodes
			}
		case ast.KindMethodDeclaration:
			methodDecl := node.AsMethodDeclaration()
			if methodDecl != nil && methodDecl.Parameters != nil {
				params = methodDecl.Parameters.Nodes
			}
		case ast.KindConstructor:
			constructor := node.AsConstructorDeclaration()
			if constructor != nil && constructor.Parameters != nil {
				params = constructor.Parameters.Nodes
			}
		case ast.KindFunctionType:
			functionType := node.AsFunctionTypeNode()
			if functionType != nil && functionType.Parameters != nil {
				params = functionType.Parameters.Nodes
			}
		case ast.KindCallSignature:
			callSignature := node.AsCallSignatureDeclaration()
			if callSignature != nil && callSignature.Parameters != nil {
				params = callSignature.Parameters.Nodes
			}
		default:
			return
		}

		// Count parameters, excluding "this" parameters (or including "this: void" based on option)
		paramCount := 0
		for _, param := range params {
			if param == nil {
				continue
			}

			// Check if this is a "this" parameter
			if isThisParam(param) {
				// If it's "this: void", only count it if countVoidThis is true
				if isVoidThisParam(param) {
					if countVoidThis {
						paramCount++
					}
					// Otherwise skip it
					continue
				}
				// For non-void "this" parameters (e.g., "this: Foo"), always count them
				paramCount++
			} else {
				// Regular parameter
				paramCount++
			}
		}

		// Report if parameter count exceeds maximum
		if paramCount > maxParams {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "exceed",
				Description: fmt.Sprintf("Function has too many parameters (%d). Maximum allowed is %d.", paramCount, maxParams),
			})
		}
	}

	return rule.RuleListeners{
		ast.KindFunctionDeclaration: checkMaxParams,
		ast.KindFunctionExpression:  checkMaxParams,
		ast.KindArrowFunction:       checkMaxParams,
		ast.KindMethodDeclaration:   checkMaxParams,
		ast.KindConstructor:         checkMaxParams,
		ast.KindFunctionType:        checkMaxParams,
		ast.KindCallSignature:       checkMaxParams,
	}
}

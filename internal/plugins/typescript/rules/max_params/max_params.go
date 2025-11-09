package max_params

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type MaxParamsOptions struct {
	Max           int  `json:"max"`
	Maximum       int  `json:"maximum"`
	CountVoidThis bool `json:"countVoidThis"`
}

var MaxParamsRule = rule.CreateRule(rule.Rule{
	Name: "max-params",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Parse options with default max of 3
		opts := MaxParamsOptions{
			Max:           3,
			Maximum:       0,
			CountVoidThis: false,
		}

		if options != nil {
			var optsMap map[string]interface{}
			if optsArray, ok := options.([]interface{}); ok && len(optsArray) > 0 {
				if opts, ok := optsArray[0].(map[string]interface{}); ok {
					optsMap = opts
				}
			} else if opts, ok := options.(map[string]interface{}); ok {
				optsMap = opts
			}

			if optsMap != nil {
				if max, ok := optsMap["max"].(float64); ok {
					opts.Max = int(max)
				} else if max, ok := optsMap["max"].(int); ok {
					opts.Max = max
				}
				if maximum, ok := optsMap["maximum"].(float64); ok {
					opts.Maximum = int(maximum)
				} else if maximum, ok := optsMap["maximum"].(int); ok {
					opts.Maximum = maximum
				}
				if countVoidThis, ok := optsMap["countVoidThis"].(bool); ok {
					opts.CountVoidThis = countVoidThis
				}
			}
		}

		// Use maximum if it's set, otherwise use max
		maxParams := opts.Max
		if opts.Maximum > 0 {
			maxParams = opts.Maximum
		}

		// Helper to check if a parameter is a void this parameter
		isVoidThisParam := func(param *ast.Node) bool {
			if param.Kind != ast.KindParameter {
				return false
			}
			p := param.AsParameter()
			if p == nil || p.Name() == nil {
				return false
			}

			// Check if the parameter name is "this"
			if p.Name().Kind == ast.KindIdentifier {
				ident := p.Name().AsIdentifier()
				if ident != nil && ident.Text == "this" {
					// Check if it has a void type annotation
					if p.Type != nil && p.Type.Kind == ast.KindVoidKeyword {
						return true
					}
				}
			}

			return false
		}

		// Helper to count parameters
		countParams := func(params []*ast.Node) int {
			count := 0
			for _, param := range params {
				// If not counting void this, skip void this parameters
				if !opts.CountVoidThis && isVoidThisParam(param) {
					continue
				}
				count++
			}
			return count
		}

		// Helper to get parameters from different function types
		getParameters := func(node *ast.Node) []*ast.Node {
			switch node.Kind {
			case ast.KindFunctionDeclaration:
				fn := node.AsFunctionDeclaration()
				if fn != nil && fn.Parameters != nil {
					return fn.Parameters.Nodes
				}
			case ast.KindFunctionExpression:
				fn := node.AsFunctionExpression()
				if fn != nil && fn.Parameters != nil {
					return fn.Parameters.Nodes
				}
			case ast.KindArrowFunction:
				fn := node.AsArrowFunction()
				if fn != nil && fn.Parameters != nil {
					return fn.Parameters.Nodes
				}
			case ast.KindMethodDeclaration:
				method := node.AsMethodDeclaration()
				if method != nil && method.Parameters != nil {
					return method.Parameters.Nodes
				}
			case ast.KindConstructor:
				constructor := node.AsConstructorDeclaration()
				if constructor != nil && constructor.Parameters != nil {
					return constructor.Parameters.Nodes
				}
			case ast.KindGetAccessor:
				accessor := node.AsGetAccessorDeclaration()
				if accessor != nil && accessor.Parameters != nil {
					return accessor.Parameters.Nodes
				}
			case ast.KindSetAccessor:
				accessor := node.AsSetAccessorDeclaration()
				if accessor != nil && accessor.Parameters != nil {
					return accessor.Parameters.Nodes
				}
			case ast.KindFunctionType:
				fnType := node.AsFunctionTypeNode()
				if fnType != nil && fnType.Parameters != nil {
					return fnType.Parameters.Nodes
				}
			case ast.KindCallSignature:
				sig := node.AsCallSignatureDeclaration()
				if sig != nil && sig.Parameters != nil {
					return sig.Parameters.Nodes
				}
			case ast.KindConstructSignature:
				sig := node.AsConstructSignatureDeclaration()
				if sig != nil && sig.Parameters != nil {
					return sig.Parameters.Nodes
				}
			case ast.KindMethodSignature:
				sig := node.AsMethodSignature()
				if sig != nil && sig.Parameters != nil {
					return sig.Parameters.Nodes
				}
			}
			return nil
		}

		// Main check function
		checkFunction := func(node *ast.Node) {
			params := getParameters(node)
			if params == nil {
				return
			}

			numParams := countParams(params)
			if numParams > maxParams {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "exceed",
					Description: fmt.Sprintf("Function has too many parameters (%d). Maximum allowed is %d.", numParams, maxParams),
				})
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration:      checkFunction,
			ast.KindFunctionExpression:       checkFunction,
			ast.KindArrowFunction:            checkFunction,
			ast.KindMethodDeclaration:        checkFunction,
			ast.KindConstructor:              checkFunction,
			ast.KindGetAccessor:              checkFunction,
			ast.KindSetAccessor:              checkFunction,
			ast.KindFunctionType:             checkFunction,
			ast.KindCallSignature:            checkFunction,
			ast.KindConstructSignature:       checkFunction,
			ast.KindMethodSignature:          checkFunction,
		}
	},
})

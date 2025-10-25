package no_dupe_args

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoDupeArgsRule implements the no-dupe-args rule
// Disallow duplicate arguments in function definitions
var NoDupeArgsRule = rule.Rule{
	Name: "no-dupe-args",
	Run:  run,
}

func buildDuplicateMessage(paramName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "duplicateParam",
		Description: "Duplicate param '" + paramName + "'.",
	}
}

// extractParamName extracts the identifier name from a parameter node
func extractParamName(ctx rule.RuleContext, param *ast.Node) string {
	if param == nil {
		return ""
	}

	// Handle simple identifier parameters
	if param.Kind == ast.KindIdentifier {
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, param)
		return ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
	}

	// Handle binding patterns (destructuring) - skip these as they can't have simple duplicates
	if param.Kind == ast.KindObjectBindingPattern || param.Kind == ast.KindArrayBindingPattern {
		return ""
	}

	// Handle parameter declarations with name property
	if param.Kind == ast.KindParameter {
		p := param.AsParameterDeclaration()
		if p != nil && p.Name() != nil {
			return extractParamName(ctx, p.Name())
		}
	}

	return ""
}

// checkDuplicateParams checks for duplicate parameter names in a function
func checkDuplicateParams(ctx rule.RuleContext, params []*ast.Node) {
	if params == nil {
		return
	}

	seen := make(map[string]*ast.Node)

	for _, param := range params {
		if param == nil {
			continue
		}

		name := extractParamName(ctx, param)
		if name == "" {
			// Skip empty names (destructuring patterns, etc.)
			continue
		}

		if firstOccurrence, exists := seen[name]; exists {
			// Report error on the duplicate parameter
			ctx.ReportNode(param, buildDuplicateMessage(name))
			// Also mark the first occurrence for consistency
			_ = firstOccurrence
		} else {
			seen[name] = param
		}
	}
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			funcDecl := node.AsFunctionDeclaration()
			if funcDecl != nil && funcDecl.Parameters != nil && len(funcDecl.Parameters.Nodes) > 0 {
				checkDuplicateParams(ctx, funcDecl.Parameters.Nodes)
			}
		},
		ast.KindFunctionExpression: func(node *ast.Node) {
			funcExpr := node.AsFunctionExpression()
			if funcExpr != nil && funcExpr.Parameters != nil && len(funcExpr.Parameters.Nodes) > 0 {
				checkDuplicateParams(ctx, funcExpr.Parameters.Nodes)
			}
		},
		ast.KindArrowFunction: func(node *ast.Node) {
			arrowFunc := node.AsArrowFunction()
			if arrowFunc != nil && arrowFunc.Parameters != nil && len(arrowFunc.Parameters.Nodes) > 0 {
				checkDuplicateParams(ctx, arrowFunc.Parameters.Nodes)
			}
		},
	}
}

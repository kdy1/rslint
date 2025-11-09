package no_unnecessary_type_parameters

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildNotUsedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "sole",
		Description: "Type parameter appears only once in function signature.",
	}
}

// countTypeParameterUsage counts how many times a type parameter is used in a node tree
func countTypeParameterUsage(node *ast.Node, typeParamName string) int {
	if node == nil {
		return 0
	}

	count := 0

	// Check if this node is a reference to the type parameter
	if node.Kind == ast.KindTypeReference {
		typeRef := node.AsTypeReference()
		if typeRef != nil && ast.IsIdentifier(typeRef.TypeName) {
			identifier := typeRef.TypeName.AsIdentifier()
			if identifier != nil && identifier.Text == typeParamName {
				count++
			}
		}
	}

	// Recursively check children
	ast.ForEachChild(node, func(child *ast.Node) {
		count += countTypeParameterUsage(child, typeParamName)
	})

	return count
}

// getTypeParameterName extracts the name from a type parameter node
func getTypeParameterName(typeParam *ast.Node) string {
	if typeParam == nil || typeParam.Kind != ast.KindTypeParameter {
		return ""
	}

	tp := typeParam.AsTypeParameterDeclaration()
	if tp == nil || tp.Name == nil {
		return ""
	}

	if ast.IsIdentifier(tp.Name) {
		identifier := tp.Name.AsIdentifier()
		if identifier != nil {
			return identifier.Text
		}
	}

	return ""
}

// checkTypeParameters checks if type parameters are used more than once
func checkTypeParameters(ctx rule.RuleContext, node *ast.Node, typeParameters *ast.NodeArray) {
	if typeParameters == nil || typeParameters.Len() == 0 {
		return
	}

	// For each type parameter, count how many times it's used in the signature
	for i := 0; i < typeParameters.Len(); i++ {
		typeParam := typeParameters.Item(i)
		typeParamName := getTypeParameterName(typeParam)

		if typeParamName == "" {
			continue
		}

		// Count uses in the entire node (parameters, return type, etc.)
		// We exclude the type parameter declaration itself from counting
		usageCount := 0

		// Count in parameters
		if node.Kind == ast.KindFunctionDeclaration ||
		   node.Kind == ast.KindFunctionExpression ||
		   node.Kind == ast.KindArrowFunction ||
		   node.Kind == ast.KindMethodDeclaration {

			var parameters *ast.NodeArray
			var returnType *ast.Node

			switch node.Kind {
			case ast.KindFunctionDeclaration:
				fn := node.AsFunctionDeclaration()
				if fn != nil {
					parameters = fn.Parameters
					returnType = fn.Type
				}
			case ast.KindFunctionExpression:
				fn := node.AsFunctionExpression()
				if fn != nil {
					parameters = fn.Parameters
					returnType = fn.Type
				}
			case ast.KindArrowFunction:
				fn := node.AsArrowFunction()
				if fn != nil {
					parameters = fn.Parameters
					returnType = fn.Type
				}
			case ast.KindMethodDeclaration:
				fn := node.AsMethodDeclaration()
				if fn != nil {
					parameters = fn.Parameters
					returnType = fn.Type
				}
			}

			// Count in parameters
			if parameters != nil {
				for j := 0; j < parameters.Len(); j++ {
					param := parameters.Item(j)
					usageCount += countTypeParameterUsage(param, typeParamName)
				}
			}

			// Count in return type
			if returnType != nil {
				usageCount += countTypeParameterUsage(returnType, typeParamName)
			}
		} else if node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression {
			// For classes, count usage in heritage clauses and members
			var heritageClauses *ast.NodeArray
			var members *ast.NodeArray

			if node.Kind == ast.KindClassDeclaration {
				cls := node.AsClassDeclaration()
				if cls != nil {
					heritageClauses = cls.HeritageClauses
					members = cls.Members
				}
			} else {
				cls := node.AsClassExpression()
				if cls != nil {
					heritageClauses = cls.HeritageClauses
					members = cls.Members
				}
			}

			if heritageClauses != nil {
				for j := 0; j < heritageClauses.Len(); j++ {
					clause := heritageClauses.Item(j)
					usageCount += countTypeParameterUsage(clause, typeParamName)
				}
			}

			if members != nil {
				for j := 0; j < members.Len(); j++ {
					member := members.Item(j)
					usageCount += countTypeParameterUsage(member, typeParamName)
				}
			}
		}

		// If type parameter is used 0 or 1 times, it's unnecessary
		if usageCount <= 1 {
			ctx.ReportNode(typeParam, buildNotUsedMessage())
		}
	}
}

// checkFunctionLike checks function-like nodes for unnecessary type parameters
func checkFunctionLike(ctx rule.RuleContext, node *ast.Node) {
	var typeParameters *ast.NodeArray

	switch node.Kind {
	case ast.KindFunctionDeclaration:
		fn := node.AsFunctionDeclaration()
		if fn != nil {
			typeParameters = fn.TypeParameters
		}
	case ast.KindFunctionExpression:
		fn := node.AsFunctionExpression()
		if fn != nil {
			typeParameters = fn.TypeParameters
		}
	case ast.KindArrowFunction:
		fn := node.AsArrowFunction()
		if fn != nil {
			typeParameters = fn.TypeParameters
		}
	case ast.KindMethodDeclaration:
		fn := node.AsMethodDeclaration()
		if fn != nil {
			typeParameters = fn.TypeParameters
		}
	}

	checkTypeParameters(ctx, node, typeParameters)
}

// checkClass checks class declarations for unnecessary type parameters
func checkClass(ctx rule.RuleContext, node *ast.Node) {
	var typeParameters *ast.NodeArray

	if node.Kind == ast.KindClassDeclaration {
		cls := node.AsClassDeclaration()
		if cls != nil {
			typeParameters = cls.TypeParameters
		}
	} else if node.Kind == ast.KindClassExpression {
		cls := node.AsClassExpression()
		if cls != nil {
			typeParameters = cls.TypeParameters
		}
	}

	checkTypeParameters(ctx, node, typeParameters)
}

var NoUnnecessaryTypeParametersRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-type-parameters",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				checkFunctionLike(ctx, node)
			},
			ast.KindFunctionExpression: func(node *ast.Node) {
				checkFunctionLike(ctx, node)
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				checkFunctionLike(ctx, node)
			},
			ast.KindMethodDeclaration: func(node *ast.Node) {
				checkFunctionLike(ctx, node)
			},
			ast.KindClassDeclaration: func(node *ast.Node) {
				checkClass(ctx, node)
			},
			ast.KindClassExpression: func(node *ast.Node) {
				checkClass(ctx, node)
			},
		}
	},
})

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

// hasTypeParameter checks if a type parameter is referenced anywhere in a node tree
func hasTypeParameter(node *ast.Node, typeParamName string) bool {
	if node == nil {
		return false
	}

	// Check if this node is a direct reference to the type parameter
	if node.Kind == ast.KindTypeReference {
		typeRef := node.AsTypeReference()
		if typeRef != nil && ast.IsIdentifier(typeRef.TypeName) {
			identifier := typeRef.TypeName.AsIdentifier()
			if identifier != nil && identifier.Text == typeParamName {
				return true
			}
		}

		// Also check type arguments
		if typeRef != nil && typeRef.TypeArguments != nil {
			for _, arg := range typeRef.TypeArguments.Nodes {
				if hasTypeParameter(arg, typeParamName) {
					return true
				}
			}
		}
	}

	// Check in union types (e.g., T | null)
	if node.Kind == ast.KindUnionType {
		unionType := node.AsUnionTypeNode()
		if unionType != nil && unionType.Types != nil {
			for _, t := range unionType.Types.Nodes {
				if hasTypeParameter(t, typeParamName) {
					return true
				}
			}
		}
	}

	// Check in intersection types
	if node.Kind == ast.KindIntersectionType {
		intersectionType := node.AsIntersectionTypeNode()
		if intersectionType != nil && intersectionType.Types != nil {
			for _, t := range intersectionType.Types.Nodes {
				if hasTypeParameter(t, typeParamName) {
					return true
				}
			}
		}
	}

	// Check in array types
	if node.Kind == ast.KindArrayType {
		arrayType := node.AsArrayTypeNode()
		if arrayType != nil && arrayType.ElementType != nil {
			return hasTypeParameter(arrayType.ElementType, typeParamName)
		}
	}

	// Check in parenthesized types
	if node.Kind == ast.KindParenthesizedType {
		parenType := node.AsParenthesizedTypeNode()
		if parenType != nil && parenType.Type != nil {
			return hasTypeParameter(parenType.Type, typeParamName)
		}
	}

	// Check in tuple types
	if node.Kind == ast.KindTupleType {
		tupleType := node.AsTupleTypeNode()
		if tupleType != nil && tupleType.Elements != nil {
			for _, elem := range tupleType.Elements.Nodes {
				if hasTypeParameter(elem, typeParamName) {
					return true
				}
			}
		}
	}

	return false
}

// countTypeParameterUsage counts how many times a type parameter is used in a node tree
// Returns a count where usage as a type argument counts as 2+ (making it valid)
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

	// Recursively check different node types
	switch node.Kind {
	case ast.KindTypeReference:
		typeRef := node.AsTypeReference()
		if typeRef != nil {
			// Check type arguments
			// If a type parameter is used as a type argument to another generic type,
			// it's considered "reused" and should be counted as 2+ (making it valid)
			// This handles cases like Promise<T>, Map<K, V>, Array<T>, etc.
			if typeRef.TypeArguments != nil {
				for _, arg := range typeRef.TypeArguments.Nodes {
					if hasTypeParameter(arg, typeParamName) {
						// Type parameter found in type arguments - count as valid usage
						count += 2
					}
				}
			}
		}

	case ast.KindArrayType:
		arrayType := node.AsArrayTypeNode()
		if arrayType != nil && arrayType.ElementType != nil {
			count += countTypeParameterUsage(arrayType.ElementType, typeParamName)
		}

	case ast.KindUnionType:
		unionType := node.AsUnionTypeNode()
		if unionType != nil && unionType.Types != nil {
			for _, t := range unionType.Types.Nodes {
				count += countTypeParameterUsage(t, typeParamName)
			}
		}

	case ast.KindIntersectionType:
		intersectionType := node.AsIntersectionTypeNode()
		if intersectionType != nil && intersectionType.Types != nil {
			for _, t := range intersectionType.Types.Nodes {
				count += countTypeParameterUsage(t, typeParamName)
			}
		}

	case ast.KindParenthesizedType:
		parenType := node.AsParenthesizedTypeNode()
		if parenType != nil && parenType.Type != nil {
			count += countTypeParameterUsage(parenType.Type, typeParamName)
		}

	case ast.KindTupleType:
		tupleType := node.AsTupleTypeNode()
		if tupleType != nil && tupleType.Elements != nil {
			for _, elem := range tupleType.Elements.Nodes {
				count += countTypeParameterUsage(elem, typeParamName)
			}
		}

	case ast.KindTypeLiteral:
		typeLiteral := node.AsTypeLiteralNode()
		if typeLiteral != nil && typeLiteral.Members != nil {
			for _, member := range typeLiteral.Members.Nodes {
				count += countTypeParameterUsageInMember(member, typeParamName)
			}
		}

	case ast.KindFunctionType:
		funcType := node.AsFunctionTypeNode()
		if funcType != nil {
			if funcType.Type != nil {
				count += countTypeParameterUsage(funcType.Type, typeParamName)
			}
			if funcType.Parameters != nil {
				for _, param := range funcType.Parameters.Nodes {
					count += countTypeParameterUsageInParameter(param, typeParamName)
				}
			}
		}

	case ast.KindParameter:
		count += countTypeParameterUsageInParameter(node, typeParamName)

	case ast.KindTypeOperator:
		// Handle keyof T, readonly T, etc.
		typeOperator := node.AsTypeOperatorNode()
		if typeOperator != nil && typeOperator.Type != nil {
			count += countTypeParameterUsage(typeOperator.Type, typeParamName)
		}

	case ast.KindIndexedAccessType:
		// Handle T[K] type
		indexedAccess := node.AsIndexedAccessTypeNode()
		if indexedAccess != nil {
			if indexedAccess.ObjectType != nil {
				count += countTypeParameterUsage(indexedAccess.ObjectType, typeParamName)
			}
			if indexedAccess.IndexType != nil {
				count += countTypeParameterUsage(indexedAccess.IndexType, typeParamName)
			}
		}

	case ast.KindConditionalType:
		// Handle T extends U ? X : Y
		conditional := node.AsConditionalTypeNode()
		if conditional != nil {
			if conditional.CheckType != nil {
				count += countTypeParameterUsage(conditional.CheckType, typeParamName)
			}
			if conditional.ExtendsType != nil {
				count += countTypeParameterUsage(conditional.ExtendsType, typeParamName)
			}
			if conditional.TrueType != nil {
				count += countTypeParameterUsage(conditional.TrueType, typeParamName)
			}
			if conditional.FalseType != nil {
				count += countTypeParameterUsage(conditional.FalseType, typeParamName)
			}
		}

	case ast.KindMappedType:
		// Handle { [K in keyof T]: ... }
		mapped := node.AsMappedTypeNode()
		if mapped != nil {
			if mapped.Type != nil {
				count += countTypeParameterUsage(mapped.Type, typeParamName)
			}
			if mapped.NameType != nil {
				count += countTypeParameterUsage(mapped.NameType, typeParamName)
			}
		}

	case ast.KindTypePredicate:
		// Handle type predicates like "x is T"
		predicate := node.AsTypePredicateNode()
		if predicate != nil && predicate.Type != nil {
			count += countTypeParameterUsage(predicate.Type, typeParamName)
		}
	}

	return count
}

// countTypeParameterUsageInParameter counts usages in a parameter node
func countTypeParameterUsageInParameter(param *ast.Node, typeParamName string) int {
	if param == nil {
		return 0
	}

	count := 0
	paramDecl := param.AsParameterDeclaration()
	if paramDecl != nil && paramDecl.Type != nil {
		count += countTypeParameterUsage(paramDecl.Type, typeParamName)
	}

	return count
}

// countTypeParameterUsageInMember counts usages in a type member
func countTypeParameterUsageInMember(member *ast.Node, typeParamName string) int {
	if member == nil {
		return 0
	}

	count := 0
	switch member.Kind {
	case ast.KindPropertySignature:
		propSig := member.AsPropertySignatureDeclaration()
		if propSig != nil && propSig.Type != nil {
			count += countTypeParameterUsage(propSig.Type, typeParamName)
		}

	case ast.KindMethodSignature:
		methodSig := member.AsMethodSignatureDeclaration()
		if methodSig != nil {
			if methodSig.Type != nil {
				count += countTypeParameterUsage(methodSig.Type, typeParamName)
			}
			if methodSig.Parameters != nil {
				for _, param := range methodSig.Parameters.Nodes {
					count += countTypeParameterUsageInParameter(param, typeParamName)
				}
			}
		}

	case ast.KindIndexSignature:
		indexSig := member.AsIndexSignatureDeclaration()
		if indexSig != nil && indexSig.Type != nil {
			count += countTypeParameterUsage(indexSig.Type, typeParamName)
		}
	}

	return count
}

// getTypeParameterName extracts the name from a type parameter node
func getTypeParameterName(typeParam *ast.Node) string {
	if typeParam == nil || typeParam.Kind != ast.KindTypeParameter {
		return ""
	}

	// Type parameters have a Name() method
	name := typeParam.Name()
	if name == nil {
		return ""
	}

	if ast.IsIdentifier(name) {
		identifier := name.AsIdentifier()
		if identifier != nil {
			return identifier.Text
		}
	}

	return ""
}

// checkTypeParameters checks if type parameters are used more than once
func checkTypeParameters(ctx rule.RuleContext, node *ast.Node, typeParameters *ast.NodeList) {
	if typeParameters == nil || len(typeParameters.Nodes) == 0 {
		return
	}

	// For each type parameter, count how many times it's used in the signature
	for _, typeParam := range typeParameters.Nodes {
		typeParamName := getTypeParameterName(typeParam)

		if typeParamName == "" {
			continue
		}

		// Count uses in the entire node (parameters, return type, etc.)
		// We exclude the type parameter declaration itself from counting
		usageCount := 0

		// Also count usage in constraints of OTHER type parameters
		// For example, in <T, K extends keyof T>, T is used in K's constraint
		for _, otherParam := range typeParameters.Nodes {
			if otherParam == typeParam {
				continue // Skip the parameter itself
			}
			// Check if this type parameter is used in another parameter's constraint
			if otherParam.Kind == ast.KindTypeParameter {
				tp := otherParam.AsTypeParameter()
				if tp != nil && tp.Constraint != nil {
					usageCount += countTypeParameterUsage(tp.Constraint, typeParamName)
				}
			}
		}

		// Count in parameters
		if node.Kind == ast.KindFunctionDeclaration ||
			node.Kind == ast.KindFunctionExpression ||
			node.Kind == ast.KindArrowFunction ||
			node.Kind == ast.KindMethodDeclaration {

			var parameters *ast.NodeList
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
				for _, param := range parameters.Nodes {
					usageCount += countTypeParameterUsageInParameter(param, typeParamName)
				}
			}

			// Count in return type
			if returnType != nil {
				usageCount += countTypeParameterUsage(returnType, typeParamName)
			}
		} else if node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression {
			// For classes, count usage in heritage clauses and members
			var heritageClauses *ast.NodeList
			var members *ast.NodeList

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
				for _, clause := range heritageClauses.Nodes {
					heritageClause := clause.AsHeritageClause()
					if heritageClause != nil && heritageClause.Types != nil {
						for _, typeExpr := range heritageClause.Types.Nodes {
							usageCount += countTypeParameterUsage(typeExpr, typeParamName)
						}
					}
				}
			}

			if members != nil {
				for _, member := range members.Nodes {
					usageCount += countTypeParameterUsageInClassMember(member, typeParamName)
				}
			}
		}

		// If type parameter is used 0 or 1 times, it's unnecessary
		if usageCount <= 1 {
			ctx.ReportNode(typeParam, buildNotUsedMessage())
		}
	}
}

// countTypeParameterUsageInClassMember counts usages in a class member
func countTypeParameterUsageInClassMember(member *ast.Node, typeParamName string) int {
	if member == nil {
		return 0
	}

	count := 0
	switch member.Kind {
	case ast.KindPropertyDeclaration:
		propDecl := member.AsPropertyDeclaration()
		if propDecl != nil && propDecl.Type != nil {
			count += countTypeParameterUsage(propDecl.Type, typeParamName)
		}

	case ast.KindMethodDeclaration:
		methodDecl := member.AsMethodDeclaration()
		if methodDecl != nil {
			if methodDecl.Type != nil {
				count += countTypeParameterUsage(methodDecl.Type, typeParamName)
			}
			if methodDecl.Parameters != nil {
				for _, param := range methodDecl.Parameters.Nodes {
					count += countTypeParameterUsageInParameter(param, typeParamName)
				}
			}
		}

	case ast.KindConstructor:
		constructor := member.AsConstructorDeclaration()
		if constructor != nil && constructor.Parameters != nil {
			for _, param := range constructor.Parameters.Nodes {
				count += countTypeParameterUsageInParameter(param, typeParamName)
			}
		}

	case ast.KindGetAccessor:
		getter := member.AsGetAccessorDeclaration()
		if getter != nil && getter.Type != nil {
			count += countTypeParameterUsage(getter.Type, typeParamName)
		}

	case ast.KindSetAccessor:
		setter := member.AsSetAccessorDeclaration()
		if setter != nil && setter.Parameters != nil {
			for _, param := range setter.Parameters.Nodes {
				count += countTypeParameterUsageInParameter(param, typeParamName)
			}
		}
	}

	return count
}

// checkFunctionLike checks function-like nodes for unnecessary type parameters
func checkFunctionLike(ctx rule.RuleContext, node *ast.Node) {
	var typeParameters *ast.NodeList

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
	var typeParameters *ast.NodeList

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

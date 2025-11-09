package no_misused_new

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var NoMisusedNewRule = rule.CreateRule(rule.Rule{
	Name: "no-misused-new",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			// Check for 'constructor' method in interfaces and type literals
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				interfaceDecl := node.AsInterfaceDeclaration()
				if interfaceDecl == nil {
					return
				}

				checkInterfaceMembers(ctx, interfaceDecl.Name(), interfaceDecl.Members)
			},
			ast.KindTypeLiteral: func(node *ast.Node) {
				typeLiteral := node.AsTypeLiteralNode()
				if typeLiteral == nil {
					return
				}

				checkInterfaceMembers(ctx, nil, typeLiteral.Members)
			},
			// Check for 'new' method signature in classes
			ast.KindClassDeclaration: func(node *ast.Node) {
				classDecl := node.AsClassDeclaration()
				if classDecl == nil {
					return
				}

				checkClassMembers(ctx, classDecl)
			},
			ast.KindClassExpression: func(node *ast.Node) {
				classExpr := node.AsClassExpression()
				if classExpr == nil {
					return
				}

				checkClassMembers(ctx, classExpr)
			},
		}
	},
})

// checkInterfaceMembers checks for misused 'new' or 'constructor' in interfaces/type literals
func checkInterfaceMembers(ctx rule.RuleContext, interfaceName *ast.Node, members *ast.NodeArray) {
	if members == nil {
		return
	}

	for _, member := range members.Nodes {
		if member.Kind != ast.KindMethodSignature && member.Kind != ast.KindConstructSignature {
			continue
		}

		if member.Kind == ast.KindConstructSignature {
			// Check if it's a constructor signature in an interface
			constructSig := member.AsConstructSignatureDeclaration()
			if constructSig == nil {
				continue
			}

			// Check if the return type matches the interface name
			if shouldReportConstructorInInterface(ctx, constructSig, interfaceName) {
				ctx.ReportNode(member, rule.RuleMessage{
					Id:          "errorMessageInterface",
					Description: "Interfaces must not contain a `constructor` method. Did you mean `new`?",
				})
			}
			continue
		}

		// Check for method signature named 'constructor' or 'new'
		methodSig := member.AsMethodSignature()
		if methodSig == nil {
			continue
		}

		if methodSig.Name == nil {
			continue
		}

		var methodName string
		if methodSig.Name.Kind == ast.KindIdentifier {
			identifier := methodSig.Name.AsIdentifier()
			if identifier != nil {
				methodName = identifier.EscapedText
			}
		}

		if methodName == "constructor" {
			// 'constructor' method in interface/type literal is always wrong
			ctx.ReportNode(member, rule.RuleMessage{
				Id:          "errorMessageInterface",
				Description: "Interfaces must not contain a `constructor` method. Did you mean `new`?",
			})
		} else if methodName == "new" {
			// 'new' method in interface is wrong if it returns the same interface type
			if shouldReportNewInInterface(ctx, methodSig, interfaceName) {
				ctx.ReportNode(member, rule.RuleMessage{
					Id:          "errorMessageInterface",
					Description: "Interfaces must not contain a `new` method. Did you mean `constructor`?",
				})
			}
		}
	}
}

// checkClassMembers checks for misused 'new' method signature in classes
func checkClassMembers(ctx rule.RuleContext, classNode ast.ClassLikeDeclaration) {
	members := classNode.Members()
	if members == nil {
		return
	}

	className := getClassName(classNode)

	for _, member := range members.Nodes {
		if member.Kind != ast.KindMethodDeclaration {
			continue
		}

		methodDecl := member.AsMethodDeclaration()
		if methodDecl == nil {
			continue
		}

		// Skip methods with body (only check signatures)
		if methodDecl.Body != nil {
			continue
		}

		if methodDecl.Name == nil {
			continue
		}

		var methodName string
		if methodDecl.Name.Kind == ast.KindIdentifier {
			identifier := methodDecl.Name.AsIdentifier()
			if identifier != nil {
				methodName = identifier.EscapedText
			}
		}

		if methodName == "new" {
			// Check if the return type matches the class name
			if shouldReportNewInClass(ctx, methodDecl, className) {
				ctx.ReportNode(member, rule.RuleMessage{
					Id:          "errorMessageClass",
					Description: "Classes must not contain a `new` method. Did you mean `constructor`?",
				})
			}
		}
	}
}

// shouldReportConstructorInInterface checks if a constructor signature should be reported
func shouldReportConstructorInInterface(ctx rule.RuleContext, constructSig *ast.ConstructSignatureDeclaration, interfaceName *ast.Node) bool {
	// Constructor signatures in interfaces are always wrong
	return true
}

// shouldReportNewInInterface checks if a 'new' method should be reported in an interface
func shouldReportNewInInterface(ctx rule.RuleContext, methodSig *ast.MethodSignature, interfaceName *ast.Node) bool {
	// If we don't have an interface name (type literal), we can't check the return type
	if interfaceName == nil {
		return false
	}

	var interfaceNameText string
	if interfaceName.Kind == ast.KindIdentifier {
		identifier := interfaceName.AsIdentifier()
		if identifier != nil {
			interfaceNameText = identifier.EscapedText
		}
	}

	if interfaceNameText == "" {
		return false
	}

	// Check the return type
	returnType := methodSig.Type
	if returnType == nil {
		return false
	}

	// Check if return type is a type reference to the interface itself
	if returnType.Kind == ast.KindTypeReference {
		typeRef := returnType.AsTypeReferenceNode()
		if typeRef == nil {
			return false
		}

		// Get the type name
		typeName := getTypeReferenceName(typeRef)

		// Check if it matches the interface name
		if typeName == interfaceNameText {
			return true
		}

		// Also check if it's a generic version (e.g., G<T> for interface G)
		// The base name should still match
		return typeName == interfaceNameText
	}

	return false
}

// shouldReportNewInClass checks if a 'new' method should be reported in a class
func shouldReportNewInClass(ctx rule.RuleContext, methodDecl *ast.MethodDeclaration, className string) bool {
	if className == "" {
		return false
	}

	// Check the return type
	returnType := methodDecl.Type
	if returnType == nil {
		return false
	}

	// Check if return type is a type reference to the class itself
	if returnType.Kind == ast.KindTypeReference {
		typeRef := returnType.AsTypeReferenceNode()
		if typeRef == nil {
			return false
		}

		// Get the type name
		typeName := getTypeReferenceName(typeRef)

		// Check if it matches the class name
		return typeName == className
	}

	return false
}

// getClassName extracts the class name from a class declaration or expression
func getClassName(classNode ast.ClassLikeDeclaration) string {
	var name *ast.Node

	if classDecl, ok := classNode.(*ast.ClassDeclaration); ok {
		name = classDecl.Name
	} else if classExpr, ok := classNode.(*ast.ClassExpression); ok {
		name = classExpr.Name
	}

	if name == nil {
		return ""
	}

	if name.Kind == ast.KindIdentifier {
		identifier := name.AsIdentifier()
		if identifier != nil {
			return identifier.EscapedText
		}
	}

	return ""
}

// getTypeReferenceName extracts the type name from a type reference
func getTypeReferenceName(typeRef *ast.TypeReferenceNode) string {
	if typeRef.TypeName == nil {
		return ""
	}

	typeName := typeRef.TypeName

	if typeName.Kind == ast.KindIdentifier {
		identifier := typeName.AsIdentifier()
		if identifier != nil {
			return identifier.EscapedText
		}
	}

	// Handle qualified names (e.g., namespace.Type)
	if typeName.Kind == ast.KindQualifiedName {
		qualifiedName := typeName.AsQualifiedName()
		if qualifiedName != nil && qualifiedName.Right != nil {
			if qualifiedName.Right.Kind == ast.KindIdentifier {
				identifier := qualifiedName.Right.AsIdentifier()
				if identifier != nil {
					return identifier.EscapedText
				}
			}
		}
	}

	return ""
}

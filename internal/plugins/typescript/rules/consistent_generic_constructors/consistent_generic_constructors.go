package consistent_generic_constructors

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ConsistentGenericConstructorsOptions struct {
	Style string `json:"style"`
}

// ConsistentGenericConstructorsRule enforces consistent generic specifier style in constructor signatures
var ConsistentGenericConstructorsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-generic-constructors",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := ConsistentGenericConstructorsOptions{
		Style: "constructor", // default
	}

	// Parse options
	if options != nil {
		// Handle array format: ["type-annotation"]
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if style, ok := optArray[0].(string); ok {
				opts.Style = style
			}
		} else if optsMap, ok := options.(map[string]interface{}); ok {
			if style, exists := optsMap["style"].(string); exists {
				opts.Style = style
			}
		} else if style, ok := options.(string); ok {
			opts.Style = style
		}
	}

	checkNode := func(node *ast.Node, typeAnnotation *ast.Node, initializer *ast.Node) {
		if initializer == nil {
			return
		}

		// Check if initializer is a NewExpression
		if initializer.Kind != ast.KindNewExpression {
			return
		}

		newExpr := initializer.AsNewExpression()
		if newExpr == nil {
			return
		}

		// Check if the callee is a simple identifier
		if newExpr.Expression == nil || newExpr.Expression.Kind != ast.KindIdentifier {
			return
		}

		calleeIdent := newExpr.Expression.AsIdentifier()
		if calleeIdent == nil {
			return
		}

		// Check if type arguments exist on constructor
		hasTypeArgsOnConstructor := newExpr.TypeArguments != nil && len(newExpr.TypeArguments.Nodes) > 0

		// Handle case where there's no type annotation
		if typeAnnotation == nil {
			// In type-annotation mode:
			// constructor with type args should have type annotation instead
			if opts.Style == "type-annotation" && hasTypeArgsOnConstructor {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "preferTypeAnnotation",
					Description: "The generic type arguments should be specified as part of the type annotation.",
				})
			}
			// For constructor mode, no type annotation is fine
			return
		}

		// Check if the type annotation is a type reference
		if typeAnnotation.Kind != ast.KindTypeReference {
			return
		}

		typeRef := typeAnnotation.AsTypeReference()
		if typeRef == nil {
			return
		}

		// Check if type reference name is an identifier
		if typeRef.TypeName == nil || typeRef.TypeName.Kind != ast.KindIdentifier {
			return
		}

		typeNameIdent := typeRef.TypeName.AsIdentifier()
		if typeNameIdent == nil {
			return
		}

		// Check if the names match
		calleeText := calleeIdent.Text
		typeNameText := typeNameIdent.Text
		if calleeText != typeNameText {
			return
		}

		// Check if type arguments exist on type annotation
		hasTypeArgsOnAnnotation := typeRef.TypeArguments != nil && len(typeRef.TypeArguments.Nodes) > 0

		// If both have type arguments or neither has type arguments, no violation
		if hasTypeArgsOnAnnotation == hasTypeArgsOnConstructor {
			return
		}

		if opts.Style == "constructor" {
			// Prefer constructor style
			if hasTypeArgsOnAnnotation && !hasTypeArgsOnConstructor {
				// Report: type args should be on constructor
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "preferConstructor",
					Description: "The generic type arguments should be specified as part of the constructor type arguments.",
				})
			}
		} else {
			// Prefer type-annotation style
			if hasTypeArgsOnConstructor && !hasTypeArgsOnAnnotation {
				// Report: type args should be on type annotation
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "preferTypeAnnotation",
					Description: "The generic type arguments should be specified as part of the type annotation.",
				})
			}
		}
	}

	return rule.RuleListeners{
		// Variable declarations
		ast.KindVariableDeclaration: func(node *ast.Node) {
			if node.Kind != ast.KindVariableDeclaration {
				return
			}
			varDecl := node.AsVariableDeclaration()
			if varDecl == nil {
				return
			}
			checkNode(node, varDecl.Type, varDecl.Initializer)
		},

		// Property declarations (class properties, including accessor properties)
		ast.KindPropertyDeclaration: func(node *ast.Node) {
			if node.Kind != ast.KindPropertyDeclaration {
				return
			}
			propDecl := node.AsPropertyDeclaration()
			if propDecl == nil {
				return
			}
			checkNode(node, propDecl.Type, propDecl.Initializer)
		},

		// Parameters (for functions, constructors, methods, arrow functions)
		ast.KindParameter: func(node *ast.Node) {
			if node.Kind != ast.KindParameter {
				return
			}
			param := node.AsParameterDeclaration()
			if param == nil {
				return
			}
			checkNode(node, param.Type, param.Initializer)
		},

		// Binding elements (for destructuring patterns)
		ast.KindBindingElement: func(node *ast.Node) {
			if node.Kind != ast.KindBindingElement {
				return
			}
			bindingElem := node.AsBindingElement()
			if bindingElem == nil {
				return
			}
			// BindingElement doesn't have a Type field, it can only have an initializer
			checkNode(node, nil, bindingElem.Initializer)
		},
	}
}

// createConstructorFixes creates fixes to move type arguments from type annotation to constructor
func createConstructorFixes(ctx rule.RuleContext, node *ast.Node, typeAnnotation *ast.Node, newExpr *ast.NewExpression, typeRef *ast.TypeReferenceNode) []rule.RuleFix {
	sourceText := ctx.SourceFile.Text()

	// Get type arguments text from type annotation
	if typeRef.TypeArguments == nil || len(typeRef.TypeArguments.Nodes) == 0 {
		return nil
	}

	typeArgsStart := typeRef.TypeArguments.Pos()
	typeArgsEnd := typeRef.TypeArguments.End()
	typeArgsText := sourceText[typeArgsStart:typeArgsEnd]

	// Find where to insert in constructor (after the constructor identifier)
	insertNode := newExpr.Expression

	return []rule.RuleFix{
		// Remove type arguments from type annotation
		rule.RuleFixReplaceRange(core.NewTextRange(typeArgsStart, typeArgsEnd), ""),
		// Add type arguments to constructor
		rule.RuleFixInsertAfter(insertNode, typeArgsText),
	}
}

// createTypeAnnotationFixes creates fixes to move type arguments from constructor to type annotation
func createTypeAnnotationFixes(ctx rule.RuleContext, node *ast.Node, typeAnnotation *ast.Node, newExpr *ast.NewExpression, typeRef *ast.TypeReferenceNode) []rule.RuleFix {
	sourceText := ctx.SourceFile.Text()

	// Get type arguments text from constructor
	if newExpr.TypeArguments == nil || len(newExpr.TypeArguments.Nodes) == 0 {
		return nil
	}

	typeArgsStart := newExpr.TypeArguments.Pos()
	typeArgsEnd := newExpr.TypeArguments.End()
	typeArgsText := sourceText[typeArgsStart:typeArgsEnd]

	// If there's no type annotation, we need to add it
	if typeAnnotation == nil {
		// Get the name node based on node kind
		var nameNode *ast.Node

		switch node.Kind {
		case ast.KindVariableDeclaration:
			varDecl := node.AsVariableDeclaration()
			if varDecl != nil {
				nameNode = varDecl.Name()
			}
		case ast.KindParameter:
			param := node.AsParameterDeclaration()
			if param != nil {
				nameNode = param.Name()
			}
		case ast.KindPropertyDeclaration:
			propDecl := node.AsPropertyDeclaration()
			if propDecl != nil {
				nameNode = propDecl.Name()
			}
		case ast.KindBindingElement:
			bindingElem := node.AsBindingElement()
			if bindingElem != nil {
				nameNode = bindingElem.Name()
			}
		}

		if nameNode == nil {
			return nil
		}

		// Get constructor name
		calleeIdent := newExpr.Expression.AsIdentifier()
		if calleeIdent == nil {
			return nil
		}
		constructorName := calleeIdent.Text

		// Create type annotation and remove type args from constructor
		return []rule.RuleFix{
			// Add type annotation after variable name
			rule.RuleFixInsertAfter(nameNode.AsNode(), ": "+constructorName+typeArgsText),
			// Remove type arguments from constructor
			rule.RuleFixReplaceRange(core.NewTextRange(typeArgsStart, typeArgsEnd), ""),
		}
	}

	// Find where to insert in type annotation (after the type name)
	insertNode := typeRef.TypeName

	// Create fixes: add to type annotation, remove from constructor
	return []rule.RuleFix{
		// Add type arguments to type annotation
		rule.RuleFixInsertAfter(insertNode, typeArgsText),
		// Remove type arguments from constructor
		rule.RuleFixReplaceRange(core.NewTextRange(typeArgsStart, typeArgsEnd), ""),
	}
}


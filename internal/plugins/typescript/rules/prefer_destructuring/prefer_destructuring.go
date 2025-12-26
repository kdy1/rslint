package prefer_destructuring

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildPreferDestructuringMessage(destructType string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferDestructuring",
		Description: "Use " + destructType + " destructuring.",
		Data: map[string]interface{}{
			"type": destructType,
		},
	}
}

type Options struct {
	VariableDeclarator    *DestructuringOptions `json:"VariableDeclarator,omitempty"`
	AssignmentExpression  *DestructuringOptions `json:"AssignmentExpression,omitempty"`
	Array                 *bool                 `json:"array,omitempty"`
	Object                *bool                 `json:"object,omitempty"`
	EnforceForRenamedProperties              *bool `json:"enforceForRenamedProperties,omitempty"`
	EnforceForDeclarationWithTypeAnnotation  *bool `json:"enforceForDeclarationWithTypeAnnotation,omitempty"`
}

type DestructuringOptions struct {
	Array  bool `json:"array"`
	Object bool `json:"object"`
}

var PreferDestructuringRule = rule.CreateRule(rule.Rule{
	Name: "prefer-destructuring",
	Run: func(ctx rule.RuleContext, optionsRaw any) rule.RuleListeners {
		options := parseOptions(optionsRaw)

		isAssignmentExpressionEnforced := func(shouldEnforce *bool) bool {
			if options.AssignmentExpression != nil {
				return shouldEnforce != nil && *shouldEnforce
			}
			return shouldEnforce != nil && *shouldEnforce
		}

		isVariableDeclaratorEnforced := func(shouldEnforce *bool) bool {
			if options.VariableDeclarator != nil {
				return shouldEnforce != nil && *shouldEnforce
			}
			return shouldEnforce != nil && *shouldEnforce
		}

		getObjectOption := func(nodeType ast.SyntaxKind) bool {
			if nodeType == ast.KindVariableDeclaration {
				if options.VariableDeclarator != nil {
					return options.VariableDeclarator.Object
				}
				return options.Object != nil && *options.Object
			}
			if options.AssignmentExpression != nil {
				return options.AssignmentExpression.Object
			}
			return options.Object != nil && *options.Object
		}

		getArrayOption := func(nodeType ast.SyntaxKind) bool {
			if nodeType == ast.KindVariableDeclaration {
				if options.VariableDeclarator != nil {
					return options.VariableDeclarator.Array
				}
				return options.Array != nil && *options.Array
			}
			if options.AssignmentExpression != nil {
				return options.AssignmentExpression.Array
			}
			return options.Array != nil && *options.Array
		}

		enforceForRenamedProperties := options.EnforceForRenamedProperties != nil && *options.EnforceForRenamedProperties
		enforceForDeclarationWithTypeAnnotation := options.EnforceForDeclarationWithTypeAnnotation != nil && *options.EnforceForDeclarationWithTypeAnnotation

		// Helper to check if a member expression is an array-like access (e.g., obj[0])
		isArrayIndexAccess := func(node *ast.Node) (bool, *ast.Node) {
			if node.Kind != ast.KindElementAccessExpression {
				return false, nil
			}
			elemAccess := node.AsElementAccessExpression()
			if elemAccess == nil || elemAccess.ArgumentExpression == nil {
				return false, nil
			}
			arg := elemAccess.ArgumentExpression
			if ast.IsNumericLiteral(arg) {
				return true, elemAccess.Expression
			}
			return false, nil
		}

		// Helper to check if a member expression is a property access (e.g., obj.foo)
		isPropertyAccess := func(node *ast.Node) (bool, string, *ast.Node) {
			if node.Kind == ast.KindPropertyAccessExpression {
				propAccess := node.AsPropertyAccessExpression()
				if propAccess != nil && propAccess.Name != nil {
					name := utils.GetIdentifierText(propAccess.Name)
					return true, name, propAccess.Expression
				}
			} else if node.Kind == ast.KindElementAccessExpression {
				elemAccess := node.AsElementAccessExpression()
				if elemAccess != nil && elemAccess.ArgumentExpression != nil {
					arg := elemAccess.ArgumentExpression
					if ast.IsStringLiteral(arg) {
						strLit := arg.AsStringLiteral()
						if strLit != nil {
							return true, strLit.Text, elemAccess.Expression
						}
					}
				}
			}
			return false, "", nil
		}

		// Helper to check if type is iterable
		isIterableType := func(node *ast.Node) bool {
			if node == nil || ctx.TypeChecker == nil {
				return false
			}
			tsType := ctx.TypeChecker.GetTypeAtLocation(node)
			if tsType == nil {
				return false
			}
			// Check if it's any
			if utils.IsTypeAnyType(tsType) {
				return true
			}
			// Check if it's an array or tuple
			if utils.IsTypeArrayTypeOrUnionOfArrayTypes(tsType) {
				return true
			}
			// Check if it has iterator symbol
			if utils.IsTypeIterable(tsType) {
				return true
			}
			return false
		}

		// Helper to determine if numeric index should be treated as array or object access
		shouldTreatAsArrayAccess := func(objectNode *ast.Node) bool {
			if objectNode == nil || ctx.TypeChecker == nil {
				return true // Default to array if no type info
			}
			tsType := ctx.TypeChecker.GetTypeAtLocation(objectNode)
			if tsType == nil {
				return true
			}
			// If it's any, treat as array
			if utils.IsTypeAnyType(tsType) {
				return true
			}
			// If it's iterable (array, tuple, has Symbol.iterator), treat as array
			if isIterableType(objectNode) {
				return true
			}
			// Otherwise, it's an object with numeric key
			return false
		}

		// Check variable declarator
		checkVariableDeclarator := func(node *ast.Node) {
			if node.Kind != ast.KindVariableDeclaration {
				return
			}
			varDecl := node.AsVariableDeclaration()
			if varDecl == nil || varDecl.Initializer == nil {
				return
			}

			// Skip if already destructured
			if varDecl.Name.Kind == ast.KindObjectBindingPattern || varDecl.Name.Kind == ast.KindArrayBindingPattern {
				return
			}

			// Skip if has type annotation unless enforceForDeclarationWithTypeAnnotation is true
			if varDecl.Type != nil && !enforceForDeclarationWithTypeAnnotation {
				return
			}

			init := varDecl.Initializer

			// Skip optional chaining
			if utils.IsOptionalChain(init) {
				return
			}

			// Skip private identifiers
			if init.Kind == ast.KindPropertyAccessExpression {
				propAccess := init.AsPropertyAccessExpression()
				if propAccess != nil && propAccess.Name != nil && propAccess.Name.Kind == ast.KindPrivateIdentifier {
					return
				}
			}

			// Check for array index access
			if isArrayAccess, objectNode := isArrayIndexAccess(init); isArrayAccess && getArrayOption(ast.KindVariableDeclaration) {
				// Determine if this should be treated as array or object destructuring
				if shouldTreatAsArrayAccess(objectNode) {
					ctx.ReportNode(node, buildPreferDestructuringMessage("array"))
					return
				} else if getObjectOption(ast.KindVariableDeclaration) {
					// It's an object with numeric key
					if enforceForRenamedProperties {
						ctx.ReportNode(node, buildPreferDestructuringMessage("object"))
					}
					return
				}
			}

			// Check for property access
			if isPropAccess, propName, objectNode := isPropertyAccess(init); isPropAccess && getObjectOption(ast.KindVariableDeclaration) {
				if objectNode == nil {
					return
				}

				// Skip super property access
				if objectNode.Kind == ast.KindSuperKeyword {
					return
				}

				// Check if variable name matches property name
				varName := utils.GetIdentifierText(varDecl.Name)
				if varName == propName {
					// Can use shorthand destructuring
					ctx.ReportNodeWithFixes(node, buildPreferDestructuringMessage("object"),
						createDestructuringFix(ctx, node, varDecl, init, propName, true))
				} else if enforceForRenamedProperties {
					// Would need renamed destructuring
					ctx.ReportNode(node, buildPreferDestructuringMessage("object"))
				}
			}
		}

		// Check assignment expression
		checkAssignmentExpression := func(node *ast.Node) {
			if node.Kind != ast.KindBinaryExpression {
				return
			}
			binExpr := node.AsBinaryExpression()
			if binExpr == nil || binExpr.OperatorToken.Kind != ast.KindEqualsToken {
				return
			}

			// Skip compound assignments (+=, -=, etc.)
			parent := node.Parent
			if parent != nil && parent.Kind == ast.KindBinaryExpression {
				parentBin := parent.AsBinaryExpression()
				if parentBin != nil && parentBin.OperatorToken.Kind != ast.KindEqualsToken {
					return
				}
			}

			left := binExpr.Left
			right := binExpr.Right

			// Skip if left is already destructured
			if left.Kind == ast.KindObjectBindingPattern || left.Kind == ast.KindArrayBindingPattern {
				return
			}

			// Skip optional chaining
			if utils.IsOptionalChain(right) {
				return
			}

			// Skip private identifiers
			if right.Kind == ast.KindPropertyAccessExpression {
				propAccess := right.AsPropertyAccessExpression()
				if propAccess != nil && propAccess.Name != nil && propAccess.Name.Kind == ast.KindPrivateIdentifier {
					return
				}
			}

			// Check for array index access
			if isArrayAccess, objectNode := isArrayIndexAccess(right); isArrayAccess {
				if shouldTreatAsArrayAccess(objectNode) {
					if isAssignmentExpressionEnforced(options.Array) {
						ctx.ReportNode(node, buildPreferDestructuringMessage("array"))
					}
					return
				} else if isAssignmentExpressionEnforced(options.Object) {
					if enforceForRenamedProperties {
						ctx.ReportNode(node, buildPreferDestructuringMessage("object"))
					}
					return
				}
			}

			// Check for property access
			if isPropAccess, propName, objectNode := isPropertyAccess(right); isPropAccess {
				if objectNode == nil {
					return
				}

				// Skip super property access
				if objectNode.Kind == ast.KindSuperKeyword {
					return
				}

				if isAssignmentExpressionEnforced(options.Object) {
					// Check if left side identifier matches property name
					if left.Kind == ast.KindIdentifier {
						leftName := utils.GetIdentifierText(left)
						if leftName != propName && !enforceForRenamedProperties {
							return
						}
					}
					ctx.ReportNode(node, buildPreferDestructuringMessage("object"))
				}
			}
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration: checkVariableDeclarator,
			ast.KindBinaryExpression:    checkAssignmentExpression,
		}
	},
})

// Helper to create auto-fix for simple object destructuring
func createDestructuringFix(ctx rule.RuleContext, node *ast.Node, varDecl *ast.VariableDeclaration, init *ast.Node, propName string, canAutoFix bool) rule.RuleFix {
	if !canAutoFix {
		return rule.RuleFix{}
	}

	// Get the object being accessed
	var objectNode *ast.Node
	if init.Kind == ast.KindPropertyAccessExpression {
		propAccess := init.AsPropertyAccessExpression()
		if propAccess != nil {
			objectNode = propAccess.Expression
		}
	} else if init.Kind == ast.KindElementAccessExpression {
		elemAccess := init.AsElementAccessExpression()
		if elemAccess != nil {
			objectNode = elemAccess.Expression
		}
	}

	if objectNode == nil {
		return rule.RuleFix{}
	}

	// Get the text of the variable declaration without the initializer
	objectRange := utils.TrimNodeTextRange(ctx.SourceFile, objectNode)
	objectText := ctx.SourceFile.Text()[objectRange.Pos():objectRange.End()]

	// Create the destructuring pattern
	newText := "{" + propName + "} = " + objectText

	// Replace the variable name and initializer
	nameRange := utils.TrimNodeTextRange(ctx.SourceFile, varDecl.Name)
	initRange := utils.TrimNodeTextRange(ctx.SourceFile, init)

	return rule.RuleFixReplaceRange(
		core.NewTextRange(nameRange.Pos(), initRange.End()),
		newText,
	)
}

func parseOptions(optionsRaw any) Options {
	options := Options{
		Array:                                    boolPtr(true),
		Object:                                   boolPtr(true),
		EnforceForRenamedProperties:              boolPtr(false),
		EnforceForDeclarationWithTypeAnnotation:  boolPtr(false),
	}

	if optionsRaw == nil {
		return options
	}

	optionsSlice, ok := optionsRaw.([]interface{})
	if !ok || len(optionsSlice) == 0 {
		return options
	}

	// First parameter: destructuring type configuration
	if len(optionsSlice) > 0 && optionsSlice[0] != nil {
		if typeConfig, ok := optionsSlice[0].(map[string]interface{}); ok {
			// Check if it's the simplified format {array: bool, object: bool}
			if arrayVal, hasArray := typeConfig["array"]; hasArray {
				if arrayBool, ok := arrayVal.(bool); ok {
					options.Array = &arrayBool
				}
			}
			if objectVal, hasObject := typeConfig["object"]; hasObject {
				if objectBool, ok := objectVal.(bool); ok {
					options.Object = &objectBool
				}
			}

			// Check for per-statement-type format
			if varDeclConfig, hasVarDecl := typeConfig["VariableDeclarator"]; hasVarDecl {
				if varDeclMap, ok := varDeclConfig.(map[string]interface{}); ok {
					varOpts := &DestructuringOptions{}
					if arrayVal, hasArray := varDeclMap["array"]; hasArray {
						if arrayBool, ok := arrayVal.(bool); ok {
							varOpts.Array = arrayBool
						}
					}
					if objectVal, hasObject := varDeclMap["object"]; hasObject {
						if objectBool, ok := objectVal.(bool); ok {
							varOpts.Object = objectBool
						}
					}
					options.VariableDeclarator = varOpts
				}
			}

			if assignConfig, hasAssign := typeConfig["AssignmentExpression"]; hasAssign {
				if assignMap, ok := assignConfig.(map[string]interface{}); ok {
					assignOpts := &DestructuringOptions{}
					if arrayVal, hasArray := assignMap["array"]; hasArray {
						if arrayBool, ok := arrayVal.(bool); ok {
							assignOpts.Array = arrayBool
						}
					}
					if objectVal, hasObject := assignMap["object"]; hasObject {
						if objectBool, ok := objectVal.(bool); ok {
							assignOpts.Object = objectBool
						}
					}
					options.AssignmentExpression = assignOpts
				}
			}
		}
	}

	// Second parameter: additional options
	if len(optionsSlice) > 1 && optionsSlice[1] != nil {
		if additionalOpts, ok := optionsSlice[1].(map[string]interface{}); ok {
			if enforceRenamedVal, hasEnforceRenamed := additionalOpts["enforceForRenamedProperties"]; hasEnforceRenamed {
				if enforceRenamedBool, ok := enforceRenamedVal.(bool); ok {
					options.EnforceForRenamedProperties = &enforceRenamedBool
				}
			}
			if enforceTypeAnnotationVal, hasEnforceTypeAnnotation := additionalOpts["enforceForDeclarationWithTypeAnnotation"]; hasEnforceTypeAnnotation {
				if enforceTypeAnnotationBool, ok := enforceTypeAnnotationVal.(bool); ok {
					options.EnforceForDeclarationWithTypeAnnotation = &enforceTypeAnnotationBool
				}
			}
		}
	}

	return options
}

func boolPtr(b bool) *bool {
	return &b
}

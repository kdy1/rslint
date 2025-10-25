package no_undef

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoUndefOptions defines the configuration options for this rule
type NoUndefOptions struct {
	Typeof bool `json:"typeof"`
}

// parseOptions parses and validates the rule options
func parseOptions(options any) NoUndefOptions {
	opts := NoUndefOptions{
		Typeof: false, // Default: check typeof operands
	}

	if options == nil {
		return opts
	}

	// Handle both array format [{ option: value }] and object format { option: value }
	var optsMap map[string]interface{}
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, _ = optArray[0].(map[string]interface{})
	} else {
		optsMap, _ = options.(map[string]interface{})
	}

	if optsMap != nil {
		if v, ok := optsMap["typeof"].(bool); ok {
			opts.Typeof = v
		}
	}

	return opts
}

// NoUndefRule implements the no-undef rule
// Disallow undeclared variables
var NoUndefRule = rule.Rule{
	Name: "no-undef",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)
	typeChecker := ctx.TypeChecker
	if typeChecker == nil {
		return rule.RuleListeners{}
	}

	return rule.RuleListeners{
		ast.KindIdentifier: func(node *ast.Node) {
			identifier := node.AsIdentifier()
			if identifier == nil {
				return
			}

			// Skip if this identifier is in a typeof expression and typeof option is false
			if !opts.Typeof && isInTypeOf(node) {
				return
			}

			// Skip if this is part of a type annotation (not a value reference)
			if isInTypeContext(node) {
				return
			}

			// Skip if this is part of a declaration (e.g., function name, parameter name)
			if isPartOfDeclaration(node) {
				return
			}

			// Skip property names in object literals and member expressions
			if isPropertyName(node) {
				return
			}

			// Try to resolve the symbol for this identifier
			symbol := utils.GetSymbolAtLocation(typeChecker, node)

			// If no symbol found, this might be an undeclared variable
			if symbol == nil {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "undef",
					Description: "'" + identifier.EscapedText + "' is not defined.",
				})
				return
			}

			// Check if the symbol has any declarations
			// Symbols from ambient declarations (like global browser APIs) will have declarations
			if symbol.Declarations == nil || len(symbol.Declarations.Nodes) == 0 {
				// No declarations means it's not properly declared
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "undef",
					Description: "'" + identifier.EscapedText + "' is not defined.",
				})
			}
		},
	}
}

// isInTypeOf checks if a node is used as the operand of a typeof expression
func isInTypeOf(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}

	parent := node.Parent
	if parent.Kind == ast.KindTypeOfExpression {
		typeofExpr := parent.AsTypeOfExpression()
		if typeofExpr != nil && typeofExpr.Expression == node {
			return true
		}
	}

	return false
}

// isInTypeContext checks if a node is in a type annotation context (not a value)
func isInTypeContext(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}

	parent := node.Parent
	switch parent.Kind {
	case ast.KindTypeReference,
		ast.KindTypeQuery,
		ast.KindTypeParameter,
		ast.KindTypeAliasDeclaration,
		ast.KindInterfaceDeclaration,
		ast.KindTypeLiteral:
		return true
	}

	return false
}

// isPartOfDeclaration checks if an identifier is part of a declaration
func isPartOfDeclaration(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}

	parent := node.Parent
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		varDecl := parent.AsVariableDeclaration()
		return varDecl != nil && varDecl.Name == node

	case ast.KindFunctionDeclaration:
		funcDecl := parent.AsFunctionDeclaration()
		return funcDecl != nil && funcDecl.Name == node

	case ast.KindParameter:
		param := parent.AsParameterDeclaration()
		return param != nil && param.Name == node

	case ast.KindClassDeclaration:
		classDecl := parent.AsClassDeclaration()
		return classDecl != nil && classDecl.Name == node

	case ast.KindInterfaceDeclaration:
		ifaceDecl := parent.AsInterfaceDeclaration()
		return ifaceDecl != nil && ifaceDecl.Name == node

	case ast.KindTypeAliasDeclaration:
		typeAlias := parent.AsTypeAliasDeclaration()
		return typeAlias != nil && typeAlias.Name == node

	case ast.KindEnumDeclaration:
		enumDecl := parent.AsEnumDeclaration()
		return enumDecl != nil && enumDecl.Name == node

	case ast.KindImportSpecifier:
		return true

	case ast.KindImportClause:
		return true

	case ast.KindNamespaceImport:
		return true

	case ast.KindBindingElement:
		bindingElem := parent.AsBindingElement()
		return bindingElem != nil && bindingElem.Name == node

	case ast.KindCatchClause:
		catchClause := parent.AsCatchClause()
		if catchClause != nil && catchClause.VariableDeclaration != nil {
			varDecl := catchClause.VariableDeclaration.AsVariableDeclaration()
			return varDecl != nil && varDecl.Name == node
		}
	}

	return false
}

// isPropertyName checks if an identifier is used as a property name
func isPropertyName(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}

	parent := node.Parent
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		propAccess := parent.AsPropertyAccessExpression()
		return propAccess != nil && propAccess.Name == node

	case ast.KindPropertyAssignment:
		propAssign := parent.AsPropertyAssignment()
		return propAssign != nil && propAssign.Name == node

	case ast.KindShorthandPropertyAssignment:
		// Shorthand properties are both declaration and usage, so don't skip them
		return false

	case ast.KindMethodDeclaration:
		methodDecl := parent.AsMethodDeclaration()
		return methodDecl != nil && methodDecl.Name == node

	case ast.KindPropertyDeclaration:
		propDecl := parent.AsPropertyDeclaration()
		return propDecl != nil && propDecl.Name == node
	}

	return false
}

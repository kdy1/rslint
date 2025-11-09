// Package no_redeclare implements the @typescript-eslint/no-redeclare rule.
// This rule disallows variable redeclaration while allowing TypeScript-specific patterns
// like function overloads and declaration merging.
package no_redeclare

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoRedeclareOptions struct {
	BuiltinGlobals         bool `json:"builtinGlobals"`
	IgnoreDeclarationMerge bool `json:"ignoreDeclarationMerge"`
}

func parseOptions(options any) NoRedeclareOptions {
	opts := NoRedeclareOptions{
		BuiltinGlobals:         false,
		IgnoreDeclarationMerge: true,
	}

	if options == nil {
		return opts
	}

	var optsMap map[string]interface{}
	if optsArray, ok := options.([]interface{}); ok && len(optsArray) > 0 {
		if m, ok := optsArray[0].(map[string]interface{}); ok {
			optsMap = m
		}
	} else if m, ok := options.(map[string]interface{}); ok {
		optsMap = m
	}

	if optsMap != nil {
		if v, ok := optsMap["builtinGlobals"].(bool); ok {
			opts.BuiltinGlobals = v
		}
		if v, ok := optsMap["ignoreDeclarationMerge"].(bool); ok {
			opts.IgnoreDeclarationMerge = v
		}
	}

	return opts
}

func buildRedeclaredMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "redeclared",
		Description: "'{{id}}' is already defined.",
	}
}

func buildRedeclaredAsBuiltinMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "redeclaredAsBuiltin",
		Description: "'{{id}}' is already defined as a built-in global variable.",
	}
}

func buildRedeclaredBySyntaxMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "redeclaredBySyntax",
		Description: "'{{id}}' is already defined by a variable declaration.",
	}
}

// Built-in global objects that are commonly available
var builtinGlobals = map[string]bool{
	"Object":            true,
	"Array":             true,
	"String":            true,
	"Number":            true,
	"Boolean":           true,
	"Function":          true,
	"Date":              true,
	"RegExp":            true,
	"Error":             true,
	"Math":              true,
	"JSON":              true,
	"Promise":           true,
	"Symbol":            true,
	"Map":               true,
	"Set":               true,
	"WeakMap":           true,
	"WeakSet":           true,
	"ArrayBuffer":       true,
	"DataView":          true,
	"Int8Array":         true,
	"Uint8Array":        true,
	"Uint8ClampedArray": true,
	"Int16Array":        true,
	"Uint16Array":       true,
	"Int32Array":        true,
	"Uint32Array":       true,
	"Float32Array":      true,
	"Float64Array":      true,
	"Proxy":             true,
	"Reflect":           true,
	// DOM types (when lib includes dom)
	"NodeListOf": true,
}

// declarationInfo tracks information about a declaration
type declarationInfo struct {
	node    *ast.Node
	name    string
	kind    ast.Kind
	isType  bool // true for type/interface declarations
	isValue bool // true for value declarations (var, let, const, function, class, enum)
}

// canMerge checks if two declarations can be merged according to TypeScript rules
func canMerge(first, second *declarationInfo, ignoreDeclarationMerge bool) bool {
	if !ignoreDeclarationMerge {
		return false
	}

	// Function overloads are always allowed
	if first.kind == ast.KindFunctionDeclaration && second.kind == ast.KindFunctionDeclaration {
		// Check if one is an overload signature (no body)
		firstFn := first.node.AsFunctionDeclaration()
		secondFn := second.node.AsFunctionDeclaration()
		if firstFn != nil && secondFn != nil {
			// If either has no body, it's an overload signature
			if firstFn.Body == nil || secondFn.Body == nil {
				return true
			}
		}
	}

	// Type-only declarations can merge with other type-only declarations
	if first.isType && second.isType {
		// Interface + Interface
		if first.kind == ast.KindInterfaceDeclaration && second.kind == ast.KindInterfaceDeclaration {
			return true
		}
	}

	// Type and value can merge in specific cases
	if first.isType != second.isType {
		// Interface + Class
		if (first.kind == ast.KindInterfaceDeclaration && second.kind == ast.KindClassDeclaration) ||
			(first.kind == ast.KindClassDeclaration && second.kind == ast.KindInterfaceDeclaration) {
			return true
		}

		// Class + Namespace
		if (first.kind == ast.KindClassDeclaration && second.kind == ast.KindModuleDeclaration) ||
			(first.kind == ast.KindModuleDeclaration && second.kind == ast.KindClassDeclaration) {
			return true
		}

		// Function + Namespace
		if (first.kind == ast.KindFunctionDeclaration && second.kind == ast.KindModuleDeclaration) ||
			(first.kind == ast.KindModuleDeclaration && second.kind == ast.KindFunctionDeclaration) {
			return true
		}

		// Enum + Namespace
		if (first.kind == ast.KindEnumDeclaration && second.kind == ast.KindModuleDeclaration) ||
			(first.kind == ast.KindModuleDeclaration && second.kind == ast.KindEnumDeclaration) {
			return true
		}

		// Interface + Namespace
		if (first.kind == ast.KindInterfaceDeclaration && second.kind == ast.KindModuleDeclaration) ||
			(first.kind == ast.KindModuleDeclaration && second.kind == ast.KindInterfaceDeclaration) {
			return true
		}
	}

	// Value declarations can merge in specific cases
	if first.isValue && second.isValue {
		// Namespace + Namespace
		if first.kind == ast.KindModuleDeclaration && second.kind == ast.KindModuleDeclaration {
			return true
		}
	}

	return false
}

// isTypeDeclaration checks if a node is a type-only declaration
func isTypeDeclaration(kind ast.Kind) bool {
	return kind == ast.KindInterfaceDeclaration ||
		kind == ast.KindTypeAliasDeclaration
}

// isValueDeclaration checks if a node is a value declaration
func isValueDeclaration(kind ast.Kind) bool {
	return kind == ast.KindVariableDeclaration ||
		kind == ast.KindFunctionDeclaration ||
		kind == ast.KindClassDeclaration ||
		kind == ast.KindEnumDeclaration ||
		kind == ast.KindModuleDeclaration
}

// getIdentifierName extracts the name from a node
func getIdentifierName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		ident := node.AsIdentifier()
		if ident != nil {
			return ident.Text
		}
	case ast.KindVariableDeclaration:
		varDecl := node.AsVariableDeclaration()
		if varDecl != nil && varDecl.Name() != nil {
			return getIdentifierName(varDecl.Name())
		}
	case ast.KindFunctionDeclaration:
		fnDecl := node.AsFunctionDeclaration()
		if fnDecl != nil && fnDecl.Name() != nil {
			return getIdentifierName(fnDecl.Name())
		}
	case ast.KindClassDeclaration:
		classDecl := node.AsClassDeclaration()
		if classDecl != nil && classDecl.Name() != nil {
			return getIdentifierName(classDecl.Name())
		}
	case ast.KindInterfaceDeclaration:
		interfaceDecl := node.AsInterfaceDeclaration()
		if interfaceDecl != nil && interfaceDecl.Name() != nil {
			return getIdentifierName(interfaceDecl.Name())
		}
	case ast.KindTypeAliasDeclaration:
		typeAlias := node.AsTypeAliasDeclaration()
		if typeAlias != nil && typeAlias.Name() != nil {
			return getIdentifierName(typeAlias.Name())
		}
	case ast.KindEnumDeclaration:
		enumDecl := node.AsEnumDeclaration()
		if enumDecl != nil && enumDecl.Name() != nil {
			return getIdentifierName(enumDecl.Name())
		}
	case ast.KindModuleDeclaration:
		moduleDecl := node.AsModuleDeclaration()
		if moduleDecl != nil && moduleDecl.Name() != nil {
			return getIdentifierName(moduleDecl.Name())
		}
	}

	return ""
}

// getReportNode gets the node to report for an error
func getReportNode(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindVariableDeclaration:
		varDecl := node.AsVariableDeclaration()
		if varDecl != nil && varDecl.Name() != nil {
			return varDecl.Name()
		}
	case ast.KindFunctionDeclaration:
		fnDecl := node.AsFunctionDeclaration()
		if fnDecl != nil && fnDecl.Name() != nil {
			return fnDecl.Name()
		}
	case ast.KindClassDeclaration:
		classDecl := node.AsClassDeclaration()
		if classDecl != nil && classDecl.Name() != nil {
			return classDecl.Name()
		}
	case ast.KindInterfaceDeclaration:
		interfaceDecl := node.AsInterfaceDeclaration()
		if interfaceDecl != nil && interfaceDecl.Name() != nil {
			return interfaceDecl.Name()
		}
	case ast.KindTypeAliasDeclaration:
		typeAlias := node.AsTypeAliasDeclaration()
		if typeAlias != nil && typeAlias.Name() != nil {
			return typeAlias.Name()
		}
	case ast.KindEnumDeclaration:
		enumDecl := node.AsEnumDeclaration()
		if enumDecl != nil && enumDecl.Name() != nil {
			return enumDecl.Name()
		}
	case ast.KindModuleDeclaration:
		moduleDecl := node.AsModuleDeclaration()
		if moduleDecl != nil && moduleDecl.Name() != nil {
			return moduleDecl.Name()
		}
	}

	return node
}

var NoRedeclareRule = rule.CreateRule(rule.Rule{
	Name: "no-redeclare",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		// Track declarations per scope
		// Map from scope node to map of identifier names to declaration info
		scopeDeclarations := make(map[*ast.Node]map[string][]*declarationInfo)

		// Get the current scope for a node (simplified - uses nearest block or source file)
		getCurrentScope := func(node *ast.Node) *ast.Node {
			current := node.Parent
			for current != nil {
				switch current.Kind {
				case ast.KindSourceFile,
					ast.KindBlock,
					ast.KindModuleBlock,
					ast.KindFunctionDeclaration,
					ast.KindFunctionExpression,
					ast.KindArrowFunction,
					ast.KindMethodDeclaration:
					return current
				}
				current = current.Parent
			}
			return nil
		}

		// Check and record a declaration
		checkDeclaration := func(node *ast.Node) {
			name := getIdentifierName(node)
			if name == "" {
				return
			}

			// Check if it's a builtin global
			if opts.BuiltinGlobals && builtinGlobals[name] {
				reportNode := getReportNode(node)
				if reportNode != nil {
					ctx.ReportNodeWithData(reportNode, buildRedeclaredAsBuiltinMessage(), map[string]string{
						"id": name,
					})
				}
				return
			}

			scope := getCurrentScope(node)
			if scope == nil {
				return
			}

			if scopeDeclarations[scope] == nil {
				scopeDeclarations[scope] = make(map[string][]*declarationInfo)
			}

			currentInfo := &declarationInfo{
				node:    node,
				name:    name,
				kind:    node.Kind,
				isType:  isTypeDeclaration(node.Kind),
				isValue: isValueDeclaration(node.Kind),
			}

			// Check for redeclaration
			if existing, found := scopeDeclarations[scope][name]; found {
				// Check if any existing declaration conflicts
				canMergeWithAll := true
				for _, existingInfo := range existing {
					if !canMerge(existingInfo, currentInfo, opts.IgnoreDeclarationMerge) {
						canMergeWithAll = false
						break
					}
				}

				if !canMergeWithAll {
					reportNode := getReportNode(node)
					if reportNode != nil {
						ctx.ReportNodeWithData(reportNode, buildRedeclaredMessage(), map[string]string{
							"id": name,
						})
					}
					return
				}
			}

			// Add to scope declarations
			scopeDeclarations[scope][name] = append(scopeDeclarations[scope][name], currentInfo)
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				checkDeclaration(node)
			},
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				checkDeclaration(node)
			},
			ast.KindClassDeclaration: func(node *ast.Node) {
				checkDeclaration(node)
			},
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				checkDeclaration(node)
			},
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				checkDeclaration(node)
			},
			ast.KindEnumDeclaration: func(node *ast.Node) {
				checkDeclaration(node)
			},
			ast.KindModuleDeclaration: func(node *ast.Node) {
				checkDeclaration(node)
			},
		}
	},
})

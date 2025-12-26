package no_use_before_define

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Options for the no-use-before-define rule
type NoUseBeforeDefineOptions struct {
	Functions            *bool `json:"functions,omitempty"`
	Classes              *bool `json:"classes,omitempty"`
	Enums                *bool `json:"enums,omitempty"`
	Variables            *bool `json:"variables,omitempty"`
	Typedefs             *bool `json:"typedefs,omitempty"`
	IgnoreTypeReferences *bool `json:"ignoreTypeReferences,omitempty"`
	AllowNamedExports    *bool `json:"allowNamedExports,omitempty"`
}

// Message builder
func buildNoUseBeforeDefineMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noUseBeforeDefine",
		Description: "'" + name + "' was used before it was defined.",
	}
}

// parseOptions parses rule options with defaults
func parseOptions(options any) NoUseBeforeDefineOptions {
	opts := NoUseBeforeDefineOptions{
		Functions:            boolPtr(true),
		Classes:              boolPtr(true),
		Enums:                boolPtr(true),
		Variables:            boolPtr(true),
		Typedefs:             boolPtr(true),
		IgnoreTypeReferences: boolPtr(true),
		AllowNamedExports:    boolPtr(false),
	}

	if options == nil {
		return opts
	}

	// Handle string option "nofunc"
	if str, ok := options.(string); ok {
		if str == "nofunc" {
			opts.Functions = boolPtr(false)
		}
		return opts
	}

	// Handle array of options
	if arr, ok := options.([]any); ok && len(arr) > 0 {
		if str, ok := arr[0].(string); ok && str == "nofunc" {
			opts.Functions = boolPtr(false)
			return opts
		}
		if optMap, ok := arr[0].(map[string]any); ok {
			applyOptionMap(&opts, optMap)
		}
		return opts
	}

	// Handle map option
	if optMap, ok := options.(map[string]any); ok {
		applyOptionMap(&opts, optMap)
	}

	return opts
}

func applyOptionMap(opts *NoUseBeforeDefineOptions, optMap map[string]any) {
	if v, ok := optMap["functions"].(bool); ok {
		opts.Functions = &v
	}
	if v, ok := optMap["classes"].(bool); ok {
		opts.Classes = &v
	}
	if v, ok := optMap["enums"].(bool); ok {
		opts.Enums = &v
	}
	if v, ok := optMap["variables"].(bool); ok {
		opts.Variables = &v
	}
	if v, ok := optMap["typedefs"].(bool); ok {
		opts.Typedefs = &v
	}
	if v, ok := optMap["ignoreTypeReferences"].(bool); ok {
		opts.IgnoreTypeReferences = &v
	}
	if v, ok := optMap["allowNamedExports"].(bool); ok {
		opts.AllowNamedExports = &v
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func getBool(ptr *bool) bool {
	if ptr == nil {
		return false
	}
	return *ptr
}

// declarationInfo tracks a declaration and its position
type declarationInfo struct {
	node       *ast.Node
	name       string
	pos        int
	kind       declarationKind
	isFunction bool
	isClass    bool
	isEnum     bool
	isTypedef  bool
	isVariable bool
}

type declarationKind int

const (
	declVariable declarationKind = iota
	declFunction
	declClass
	declEnum
	declTypedef
)

// referenceInfo tracks a reference to a name
type referenceInfo struct {
	node      *ast.Node
	name      string
	pos       int
	isTypeRef bool
	isExport  bool
}

// NoUseBeforeDefineRule implements the no-use-before-define rule
var NoUseBeforeDefineRule = rule.CreateRule(rule.Rule{
	Name: "no-use-before-define",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		// Track declarations and references
		declarations := make(map[string][]*declarationInfo)
		references := make([]*referenceInfo, 0)

		// Helper to get text position
		getPos := func(node *ast.Node) int {
			if node == nil {
				return 0
			}
			return node.Pos()
		}

		// Helper to get identifier name
		getName := func(node *ast.Node) string {
			if node == nil || node.Kind != ast.KindIdentifier {
				return ""
			}
			return node.Text()
		}

		// Helper to check if node is in type position
		var isInTypePosition func(node *ast.Node) bool
		isInTypePosition = func(node *ast.Node) bool {
			if node == nil || node.Parent == nil {
				return false
			}

			parent := node.Parent
			switch parent.Kind {
			case ast.KindTypeReference, ast.KindTypeQuery, ast.KindTypeParameter,
				ast.KindTypeAliasDeclaration, ast.KindInterfaceDeclaration,
				ast.KindAsExpression, ast.KindTypeAssertionExpression:
				return true
			case ast.KindQualifiedName:
				// For qualified names like Foo.Bar in typeof Foo.Bar
				return isInTypePosition(parent)
			}
			return false
		}

		// Helper to check if identifier is in export specifier
		isInExportSpecifier := func(node *ast.Node) bool {
			if node == nil || node.Parent == nil {
				return false
			}
			parent := node.Parent
			if parent.Kind == ast.KindExportSpecifier {
				spec := parent.AsExportSpecifier()
				if spec != nil && spec.Name != nil {
					return spec.Name() == node
				}
			}
			return false
		}

		// Helper to check if in initializer
		isInInitializer := func(node *ast.Node, declNode *ast.Node) bool {
			if node == nil || declNode == nil {
				return false
			}

			// Walk up from the reference to see if it's in the initializer
			current := node.Parent
			for current != nil {
				if current.Kind == ast.KindVariableDeclaration {
					varDecl := current.AsVariableDeclaration()
					if varDecl != nil && varDecl.Initializer != nil {
						// Check if node is within the initializer
						refPos := getPos(node)
						initStart := getPos(varDecl.Initializer)
						initEnd := varDecl.Initializer.End()
						if refPos >= initStart && refPos <= initEnd {
							return current == declNode
						}
					}
					break
				} else if current.Kind == ast.KindBindingElement {
					bindingElem := current.AsBindingElement()
					if bindingElem != nil && bindingElem.Initializer != nil {
						refPos := getPos(node)
						rightStart := getPos(bindingElem.Initializer)
						rightEnd := bindingElem.Initializer.End()
						return refPos >= rightStart && refPos <= rightEnd
					}
				}
				current = current.Parent
			}
			return false
		}

		// Collect declarations
		collectDeclarations := func(node *ast.Node) {
			switch node.Kind {
			case ast.KindVariableDeclaration:
				varDecl := node.AsVariableDeclaration()
				if varDecl != nil && varDecl.Name() != nil {
					name := getName(varDecl.Name())
					if name != "" {
						declarations[name] = append(declarations[name], &declarationInfo{
							node:       node,
							name:       name,
							pos:        getPos(varDecl.Name()),
							kind:       declVariable,
							isVariable: true,
						})
					}
				}

			case ast.KindFunctionDeclaration:
				funcDecl := node.AsFunctionDeclaration()
				if funcDecl != nil && funcDecl.Name() != nil {
					name := getName(funcDecl.Name())
					if name != "" {
						declarations[name] = append(declarations[name], &declarationInfo{
							node:       node,
							name:       name,
							pos:        getPos(funcDecl.Name()),
							kind:       declFunction,
							isFunction: true,
						})
					}
				}

			case ast.KindClassDeclaration:
				classDecl := node.AsClassDeclaration()
				if classDecl != nil && classDecl.Name() != nil {
					name := getName(classDecl.Name())
					if name != "" {
						declarations[name] = append(declarations[name], &declarationInfo{
							node:       node,
							name:       name,
							pos:        getPos(classDecl.Name()),
							kind:       declClass,
							isClass:    true,
						})
					}
				}

			case ast.KindEnumDeclaration:
				enumDecl := node.AsEnumDeclaration()
				if enumDecl != nil && enumDecl.Name() != nil {
					name := getName(enumDecl.Name())
					if name != "" {
						declarations[name] = append(declarations[name], &declarationInfo{
							node:       node,
							name:       name,
							pos:        getPos(enumDecl.Name()),
							kind:       declEnum,
							isEnum:     true,
						})
					}
				}

			case ast.KindTypeAliasDeclaration:
				typeDecl := node.AsTypeAliasDeclaration()
				if typeDecl != nil && typeDecl.Name() != nil {
					name := getName(typeDecl.Name())
					if name != "" {
						declarations[name] = append(declarations[name], &declarationInfo{
							node:       node,
							name:       name,
							pos:        getPos(typeDecl.Name()),
							kind:       declTypedef,
							isTypedef:  true,
						})
					}
				}

			case ast.KindInterfaceDeclaration:
				interfaceDecl := node.AsInterfaceDeclaration()
				if interfaceDecl != nil && interfaceDecl.Name() != nil {
					name := getName(interfaceDecl.Name())
					if name != "" {
						declarations[name] = append(declarations[name], &declarationInfo{
							node:       node,
							name:       name,
							pos:        getPos(interfaceDecl.Name()),
							kind:       declTypedef,
							isTypedef:  true,
						})
					}
				}
			}
		}

		// Collect references
		collectReferences := func(node *ast.Node) {
			if node.Kind == ast.KindIdentifier {
				// Skip if it's a declaration
				if node.Parent != nil {
					parent := node.Parent
					switch parent.Kind {
					case ast.KindVariableDeclaration:
						varDecl := parent.AsVariableDeclaration()
						if varDecl != nil && varDecl.Name() == node {
							return
						}
					case ast.KindFunctionDeclaration:
						funcDecl := parent.AsFunctionDeclaration()
						if funcDecl != nil && funcDecl.Name() == node {
							return
						}
					case ast.KindClassDeclaration:
						classDecl := parent.AsClassDeclaration()
						if classDecl != nil && classDecl.Name() == node {
							return
						}
					case ast.KindEnumDeclaration:
						enumDecl := parent.AsEnumDeclaration()
						if enumDecl != nil && enumDecl.Name() == node {
							return
						}
					case ast.KindTypeAliasDeclaration:
						typeDecl := parent.AsTypeAliasDeclaration()
						if typeDecl != nil && typeDecl.Name() == node {
							return
						}
					case ast.KindInterfaceDeclaration:
						interfaceDecl := parent.AsInterfaceDeclaration()
						if interfaceDecl != nil && interfaceDecl.Name() == node {
							return
						}
					case ast.KindParameter:
						param := parent.AsParameterDeclaration()
						if param != nil && param.Name() == node {
							return
						}
					case ast.KindPropertyDeclaration, ast.KindPropertySignature,
						ast.KindMethodDeclaration, ast.KindMethodSignature:
						// Skip property/method names
						return
					}
				}

				name := getName(node)
				if name != "" {
					references = append(references, &referenceInfo{
						node:      node,
						name:      name,
						pos:       getPos(node),
						isTypeRef: isInTypePosition(node),
						isExport:  isInExportSpecifier(node),
					})
				}
			}
		}

		// Walk the tree on SourceFile node
		return rule.RuleListeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// First pass: collect all declarations
				var walkDecl func(*ast.Node)
				walkDecl = func(n *ast.Node) {
					if n == nil {
						return
					}
					collectDeclarations(n)
					n.ForEachChild(func(child *ast.Node) bool {
						walkDecl(child)
						return false
					})
				}
				walkDecl(node)

				// Second pass: collect all references
				var walkRef func(*ast.Node)
				walkRef = func(n *ast.Node) {
					if n == nil {
						return
					}
					collectReferences(n)
					n.ForEachChild(func(child *ast.Node) bool {
						walkRef(child)
						return false
					})
				}
				walkRef(node)

				// Check references against declarations
				for _, ref := range references {
					decls, ok := declarations[ref.name]
					if !ok || len(decls) == 0 {
						continue
					}

					// Check if this reference should be ignored
					if getBool(opts.IgnoreTypeReferences) && ref.isTypeRef {
						continue
					}

					if ref.isExport && !getBool(opts.AllowNamedExports) {
						// Check if used before definition for exports
						for _, decl := range decls {
							if ref.pos < decl.pos {
								ctx.ReportNode(ref.node, buildNoUseBeforeDefineMessage(ref.name))
								break
							}
						}
						continue
					}

					if ref.isExport {
						continue
					}

					// Find matching declaration
					for _, decl := range decls {
						// Skip if reference is after declaration
						if ref.pos >= decl.pos {
							// But check if it's in initializer
							if !isInInitializer(ref.node, decl.node) {
								continue
							}
						}

						// Check if we should report based on declaration type
						shouldReport := false

						if decl.isFunction && getBool(opts.Functions) {
							shouldReport = true
						} else if decl.isClass && getBool(opts.Classes) {
							shouldReport = true
						} else if decl.isEnum && getBool(opts.Enums) {
							shouldReport = true
						} else if decl.isTypedef && getBool(opts.Typedefs) {
							shouldReport = true
						} else if decl.isVariable && getBool(opts.Variables) {
							shouldReport = true
						}

						if shouldReport {
							ctx.ReportNode(ref.node, buildNoUseBeforeDefineMessage(ref.name))
							break
						}
					}
				}
			},
		}
	},
})

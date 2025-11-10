package no_extraneous_class

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoExtraneousClassOptions struct {
	AllowConstructorOnly bool `json:"allowConstructorOnly"`
	AllowEmpty           bool `json:"allowEmpty"`
	AllowStaticOnly      bool `json:"allowStaticOnly"`
	AllowWithDecorator   bool `json:"allowWithDecorator"`
}

var NoExtraneousClassRule = rule.CreateRule(rule.Rule{
	Name: "no-extraneous-class",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoExtraneousClassOptions{
			AllowConstructorOnly: false,
			AllowEmpty:           false,
			AllowStaticOnly:      false,
			AllowWithDecorator:   false,
		}

		// Parse options with dual-format support
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			// Handle array format: [{ option: value }]
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				// Handle direct object format: { option: value }
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if allowConstructorOnly, ok := optsMap["allowConstructorOnly"].(bool); ok {
					opts.AllowConstructorOnly = allowConstructorOnly
				}
				if allowEmpty, ok := optsMap["allowEmpty"].(bool); ok {
					opts.AllowEmpty = allowEmpty
				}
				if allowStaticOnly, ok := optsMap["allowStaticOnly"].(bool); ok {
					opts.AllowStaticOnly = allowStaticOnly
				}
				if allowWithDecorator, ok := optsMap["allowWithDecorator"].(bool); ok {
					opts.AllowWithDecorator = allowWithDecorator
				}
			}
		}

		checkClass := func(node *ast.Node) {
			// Get the node to report on (prefer name, fallback to class node)
			reportNode := node.Name()
			if reportNode == nil {
				reportNode = node
			}

			// Check if class extends another class - these are always valid
			var heritageClauses *ast.NodeList
			if classDecl := node.AsClassDeclaration(); classDecl != nil {
				heritageClauses = classDecl.HeritageClauses
			} else if classExpr := node.AsClassExpression(); classExpr != nil {
				heritageClauses = classExpr.HeritageClauses
			}

			if heritageClauses != nil {
				for _, clause := range heritageClauses.Nodes {
					heritageClause := clause.AsHeritageClause()
					if heritageClause != nil && heritageClause.Token == ast.KindExtendsKeyword {
						// Check if there are actually types being extended
						if heritageClause.Types != nil && len(heritageClause.Types.Nodes) > 0 {
							// Class extends another class - always valid
							return
						}
					}
				}
			}

			// Check for decorators
			hasDecorators := false
			if node.Modifiers() != nil {
				for _, modifier := range node.Modifiers().Nodes {
					if modifier.Kind == ast.KindDecorator {
						hasDecorators = true
						break
					}
				}
			}

			if hasDecorators && opts.AllowWithDecorator {
				return
			}

			// Check class members
			hasNonStaticMember := false
			hasConstructor := false
			hasStaticMember := false
			isEmpty := true

			members := node.Members()
			if members != nil {
				isEmpty = len(members) == 0

				for _, member := range members {
					// Check if it's a constructor
					if member.Kind == ast.KindConstructor {
						hasConstructor = true
						isEmpty = false

						// Check if constructor has parameter properties (public, private, protected params)
						// These act as class members
						constructor := member.AsConstructorDeclaration()
						if constructor != nil && constructor.Parameters != nil {
							for _, param := range constructor.Parameters.Nodes {
								if param.Kind == ast.KindParameter {
									paramDecl := param.AsParameterDeclaration()
									if paramDecl != nil && paramDecl.Modifiers() != nil {
										for _, mod := range paramDecl.Modifiers().Nodes {
											if mod.Kind == ast.KindPublicKeyword ||
												mod.Kind == ast.KindPrivateKeyword ||
												mod.Kind == ast.KindProtectedKeyword ||
												mod.Kind == ast.KindReadonlyKeyword {
												// This is a parameter property, counts as a non-static member
												hasNonStaticMember = true
												break
											}
										}
									}
								}
							}
						}
						continue
					}

					// Check for static members (properties, methods, and accessors)
					isStatic := false
					isAbstractMember := false

					// Helper to check modifiers
					checkModifiers := func(modifiers *ast.ModifierList) {
						if modifiers != nil {
							for _, mod := range modifiers.Nodes {
								if mod.Kind == ast.KindStaticKeyword {
									isStatic = true
								}
								if mod.Kind == ast.KindAbstractKeyword {
									isAbstractMember = true
								}
							}
						}
					}

					switch member.Kind {
					case ast.KindPropertyDeclaration:
						prop := member.AsPropertyDeclaration()
						if prop != nil {
							checkModifiers(prop.Modifiers())
						}
					case ast.KindMethodDeclaration:
						method := member.AsMethodDeclaration()
						if method != nil {
							checkModifiers(method.Modifiers())
						}
					case ast.KindGetAccessor:
						getter := member.AsGetAccessorDeclaration()
						if getter != nil {
							checkModifiers(getter.Modifiers())
						}
					case ast.KindSetAccessor:
						setter := member.AsSetAccessorDeclaration()
						if setter != nil {
							checkModifiers(setter.Modifiers())
						}
					default:
						// For any other member types, treat as non-static
						hasNonStaticMember = true
						isEmpty = false
						continue
					}

					// Abstract members (non-static) make the class valid
					if isAbstractMember && !isStatic {
						hasNonStaticMember = true
						isEmpty = false
					} else if isStatic {
						hasStaticMember = true
						isEmpty = false
					} else {
						hasNonStaticMember = true
						isEmpty = false
					}
				}
			}

			// Report empty class
			if isEmpty {
				if !opts.AllowEmpty {
					ctx.ReportNode(reportNode, rule.RuleMessage{
						Id:          "empty",
						Description: "Unexpected empty class.",
					})
				}
				return
			}

			// Report constructor-only class
			if hasConstructor && !hasNonStaticMember && !hasStaticMember {
				if !opts.AllowConstructorOnly {
					ctx.ReportNode(reportNode, rule.RuleMessage{
						Id:          "onlyConstructor",
						Description: "Unexpected class with only a constructor.",
					})
				}
				return
			}

			// Report static-only class
			// A class is static-only if it has static members but no non-static members
			// (constructor without parameter properties doesn't count as non-static)
			if hasStaticMember && !hasNonStaticMember {
				if !opts.AllowStaticOnly {
					ctx.ReportNode(reportNode, rule.RuleMessage{
						Id:          "onlyStatic",
						Description: "Unexpected class with only static properties.",
					})
				}
				return
			}
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: checkClass,
			ast.KindClassExpression:  checkClass,
		}
	},
})

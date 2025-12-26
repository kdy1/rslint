package prefer_readonly

import (
	"encoding/json"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type PreferReadonlyOptions struct {
	OnlyInlineLambdas *bool `json:"onlyInlineLambdas"`
}

func buildPreferReadonlyMessage(memberName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferReadonly",
		Description: "Member '" + memberName + "' is never reassigned; mark it as `readonly`.",
		Data: map[string]interface{}{
			"name": memberName,
		},
	}
}

type memberInfo struct {
	node                  *ast.Node
	name                  string
	isStatic              bool
	hasInitializer        bool
	isLambda              bool
	assignedInConstructor bool
	modifiedElsewhere     bool
	symbol                *ast.Symbol
}

var PreferReadonlyRule = rule.CreateRule(rule.Rule{
	Name: "prefer-readonly",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := PreferReadonlyOptions{
			OnlyInlineLambdas: utils.Ref(false),
		}

		if options != nil {
			optionsArray, _ := options.([]interface{})
			if len(optionsArray) > 0 {
				optsJSON, err := json.Marshal(optionsArray[0])
				if err == nil {
					json.Unmarshal(optsJSON, &opts)
				}
			}
		}

		onlyInlineLambdas := opts.OnlyInlineLambdas != nil && *opts.OnlyInlineLambdas

		return rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				if !ast.IsClassDeclaration(node) && !ast.IsClassExpression(node) {
					return
				}

				classNode := node
				var members []*ast.Node
				if ast.IsClassDeclaration(node) {
					members = node.AsClassDeclaration().Members
				} else {
					members = node.AsClassExpression().Members
				}

				// Track private members
				privateMembers := make(map[string]*memberInfo)

				// First pass: identify all private members
				for _, member := range members {
					if ast.IsPropertyDeclaration(member) {
						prop := member.AsPropertyDeclaration()

						// Skip if not private
						isPrivate := false
						for _, modifier := range prop.Modifiers {
							if modifier.Kind == ast.KindPrivateKeyword {
								isPrivate = true
								break
							}
						}

						// Also check for private identifier (#field)
						if !isPrivate && prop.Name != nil && ast.IsPrivateIdentifier(prop.Name) {
							isPrivate = true
						}

						if !isPrivate {
							continue
						}

						// Skip if already readonly
						hasReadonly := false
						for _, modifier := range prop.Modifiers {
							if modifier.Kind == ast.KindReadonlyKeyword {
								hasReadonly = true
								break
							}
						}
						if hasReadonly {
							continue
						}

						// Skip accessor fields
						hasAccessor := false
						for _, modifier := range prop.Modifiers {
							if modifier.Kind == ast.KindAccessorKeyword {
								hasAccessor = true
								break
							}
						}
						if hasAccessor {
							continue
						}

						// Skip computed property names
						if prop.Name != nil && ast.IsComputedPropertyName(prop.Name) {
							continue
						}

						// Get member name
						memberName := ""
						if prop.Name != nil {
							if ast.IsIdentifier(prop.Name) {
								memberName = prop.Name.AsIdentifier().EscapedText
							} else if ast.IsPrivateIdentifier(prop.Name) {
								memberName = "#" + prop.Name.AsPrivateIdentifier().EscapedText
							}
						}

						if memberName == "" {
							continue
						}

						// Check if static
						isStatic := false
						for _, modifier := range prop.Modifiers {
							if modifier.Kind == ast.KindStaticKeyword {
								isStatic = true
								break
							}
						}

						// Check if has initializer
						hasInitializer := prop.Initializer != nil

						// Check if initializer is a lambda/arrow function
						isLambda := false
						if hasInitializer && ast.IsArrowFunction(prop.Initializer) {
							isLambda = true
						}

						// Skip if onlyInlineLambdas is true and this is not a lambda
						if onlyInlineLambdas && !isLambda {
							continue
						}

						// Get symbol for this property
						var symbol *ast.Symbol
						if ctx.TypeChecker != nil && prop.Name != nil {
							symbol = ctx.TypeChecker.GetSymbolAtLocation(prop.Name)
						}

						privateMembers[memberName] = &memberInfo{
							node:                  member,
							name:                  memberName,
							isStatic:              isStatic,
							hasInitializer:        hasInitializer,
							isLambda:              isLambda,
							assignedInConstructor: false,
							modifiedElsewhere:     false,
							symbol:                symbol,
						}
					} else if ast.IsConstructorDeclaration(member) {
						// Check for parameter properties
						ctor := member.AsConstructorDeclaration()
						for _, param := range ctor.Parameters {
							if !ast.IsParameterDeclaration(param) {
								continue
							}

							paramDecl := param.AsParameterDeclaration()

							// Check if parameter has private modifier
							isPrivate := false
							hasReadonly := false
							for _, modifier := range paramDecl.Modifiers {
								if modifier.Kind == ast.KindPrivateKeyword {
									isPrivate = true
								}
								if modifier.Kind == ast.KindReadonlyKeyword {
									hasReadonly = true
								}
							}

							if !isPrivate || hasReadonly {
								continue
							}

							// Get parameter name
							paramName := ""
							if paramDecl.Name != nil && ast.IsIdentifier(paramDecl.Name) {
								paramName = paramDecl.Name.AsIdentifier().EscapedText
							}

							if paramName == "" {
								continue
							}

							// Get symbol for this parameter
							var symbol *ast.Symbol
							if ctx.TypeChecker != nil && paramDecl.Name != nil {
								symbol = ctx.TypeChecker.GetSymbolAtLocation(paramDecl.Name)
							}

							privateMembers[paramName] = &memberInfo{
								node:                  param,
								name:                  paramName,
								isStatic:              false,
								hasInitializer:        paramDecl.Initializer != nil,
								isLambda:              false,
								assignedInConstructor: false,
								modifiedElsewhere:     false,
								symbol:                symbol,
							}
						}
					}
				}

				if len(privateMembers) == 0 {
					return
				}

				// Second pass: check for modifications
				var checkNode func(*ast.Node, bool)
				checkNode = func(node *ast.Node, inConstructor bool) {
					if node == nil {
						return
					}

					// Check for assignments
					if ast.IsBinaryExpression(node) {
						binExpr := node.AsBinaryExpression()
						if isAssignmentOperator(binExpr.OperatorToken.Kind) {
							checkModification(binExpr.Left, privateMembers, classNode, inConstructor)
						}
					}

					// Check for increment/decrement operators
					if ast.IsPrefixUnaryExpression(node) {
						unaryExpr := node.AsPrefixUnaryExpression()
						if unaryExpr.Operator == ast.KindPlusPlusToken || unaryExpr.Operator == ast.KindMinusMinusToken {
							checkModification(unaryExpr.Operand, privateMembers, classNode, inConstructor)
						}
					}

					if ast.IsPostfixUnaryExpression(node) {
						unaryExpr := node.AsPostfixUnaryExpression()
						if unaryExpr.Operator == ast.KindPlusPlusToken || unaryExpr.Operator == ast.KindMinusMinusToken {
							checkModification(unaryExpr.Operand, privateMembers, classNode, inConstructor)
						}
					}

					// Check for delete operator
					if ast.IsDeleteExpression(node) {
						deleteExpr := node.AsDeleteExpression()
						checkModification(deleteExpr.Expression, privateMembers, classNode, inConstructor)
					}

					// Check for destructuring
					if ast.IsArrayBindingPattern(node) || ast.IsObjectBindingPattern(node) {
						// In destructuring patterns, if a private member appears, it's being modified
						checkDestructuringPattern(node, privateMembers, classNode, inConstructor)
					}

					// Recursively check children
					ast.ForEachChild(node, func(child *ast.Node) {
						checkNode(child, inConstructor)
					})
				}

				// Check constructors separately
				for _, member := range members {
					if ast.IsConstructorDeclaration(member) {
						ctor := member.AsConstructorDeclaration()
						if ctor.Body != nil {
							checkNode(ctor.Body, true)
						}
					}
				}

				// Check other members
				for _, member := range members {
					if ast.IsConstructorDeclaration(member) {
						continue
					}
					checkNode(member, false)
				}

				// Third pass: report members that are never modified (or only in constructor)
				for _, info := range privateMembers {
					if !info.modifiedElsewhere {
						// Report and suggest adding readonly
						addReadonlyFix(ctx, info)
					}
				}
			},
			ast.KindClassExpression: func(node *ast.Node) {
				// Reuse the same logic for class expressions
				if listener := PreferReadonlyRule.Run(ctx, options)[ast.KindClassDeclaration]; listener != nil {
					listener(node)
				}
			},
		}
	},
})

func isAssignmentOperator(kind ast.SyntaxKind) bool {
	switch kind {
	case ast.KindEqualsToken,
		ast.KindPlusEqualsToken,
		ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken,
		ast.KindAsteriskAsteriskEqualsToken,
		ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken:
		return true
	}
	return false
}

func checkModification(expr *ast.Expression, members map[string]*memberInfo, classNode *ast.Node, inConstructor bool) {
	if expr == nil {
		return
	}

	memberName, isStatic := getMemberAccess(expr, classNode)
	if memberName == "" {
		return
	}

	if info, ok := members[memberName]; ok {
		if info.isStatic != isStatic {
			return
		}

		if inConstructor {
			info.assignedInConstructor = true
		} else {
			info.modifiedElsewhere = true
		}
	}
}

func checkDestructuringPattern(pattern *ast.Node, members map[string]*memberInfo, classNode *ast.Node, inConstructor bool) {
	if pattern == nil {
		return
	}

	ast.ForEachChild(pattern, func(child *ast.Node) {
		if ast.IsBindingElement(child) {
			elem := child.AsBindingElement()
			if elem.PropertyName != nil {
				// Check if property name references a member
				if ast.IsPropertyAccessExpression(elem.PropertyName) || ast.IsElementAccessExpression(elem.PropertyName) {
					memberName, isStatic := getMemberAccess(elem.PropertyName, classNode)
					if memberName != "" {
						if info, ok := members[memberName]; ok && info.isStatic == isStatic {
							if inConstructor {
								info.assignedInConstructor = true
							} else {
								info.modifiedElsewhere = true
							}
						}
					}
				}
			}
		}
		checkDestructuringPattern(child, members, classNode, inConstructor)
	})
}

func getMemberAccess(expr *ast.Expression, classNode *ast.Node) (memberName string, isStatic bool) {
	if expr == nil {
		return "", false
	}

	// Handle property access (this.member or ClassName.member)
	if ast.IsPropertyAccessExpression(expr) {
		propAccess := expr.AsPropertyAccessExpression()

		// Check for this.member
		if ast.IsThisKeyword(propAccess.Expression) {
			if ast.IsIdentifier(propAccess.Name) {
				return propAccess.Name.AsIdentifier().EscapedText, false
			} else if ast.IsPrivateIdentifier(propAccess.Name) {
				return "#" + propAccess.Name.AsPrivateIdentifier().EscapedText, false
			}
		}

		// Check for ClassName.staticMember
		if ast.IsIdentifier(propAccess.Expression) {
			// Get class name
			className := ""
			if ast.IsClassDeclaration(classNode) {
				classDecl := classNode.AsClassDeclaration()
				if classDecl.Name != nil && ast.IsIdentifier(classDecl.Name) {
					className = classDecl.Name.AsIdentifier().EscapedText
				}
			}

			if className != "" && propAccess.Expression.AsIdentifier().EscapedText == className {
				if ast.IsIdentifier(propAccess.Name) {
					return propAccess.Name.AsIdentifier().EscapedText, true
				} else if ast.IsPrivateIdentifier(propAccess.Name) {
					return "#" + propAccess.Name.AsPrivateIdentifier().EscapedText, true
				}
			}
		}
	}

	// Handle element access (this['member'] or this.#member)
	if ast.IsElementAccessExpression(expr) {
		elemAccess := expr.AsElementAccessExpression()

		if ast.IsThisKeyword(elemAccess.Expression) {
			// Try to get the property name from the argument
			if elemAccess.ArgumentExpression != nil && ast.IsStringLiteral(elemAccess.ArgumentExpression) {
				return elemAccess.ArgumentExpression.AsStringLiteral().Text, false
			}
		}
	}

	return "", false
}

func addReadonlyFix(ctx rule.RuleContext, info *memberInfo) {
	if info.node == nil {
		return
	}

	var fixes []rule.RuleFix
	var prop *ast.PropertyDeclaration
	var paramDecl *ast.ParameterDeclaration
	isParam := false

	if ast.IsPropertyDeclaration(info.node) {
		prop = info.node.AsPropertyDeclaration()
	} else if ast.IsParameterDeclaration(info.node) {
		paramDecl = info.node.AsParameterDeclaration()
		isParam = true
	} else {
		return
	}

	if isParam {
		// For parameter properties, insert readonly after private
		insertPos := paramDecl.Pos()

		// Find the private keyword position
		for _, modifier := range paramDecl.Modifiers {
			if modifier.Kind == ast.KindPrivateKeyword {
				insertPos = modifier.End()
				break
			}
		}

		fixes = append(fixes, rule.RuleFixInsertTextAt(insertPos, " readonly"))
	} else {
		// For property declarations
		if info.isStatic {
			// For static properties, insert readonly after static
			insertPos := prop.Pos()

			// Find position after static keyword
			for _, modifier := range prop.Modifiers {
				if modifier.Kind == ast.KindStaticKeyword {
					insertPos = modifier.End()
					break
				}
			}

			// For private identifier (#field), don't include private keyword
			if ast.IsPrivateIdentifier(prop.Name) {
				fixes = append(fixes, rule.RuleFixInsertTextAt(insertPos, " readonly"))
			} else {
				// Find position after private keyword
				for _, modifier := range prop.Modifiers {
					if modifier.Kind == ast.KindPrivateKeyword {
						insertPos = modifier.End()
						break
					}
				}
				fixes = append(fixes, rule.RuleFixInsertTextAt(insertPos, " readonly"))
			}
		} else {
			// For instance properties
			if ast.IsPrivateIdentifier(prop.Name) {
				// For #field syntax, insert readonly before the #
				insertPos := prop.Pos()
				// Skip past any modifiers
				if len(prop.Modifiers) > 0 {
					insertPos = prop.Modifiers[len(prop.Modifiers)-1].End()
					fixes = append(fixes, rule.RuleFixInsertTextAt(insertPos, " readonly"))
				} else {
					fixes = append(fixes, rule.RuleFixInsertTextAt(insertPos, "readonly "))
				}
			} else {
				// For private keyword syntax, insert readonly after private
				insertPos := prop.Pos()

				for _, modifier := range prop.Modifiers {
					if modifier.Kind == ast.KindPrivateKeyword {
						insertPos = modifier.End()
						break
					}
				}

				fixes = append(fixes, rule.RuleFixInsertTextAt(insertPos, " readonly"))
			}
		}

		// If the property is modified in constructor and has an initializer, add type annotation
		if info.assignedInConstructor && info.hasInitializer && prop.Type == nil {
			// We need to add a type annotation
			typeAnnotation := inferTypeFromInitializer(ctx, prop.Initializer)
			if typeAnnotation != "" {
				// Find position after property name
				nameEnd := prop.Name.End()
				fixes = append(fixes, rule.RuleFixInsertTextAt(nameEnd, ": "+typeAnnotation))
			}
		}
	}

	ctx.ReportNodeWithFixes(
		info.node,
		buildPreferReadonlyMessage(info.name),
		fixes...,
	)
}

func inferTypeFromInitializer(ctx rule.RuleContext, initializer *ast.Expression) string {
	if initializer == nil {
		return ""
	}

	// Use type checker to infer type
	if ctx.TypeChecker != nil {
		t := ctx.TypeChecker.GetTypeAtLocation(initializer)
		if t != nil {
			typeString := ctx.TypeChecker.TypeToString(t, nil, 0)
			// Clean up the type string
			typeString = strings.TrimSpace(typeString)
			return typeString
		}
	}

	// Fallback to simple inference
	if ast.IsNumericLiteral(initializer) {
		return "number"
	}
	if ast.IsStringLiteral(initializer) {
		return "string"
	}
	if initializer.Kind == ast.KindTrueKeyword || initializer.Kind == ast.KindFalseKeyword {
		return "boolean"
	}
	if initializer.Kind == ast.KindNullKeyword {
		return "null"
	}

	return ""
}

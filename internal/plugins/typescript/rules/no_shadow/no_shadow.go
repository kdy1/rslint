package no_shadow

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type NoShadowOptions struct {
	BuiltinGlobals                               bool   `json:"builtinGlobals"`
	Hoist                                        string `json:"hoist"`
	IgnoreTypeValueShadow                        bool   `json:"ignoreTypeValueShadow"`
	IgnoreFunctionTypeParameterNameValueShadow   bool   `json:"ignoreFunctionTypeParameterNameValueShadow"`
	IgnoreOnInitialization                       bool   `json:"ignoreOnInitialization"`
}

type variable struct {
	name         string
	identifierNode *ast.Node
	isType       bool
	isFunctionTypeParameter bool
}

type scope struct {
	variables map[string]*variable
	upper     *scope
	isType    bool
}

var NoShadowRule = rule.CreateRule(rule.Rule{
	Name: "no-shadow",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoShadowOptions{
			BuiltinGlobals:                             false,
			Hoist:                                      "functions-and-types",
			IgnoreTypeValueShadow:                      true,
			IgnoreFunctionTypeParameterNameValueShadow: true,
			IgnoreOnInitialization:                     false,
		}

		// Parse options
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if builtinGlobals, ok := optsMap["builtinGlobals"].(bool); ok {
					opts.BuiltinGlobals = builtinGlobals
				}
				if hoist, ok := optsMap["hoist"].(string); ok {
					opts.Hoist = hoist
				}
				if ignoreTypeValueShadow, ok := optsMap["ignoreTypeValueShadow"].(bool); ok {
					opts.IgnoreTypeValueShadow = ignoreTypeValueShadow
				}
				if ignoreFunctionTypeParameterNameValueShadow, ok := optsMap["ignoreFunctionTypeParameterNameValueShadow"].(bool); ok {
					opts.IgnoreFunctionTypeParameterNameValueShadow = ignoreFunctionTypeParameterNameValueShadow
				}
				if ignoreOnInitialization, ok := optsMap["ignoreOnInitialization"].(bool); ok {
					opts.IgnoreOnInitialization = ignoreOnInitialization
				}
			}
		}

		var currentScope *scope

		// Track global variables from languageOptions.globals in test cases
		globalVars := make(map[string]bool)

		// Helper to get variable name from identifier
		getIdentifierName := func(node *ast.Node) string {
			if node == nil {
				return ""
			}
			if node.Kind == ast.KindIdentifier {
				if ident := node.AsIdentifier(); ident != nil {
					textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
					return ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
				}
			}
			return ""
		}

		// Helper to create a new scope
		createScope := func(isTypeScope bool) {
			currentScope = &scope{
				variables: make(map[string]*variable),
				upper:     currentScope,
				isType:    isTypeScope,
			}
		}

		// Helper to exit a scope
		exitScope := func() {
			if currentScope != nil {
				currentScope = currentScope.upper
			}
		}

		// Helper to declare a variable in current scope
		declareVariable := func(name string, node *ast.Node, isType bool, isFunctionTypeParam bool) {
			if currentScope == nil || name == "" {
				return
			}
			currentScope.variables[name] = &variable{
				name:                    name,
				identifierNode:          node,
				isType:                  isType,
				isFunctionTypeParameter: isFunctionTypeParam,
			}
		}

		// Helper to check if variable shadows an outer scope variable
		checkShadowing := func(name string, node *ast.Node, isType bool, isFunctionTypeParam bool) {
			if name == "" || currentScope == nil {
				return
			}

			// Check if it's in an initialization context
			if opts.IgnoreOnInitialization && isInInitialization(node) {
				return
			}

			// Look for shadowed variable in outer scopes
			var shadowedVar *variable
			upperScope := currentScope.upper

			for upperScope != nil {
				if v, exists := upperScope.variables[name]; exists {
					shadowedVar = v
					break
				}
				upperScope = upperScope.upper
			}

			// Also check global variables
			isGlobal := false
			if shadowedVar == nil && globalVars[name] {
				isGlobal = true
			}

			if shadowedVar != nil || isGlobal {
				// Apply type-value shadowing rules
				if opts.IgnoreTypeValueShadow {
					if shadowedVar != nil && isType != shadowedVar.isType {
						// Type shadowing value or value shadowing type is allowed
						return
					}
				}

				// Apply function type parameter shadowing rules
				if opts.IgnoreFunctionTypeParameterNameValueShadow && isFunctionTypeParam {
					// Function type parameters can shadow outer values
					return
				}

				// Report shadowing
				if isGlobal {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "noShadowGlobal",
						Description: "'" + name + "' is already declared in the upper scope.",
					})
				} else if shadowedVar != nil {
					// Get position of shadowed variable
					shadowedRange := utils.TrimNodeTextRange(ctx.SourceFile, shadowedVar.identifierNode)
					shadowedLine := getLineNumber(ctx.SourceFile, shadowedRange.Pos())
					shadowedColumn := getColumnNumber(ctx.SourceFile, shadowedRange.Pos())

					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "noShadow",
						Description: "'" + name + "' is already declared in the upper scope on line " +
									 string(rune(shadowedLine)) + ":" + string(rune(shadowedColumn)) + ".",
					})
				}
			}
		}

		// Helper to check if node is in initialization context
		isInInitialization := func(node *ast.Node) bool {
			// Walk up the tree to check if we're in an initialization
			parent := node.Parent
			for parent != nil {
				switch parent.Kind {
				case ast.KindVariableDeclaration:
					// Check if this is part of the initializer
					varDecl := parent.AsVariableDeclaration()
					if varDecl != nil && varDecl.DeclarationList != nil {
						for _, decl := range varDecl.DeclarationList.Declarations.Nodes {
							if vd := decl.AsVariableDeclaration(); vd != nil {
								if vd.Initializer != nil && contains(vd.Initializer, node) {
									return true
								}
							}
						}
					}
				case ast.KindCallExpression:
					// Check for array methods like .map, .filter, .find
					callExpr := parent.AsCallExpression()
					if callExpr != nil {
						if propAccess := callExpr.Expression.AsPropertyAccessExpression(); propAccess != nil {
							nameNode := propAccess.Name
							if nameNode != nil {
								methodName := getIdentifierName(nameNode)
								if methodName == "map" || methodName == "filter" || methodName == "find" ||
									methodName == "forEach" || methodName == "some" || methodName == "every" ||
									methodName == "reduce" || methodName == "reduceRight" {
									return true
								}
							}
						}
					}
				case ast.KindBinaryExpression:
					// Check for logical operators
					binExpr := parent.AsBinaryExpression()
					if binExpr != nil {
						if binExpr.OperatorToken.Kind == ast.KindBarBarToken ||
							binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
							return true
						}
					}
				}
				parent = parent.Parent
			}
			return false
		}

		// Helper to check if a node contains another node
		contains := func(parent *ast.Node, child *ast.Node) bool {
			if parent == nil || child == nil {
				return false
			}
			current := child
			for current != nil {
				if current == parent {
					return true
				}
				current = current.Parent
			}
			return false
		}

		// Helper functions for line/column numbers
		getLineNumber := func(file *ast.SourceFile, pos int) int {
			lineStarts := file.GetLineStarts()
			for i := len(lineStarts) - 1; i >= 0; i-- {
				if pos >= lineStarts[i] {
					return i + 1
				}
			}
			return 1
		}

		getColumnNumber := func(file *ast.SourceFile, pos int) int {
			lineStarts := file.GetLineStarts()
			for i := len(lineStarts) - 1; i >= 0; i-- {
				if pos >= lineStarts[i] {
					return pos - lineStarts[i] + 1
				}
			}
			return 1
		}

		// Helper to handle hoisting
		shouldHoist := func(isType bool, isFunction bool) bool {
			switch opts.Hoist {
			case "all":
				return true
			case "never":
				return false
			case "functions":
				return isFunction
			case "types":
				return isType
			case "functions-and-types":
				return isFunction || isType
			default:
				return isFunction || isType
			}
		}

		// Check if we're in a .d.ts file
		isInDtsFile := func() bool {
			return strings.HasSuffix(ctx.SourceFile.FileName(), ".d.ts")
		}

		// Check if we're in a global augmentation
		isInGlobalAugmentation := func(node *ast.Node) bool {
			parent := node.Parent
			for parent != nil {
				if parent.Kind == ast.KindModuleDeclaration {
					modDecl := parent.AsModuleDeclaration()
					if modDecl != nil && modDecl.Name != nil {
						name := getIdentifierName(modDecl.Name)
						if name == "global" {
							return true
						}
					}
				}
				parent = parent.Parent
			}
			return false
		}

		// Check if identifier is a function type parameter
		isFunctionTypeParameter := func(node *ast.Node) bool {
			if node == nil || node.Parent == nil {
				return false
			}

			parent := node.Parent

			// Check various function type signature contexts
			switch parent.Kind {
			case ast.KindParameter:
				// Walk up to find the function signature
				current := parent.Parent
				for current != nil {
					switch current.Kind {
					case ast.KindFunctionType,
						ast.KindConstructorType,
						ast.KindCallSignature,
						ast.KindConstructSignature,
						ast.KindMethodSignature,
						ast.KindFunctionDeclaration:
						// Check if this is a declare function
						if current.Kind == ast.KindFunctionDeclaration {
							funcDecl := current.AsFunctionDeclaration()
							if funcDecl != nil {
								modifiers := funcDecl.Modifiers()
								if modifiers != nil {
									for _, mod := range modifiers.Nodes {
										if mod.Kind == ast.KindDeclareKeyword {
											return true
										}
									}
								}
							}
							return false
						}
						return true
					}
					current = current.Parent
				}
			}
			return false
		}

		// Create global scope
		createScope(false)

		return rule.RuleListeners{
			// Block scopes
			ast.KindBlock: func(node *ast.Node) {
				createScope(false)
			},
			rule.ListenerOnExit(ast.KindBlock): func(node *ast.Node) {
				exitScope()
			},

			// Function scopes
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				funcDecl := node.AsFunctionDeclaration()
				if funcDecl == nil {
					return
				}

				// Declare function name in current scope before entering function scope
				if funcDecl.Name != nil && shouldHoist(false, true) {
					name := getIdentifierName(funcDecl.Name)
					checkShadowing(name, funcDecl.Name, false, false)
					declareVariable(name, funcDecl.Name, false, false)
				}

				createScope(false)

				// Declare type parameters
				if funcDecl.TypeParameters != nil {
					for _, tp := range funcDecl.TypeParameters.Nodes {
						if typeParam := tp.AsTypeParameterDeclaration(); typeParam != nil {
							if typeParam.Name != nil {
								name := getIdentifierName(typeParam.Name)
								checkShadowing(name, typeParam.Name, true, false)
								declareVariable(name, typeParam.Name, true, false)
							}
						}
					}
				}

				// Declare parameters
				if funcDecl.Parameters != nil {
					for _, p := range funcDecl.Parameters.Nodes {
						if param := p.AsParameterDeclaration(); param != nil {
							if param.Name != nil {
								name := getIdentifierName(param.Name)
								checkShadowing(name, param.Name, false, false)
								declareVariable(name, param.Name, false, false)
							}
						}
					}
				}
			},
			rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) {
				exitScope()
			},

			ast.KindFunctionExpression: func(node *ast.Node) {
				funcExpr := node.AsFunctionExpression()
				if funcExpr == nil {
					return
				}

				createScope(false)

				// Declare function name if present
				if funcExpr.Name != nil {
					name := getIdentifierName(funcExpr.Name)
					declareVariable(name, funcExpr.Name, false, false)
				}

				// Declare type parameters
				if funcExpr.TypeParameters != nil {
					for _, tp := range funcExpr.TypeParameters.Nodes {
						if typeParam := tp.AsTypeParameterDeclaration(); typeParam != nil {
							if typeParam.Name != nil {
								name := getIdentifierName(typeParam.Name)
								checkShadowing(name, typeParam.Name, true, false)
								declareVariable(name, typeParam.Name, true, false)
							}
						}
					}
				}

				// Declare parameters
				if funcExpr.Parameters != nil {
					for _, p := range funcExpr.Parameters.Nodes {
						if param := p.AsParameterDeclaration(); param != nil {
							if param.Name != nil {
								name := getIdentifierName(param.Name)
								checkShadowing(name, param.Name, false, false)
								declareVariable(name, param.Name, false, false)
							}
						}
					}
				}
			},
			rule.ListenerOnExit(ast.KindFunctionExpression): func(node *ast.Node) {
				exitScope()
			},

			ast.KindArrowFunction: func(node *ast.Node) {
				arrowFunc := node.AsArrowFunction()
				if arrowFunc == nil {
					return
				}

				createScope(false)

				// Declare type parameters
				if arrowFunc.TypeParameters != nil {
					for _, tp := range arrowFunc.TypeParameters.Nodes {
						if typeParam := tp.AsTypeParameterDeclaration(); typeParam != nil {
							if typeParam.Name != nil {
								name := getIdentifierName(typeParam.Name)
								checkShadowing(name, typeParam.Name, true, false)
								declareVariable(name, typeParam.Name, true, false)
							}
						}
					}
				}

				// Declare parameters
				if arrowFunc.Parameters != nil {
					for _, p := range arrowFunc.Parameters.Nodes {
						if param := p.AsParameterDeclaration(); param != nil {
							if param.Name != nil {
								name := getIdentifierName(param.Name)
								checkShadowing(name, param.Name, false, false)
								declareVariable(name, param.Name, false, false)
							}
						}
					}
				}
			},
			rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
				exitScope()
			},

			// Method declarations
			ast.KindMethodDeclaration: func(node *ast.Node) {
				methodDecl := node.AsMethodDeclaration()
				if methodDecl == nil {
					return
				}

				createScope(false)

				// Declare type parameters
				if methodDecl.TypeParameters != nil {
					for _, tp := range methodDecl.TypeParameters.Nodes {
						if typeParam := tp.AsTypeParameterDeclaration(); typeParam != nil {
							if typeParam.Name != nil {
								name := getIdentifierName(typeParam.Name)
								checkShadowing(name, typeParam.Name, true, false)
								declareVariable(name, typeParam.Name, true, false)
							}
						}
					}
				}

				// Declare parameters
				if methodDecl.Parameters != nil {
					for _, p := range methodDecl.Parameters.Nodes {
						if param := p.AsParameterDeclaration(); param != nil {
							if param.Name != nil {
								name := getIdentifierName(param.Name)
								checkShadowing(name, param.Name, false, false)
								declareVariable(name, param.Name, false, false)
							}
						}
					}
				}
			},
			rule.ListenerOnExit(ast.KindMethodDeclaration): func(node *ast.Node) {
				exitScope()
			},

			// Variable declarations
			ast.KindVariableStatement: func(node *ast.Node) {
				varStmt := node.AsVariableStatement()
				if varStmt == nil || varStmt.DeclarationList == nil {
					return
				}

				for _, decl := range varStmt.DeclarationList.Declarations.Nodes {
					if vd := decl.AsVariableDeclaration(); vd != nil {
						if vd.Name != nil {
							name := getIdentifierName(vd.Name)

							// Skip if in global augmentation
							if isInGlobalAugmentation(node) {
								declareVariable(name, vd.Name, false, false)
								continue
							}

							// Skip global checks in .d.ts files with builtinGlobals
							if isInDtsFile() && opts.BuiltinGlobals && globalVars[name] {
								declareVariable(name, vd.Name, false, false)
								continue
							}

							checkShadowing(name, vd.Name, false, false)
							declareVariable(name, vd.Name, false, false)
						}
					}
				}
			},

			// Type aliases
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				typeAlias := node.AsTypeAliasDeclaration()
				if typeAlias == nil || typeAlias.Name == nil {
					return
				}

				name := getIdentifierName(typeAlias.Name)

				// Skip if in global augmentation
				if isInGlobalAugmentation(node) {
					return
				}

				// Skip global checks in .d.ts files with builtinGlobals
				if isInDtsFile() && opts.BuiltinGlobals && globalVars[name] {
					if shouldHoist(true, false) {
						declareVariable(name, typeAlias.Name, true, false)
					}
					return
				}

				if shouldHoist(true, false) {
					checkShadowing(name, typeAlias.Name, true, false)
					declareVariable(name, typeAlias.Name, true, false)
				}

				// Create scope for type parameters
				createScope(true)

				if typeAlias.TypeParameters != nil {
					for _, tp := range typeAlias.TypeParameters.Nodes {
						if typeParam := tp.AsTypeParameterDeclaration(); typeParam != nil {
							if typeParam.Name != nil {
								tpName := getIdentifierName(typeParam.Name)
								checkShadowing(tpName, typeParam.Name, true, false)
								declareVariable(tpName, typeParam.Name, true, false)
							}
						}
					}
				}
			},
			rule.ListenerOnExit(ast.KindTypeAliasDeclaration): func(node *ast.Node) {
				exitScope()
			},

			// Interface declarations
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				interfaceDecl := node.AsInterfaceDeclaration()
				if interfaceDecl == nil || interfaceDecl.Name == nil {
					return
				}

				name := getIdentifierName(interfaceDecl.Name)

				// Skip if in global augmentation
				if isInGlobalAugmentation(node) {
					return
				}

				// Skip global checks in .d.ts files with builtinGlobals
				if isInDtsFile() && opts.BuiltinGlobals && globalVars[name] {
					if shouldHoist(true, false) {
						declareVariable(name, interfaceDecl.Name, true, false)
					}
					return
				}

				if shouldHoist(true, false) {
					checkShadowing(name, interfaceDecl.Name, true, false)
					declareVariable(name, interfaceDecl.Name, true, false)
				}

				// Create scope for type parameters
				createScope(true)

				if interfaceDecl.TypeParameters != nil {
					for _, tp := range interfaceDecl.TypeParameters.Nodes {
						if typeParam := tp.AsTypeParameterDeclaration(); typeParam != nil {
							if typeParam.Name != nil {
								tpName := getIdentifierName(typeParam.Name)
								checkShadowing(tpName, typeParam.Name, true, false)
								declareVariable(tpName, typeParam.Name, true, false)
							}
						}
					}
				}
			},
			rule.ListenerOnExit(ast.KindInterfaceDeclaration): func(node *ast.Node) {
				exitScope()
			},

			// Class declarations
			ast.KindClassDeclaration: func(node *ast.Node) {
				classDecl := node.AsClassDeclaration()
				if classDecl == nil {
					return
				}

				// Declare class name in current scope
				if classDecl.Name != nil && shouldHoist(false, false) {
					name := getIdentifierName(classDecl.Name)

					// Skip if in global augmentation
					if !isInGlobalAugmentation(node) {
						checkShadowing(name, classDecl.Name, false, false)
						declareVariable(name, classDecl.Name, false, false)
					}
				}

				createScope(false)

				// Declare type parameters
				if classDecl.TypeParameters != nil {
					for _, tp := range classDecl.TypeParameters.Nodes {
						if typeParam := tp.AsTypeParameterDeclaration(); typeParam != nil {
							if typeParam.Name != nil {
								name := getIdentifierName(typeParam.Name)
								checkShadowing(name, typeParam.Name, true, false)
								declareVariable(name, typeParam.Name, true, false)
							}
						}
					}
				}
			},
			rule.ListenerOnExit(ast.KindClassDeclaration): func(node *ast.Node) {
				exitScope()
			},

			// Function type (for function type parameters)
			ast.KindFunctionType: func(node *ast.Node) {
				funcType := node.AsFunctionType()
				if funcType == nil {
					return
				}

				createScope(false)

				// Declare type parameters
				if funcType.TypeParameters != nil {
					for _, tp := range funcType.TypeParameters.Nodes {
						if typeParam := tp.AsTypeParameterDeclaration(); typeParam != nil {
							if typeParam.Name != nil {
								name := getIdentifierName(typeParam.Name)
								checkShadowing(name, typeParam.Name, true, false)
								declareVariable(name, typeParam.Name, true, false)
							}
						}
					}
				}

				// Declare parameters (as function type parameters)
				if funcType.Parameters != nil {
					for _, p := range funcType.Parameters.Nodes {
						if param := p.AsParameterDeclaration(); param != nil {
							if param.Name != nil {
								name := getIdentifierName(param.Name)
								isFuncTypeParam := isFunctionTypeParameter(param.Name)
								checkShadowing(name, param.Name, false, isFuncTypeParam)
								declareVariable(name, param.Name, false, isFuncTypeParam)
							}
						}
					}
				}
			},
			rule.ListenerOnExit(ast.KindFunctionType): func(node *ast.Node) {
				exitScope()
			},

			// Constructor type
			ast.KindConstructorType: func(node *ast.Node) {
				constructorType := node.AsConstructorType()
				if constructorType == nil {
					return
				}

				createScope(false)

				// Declare type parameters
				if constructorType.TypeParameters != nil {
					for _, tp := range constructorType.TypeParameters.Nodes {
						if typeParam := tp.AsTypeParameterDeclaration(); typeParam != nil {
							if typeParam.Name != nil {
								name := getIdentifierName(typeParam.Name)
								checkShadowing(name, typeParam.Name, true, false)
								declareVariable(name, typeParam.Name, true, false)
							}
						}
					}
				}

				// Declare parameters
				if constructorType.Parameters != nil {
					for _, p := range constructorType.Parameters.Nodes {
						if param := p.AsParameterDeclaration(); param != nil {
							if param.Name != nil {
								name := getIdentifierName(param.Name)
								isFuncTypeParam := isFunctionTypeParameter(param.Name)
								checkShadowing(name, param.Name, false, isFuncTypeParam)
								declareVariable(name, param.Name, false, isFuncTypeParam)
							}
						}
					}
				}
			},
			rule.ListenerOnExit(ast.KindConstructorType): func(node *ast.Node) {
				exitScope()
			},

			// Call signature
			ast.KindCallSignature: func(node *ast.Node) {
				callSig := node.AsCallSignature()
				if callSig == nil {
					return
				}

				createScope(false)

				// Declare type parameters
				if callSig.TypeParameters != nil {
					for _, tp := range callSig.TypeParameters.Nodes {
						if typeParam := tp.AsTypeParameterDeclaration(); typeParam != nil {
							if typeParam.Name != nil {
								name := getIdentifierName(typeParam.Name)
								checkShadowing(name, typeParam.Name, true, false)
								declareVariable(name, typeParam.Name, true, false)
							}
						}
					}
				}

				// Declare parameters
				if callSig.Parameters != nil {
					for _, p := range callSig.Parameters.Nodes {
						if param := p.AsParameterDeclaration(); param != nil {
							if param.Name != nil {
								name := getIdentifierName(param.Name)
								isFuncTypeParam := isFunctionTypeParameter(param.Name)
								checkShadowing(name, param.Name, false, isFuncTypeParam)
								declareVariable(name, param.Name, false, isFuncTypeParam)
							}
						}
					}
				}
			},
			rule.ListenerOnExit(ast.KindCallSignature): func(node *ast.Node) {
				exitScope()
			},

			// Constructor signature
			ast.KindConstructSignature: func(node *ast.Node) {
				constructSig := node.AsConstructSignature()
				if constructSig == nil {
					return
				}

				createScope(false)

				// Declare type parameters
				if constructSig.TypeParameters != nil {
					for _, tp := range constructSig.TypeParameters.Nodes {
						if typeParam := tp.AsTypeParameterDeclaration(); typeParam != nil {
							if typeParam.Name != nil {
								name := getIdentifierName(typeParam.Name)
								checkShadowing(name, typeParam.Name, true, false)
								declareVariable(name, typeParam.Name, true, false)
							}
						}
					}
				}

				// Declare parameters
				if constructSig.Parameters != nil {
					for _, p := range constructSig.Parameters.Nodes {
						if param := p.AsParameterDeclaration(); param != nil {
							if param.Name != nil {
								name := getIdentifierName(param.Name)
								isFuncTypeParam := isFunctionTypeParameter(param.Name)
								checkShadowing(name, param.Name, false, isFuncTypeParam)
								declareVariable(name, param.Name, false, isFuncTypeParam)
							}
						}
					}
				}
			},
			rule.ListenerOnExit(ast.KindConstructSignature): func(node *ast.Node) {
				exitScope()
			},

			// Method signature
			ast.KindMethodSignature: func(node *ast.Node) {
				methodSig := node.AsMethodSignature()
				if methodSig == nil {
					return
				}

				createScope(false)

				// Declare type parameters
				if methodSig.TypeParameters != nil {
					for _, tp := range methodSig.TypeParameters.Nodes {
						if typeParam := tp.AsTypeParameterDeclaration(); typeParam != nil {
							if typeParam.Name != nil {
								name := getIdentifierName(typeParam.Name)
								checkShadowing(name, typeParam.Name, true, false)
								declareVariable(name, typeParam.Name, true, false)
							}
						}
					}
				}

				// Declare parameters
				if methodSig.Parameters != nil {
					for _, p := range methodSig.Parameters.Nodes {
						if param := p.AsParameterDeclaration(); param != nil {
							if param.Name != nil {
								name := getIdentifierName(param.Name)
								isFuncTypeParam := isFunctionTypeParameter(param.Name)
								checkShadowing(name, param.Name, false, isFuncTypeParam)
								declareVariable(name, param.Name, false, isFuncTypeParam)
							}
						}
					}
				}
			},
			rule.ListenerOnExit(ast.KindMethodSignature): func(node *ast.Node) {
				exitScope()
			},

			// Import declarations
			ast.KindImportDeclaration: func(node *ast.Node) {
				importDecl := node.AsImportDeclaration()
				if importDecl == nil || importDecl.ImportClause == nil {
					return
				}

				clause := importDecl.ImportClause

				// Default import
				if clause.Name != nil {
					name := getIdentifierName(clause.Name)
					declareVariable(name, clause.Name, false, false)
				}

				// Named imports
				if clause.NamedBindings != nil {
					nb := clause.NamedBindings

					// Namespace import
					if nb.Kind == ast.KindNamespaceImport {
						if nsImport := nb.AsNamespaceImport(); nsImport != nil {
							if nsImport.Name != nil {
								name := getIdentifierName(nsImport.Name)
								declareVariable(name, nsImport.Name, false, false)
							}
						}
					}

					// Named imports
					if nb.Kind == ast.KindNamedImports {
						if namedImports := nb.AsNamedImports(); namedImports != nil {
							if namedImports.Elements != nil {
								for _, elem := range namedImports.Elements.Nodes {
									if importSpec := elem.AsImportSpecifier(); importSpec != nil {
										// Determine if this is a type-only import
										isTypeImport := false

										// Check for inline `type` keyword (import { type Foo })
										if importSpec.IsTypeOnly {
											isTypeImport = true
										}

										// Check if parent import is type-only (import type { Foo })
										if importDecl.ImportClause.IsTypeOnly {
											isTypeImport = true
										}

										var nameNode *ast.Node
										if importSpec.PropertyName != nil {
											nameNode = importSpec.Name
										} else {
											nameNode = importSpec.Name
										}

										if nameNode != nil {
											name := getIdentifierName(nameNode)
											declareVariable(name, nameNode, isTypeImport, false)
										}
									}
								}
							}
						}
					}
				}
			},

			// Module declarations
			ast.KindModuleDeclaration: func(node *ast.Node) {
				createScope(false)
			},
			rule.ListenerOnExit(ast.KindModuleDeclaration): func(node *ast.Node) {
				exitScope()
			},

			// Enum declarations
			ast.KindEnumDeclaration: func(node *ast.Node) {
				enumDecl := node.AsEnumDeclaration()
				if enumDecl == nil || enumDecl.Name == nil {
					return
				}

				name := getIdentifierName(enumDecl.Name)

				// Skip if in global augmentation
				if !isInGlobalAugmentation(node) {
					checkShadowing(name, enumDecl.Name, false, false)
					declareVariable(name, enumDecl.Name, false, false)
				}

				createScope(false)
			},
			rule.ListenerOnExit(ast.KindEnumDeclaration): func(node *ast.Node) {
				exitScope()
			},
		}
	},
})

package no_useless_constructor

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Message builders
func buildNoUselessConstructor() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noUselessConstructor",
		Description: "Useless constructor.",
	}
}

func buildRemoveConstructor() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeConstructor",
		Description: "Remove the useless constructor.",
	}
}

// hasParameterProperty checks if any parameter has a TypeScript property modifier
func hasParameterProperty(params *ast.NodeArray) bool {
	if params == nil || len(params.Nodes) == 0 {
		return false
	}

	for _, param := range params.Nodes {
		if param == nil {
			continue
		}

		paramNode := param.AsParameterDeclaration()
		if paramNode == nil {
			continue
		}

		// Check for parameter property modifiers (public, private, protected, readonly)
		if hasModifiers(paramNode.Modifiers, []ast.Kind{
			ast.KindPublicKeyword,
			ast.KindPrivateKeyword,
			ast.KindProtectedKeyword,
			ast.KindReadonlyKeyword,
		}) {
			return true
		}

		// Check for decorators
		if paramNode.Decorators != nil && len(paramNode.Decorators.Nodes) > 0 {
			return true
		}
	}

	return false
}

// hasModifiers checks if the modifiers array contains any of the specified kinds
func hasModifiers(modifiers *ast.NodeArray, kinds []ast.Kind) bool {
	if modifiers == nil || len(modifiers.Nodes) == 0 {
		return false
	}

	for _, modifier := range modifiers.Nodes {
		if modifier == nil {
			continue
		}
		for _, kind := range kinds {
			if modifier.Kind == kind {
				return true
			}
		}
	}

	return false
}

// getClassNode gets the class declaration/expression for a constructor
func getClassNode(constructorNode *ast.Node) *ast.Node {
	if constructorNode == nil {
		return nil
	}

	// Walk up the parent chain to find the class
	current := constructorNode.Parent
	for current != nil {
		if current.Kind == ast.KindClassDeclaration || current.Kind == ast.KindClassExpression {
			return current
		}
		current = current.Parent
	}

	return nil
}

// hasValidExtends checks if a class extends a valid constructor
func hasValidExtends(classNode *ast.Node) bool {
	if classNode == nil {
		return false
	}

	// Get heritage clause (extends clause)
	heritageClauses := utils.GetHeritageClauses(classNode)
	if heritageClauses == nil || len(heritageClauses.Nodes) == 0 {
		return false
	}

	// Look for extends clause (token = ExtendsKeyword)
	for _, clause := range heritageClauses.Nodes {
		if clause == nil {
			continue
		}
		heritageClause := clause.AsHeritageClause()
		if heritageClause == nil {
			continue
		}
		// Check if this is an extends clause
		if heritageClause.Token == ast.KindExtendsKeyword {
			// Check if it has types
			if heritageClause.Types != nil && len(heritageClause.Types.Nodes) > 0 {
				return true
			}
		}
	}

	return false
}

// argsMatch checks if super arguments match constructor parameters exactly
func argsMatch(superArgs *ast.NodeArray, params *ast.NodeArray) bool {
	// Handle super(...arguments) case
	if superArgs != nil && len(superArgs.Nodes) == 1 {
		arg := superArgs.Nodes[0]
		if arg != nil && arg.Kind == ast.KindSpreadElement {
			spread := arg.AsSpreadElement()
			if spread != nil && spread.Expression != nil {
				// Check for "arguments" identifier
				if spread.Expression.Kind == ast.KindIdentifier {
					ident := spread.Expression.AsIdentifier()
					if ident != nil && ident.Text == "arguments" {
						return true
					}
				}
			}
		}
	}

	// If no params, super must have no args
	if params == nil || len(params.Nodes) == 0 {
		return superArgs == nil || len(superArgs.Nodes) == 0
	}

	// If no super args, params must be empty
	if superArgs == nil || len(superArgs.Nodes) == 0 {
		return false
	}

	// Count regular params vs rest params
	regularParams := 0
	hasRest := false
	for i, param := range params.Nodes {
		if param == nil {
			continue
		}
		paramNode := param.AsParameterDeclaration()
		if paramNode != nil && paramNode.DotDotDotToken != nil {
			hasRest = true
			// Rest param should be last
			if i != len(params.Nodes)-1 {
				return false
			}
		} else {
			regularParams++
		}
	}

	// Count regular args vs spread args
	regularArgs := 0
	hasSpread := false
	for i, arg := range superArgs.Nodes {
		if arg == nil {
			continue
		}
		if arg.Kind == ast.KindSpreadElement {
			hasSpread = true
			// Spread should be last
			if i != len(superArgs.Nodes)-1 {
				return false
			}
		} else {
			regularArgs++
		}
	}

	// If rest/spread mismatch, not a match
	if hasRest != hasSpread {
		return false
	}

	// If regular counts don't match, not a match
	if regularParams != regularArgs {
		return false
	}

	// Check if each regular arg matches the corresponding param
	for i := 0; i < regularParams; i++ {
		param := params.Nodes[i]
		arg := superArgs.Nodes[i]

		if param == nil || arg == nil {
			return false
		}

		paramNode := param.AsParameterDeclaration()
		if paramNode == nil || paramNode.Name == nil {
			return false
		}

		// Get parameter name
		paramName := ""
		if paramNode.Name.Kind == ast.KindIdentifier {
			paramIdent := paramNode.Name.AsIdentifier()
			if paramIdent != nil {
				paramName = paramIdent.Text
			}
		}

		// Get argument name (should be an identifier)
		if arg.Kind != ast.KindIdentifier {
			return false
		}
		argIdent := arg.AsIdentifier()
		if argIdent == nil {
			return false
		}

		if paramName != argIdent.Text {
			return false
		}
	}

	// If there's a rest/spread, check if they match
	if hasRest && hasSpread {
		restParam := params.Nodes[len(params.Nodes)-1]
		spreadArg := superArgs.Nodes[len(superArgs.Nodes)-1]

		if restParam == nil || spreadArg == nil {
			return false
		}

		restParamNode := restParam.AsParameterDeclaration()
		if restParamNode == nil || restParamNode.Name == nil {
			return false
		}

		restParamName := ""
		if restParamNode.Name.Kind == ast.KindIdentifier {
			restIdent := restParamNode.Name.AsIdentifier()
			if restIdent != nil {
				restParamName = restIdent.Text
			}
		}

		spreadElement := spreadArg.AsSpreadElement()
		if spreadElement == nil || spreadElement.Expression == nil {
			return false
		}

		if spreadElement.Expression.Kind != ast.KindIdentifier {
			return false
		}

		spreadIdent := spreadElement.Expression.AsIdentifier()
		if spreadIdent == nil {
			return false
		}

		if restParamName != spreadIdent.Text {
			return false
		}
	}

	return true
}

// isUselessConstructor checks if a constructor is useless
func isUselessConstructor(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindConstructor {
		return false
	}

	constructor := node.AsConstructor()
	if constructor == nil {
		return false
	}

	// Check for abstract/declare/overload - these are not useless
	if constructor.Body == nil {
		return false
	}

	// Check for parameter properties - constructors with parameter properties are not useless
	if hasParameterProperty(constructor.Parameters) {
		return false
	}

	// Get modifiers
	modifiers := constructor.Modifiers

	// Get the class node
	classNode := getClassNode(node)
	if classNode == nil {
		return false
	}

	// Check if class extends anything
	hasExtends := hasValidExtends(classNode)

	// Check constructor accessibility
	isPrivate := hasModifiers(modifiers, []ast.Kind{ast.KindPrivateKeyword})
	isProtected := hasModifiers(modifiers, []ast.Kind{ast.KindProtectedKeyword})
	isPublic := hasModifiers(modifiers, []ast.Kind{ast.KindPublicKeyword})

	// Private or protected constructors can be useful even if empty
	// They control instantiation of the class
	if hasExtends {
		// For classes with extends, private/protected constructors are useful
		// Public constructors may still be useless if they just forward to super
		if isPrivate || isProtected {
			return false
		}
	} else {
		// For classes without extends, private/protected constructors are useful
		// Public or no modifier is useless if empty
		if isPrivate || isProtected {
			return false
		}
	}

	// Get constructor body
	body := constructor.Body
	if body == nil {
		return false
	}

	// Get statements from body
	statements := body.AsBlock()
	if statements == nil {
		return false
	}

	// If class doesn't extend, empty or public empty constructor is useless
	if !hasExtends {
		// Empty constructor
		if statements.Statements == nil || len(statements.Statements.Nodes) == 0 {
			return true
		}
		return false
	}

	// Class extends another class
	// Empty constructor in derived class is NOT useless (it calls super implicitly)
	if statements.Statements == nil || len(statements.Statements.Nodes) == 0 {
		return false
	}

	// If there's more than one statement, not useless
	if len(statements.Statements.Nodes) > 1 {
		return false
	}

	// Check if the single statement is a super call
	stmt := statements.Statements.Nodes[0]
	if stmt == nil {
		return false
	}

	// Must be an expression statement containing a super call
	if stmt.Kind != ast.KindExpressionStatement {
		return false
	}

	exprStmt := stmt.AsExpressionStatement()
	if exprStmt == nil || exprStmt.Expression == nil {
		return false
	}

	// Check if it's a call expression
	if exprStmt.Expression.Kind != ast.KindCallExpression {
		return false
	}

	callExpr := exprStmt.Expression.AsCallExpression()
	if callExpr == nil || callExpr.Expression == nil {
		return false
	}

	// Check if it's calling super
	if callExpr.Expression.Kind != ast.KindSuperKeyword {
		return false
	}

	// Check if super arguments match constructor parameters exactly
	return argsMatch(callExpr.Arguments, constructor.Parameters)
}

// NoUselessConstructorRule prevents useless constructors
var NoUselessConstructorRule = rule.CreateRule(rule.Rule{
	Name: "no-useless-constructor",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindConstructor: func(node *ast.Node) {
				if isUselessConstructor(node) {
					// Create suggestion to remove the constructor
					suggestion := rule.RuleSuggestion{
						MessageId:   "removeConstructor",
						Description: buildRemoveConstructor().Description,
						Fix: func() []rule.RuleFix {
							return []rule.RuleFix{
								{
									Range: rule.RuleFixRange{
										Start: node.Pos(),
										End:   node.End(),
									},
									Text: "",
								},
							}
						},
					}

					ctx.ReportNodeWithSuggestions(node, buildNoUselessConstructor(), suggestion)
				}
			},
		}
	},
})

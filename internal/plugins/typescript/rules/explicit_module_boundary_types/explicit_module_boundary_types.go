// Package explicit_module_boundary_types implements the @typescript-eslint/explicit-module-boundary-types rule.
// This rule enforces explicit return and argument types on exported functions' and classes' public class methods,
// clarifying module boundaries and improving code readability and TypeScript performance.
package explicit_module_boundary_types

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type ExplicitModuleBoundaryTypesOptions struct {
	AllowArgumentsExplicitlyTypedAsAny            bool     `json:"allowArgumentsExplicitlyTypedAsAny"`
	AllowDirectConstAssertionInArrowFunctions     bool     `json:"allowDirectConstAssertionInArrowFunctions"`
	AllowedNames                                  []string `json:"allowedNames"`
	AllowHigherOrderFunctions                     bool     `json:"allowHigherOrderFunctions"`
	AllowTypedFunctionExpressions                 bool     `json:"allowTypedFunctionExpressions"`
	AllowOverloadFunctions                        bool     `json:"allowOverloadFunctions"`
}

func parseOptions(options any) ExplicitModuleBoundaryTypesOptions {
	opts := ExplicitModuleBoundaryTypesOptions{
		AllowArgumentsExplicitlyTypedAsAny:        false,
		AllowDirectConstAssertionInArrowFunctions: true,
		AllowedNames:                              []string{},
		AllowHigherOrderFunctions:                 true,
		AllowTypedFunctionExpressions:             true,
		AllowOverloadFunctions:                    false,
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
		if v, ok := optsMap["allowArgumentsExplicitlyTypedAsAny"].(bool); ok {
			opts.AllowArgumentsExplicitlyTypedAsAny = v
		}
		if v, ok := optsMap["allowDirectConstAssertionInArrowFunctions"].(bool); ok {
			opts.AllowDirectConstAssertionInArrowFunctions = v
		}
		if v, ok := optsMap["allowHigherOrderFunctions"].(bool); ok {
			opts.AllowHigherOrderFunctions = v
		}
		if v, ok := optsMap["allowTypedFunctionExpressions"].(bool); ok {
			opts.AllowTypedFunctionExpressions = v
		}
		if v, ok := optsMap["allowOverloadFunctions"].(bool); ok {
			opts.AllowOverloadFunctions = v
		}
		if allowedNames, ok := optsMap["allowedNames"].([]interface{}); ok {
			for _, name := range allowedNames {
				if str, ok := name.(string); ok {
					opts.AllowedNames = append(opts.AllowedNames, str)
				}
			}
		}
	}

	return opts
}

func buildMissingReturnTypeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingReturnType",
		Description: "Missing return type on function.",
	}
}

func buildMissingArgTypeMessage(argName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingArgType",
		Description: "Argument '" + argName + "' should be typed.",
		Data: map[string]interface{}{
			"name": argName,
		},
	}
}

func buildAnyTypedArgMessage(argName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "anyTypedArg",
		Description: "Argument '" + argName + "' should be typed with a non-any type.",
		Data: map[string]interface{}{
			"name": argName,
		},
	}
}

// Check if a function has an explicit return type
func hasReturnType(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		fn := node.AsFunctionDeclaration()
		return fn != nil && fn.Type != nil
	case ast.KindFunctionExpression:
		fn := node.AsFunctionExpression()
		return fn != nil && fn.Type != nil
	case ast.KindArrowFunction:
		fn := node.AsArrowFunction()
		return fn != nil && fn.Type != nil
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		return method != nil && method.Type != nil
	case ast.KindGetAccessor:
		accessor := node.AsGetAccessorDeclaration()
		return accessor != nil && accessor.Type != nil
	}
	return false
}

// Check if arrow function body is a const assertion
func isConstAssertion(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check for direct const assertion: () => x as const
	if node.Kind == ast.KindAsExpression {
		asExpr := node.AsAsExpression()
		if asExpr != nil && asExpr.Type != nil && asExpr.Type.Kind == ast.KindTypeReference {
			typeRef := asExpr.Type.AsTypeReference()
			if typeRef != nil && ast.IsIdentifier(typeRef.TypeName) {
				ident := typeRef.TypeName.AsIdentifier()
				if ident != nil && ident.Text == "const" {
					return true
				}
			}
		}
	}

	// Check for satisfies with const: () => x as const satisfies R
	if node.Kind == ast.KindSatisfiesExpression {
		satisfiesExpr := node.AsSatisfiesExpression()
		if satisfiesExpr != nil && satisfiesExpr.Expression != nil {
			return isConstAssertion(satisfiesExpr.Expression)
		}
	}

	return false
}

// Check if node is a typed function expression
func isTypedFunctionExpression(ctx rule.RuleContext, node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// Check for variable declaration with type: const x: Foo = () => {}
	if parent.Kind == ast.KindVariableDeclaration {
		varDecl := parent.AsVariableDeclaration()
		if varDecl != nil && varDecl.Type != nil {
			return true
		}
	}

	// Check for type assertion: (() => {}) as Foo or <Foo>(() => {})
	if parent.Kind == ast.KindAsExpression || parent.Kind == ast.KindTypeAssertionExpression {
		return true
	}

	// Check for property assignment in typed object literal
	if parent.Kind == ast.KindPropertyAssignment {
		// Walk up to find if object literal has type assertion
		for p := parent.Parent; p != nil; p = p.Parent {
			if p.Kind == ast.KindAsExpression || p.Kind == ast.KindTypeAssertionExpression {
				return true
			}
			if p.Kind == ast.KindVariableDeclaration {
				varDecl := p.AsVariableDeclaration()
				if varDecl != nil && varDecl.Type != nil {
					return true
				}
			}
			// Stop at certain boundaries
			if p.Kind == ast.KindSourceFile || p.Kind == ast.KindBlock {
				break
			}
		}
	}

	// Check for property declaration with type: private method: MethodType = () => {}
	if parent.Kind == ast.KindPropertyDeclaration {
		propDecl := parent.AsPropertyDeclaration()
		if propDecl != nil && propDecl.Type != nil {
			return true
		}
	}

	return false
}

// Check if node is a higher-order function (returns a function)
func isHigherOrderFunction(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindArrowFunction:
		arrowFn := node.AsArrowFunction()
		if arrowFn == nil || arrowFn.Body == nil {
			return false
		}

		// Direct return of arrow or function: () => () => {}
		bodyKind := arrowFn.Body.Kind
		if bodyKind == ast.KindArrowFunction || bodyKind == ast.KindFunctionExpression {
			return true
		}

		// Block with return statement
		if bodyKind == ast.KindBlock {
			block := arrowFn.Body.AsBlock()
			if block != nil && block.Statements != nil && block.Statements.Nodes != nil {
				for _, stmt := range block.Statements.Nodes {
					if stmt.Kind == ast.KindReturnStatement {
						retStmt := stmt.AsReturnStatement()
						if retStmt != nil && retStmt.Expression != nil {
							exprKind := retStmt.Expression.Kind
							if exprKind == ast.KindArrowFunction || exprKind == ast.KindFunctionExpression {
								return true
							}
						}
					}
				}
			}
		}

	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		var body *ast.Node
		if node.Kind == ast.KindFunctionDeclaration {
			fn := node.AsFunctionDeclaration()
			if fn != nil {
				body = fn.Body
			}
		} else {
			fn := node.AsFunctionExpression()
			if fn != nil {
				body = fn.Body
			}
		}

		if body == nil || body.Kind != ast.KindBlock {
			return false
		}

		block := body.AsBlock()
		if block != nil && block.Statements != nil && block.Statements.Nodes != nil {
			for _, stmt := range block.Statements.Nodes {
				if stmt.Kind == ast.KindReturnStatement {
					retStmt := stmt.AsReturnStatement()
					if retStmt != nil && retStmt.Expression != nil {
						exprKind := retStmt.Expression.Kind
						if exprKind == ast.KindArrowFunction || exprKind == ast.KindFunctionExpression {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// Get function name for reporting
func getFunctionName(ctx rule.RuleContext, node *ast.Node) string {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		fn := node.AsFunctionDeclaration()
		if fn != nil && fn.Name() != nil && fn.Name().Kind == ast.KindIdentifier {
			ident := fn.Name().AsIdentifier()
			if ident != nil {
				return ident.Text
			}
		}
	case ast.KindFunctionExpression:
		fn := node.AsFunctionExpression()
		if fn != nil && fn.Name() != nil && fn.Name().Kind == ast.KindIdentifier {
			ident := fn.Name().AsIdentifier()
			if ident != nil {
				return ident.Text
			}
		}
		// Check parent for name
		if node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclaration {
			varDecl := node.Parent.AsVariableDeclaration()
			if varDecl != nil && varDecl.Name() != nil && varDecl.Name().Kind == ast.KindIdentifier {
				ident := varDecl.Name().AsIdentifier()
				if ident != nil {
					return ident.Text
				}
			}
		}
	case ast.KindArrowFunction:
		// Check parent for name
		if node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclaration {
			varDecl := node.Parent.AsVariableDeclaration()
			if varDecl != nil && varDecl.Name() != nil && varDecl.Name().Kind == ast.KindIdentifier {
				ident := varDecl.Name().AsIdentifier()
				if ident != nil {
					return ident.Text
				}
			}
		}
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		if method != nil && method.Name() != nil {
			name, _ := utils.GetNameFromMember(ctx.SourceFile, method.Name())
			return name
		}
	case ast.KindGetAccessor:
		accessor := node.AsGetAccessorDeclaration()
		if accessor != nil && accessor.Name() != nil {
			name, _ := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
			return name
		}
	}
	return ""
}

// Check if function name is in allowed list
func isAllowedName(ctx rule.RuleContext, node *ast.Node, allowedNames []string) bool {
	if len(allowedNames) == 0 {
		return false
	}

	name := getFunctionName(ctx, node)
	if name == "" {
		return false
	}

	for _, allowed := range allowedNames {
		if name == allowed {
			return true
		}
	}
	return false
}

// Check if a node or its ancestors are exported
func isExported(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check for export modifiers on the node itself
	if hasExportModifier(node) {
		return true
	}

	// Check parent contexts
	parent := node.Parent
	for parent != nil {
		// If this is a variable declaration, check its parent (VariableStatement)
		if node.Kind == ast.KindVariableDeclaration && parent.Kind == ast.KindVariableDeclarationList {
			// Check the grandparent (VariableStatement) for export
			if parent.Parent != nil && hasExportModifier(parent.Parent) {
				return true
			}
		}

		// Check if parent is an export declaration
		if parent.Kind == ast.KindExportAssignment || parent.Kind == ast.KindExportDeclaration {
			return true
		}

		// For class members, check if the class is exported
		if parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindClassExpression {
			return isExported(parent)
		}

		parent = parent.Parent
	}

	return false
}

// Check if a node has export modifier
func hasExportModifier(node *ast.Node) bool {
	if node == nil {
		return false
	}

	var modifiers *ast.NodeArray
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		fn := node.AsFunctionDeclaration()
		if fn != nil {
			modifiers = fn.Modifiers
		}
	case ast.KindClassDeclaration:
		class := node.AsClassDeclaration()
		if class != nil {
			modifiers = class.Modifiers
		}
	case ast.KindVariableStatement:
		stmt := node.AsVariableStatement()
		if stmt != nil {
			modifiers = stmt.Modifiers
		}
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		if method != nil {
			modifiers = method.Modifiers
		}
	case ast.KindPropertyDeclaration:
		prop := node.AsPropertyDeclaration()
		if prop != nil {
			modifiers = prop.Modifiers
		}
	}

	if modifiers != nil && modifiers.Nodes != nil {
		for _, mod := range modifiers.Nodes {
			if mod.Kind == ast.KindExportKeyword || mod.Kind == ast.KindDefaultKeyword {
				return true
			}
		}
	}

	return false
}

// Check if a class member is public (not private/protected)
func isPublicMember(node *ast.Node) bool {
	if node == nil {
		return false
	}

	var modifiers *ast.NodeArray
	var name *ast.Node

	switch node.Kind {
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		if method != nil {
			modifiers = method.Modifiers
			name = method.Name()
		}
	case ast.KindPropertyDeclaration:
		prop := node.AsPropertyDeclaration()
		if prop != nil {
			modifiers = prop.Modifiers
			name = prop.Name()
		}
	case ast.KindGetAccessor:
		accessor := node.AsGetAccessorDeclaration()
		if accessor != nil {
			modifiers = accessor.Modifiers
			name = accessor.Name()
		}
	case ast.KindSetAccessor:
		accessor := node.AsSetAccessorDeclaration()
		if accessor != nil {
			modifiers = accessor.Modifiers
		}
	case ast.KindConstructorType:
		// Constructors are always "public" in the sense that they're part of the module boundary
		return true
	}

	// Check if it's a private identifier (#property)
	if name != nil && name.Kind == ast.KindPrivateIdentifier {
		return false
	}

	// Check modifiers for private/protected
	if modifiers != nil && modifiers.Nodes != nil {
		for _, mod := range modifiers.Nodes {
			if mod.Kind == ast.KindPrivateKeyword || mod.Kind == ast.KindProtectedKeyword {
				return false
			}
		}
	}

	return true
}

// Check if a node is within an exported class
func isInExportedClass(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil {
		if parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindClassExpression {
			return isExported(parent)
		}
		parent = parent.Parent
	}
	return false
}

// Check if function should be checked based on export status
func shouldCheckFunction(node *ast.Node) bool {
	// Direct exports
	if isExported(node) {
		return true
	}

	// Class members in exported classes
	if isInExportedClass(node) {
		return isPublicMember(node)
	}

	return false
}

// Check if this is a constructor
func isConstructor(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindConstructorType {
		return false
	}
	return true
}

// Check if this is an overload signature
func isOverloadSignature(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindFunctionDeclaration:
		fn := node.AsFunctionDeclaration()
		return fn != nil && fn.Body == nil
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		return method != nil && method.Body == nil
	}

	return false
}

// Get the node to report (the function signature part)
func getReportNode(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		fn := node.AsFunctionDeclaration()
		if fn != nil && fn.Name() != nil {
			return fn.Name()
		}
	case ast.KindFunctionExpression:
		return node
	case ast.KindArrowFunction:
		return node
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		if method != nil && method.Name() != nil {
			return method.Name()
		}
	case ast.KindGetAccessor:
		accessor := node.AsGetAccessorDeclaration()
		if accessor != nil && accessor.Name() != nil {
			return accessor.Name()
		}
	}
	return node
}

// Get parameter name
func getParameterName(param *ast.Node) string {
	if param == nil {
		return ""
	}

	// Handle binding patterns (destructuring)
	if param.Kind == ast.KindParameter {
		p := param.AsParameter()
		if p != nil && p.Name() != nil {
			name := p.Name()
			if name.Kind == ast.KindIdentifier {
				ident := name.AsIdentifier()
				if ident != nil {
					return ident.Text
				}
			}
			// For destructuring patterns, we can't easily get a single name
			if name.Kind == ast.KindObjectBindingPattern || name.Kind == ast.KindArrayBindingPattern {
				// Return a generic name for destructured params
				return "<destructured>"
			}
		}
	}

	return ""
}

// Check if parameter has explicit type
func parameterHasType(param *ast.Node) bool {
	if param == nil || param.Kind != ast.KindParameter {
		return false
	}

	p := param.AsParameter()
	return p != nil && p.Type != nil
}

// Check if parameter type is "any"
func parameterIsAny(param *ast.Node) bool {
	if param == nil || param.Kind != ast.KindParameter {
		return false
	}

	p := param.AsParameter()
	if p == nil || p.Type == nil {
		return false
	}

	return p.Type.Kind == ast.KindAnyKeyword
}

// Get function parameters
func getFunctionParameters(node *ast.Node) *ast.NodeArray {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		fn := node.AsFunctionDeclaration()
		if fn != nil {
			return fn.Parameters
		}
	case ast.KindFunctionExpression:
		fn := node.AsFunctionExpression()
		if fn != nil {
			return fn.Parameters
		}
	case ast.KindArrowFunction:
		fn := node.AsArrowFunction()
		if fn != nil {
			return fn.Parameters
		}
	case ast.KindMethodDeclaration:
		method := node.AsMethodDeclaration()
		if method != nil {
			return method.Parameters
		}
	case ast.KindGetAccessor:
		accessor := node.AsGetAccessorDeclaration()
		if accessor != nil {
			return accessor.Parameters
		}
	case ast.KindSetAccessor:
		accessor := node.AsSetAccessorDeclaration()
		if accessor != nil {
			return accessor.Parameters
		}
	}
	return nil
}

// Check parameters for missing types
func checkParameters(ctx rule.RuleContext, node *ast.Node, opts ExplicitModuleBoundaryTypesOptions) {
	params := getFunctionParameters(node)
	if params == nil || params.Nodes == nil {
		return
	}

	for _, param := range params.Nodes {
		if param.Kind != ast.KindParameter {
			continue
		}

		// Skip if parameter already has type
		if parameterHasType(param) {
			// Check if it's explicitly typed as "any"
			if !opts.AllowArgumentsExplicitlyTypedAsAny && parameterIsAny(param) {
				paramName := getParameterName(param)
				if paramName != "" && paramName != "<destructured>" {
					p := param.AsParameter()
					if p != nil && p.Name() != nil {
						ctx.ReportNode(p.Name(), buildAnyTypedArgMessage(paramName))
					}
				}
			}
			continue
		}

		// Report missing type
		paramName := getParameterName(param)
		if paramName != "" && paramName != "<destructured>" {
			p := param.AsParameter()
			if p != nil && p.Name() != nil {
				ctx.ReportNode(p.Name(), buildMissingArgTypeMessage(paramName))
			}
		}
	}
}

// Check if function should be skipped based on options
func shouldSkipFunction(ctx rule.RuleContext, node *ast.Node, opts ExplicitModuleBoundaryTypesOptions) bool {
	// Check allowedNames option
	if isAllowedName(ctx, node, opts.AllowedNames) {
		return true
	}

	// Check allowTypedFunctionExpressions option
	if opts.AllowTypedFunctionExpressions &&
		(node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindArrowFunction) &&
		isTypedFunctionExpression(ctx, node) {
		return true
	}

	// Check allowHigherOrderFunctions option
	if opts.AllowHigherOrderFunctions && isHigherOrderFunction(node) {
		return true
	}

	// Check allowDirectConstAssertionInArrowFunctions option
	if opts.AllowDirectConstAssertionInArrowFunctions && node.Kind == ast.KindArrowFunction {
		arrowFn := node.AsArrowFunction()
		if arrowFn != nil && isConstAssertion(arrowFn.Body) {
			return true
		}
	}

	// Check allowOverloadFunctions option
	if opts.AllowOverloadFunctions && isOverloadSignature(node) {
		return true
	}

	return false
}

var ExplicitModuleBoundaryTypesRule = rule.CreateRule(rule.Rule{
	Name: "explicit-module-boundary-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		checkFunction := func(node *ast.Node) {
			// Only check exported functions and public class members
			if !shouldCheckFunction(node) {
				return
			}

			// Skip constructors - they don't need return types
			if isConstructor(node) {
				return
			}

			// Skip if any of the options indicate this function should be ignored
			if shouldSkipFunction(ctx, node, opts) {
				return
			}

			// Check return type
			if !hasReturnType(node) {
				reportNode := getReportNode(node)
				ctx.ReportNode(reportNode, buildMissingReturnTypeMessage())
			}

			// Check parameter types
			checkParameters(ctx, node, opts)
		}

		checkSetAccessor := func(node *ast.Node) {
			// Only check exported setters in exported classes
			if !shouldCheckFunction(node) {
				return
			}

			// Setters don't need return types, only parameter types
			checkParameters(ctx, node, opts)
		}

		checkConstructor := func(node *ast.Node) {
			// Only check constructors in exported classes
			if !isInExportedClass(node) {
				return
			}

			// Constructors don't need return types, only parameter types
			checkParameters(ctx, node, opts)
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: checkFunction,
			ast.KindFunctionExpression:  checkFunction,
			ast.KindArrowFunction:       checkFunction,
			ast.KindMethodDeclaration:   checkFunction,
			ast.KindGetAccessor:         checkFunction,
			ast.KindSetAccessor:         checkSetAccessor,
			ast.KindConstructorType:     checkConstructor,
		}
	},
})

package typedef

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type TypedefOptions struct {
	ArrayDestructuring              bool `json:"arrayDestructuring"`
	ArrowParameter                  bool `json:"arrowParameter"`
	MemberVariableDeclaration       bool `json:"memberVariableDeclaration"`
	ObjectDestructuring             bool `json:"objectDestructuring"`
	Parameter                       bool `json:"parameter"`
	PropertyDeclaration             bool `json:"propertyDeclaration"`
	VariableDeclaration             bool `json:"variableDeclaration"`
	VariableDeclarationIgnoreFunction bool `json:"variableDeclarationIgnoreFunction"`
}

func buildExpectedTypedefMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedTypedef",
		Description: "Expected a type annotation.",
	}
}

func buildExpectedTypedefNamedMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedTypedefNamed",
		Description: "Expected " + name + " to have a type annotation.",
	}
}

var TypedefRule = rule.CreateRule(rule.Rule{
	Name: "typedef",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := TypedefOptions{
		ArrayDestructuring:              false,
		ArrowParameter:                  false,
		MemberVariableDeclaration:       false,
		ObjectDestructuring:             false,
		Parameter:                       false,
		PropertyDeclaration:             false,
		VariableDeclaration:             false,
		VariableDeclarationIgnoreFunction: false,
	}

	// Parse options
	if options != nil {
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if optMap, isMap := optArray[0].(map[string]interface{}); isMap {
				if val, ok := optMap["arrayDestructuring"].(bool); ok {
					opts.ArrayDestructuring = val
				}
				if val, ok := optMap["arrowParameter"].(bool); ok {
					opts.ArrowParameter = val
				}
				if val, ok := optMap["memberVariableDeclaration"].(bool); ok {
					opts.MemberVariableDeclaration = val
				}
				if val, ok := optMap["objectDestructuring"].(bool); ok {
					opts.ObjectDestructuring = val
				}
				if val, ok := optMap["parameter"].(bool); ok {
					opts.Parameter = val
				}
				if val, ok := optMap["propertyDeclaration"].(bool); ok {
					opts.PropertyDeclaration = val
				}
				if val, ok := optMap["variableDeclaration"].(bool); ok {
					opts.VariableDeclaration = val
				}
				if val, ok := optMap["variableDeclarationIgnoreFunction"].(bool); ok {
					opts.VariableDeclarationIgnoreFunction = val
				}
			}
		}
	}

	// Helper to check if a binding pattern has type annotation
	hasTypeAnnotation := func(node *ast.Node) bool {
		if node == nil {
			return false
		}

		switch node.Kind {
		case ast.KindParameter:
			param := node.AsParameterDeclaration()
			return param != nil && param.Type != nil
		case ast.KindVariableDeclaration:
			varDecl := node.AsVariableDeclaration()
			return varDecl != nil && varDecl.Type != nil
		case ast.KindPropertyDeclaration:
			propDecl := node.AsPropertyDeclaration()
			return propDecl != nil && propDecl.Type != nil
		case ast.KindPropertySignature:
			propSig := node.AsPropertySignature()
			return propSig != nil && propSig.Type != nil
		}
		return false
	}

	// Helper to check if a binding element is part of an assignment pattern in for-in/for-of
	isForInOrForOfLoop := func(node *ast.Node) bool {
		current := node.Parent
		for current != nil {
			if current.Kind == ast.KindForOfStatement || current.Kind == ast.KindForInStatement {
				return true
			}
			// Stop if we reach a function or other scope boundary
			if current.Kind == ast.KindFunctionDeclaration ||
				current.Kind == ast.KindArrowFunction ||
				current.Kind == ast.KindMethodDeclaration {
				return false
			}
			current = current.Parent
		}
		return false
	}

	// Helper to check if initializer is a function expression
	isFunctionExpression := func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		return node.Kind == ast.KindFunctionExpression ||
			node.Kind == ast.KindArrowFunction
	}

	// Helper to check binding elements in array/object destructuring
	checkBindingElement := func(node *ast.Node, isArray bool) {
		if node == nil {
			return
		}

		switch node.Kind {
		case ast.KindBindingElement:
			bindingElem := node.AsBindingElement()
			if bindingElem == nil {
				return
			}

			// Check if in for-in/for-of loop
			if isForInOrForOfLoop(node) {
				return
			}

			name := bindingElem.Name()
			if name == nil {
				return
			}

			// If the name is a binding pattern, recursively check it
			if name.Kind == ast.KindObjectBindingPattern {
				if opts.ObjectDestructuring {
					checkObjectBindingPattern(name)
				}
			} else if name.Kind == ast.KindArrayBindingPattern {
				if opts.ArrayDestructuring {
					checkArrayBindingPattern(name)
				}
			} else if ast.IsIdentifier(name) {
				// Check if the binding element has a type annotation
				if !hasTypeAnnotation(node.Parent) {
					ident := name.AsIdentifier()
					if ident != nil && ident.Text != "" {
						ctx.ReportNode(name, buildExpectedTypedefMessage())
					}
				}
			}

		case ast.KindIdentifier:
			// Direct identifier in destructuring
			if !hasTypeAnnotation(node.Parent) {
				ctx.ReportNode(node, buildExpectedTypedefMessage())
			}
		}
	}

	checkArrayBindingPattern := func(node *ast.Node) {
		if node == nil || node.Kind != ast.KindArrayBindingPattern {
			return
		}

		arrayPattern := node.AsArrayBindingPattern()
		if arrayPattern == nil {
			return
		}

		for _, elem := range arrayPattern.Elements.Nodes {
			if elem != nil {
				checkBindingElement(elem, true)
			}
		}
	}

	checkObjectBindingPattern := func(node *ast.Node) {
		if node == nil || node.Kind != ast.KindObjectBindingPattern {
			return
		}

		objPattern := node.AsObjectBindingPattern()
		if objPattern == nil {
			return
		}

		for _, elem := range objPattern.Elements.Nodes {
			if elem != nil {
				checkBindingElement(elem, false)
			}
		}
	}

	checkVariableDeclaration := func(node *ast.Node) {
		if node == nil {
			return
		}

		varDecl := node.AsVariableDeclaration()
		if varDecl == nil {
			return
		}

		// Skip if already has type annotation
		if varDecl.Type != nil {
			return
		}

		// Check if this is part of a for-in/for-of loop
		if isForInOrForOfLoop(node) {
			return
		}

		name := varDecl.Name()
		if name == nil {
			return
		}

		// Handle destructuring patterns
		if name.Kind == ast.KindArrayBindingPattern {
			if opts.ArrayDestructuring {
				checkArrayBindingPattern(name)
			}
			return
		}

		if name.Kind == ast.KindObjectBindingPattern {
			if opts.ObjectDestructuring {
				checkObjectBindingPattern(name)
			}
			return
		}

		// Handle regular variable declarations
		if opts.VariableDeclaration {
			// Check if we should ignore function expressions
			if opts.VariableDeclarationIgnoreFunction && isFunctionExpression(varDecl.Initializer) {
				return
			}

			if ast.IsIdentifier(name) {
				ident := name.AsIdentifier()
				if ident != nil && ident.Text != "" {
					ctx.ReportNode(name, buildExpectedTypedefNamedMessage(ident.Text))
				}
			}
		}
	}

	checkParameter := func(node *ast.Node, isArrowFunction bool) {
		if node == nil {
			return
		}

		param := node.AsParameterDeclaration()
		if param == nil {
			return
		}

		// Skip if already has type annotation
		if param.Type != nil {
			return
		}

		// Skip rest parameters with destructuring - they get type from rest type
		if param.DotDotDotToken != nil {
			name := param.Name()
			if name != nil {
				if name.Kind == ast.KindArrayBindingPattern || name.Kind == ast.KindObjectBindingPattern {
					return
				}
			}
		}

		name := param.Name()
		if name == nil {
			return
		}

		// Handle destructuring patterns in parameters
		if name.Kind == ast.KindArrayBindingPattern {
			if opts.ArrayDestructuring {
				checkArrayBindingPattern(name)
			}
			return
		}

		if name.Kind == ast.KindObjectBindingPattern {
			if opts.ObjectDestructuring {
				checkObjectBindingPattern(name)
			}
			return
		}

		// Check regular parameters
		shouldCheck := false
		if isArrowFunction && opts.ArrowParameter {
			shouldCheck = true
		} else if !isArrowFunction && opts.Parameter {
			shouldCheck = true
		}

		if shouldCheck && ast.IsIdentifier(name) {
			ident := name.AsIdentifier()
			if ident != nil && ident.Text != "" {
				ctx.ReportNode(name, buildExpectedTypedefNamedMessage(ident.Text))
			}
		}
	}

	return rule.RuleListeners{
		ast.KindVariableDeclaration: func(node *ast.Node) {
			checkVariableDeclaration(node)
		},

		ast.KindArrowFunction: func(node *ast.Node) {
			if !opts.ArrowParameter {
				return
			}

			arrowFunc := node.AsArrowFunction()
			if arrowFunc == nil {
				return
			}

			for _, param := range arrowFunc.Parameters.Nodes {
				checkParameter(param, true)
			}
		},

		ast.KindFunctionDeclaration: func(node *ast.Node) {
			if !opts.Parameter {
				return
			}

			funcDecl := node.AsFunctionDeclaration()
			if funcDecl == nil {
				return
			}

			for _, param := range funcDecl.Parameters.Nodes {
				checkParameter(param, false)
			}
		},

		ast.KindMethodDeclaration: func(node *ast.Node) {
			if !opts.Parameter {
				return
			}

			method := node.AsMethodDeclaration()
			if method == nil {
				return
			}

			for _, param := range method.Parameters.Nodes {
				checkParameter(param, false)
			}
		},

		ast.KindConstructor: func(node *ast.Node) {
			if !opts.Parameter {
				return
			}

			constructor := node.AsConstructorDeclaration()
			if constructor == nil {
				return
			}

			for _, param := range constructor.Parameters.Nodes {
				// Skip parameters with accessibility modifiers (public, private, protected)
				// as they automatically become class properties
				paramDecl := param.AsParameterDeclaration()
				if paramDecl != nil {
					// Check if parameter has modifiers
					if paramDecl.Modifiers != nil && len(paramDecl.Modifiers.Nodes) > 0 {
						for _, mod := range paramDecl.Modifiers.Nodes {
							if mod.Kind == ast.KindPublicKeyword ||
								mod.Kind == ast.KindPrivateKeyword ||
								mod.Kind == ast.KindProtectedKeyword ||
								mod.Kind == ast.KindReadonlyKeyword {
								continue
							}
						}
					}
				}
				checkParameter(param, false)
			}
		},

		ast.KindPropertyDeclaration: func(node *ast.Node) {
			if !opts.MemberVariableDeclaration {
				return
			}

			propDecl := node.AsPropertyDeclaration()
			if propDecl == nil {
				return
			}

			// Skip if already has type annotation
			if propDecl.Type != nil {
				return
			}

			name := propDecl.Name()
			if name != nil && ast.IsIdentifier(name) {
				ident := name.AsIdentifier()
				if ident != nil && ident.Text != "" {
					ctx.ReportNode(name, buildExpectedTypedefNamedMessage(ident.Text))
				}
			}
		},

		ast.KindPropertySignature: func(node *ast.Node) {
			if !opts.PropertyDeclaration {
				return
			}

			propSig := node.AsPropertySignature()
			if propSig == nil {
				return
			}

			// Skip if already has type annotation
			if propSig.Type != nil {
				return
			}

			name := propSig.Name()
			if name != nil && ast.IsIdentifier(name) {
				ident := name.AsIdentifier()
				if ident != nil && ident.Text != "" {
					ctx.ReportNode(name, buildExpectedTypedefNamedMessage(ident.Text))
				}
			}
		},

		ast.KindFunctionExpression: func(node *ast.Node) {
			if !opts.Parameter {
				return
			}

			funcExpr := node.AsFunctionExpression()
			if funcExpr == nil {
				return
			}

			for _, param := range funcExpr.Parameters.Nodes {
				checkParameter(param, false)
			}
		},
	}
}

package prefer_readonly

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type PreferReadonlyOptions struct {
	OnlyInlineLambdas bool `json:"onlyInlineLambdas"`
}

// PreferReadonlyRule implements the prefer-readonly rule
// Require private members to be readonly if never modified
var PreferReadonlyRule = rule.CreateRule(rule.Rule{
	Name:         "prefer-readonly",
	RequiresType: true,
	Run:          run,
})

type propertyInfo struct {
	node          *ast.Node
	name          string
	isPrivate     bool
	isReadonly    bool
	isModified    bool
	isLambda      bool
	propertyDecl  *ast.PropertyDeclaration
	parameterDecl *ast.ParameterDeclaration
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := PreferReadonlyOptions{
		OnlyInlineLambdas: false,
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
			if onlyInlineLambdas, ok := optsMap["onlyInlineLambdas"].(bool); ok {
				opts.OnlyInlineLambdas = onlyInlineLambdas
			}
		}
	}

	// Track properties per class
	classPropertiesStack := make([]*map[string]*propertyInfo, 0)

	return rule.RuleListeners{
		ast.KindClassDeclaration: func(node *ast.Node) {
			// Enter a new class scope
			properties := make(map[string]*propertyInfo)
			classPropertiesStack = append(classPropertiesStack, &properties)

			// Collect all private properties and parameter properties
			classDecl := node.AsClassDeclaration()
			if classDecl == nil || classDecl.Members == nil {
				return
			}

			for _, member := range classDecl.Members.Nodes {
				switch member.Kind {
				case ast.KindPropertyDeclaration:
					propDecl := member.AsPropertyDeclaration()
					if propDecl == nil {
						continue
					}

					// Check if private
					isPrivate := false
					isReadonly := false
					isStatic := false

					if propDecl.Modifiers() != nil {
						for _, modifier := range propDecl.Modifiers().Nodes {
							if modifier.Kind == ast.KindPrivateKeyword {
								isPrivate = true
							} else if modifier.Kind == ast.KindReadonlyKeyword {
								isReadonly = true
							} else if modifier.Kind == ast.KindStaticKeyword {
								isStatic = true
							}
						}
					}

					// Check for private field syntax (#field)
					if !isPrivate && propDecl.Name != nil && propDecl.Name.Kind == ast.KindPrivateIdentifier {
						isPrivate = true
					}

					if !isPrivate || isReadonly {
						continue
					}

					// Get property name
					var name string
					if propDecl.Name != nil {
						nameRange := utils.TrimNodeTextRange(ctx.SourceFile, propDecl.Name)
						name = ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
					}

					// Check if it's a lambda function
					isLambda := false
					if propDecl.Initializer != nil {
						if propDecl.Initializer.Kind == ast.KindArrowFunction ||
							propDecl.Initializer.Kind == ast.KindFunctionExpression {
							isLambda = true
						}
					}

					if opts.OnlyInlineLambdas && !isLambda {
						continue
					}

					properties[name] = &propertyInfo{
						node:         member,
						name:         name,
						isPrivate:    true,
						isReadonly:   false,
						isModified:   false,
						isLambda:     isLambda,
						propertyDecl: propDecl,
					}

				case ast.KindConstructor:
					// Check for parameter properties
					constructor := member.AsConstructorDeclaration()
					if constructor == nil || constructor.Parameters == nil {
						continue
					}

					for _, param := range constructor.Parameters.Nodes {
						paramDecl := param.AsParameterDeclaration()
						if paramDecl == nil || paramDecl.Modifiers() == nil {
							continue
						}

						isPrivate := false
						isReadonly := false

						for _, modifier := range paramDecl.Modifiers().Nodes {
							if modifier.Kind == ast.KindPrivateKeyword {
								isPrivate = true
							} else if modifier.Kind == ast.KindReadonlyKeyword {
								isReadonly = true
							}
						}

						if !isPrivate || isReadonly {
							continue
						}

						// Get parameter name
						var name string
						if paramDecl.Name != nil && paramDecl.Name.Kind == ast.KindIdentifier {
							id := paramDecl.Name.AsIdentifier()
							if id != nil {
								name = id.Text()
							}
						}

						properties[name] = &propertyInfo{
							node:          param,
							name:          name,
							isPrivate:     true,
							isReadonly:    false,
							isModified:    false,
							isLambda:      false,
							parameterDecl: paramDecl,
						}
					}
				}
			}

			// Now traverse the class body to find assignments
			traverseForAssignments(node, &properties)

			// Report unmodified properties
			for _, prop := range properties {
				if !prop.isModified {
					reportReadonlyViolation(ctx, prop)
				}
			}

			// Pop the class scope
			if len(classPropertiesStack) > 0 {
				classPropertiesStack = classPropertiesStack[:len(classPropertiesStack)-1]
			}
		},
	}
}

// traverseForAssignments walks the AST to find property assignments
func traverseForAssignments(node *ast.Node, properties *map[string]*propertyInfo) {
	if node == nil {
		return
	}

	// Check for assignment expressions
	if node.Kind == ast.KindBinaryExpression {
		binExpr := node.AsBinaryExpression()
		if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindEqualsToken {
			// Check if left side is a property access
			if binExpr.Left != nil && binExpr.Left.Kind == ast.KindPropertyAccessExpression {
				propAccess := binExpr.Left.AsPropertyAccessExpression()
				if propAccess != nil {
					// Check if it's 'this.propertyName'
					if propAccess.Expression != nil && propAccess.Expression.Kind == ast.KindThisKeyword {
						if propAccess.Name != nil {
							name := propAccess.Name.Text()
							if prop, exists := (*properties)[name]; exists {
								// Check if we're inside the constructor
								if !isInConstructor(node) {
									prop.isModified = true
								}
							}
						}
					}
				}
			}
		}
	}

	// Recurse into children
	ast.ForEachChild(node, func(child *ast.Node) {
		traverseForAssignments(child, properties)
	})
}

// isInConstructor checks if a node is inside a constructor
func isInConstructor(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil {
		if parent.Kind == ast.KindConstructor {
			return true
		}
		if parent.Kind == ast.KindClassDeclaration {
			// Stop at class boundary
			return false
		}
		parent = parent.Parent
	}
	return false
}

// reportReadonlyViolation reports a property that should be readonly
func reportReadonlyViolation(ctx rule.RuleContext, prop *propertyInfo) {
	message := rule.RuleMessage{
		Id:          "preferReadonly",
		Description: "Member '" + prop.name + "' is never reassigned; mark it as `readonly`.",
	}

	if prop.propertyDecl != nil {
		// Build fix for property declaration
		fix := buildPropertyFix(ctx, prop.propertyDecl)
		if fix != nil {
			ctx.ReportNodeWithFixes(prop.node, message, *fix)
		} else {
			ctx.ReportNode(prop.node, message)
		}
	} else if prop.parameterDecl != nil {
		// Build fix for parameter property
		fix := buildParameterFix(ctx, prop.parameterDecl)
		if fix != nil {
			ctx.ReportNodeWithFixes(prop.node, message, *fix)
		} else {
			ctx.ReportNode(prop.node, message)
		}
	} else {
		ctx.ReportNode(prop.node, message)
	}
}

// buildPropertyFix creates a fix to add readonly modifier to a property
func buildPropertyFix(ctx rule.RuleContext, propDecl *ast.PropertyDeclaration) *rule.RuleFix {
	if propDecl == nil {
		return nil
	}

	nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, propDecl)
	nodeText := ctx.SourceFile.Text()[nodeRange.Pos():nodeRange.End()]

	// Find where to insert 'readonly'
	// It should go after visibility modifiers but before the property name

	var insertPos int
	var insertedReadonly string

	// Check for modifiers
	if propDecl.Modifiers() != nil && len(propDecl.Modifiers().Nodes) > 0 {
		// Insert after the last modifier
		lastModifier := propDecl.Modifiers().Nodes[len(propDecl.Modifiers().Nodes)-1]
		lastModRange := utils.TrimNodeTextRange(ctx.SourceFile, lastModifier)
		// Find the position after the last modifier in the original text
		relativePos := lastModRange.End() - nodeRange.Pos()
		insertPos = relativePos
		insertedReadonly = " readonly"
	} else {
		// Insert at the beginning
		insertPos = 0
		insertedReadonly = "readonly "
	}

	// Build the replacement text
	newText := nodeText[:insertPos] + insertedReadonly + nodeText[insertPos:]

	return &rule.RuleFix{
		Range: utils.TextRange{
			Pos: nodeRange.Pos(),
			End: nodeRange.End(),
		},
		Text: newText,
	}
}

// buildParameterFix creates a fix to add readonly modifier to a parameter property
func buildParameterFix(ctx rule.RuleContext, paramDecl *ast.ParameterDeclaration) *rule.RuleFix {
	if paramDecl == nil {
		return nil
	}

	nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, paramDecl)
	nodeText := ctx.SourceFile.Text()[nodeRange.Pos():nodeRange.End()]

	// Find where to insert 'readonly' after 'private'
	var insertPos int
	if paramDecl.Modifiers() != nil && len(paramDecl.Modifiers().Nodes) > 0 {
		lastModifier := paramDecl.Modifiers().Nodes[len(paramDecl.Modifiers().Nodes)-1]
		lastModRange := utils.TrimNodeTextRange(ctx.SourceFile, lastModifier)
		relativePos := lastModRange.End() - nodeRange.Pos()
		insertPos = relativePos
	}

	// Build the replacement text
	newText := nodeText[:insertPos] + " readonly" + nodeText[insertPos:]

	return &rule.RuleFix{
		Range: utils.TextRange{
			Pos: nodeRange.Pos(),
			End: nodeRange.End(),
		},
		Text: newText,
	}
}

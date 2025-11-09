package no_dupe_class_members

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildUnexpectedMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: fmt.Sprintf("Duplicate name '%s'.", name),
	}
}

type memberInfo struct {
	name       string
	isStatic   bool
	isMethod   bool
	isGetter   bool
	isSetter   bool
	node       *ast.Node
	isOverload bool
}

// getMemberName extracts the name from a class member node
func getMemberName(ctx rule.RuleContext, node *ast.Node) string {
	var nameNode *ast.Node

	if ast.IsPropertyDeclaration(node) {
		nameNode = node.AsPropertyDeclaration().Name()
	} else if ast.IsMethodDeclaration(node) {
		nameNode = node.AsMethodDeclaration().Name()
	} else if ast.IsGetAccessorDeclaration(node) {
		nameNode = node.AsGetAccessorDeclaration().Name()
	} else if ast.IsSetAccessorDeclaration(node) {
		nameNode = node.AsSetAccessorDeclaration().Name()
	} else {
		return ""
	}

	if nameNode == nil {
		return ""
	}

	return extractPropertyName(ctx, nameNode)
}

// extractPropertyName extracts the property name from a name node
func extractPropertyName(ctx rule.RuleContext, nameNode *ast.Node) string {
	// Handle computed property names - these are not duplicates with regular properties
	if nameNode.Kind == ast.KindComputedPropertyName {
		return "" // Computed properties are dynamic and can't be checked for duplicates
	}

	// Handle regular identifiers
	if nameNode.Kind == ast.KindIdentifier {
		return nameNode.AsIdentifier().Text
	}

	// Handle string literals as property names
	if ast.IsLiteralExpression(nameNode) {
		text := nameNode.Text()
		// Remove quotes for string literals to normalize the name
		if len(text) >= 2 && ((text[0] == '"' && text[len(text)-1] == '"') || (text[0] == '\'' && text[len(text)-1] == '\'')) {
			return text[1 : len(text)-1]
		}
		return text
	}

	return ""
}

// isMethodOverloadSignature checks if a method is an overload signature (no body)
func isMethodOverloadSignature(node *ast.Node) bool {
	if !ast.IsMethodDeclaration(node) {
		return false
	}
	method := node.AsMethodDeclaration()
	return method.Body == nil
}

// createMemberInfo creates a memberInfo struct from a class member node
func createMemberInfo(ctx rule.RuleContext, node *ast.Node) *memberInfo {
	name := getMemberName(ctx, node)
	if name == "" {
		return nil
	}

	info := &memberInfo{
		name:       name,
		isStatic:   ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic),
		node:       node,
		isOverload: isMethodOverloadSignature(node),
	}

	// Determine member type
	if ast.IsMethodDeclaration(node) {
		info.isMethod = true
	} else if ast.IsGetAccessorDeclaration(node) {
		info.isGetter = true
	} else if ast.IsSetAccessorDeclaration(node) {
		info.isSetter = true
	}

	return info
}

// isDuplicate checks if two members are duplicates according to the rule
func isDuplicate(m1, m2 *memberInfo) bool {
	// Different names are not duplicates
	if m1.name != m2.name {
		return false
	}

	// Different static/instance are not duplicates
	if m1.isStatic != m2.isStatic {
		return false
	}

	// Getter and setter with the same name are allowed
	if (m1.isGetter && m2.isSetter) || (m1.isSetter && m2.isGetter) {
		return false
	}

	// Method overload signatures are allowed (TypeScript feature)
	// Multiple overload signatures can exist, and one implementation with a body
	if m1.isMethod && m2.isMethod {
		// If either is an overload signature (no body), they can coexist
		if m1.isOverload || m2.isOverload {
			return false
		}
	}

	// Everything else is a duplicate
	return true
}

var NoDupeClassMembersRule = rule.CreateRule(rule.Rule{
	Name: "no-dupe-class-members",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		var memberStack [][]*memberInfo

		enterClass := func() {
			memberStack = append(memberStack, []*memberInfo{})
		}

		exitClass := func() {
			if len(memberStack) > 0 {
				memberStack = memberStack[:len(memberStack)-1]
			}
		}

		checkMember := func(node *ast.Node) {
			if len(memberStack) == 0 {
				return
			}

			newMember := createMemberInfo(ctx, node)
			if newMember == nil {
				return
			}

			members := memberStack[len(memberStack)-1]

			// Check against all existing members
			for _, existingMember := range members {
				if isDuplicate(newMember, existingMember) {
					ctx.ReportNode(node, buildUnexpectedMessage(newMember.name))
					return
				}
			}

			// Add this member to the list
			memberStack[len(memberStack)-1] = append(memberStack[len(memberStack)-1], newMember)
		}

		return rule.RuleListeners{
			// Track class declarations
			ast.KindClassDeclaration: func(node *ast.Node) {
				enterClass()
			},
			rule.ListenerOnExit(ast.KindClassDeclaration): func(node *ast.Node) {
				exitClass()
			},

			// Track class expressions
			ast.KindClassExpression: func(node *ast.Node) {
				enterClass()
			},
			rule.ListenerOnExit(ast.KindClassExpression): func(node *ast.Node) {
				exitClass()
			},

			// Check class members
			ast.KindPropertyDeclaration: checkMember,
			ast.KindMethodDeclaration:   checkMember,
			ast.KindGetAccessor:         checkMember,
			ast.KindSetAccessor:         checkMember,
		}
	},
})

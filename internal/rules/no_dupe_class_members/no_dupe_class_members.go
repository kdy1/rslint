package no_dupe_class_members

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoDupeClassMembersRule implements the no-dupe-class-members rule
// Disallow duplicate class members
var NoDupeClassMembersRule = rule.Rule{
	Name: "no-dupe-class-members",
	Run:  run,
}

func buildDuplicateMessage(memberName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "duplicateMember",
		Description: "Duplicate name '" + memberName + "'.",
	}
}

// getMemberKey generates a key for a class member based on its name and type
func getMemberKey(ctx rule.RuleContext, member *ast.Node) (string, string) {
	if member == nil {
		return "", ""
	}

	var memberText string
	var isStatic bool

	switch member.Kind {
	case ast.KindMethodDeclaration:
		method := member.AsMethodDeclaration()
		if method == nil || method.Name() == nil {
			return "", ""
		}
		// Use the name's range to extract text
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, method.Name())
		memberText = ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
		if method.Modifiers() != nil {
			for _, mod := range method.Modifiers().Nodes {
				if mod != nil && mod.Kind == ast.KindStaticKeyword {
					isStatic = true
					break
				}
			}
		}

	case ast.KindPropertyDeclaration:
		prop := member.AsPropertyDeclaration()
		if prop == nil || prop.Name() == nil {
			return "", ""
		}
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, prop.Name())
		memberText = ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
		if prop.Modifiers() != nil {
			for _, mod := range prop.Modifiers().Nodes {
				if mod != nil && mod.Kind == ast.KindStaticKeyword {
					isStatic = true
					break
				}
			}
		}

	case ast.KindGetAccessor:
		accessor := member.AsGetAccessorDeclaration()
		if accessor == nil || accessor.Name() == nil {
			return "", ""
		}
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, accessor.Name())
		memberText = ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
		if accessor.Modifiers() != nil {
			for _, mod := range accessor.Modifiers().Nodes {
				if mod != nil && mod.Kind == ast.KindStaticKeyword {
					isStatic = true
					break
				}
			}
		}

	case ast.KindSetAccessor:
		accessor := member.AsSetAccessorDeclaration()
		if accessor == nil || accessor.Name() == nil {
			return "", ""
		}
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, accessor.Name())
		memberText = ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()]
		if accessor.Modifiers() != nil {
			for _, mod := range accessor.Modifiers().Nodes {
				if mod != nil && mod.Kind == ast.KindStaticKeyword {
					isStatic = true
					break
				}
			}
		}

	case ast.KindConstructor:
		return "", "" // Constructor is always unique

	default:
		return "", ""
	}

	if memberText == "" {
		return "", ""
	}

	// Include static modifier in the key to distinguish static from instance members
	key := memberText
	if isStatic {
		key = "static:" + memberText
	} else {
		key = "instance:" + memberText
	}

	return key, memberText
}

// checkDuplicateClassMembers checks for duplicate member names in a class
func checkDuplicateClassMembers(ctx rule.RuleContext, members []*ast.Node) {
	if members == nil {
		return
	}

	seen := make(map[string]*ast.Node)

	for _, member := range members {
		if member == nil {
			continue
		}

		key, memberName := getMemberKey(ctx, member)
		if key == "" {
			// Skip members we can't statically analyze (computed properties, constructor)
			continue
		}

		if _, exists := seen[key]; exists {
			// Report error on the duplicate member
			ctx.ReportNode(member, buildDuplicateMessage(memberName))
		} else {
			seen[key] = member
		}
	}
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindClassDeclaration: func(node *ast.Node) {
			classDecl := node.AsClassDeclaration()
			if classDecl != nil && classDecl.Members != nil && len(classDecl.Members.Nodes) > 0 {
				checkDuplicateClassMembers(ctx, classDecl.Members.Nodes)
			}
		},
		ast.KindClassExpression: func(node *ast.Node) {
			classExpr := node.AsClassExpression()
			if classExpr != nil && classExpr.Members != nil && len(classExpr.Members.Nodes) > 0 {
				checkDuplicateClassMembers(ctx, classExpr.Members.Nodes)
			}
		},
	}
}

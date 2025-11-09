package no_restricted_types

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type NoRestrictedTypesOptions struct {
	Types map[string]interface{} `json:"types"`
}

type RestrictedTypeConfig struct {
	Message string      `json:"message"`
	FixWith interface{} `json:"fixWith"`
}

var NoRestrictedTypesRule = rule.CreateRule(rule.Rule{
	Name: "no-restricted-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoRestrictedTypesOptions{
			Types: make(map[string]interface{}),
		}

		// Parse options with dual-format support
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if types, ok := optsMap["types"].(map[string]interface{}); ok {
					opts.Types = types
				}
			}
		}

		// Build restricted types map - normalize by trimming whitespace
		restrictedTypes := make(map[string]RestrictedTypeConfig)
		for typeName, typeConfig := range opts.Types {
			normalizedName := strings.TrimSpace(typeName)
			
			if typeConfig == nil || typeConfig == false {
				// Explicit null or false allows the type
				continue
			}

			config := RestrictedTypeConfig{}
			switch v := typeConfig.(type) {
			case bool:
				// true means restrict with default message
				if v {
					config.Message = ""
				}
			case string:
				config.Message = v
			case map[string]interface{}:
				if msg, ok := v["message"].(string); ok {
					config.Message = msg
				}
				if fix, ok := v["fixWith"]; ok {
					config.FixWith = fix
				}
			}
			restrictedTypes[normalizedName] = config
		}

		checkType := func(node *ast.Node) {
			var typeName string
			
			// Handle different type node kinds
			switch node.Kind {
			case ast.KindTypeReference:
				typeRef := node.AsTypeReferenceNode()
				if typeRef == nil || typeRef.TypeName == nil {
					return
				}
				typeName = getTypeNameText(ctx, typeRef)
				
			case ast.KindTupleType:
				// Check for empty tuple []
				tupleType := node.AsTupleTypeNode()
				if tupleType != nil && tupleType.Elements != nil && len(tupleType.Elements.Nodes) == 0 {
					typeName = "[]"
				}
				
			case ast.KindTypeLiteral:
				// Check for empty object type {}
				typeLiteral := node.AsTypeLiteralNode()
				if typeLiteral != nil && typeLiteral.Members != nil && len(typeLiteral.Members.Nodes) == 0 {
					typeName = "{}"
				}
			}
			
			if typeName == "" {
				return
			}

			// Check if this type is restricted
			config, isRestricted := restrictedTypes[typeName]
			if !isRestricted {
				return
			}

			customMessage := config.Message
			message := rule.RuleMessage{
				Id:          "bannedTypeMessage",
				Description: fmt.Sprintf("Don't use `%s` as a type.%s", typeName, customMessage),
			}

			// Create fix if available
			if config.FixWith != nil {
				if fixStr, ok := config.FixWith.(string); ok {
					// Get the text range and replace with fix
					ctx.ReportNodeWithFixes(node, message, rule.RuleFixReplace(ctx.SourceFile, node, fixStr))
					return
				}
			}

			// Report without fix
			ctx.ReportNode(node, message)
		}

		return rule.RuleListeners{
			ast.KindTypeReference: checkType,
			ast.KindTupleType:     checkType,
			ast.KindTypeLiteral:   checkType,
		}
	},
})

// getTypeNameText extracts the type name text from a type reference node
// This handles simple identifiers, qualified names (NS.Type), and generic types (Type<T>)
func getTypeNameText(ctx rule.RuleContext, typeRef *ast.TypeReferenceNode) string {
	if typeRef == nil || typeRef.TypeName == nil {
		return ""
	}

	typeName := getIdentifierOrQualifiedName(ctx, typeRef.TypeName)

	// Handle generic type parameters if present
	if typeRef.TypeArguments != nil && len(typeRef.TypeArguments.Nodes) > 0 {
		// Build the generic type string like "Type<Arg1, Arg2>"
		var args []string
		for _, arg := range typeRef.TypeArguments.Nodes {
			argText := getTypeText(ctx, arg)
			args = append(args, argText)
		}

		// Join without spaces to match test expectations for patterns like "Banned<A,B>"
		argsStr := strings.Join(args, ",")
		typeName = fmt.Sprintf("%s<%s>", typeName, argsStr)
	}

	return typeName
}

// getIdentifierOrQualifiedName extracts the text from an identifier or qualified name node
func getIdentifierOrQualifiedName(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		ident := node.AsIdentifier()
		if ident != nil {
			return ident.Text
		}
	case ast.KindQualifiedName:
		// For qualified names like A.B.C, we want the full path
		qual := node.AsQualifiedName()
		if qual != nil {
			left := getIdentifierOrQualifiedName(ctx, qual.Left)
			right := getIdentifierOrQualifiedName(ctx, qual.Right)
			if left != "" && right != "" {
				return left + "." + right
			}
		}
	}

	// Fallback: get text from source
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	text := ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
	return strings.TrimSpace(text)
}

// getTypeText extracts the text representation of any type node
func getTypeText(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil {
		return ""
	}

	// Use source text for type arguments to preserve formatting
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	text := ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
	
	// Trim whitespace from the text
	return strings.TrimSpace(text)
}

package no_invalid_void_type

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoInvalidVoidTypeOptions struct {
	AllowInGenericTypeArguments interface{} `json:"allowInGenericTypeArguments"`
	AllowAsThisParameter        bool        `json:"allowAsThisParameter"`
}

func buildInvalidVoidNotReturnMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidNotReturn",
		Description: "`void` is only valid as a return type.",
	}
}

func buildInvalidVoidNotReturnOrGenericMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidNotReturnOrGeneric",
		Description: "`void` is only valid as a return type or generic type argument.",
	}
}

func buildInvalidVoidNotReturnOrThisParamMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidNotReturnOrThisParam",
		Description: "`void` is only valid as return type or type of `this` parameter.",
	}
}

func buildInvalidVoidNotReturnOrThisParamOrGenericMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidNotReturnOrThisParamOrGeneric",
		Description: "`void` is only valid as a return type, generic type argument, or type of `this` parameter.",
	}
}

func buildInvalidVoidUnionConstituentMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidUnionConstituent",
		Description: "`void` is not valid as a constituent in a union type.",
	}
}

func buildInvalidVoidForGenericMessage(generic string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidForGeneric",
		Description: "`void` is not valid for the generic argument `" + generic + "`.",
	}
}

var NoInvalidVoidTypeRule = rule.CreateRule(rule.Rule{
	Name: "no-invalid-void-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoInvalidVoidTypeOptions{
			AllowInGenericTypeArguments: true,
			AllowAsThisParameter:        false,
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
				if allowInGenericTypeArguments, ok := optsMap["allowInGenericTypeArguments"]; ok {
					opts.AllowInGenericTypeArguments = allowInGenericTypeArguments
				}
				if allowAsThisParameter, ok := optsMap["allowAsThisParameter"].(bool); ok {
					opts.AllowAsThisParameter = allowAsThisParameter
				}
			}
		}

		// Helper to normalize type name (remove spaces)
		normalizeTypeName := func(name string) string {
			return strings.ReplaceAll(name, " ", "")
		}

		// Helper to get entity name from identifier or qualified name
		// Declare this first since it's used by getTypeReferenceName
		var getEntityName func(node *ast.Node) string
		getEntityName = func(node *ast.Node) string {
			if node == nil {
				return ""
			}

			switch node.Kind {
			case ast.KindIdentifier:
				if id := node.AsIdentifier(); id != nil {
					return id.Text
				}
			case ast.KindQualifiedName:
				if qn := node.AsQualifiedName(); qn != nil {
					left := getEntityName(qn.Left)
					right := ""
					if qn.Right != nil {
						if id := qn.Right.AsIdentifier(); id != nil {
							right = id.Text
						}
					}
					if left != "" && right != "" {
						return left + "." + right
					}
				}
			}
			return ""
		}

		// Helper to get the qualified name from a type reference
		getTypeReferenceName := func(node *ast.Node) string {
			if node.Kind != ast.KindTypeReference {
				return ""
			}
			typeRef := node.AsTypeReferenceNode()
			if typeRef == nil || typeRef.TypeName == nil {
				return ""
			}

			return getEntityName(typeRef.TypeName)
		}

		// Helper to check if a node has void as a type argument
		// Declare this first since it's used by isValidVoidUnion
		var hasVoidTypeArgument func(node *ast.Node) bool
		hasVoidTypeArgument = func(node *ast.Node) bool {
			if node.Kind != ast.KindTypeReference {
				return false
			}
			typeRef := node.AsTypeReferenceNode()
			if typeRef == nil || typeRef.TypeArguments == nil {
				return false
			}

			for _, typeArg := range typeRef.TypeArguments.Nodes {
				if typeArg.Kind == ast.KindVoidKeyword {
					return true
				}
			}
			return false
		}

		// Helper to check if void is allowed in generic context
		var isAllowedInGeneric func(typeRefNode *ast.Node) (bool, string)
		isAllowedInGeneric = func(typeRefNode *ast.Node) (bool, string) {
			// If allowInGenericTypeArguments is false, never allow
			if allow, ok := opts.AllowInGenericTypeArguments.(bool); ok && !allow {
				return false, ""
			}

			// If it's true (default), allow all generics
			if allow, ok := opts.AllowInGenericTypeArguments.(bool); ok && allow {
				return true, ""
			}

			// If it's an array/whitelist, check if the type is in the whitelist
			if allowList, ok := opts.AllowInGenericTypeArguments.([]interface{}); ok {
				typeName := getTypeReferenceName(typeRefNode)
				normalizedTypeName := normalizeTypeName(typeName)

				for _, allowed := range allowList {
					if allowedStr, ok := allowed.(string); ok {
						normalizedAllowed := normalizeTypeName(allowedStr)
						if normalizedTypeName == normalizedAllowed {
							return true, ""
						}
					}
				}
				return false, typeName
			}

			// Default to allow
			return true, ""
		}

		// Helper to check if union contains only void and never (and optionally allowed generics)
		isValidVoidUnion := func(node *ast.Node) bool {
			if node.Kind != ast.KindUnionType {
				return false
			}
			union := node.AsUnionTypeNode()
			if union == nil || union.Types == nil {
				return false
			}

			hasVoid := false

			for _, typeNode := range union.Types.Nodes {
				switch typeNode.Kind {
				case ast.KindVoidKeyword:
					hasVoid = true
				case ast.KindNeverKeyword:
					// never is always allowed in unions with void
					continue
				case ast.KindTypeReference:
					// Check if it's an allowed generic with void
					if !hasVoidTypeArgument(typeNode) {
						return false
					}
					// The type reference must be an allowed generic
					allowed, _ := isAllowedInGeneric(typeNode)
					if !allowed {
						return false
					}
				default:
					// Has other types
					return false
				}
			}

			return hasVoid
		}

		// Helper to check if node is in a function overload signature (not implementation)
		isInOverloadSignature := func(node *ast.Node) bool {
			current := node.Parent

			for current != nil {
				switch current.Kind {
				case ast.KindFunctionDeclaration, ast.KindMethodDeclaration, ast.KindMethodSignature:
					// Check if this is an overload signature (no body)
					switch current.Kind {
					case ast.KindFunctionDeclaration:
						if fn := current.AsFunctionDeclaration(); fn != nil {
							return fn.Body == nil
						}
					case ast.KindMethodDeclaration:
						if method := current.AsMethodDeclaration(); method != nil {
							return method.Body == nil
						}
					case ast.KindMethodSignature:
						// Method signatures in interfaces don't have bodies
						return true
					}
				}
				current = current.Parent
			}
			return false
		}

		// Helper to get the appropriate error message based on options
		var getInvalidVoidMessage func() rule.RuleMessage
		getInvalidVoidMessage = func() rule.RuleMessage {
			allowGeneric := false
			if allow, ok := opts.AllowInGenericTypeArguments.(bool); ok {
				allowGeneric = allow
			} else if allowList, ok := opts.AllowInGenericTypeArguments.([]interface{}); ok {
				allowGeneric = len(allowList) > 0
			}

			if opts.AllowAsThisParameter && allowGeneric {
				return buildInvalidVoidNotReturnOrThisParamOrGenericMessage()
			} else if opts.AllowAsThisParameter {
				return buildInvalidVoidNotReturnOrThisParamMessage()
			} else if allowGeneric {
				return buildInvalidVoidNotReturnOrGenericMessage()
			}
			return buildInvalidVoidNotReturnMessage()
		}

		// Helper to check if node is in valid context
		isValidVoidContext := func(node *ast.Node) (bool, rule.RuleMessage) {
			current := node.Parent

			// Walk up the tree to understand context
			for current != nil {
				switch current.Kind {
				// Check for union types first
				case ast.KindUnionType:
					// If this is a valid void union (void | never or void | Promise<void>), continue checking
					if isValidVoidUnion(current) {
						current = current.Parent
						continue
					}

					// Check if we're in a function overload signature
					if isInOverloadSignature(current) {
						current = current.Parent
						continue
					}

					// Otherwise, this is an invalid union constituent
					return false, buildInvalidVoidUnionConstituentMessage()

				// Allow in this parameter if option is enabled
				case ast.KindParameter:
					param := current.AsParameterDeclaration()
					if param != nil && param.Name() != nil {
						if id := param.Name().AsIdentifier(); id != nil && id.Text == "this" {
							if opts.AllowAsThisParameter {
								return true, rule.RuleMessage{}
							}
						}
					}
					// Regular parameter - not allowed
					return false, getInvalidVoidMessage()

				// Allow in function return types
				case ast.KindFunctionType, ast.KindConstructorType:
					return true, rule.RuleMessage{}

				case ast.KindFunctionDeclaration, ast.KindMethodDeclaration,
					ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindMethodSignature:
					// Need to check if we're in the return type or parameter
					// If we got here without hitting a parameter case, we're in return type
					return true, rule.RuleMessage{}

				// Allow in generic type arguments if enabled
				case ast.KindTypeReference:
					allowed, typeName := isAllowedInGeneric(current)
					if allowed {
						return true, rule.RuleMessage{}
					}
					// Not allowed for this specific generic
					if typeName != "" {
						return false, buildInvalidVoidForGenericMessage(typeName)
					}
					return false, getInvalidVoidMessage()

				// Continue checking for other contexts
				case ast.KindTypeAliasDeclaration, ast.KindPropertySignature,
					ast.KindPropertyDeclaration, ast.KindVariableDeclaration,
					ast.KindArrayType, ast.KindTypeOperator, ast.KindIntersectionType,
					ast.KindMappedType, ast.KindConditionalType, ast.KindTypeAssertionExpression,
					ast.KindAsExpression, ast.KindRestType:
					// These are invalid contexts - continue to determine the right message
					current = current.Parent
					continue
				}

				// Move up the tree
				current = current.Parent
			}

			// If we get here, it's invalid
			return false, getInvalidVoidMessage()
		}

		return rule.RuleListeners{
			ast.KindVoidKeyword: func(node *ast.Node) {
				// Check if in valid context
				valid, message := isValidVoidContext(node)
				if valid {
					return
				}

				// Report invalid void usage
				ctx.ReportNode(node, message)
			},
		}
	},
})

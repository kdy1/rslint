package prefer_readonly_parameter_types

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type AllowSpecifier struct {
	From    string `json:"from"`
	Name    string `json:"name"`
	Package string `json:"package"`
	Path    string `json:"path"`
}

type PreferReadonlyParameterTypesOptions struct {
	CheckParameterProperties bool             `json:"checkParameterProperties"`
	IgnoreInferredTypes      bool             `json:"ignoreInferredTypes"`
	TreatMethodsAsReadonly   bool             `json:"treatMethodsAsReadonly"`
	Allow                    []AllowSpecifier `json:"allow"`
}

func buildShouldBeReadonlyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "shouldBeReadonly",
		Description: "Parameter should be a read only type.",
	}
}

func parseOptions(options any) PreferReadonlyParameterTypesOptions {
	opts := PreferReadonlyParameterTypesOptions{
		CheckParameterProperties: true,
		IgnoreInferredTypes:      false,
		TreatMethodsAsReadonly:   false,
		Allow:                    []AllowSpecifier{},
	}
	if options == nil {
		return opts
	}
	// Handle array format: [{ option: value }]
	if arr, ok := options.([]interface{}); ok {
		if len(arr) > 0 {
			if m, ok := arr[0].(map[string]interface{}); ok {
				if v, ok := m["checkParameterProperties"].(bool); ok {
					opts.CheckParameterProperties = v
				}
				if v, ok := m["ignoreInferredTypes"].(bool); ok {
					opts.IgnoreInferredTypes = v
				}
				if v, ok := m["treatMethodsAsReadonly"].(bool); ok {
					opts.TreatMethodsAsReadonly = v
				}
				if v, ok := m["allow"].([]interface{}); ok {
					opts.Allow = parseAllowList(v)
				}
			}
		}
		return opts
	}
	// Handle direct object format
	if m, ok := options.(map[string]interface{}); ok {
		if v, ok := m["checkParameterProperties"].(bool); ok {
			opts.CheckParameterProperties = v
		}
		if v, ok := m["ignoreInferredTypes"].(bool); ok {
			opts.IgnoreInferredTypes = v
		}
		if v, ok := m["treatMethodsAsReadonly"].(bool); ok {
			opts.TreatMethodsAsReadonly = v
		}
		if v, ok := m["allow"].([]interface{}); ok {
			opts.Allow = parseAllowList(v)
		}
	}
	return opts
}

func parseAllowList(allowList []interface{}) []AllowSpecifier {
	result := make([]AllowSpecifier, 0, len(allowList))
	for _, item := range allowList {
		if m, ok := item.(map[string]interface{}); ok {
			spec := AllowSpecifier{}
			if v, ok := m["from"].(string); ok {
				spec.From = v
			}
			if v, ok := m["name"].(string); ok {
				spec.Name = v
			}
			if v, ok := m["package"].(string); ok {
				spec.Package = v
			}
			if v, ok := m["path"].(string); ok {
				spec.Path = v
			}
			result = append(result, spec)
		}
	}
	return result
}

// isTypeBrandedLiteralLike checks if a type is a branded literal-like type
// (e.g., string & { __brand: unique symbol })
func isTypeBrandedLiteralLike(typeChecker *checker.Checker, t *checker.Type) bool {
	if t == nil {
		return false
	}

	flags := checker.Type_flags(t)

	// Check if it's an intersection type
	if flags&checker.TypeFlagsIntersection == 0 {
		return false
	}

	types := t.Types()
	if len(types) == 0 {
		return false
	}

	// Check if one part is a primitive or literal type
	hasPrimitiveOrLiteral := false
	hasObjectType := false

	for _, subType := range types {
		subFlags := checker.Type_flags(subType)
		if subFlags&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|
			checker.TypeFlagsBooleanLike|checker.TypeFlagsBigIntLike|
			checker.TypeFlagsESSymbolLike|checker.TypeFlagsTemplateLiteral) != 0 {
			hasPrimitiveOrLiteral = true
		}
		if subFlags&checker.TypeFlagsObject != 0 {
			hasObjectType = true
		}
	}

	return hasPrimitiveOrLiteral && hasObjectType
}

// isReadonlyType checks if a type is readonly
func isReadonlyType(program *compiler.Program, typeChecker *checker.Checker, t *checker.Type, opts PreferReadonlyParameterTypesOptions, seenTypes map[*checker.Type]bool) bool {
	if t == nil {
		return false
	}

	// Prevent infinite recursion for circular types
	if seenTypes[t] {
		return true
	}
	seenTypes[t] = true

	flags := checker.Type_flags(t)

	// Primitives are always readonly
	if flags&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|
		checker.TypeFlagsBooleanLike|checker.TypeFlagsBigIntLike|
		checker.TypeFlagsVoidLike|checker.TypeFlagsUndefined|
		checker.TypeFlagsNull|checker.TypeFlagsNever|
		checker.TypeFlagsESSymbolLike|checker.TypeFlagsAny|
		checker.TypeFlagsUnknown) != 0 {
		return true
	}

	// Enum types are readonly
	if flags&checker.TypeFlagsEnumLike != 0 {
		return true
	}

	// Function types are readonly (check for call/construct signatures)
	callSignatures := utils.GetCallSignatures(typeChecker, t)
	constructSignatures := utils.GetConstructSignatures(typeChecker, t)
	if len(callSignatures) > 0 || len(constructSignatures) > 0 {
		// If there are only signatures and no properties, it's a function type
		props := checker.Checker_getPropertiesOfType(typeChecker, t)
		if len(props) == 0 {
			return true
		}
		// If treatMethodsAsReadonly is enabled and it has call signatures, treat as readonly
		if opts.TreatMethodsAsReadonly {
			return true
		}
	}

	// Union types - all members must be readonly
	if flags&checker.TypeFlagsUnion != 0 {
		types := t.Types()
		for _, memberType := range types {
			if !isReadonlyType(program, typeChecker, memberType, opts, seenTypes) {
				return false
			}
		}
		return true
	}

	// Intersection types - check if it's a branded literal or all parts are readonly
	if flags&checker.TypeFlagsIntersection != 0 {
		// Check for branded literal types (e.g., string & { __brand: symbol })
		if isTypeBrandedLiteralLike(typeChecker, t) {
			return true
		}

		// All intersection parts must be readonly
		types := t.Types()
		for _, memberType := range types {
			if !isReadonlyType(program, typeChecker, memberType, opts, seenTypes) {
				return false
			}
		}
		return true
	}

	// Object types - simplified check for empty interfaces/types
	if flags&checker.TypeFlagsObject != 0 {
		props := checker.Checker_getPropertiesOfType(typeChecker, t)
		callSigs := utils.GetCallSignatures(typeChecker, t)
		constructSigs := utils.GetConstructSignatures(typeChecker, t)

		// Empty interfaces are considered readonly
		if len(props) == 0 && len(callSigs) == 0 && len(constructSigs) == 0 {
			return true
		}

		// For now, conservatively treat object types as mutable
		// A full implementation would check each property for readonly modifiers
		return false
	}

	// Type parameters - check constraint
	if flags&checker.TypeFlagsTypeParameter != 0 {
		constraint := checker.Checker_getBaseConstraintOfType(typeChecker, t)
		if constraint != nil {
			return isReadonlyType(program, typeChecker, constraint, opts, seenTypes)
		}
		// Unconstrained type parameters are considered mutable
		return false
	}

	// Default to not readonly for safety
	return false
}

// checkParameter validates a parameter node
func checkParameter(program *compiler.Program, ctx rule.RuleContext, param *ast.Node, opts PreferReadonlyParameterTypesOptions) {
	var actualParam *ast.Node

	// Handle TSParameterProperty
	if param.Kind == ast.KindParameter {
		paramDecl := param.AsParameterDeclaration()
		if paramDecl != nil {
			// Check for parameter properties (constructor parameters with modifiers)
			if !opts.CheckParameterProperties {
				// Skip parameter properties if not checking them
				modifiers := paramDecl.Modifiers()
				if modifiers != nil && len(modifiers.Nodes) > 0 {
					for _, mod := range modifiers.Nodes {
						if mod.Kind == ast.KindPublicKeyword ||
						   mod.Kind == ast.KindPrivateKeyword ||
						   mod.Kind == ast.KindProtectedKeyword ||
						   mod.Kind == ast.KindReadonlyKeyword {
							return
						}
					}
				}
			}
			actualParam = param
		}
	} else {
		actualParam = param
	}

	if actualParam == nil {
		return
	}

	paramDecl := actualParam.AsParameterDeclaration()
	if paramDecl == nil {
		return
	}

	// Skip if ignoring inferred types and parameter has no explicit type annotation
	if opts.IgnoreInferredTypes && paramDecl.Type == nil {
		return
	}

	// Get the type of the parameter
	paramType := ctx.TypeChecker.GetTypeAtLocation(actualParam)
	if paramType == nil {
		return
	}

	// Check if it's a branded literal-like type (exempt from the rule)
	if isTypeBrandedLiteralLike(ctx.TypeChecker, paramType) {
		return
	}

	// Check if the parameter type is readonly
	seenTypes := make(map[*checker.Type]bool)
	if !isReadonlyType(program, ctx.TypeChecker, paramType, opts, seenTypes) {
		ctx.ReportNode(actualParam, buildShouldBeReadonlyMessage())
	}
}

var PreferReadonlyParameterTypesRule = rule.CreateRule(rule.Rule{
	Name: "prefer-readonly-parameter-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		program := ctx.Program

		checkParameters := func(node *ast.Node) {
			params := node.Parameters()
			if params == nil {
				return
			}

			for _, param := range params {
				checkParameter(program, ctx, param, opts)
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration:       checkParameters,
			ast.KindFunctionExpression:        checkParameters,
			ast.KindArrowFunction:             checkParameters,
			ast.KindMethodDeclaration:         checkParameters,
			ast.KindMethodSignature:           checkParameters,
			ast.KindConstructor:               checkParameters,
			ast.KindCallSignature:             checkParameters,
			ast.KindConstructSignature:        checkParameters,
			ast.KindFunctionType:              checkParameters,
		}
	},
})

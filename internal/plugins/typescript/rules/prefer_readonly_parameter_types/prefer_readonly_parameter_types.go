package prefer_readonly_parameter_types

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type PreferReadonlyParameterTypesOptions struct {
	CheckParameterProperties bool     `json:"checkParameterProperties"`
	IgnoreInferredTypes      bool     `json:"ignoreInferredTypes"`
	TreatMethodsAsReadonly   bool     `json:"treatMethodsAsReadonly"`
	Allow                    []string `json:"allow"`
}

func buildShouldBeReadonlyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "shouldBeReadonly",
		Description: "Parameter should be a readonly type.",
	}
}

func parseOptions(options any) PreferReadonlyParameterTypesOptions {
	opts := PreferReadonlyParameterTypesOptions{
		CheckParameterProperties: true,
		IgnoreInferredTypes:      false,
		TreatMethodsAsReadonly:   false,
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
					opts.Allow = make([]string, 0, len(v))
					for _, item := range v {
						if s, ok := item.(string); ok {
							opts.Allow = append(opts.Allow, s)
						}
					}
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
			opts.Allow = make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok {
					opts.Allow = append(opts.Allow, s)
				}
			}
		}
	}
	return opts
}

type readonlynessResult int

const (
	readonlynessUnknown readonlynessResult = iota
	readonlynessMutable
	readonlynessReadonly
)

// isPropertyReadonly checks if a property is readonly in a type
func isPropertyReadonly(t *checker.Type, propName string, typeChecker *checker.Checker) bool {
	// Get all properties of the type
	properties := t.GetProperties()
	for _, prop := range properties {
		if prop.EscapedName == propName {
			// Check if the property has readonly modifier
			if len(prop.Declarations) > 0 {
				decl := prop.Declarations[0]
				// Check for readonly modifier on property declaration
				if decl.Modifiers() != nil {
					for _, mod := range decl.Modifiers() {
						if mod.Kind == ast.KindReadonlyKeyword {
							return true
						}
					}
				}
			}
			return false
		}
	}
	// If property not found, conservatively return false
	return false
}

// isReadonlyType checks if a type is readonly
func isReadonlyType(
	program *compiler.Program,
	typeChecker *checker.Checker,
	t *checker.Type,
	opts PreferReadonlyParameterTypesOptions,
	seenTypes map[*checker.Type]bool,
) readonlynessResult {
	if t == nil {
		return readonlynessUnknown
	}

	// Check if we've seen this type before (recursive types)
	if seenTypes[t] {
		return readonlynessReadonly
	}

	// Check allow list
	if len(opts.Allow) > 0 {
		typeString := typeChecker.TypeToString(t)
		for _, allowed := range opts.Allow {
			if typeString == allowed {
				return readonlynessReadonly
			}
		}
	}

	flags := checker.Type_flags(t)

	// Primitives are always readonly
	if utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|
		checker.TypeFlagsBooleanLike|checker.TypeFlagsBigIntLike|
		checker.TypeFlagsVoidLike|checker.TypeFlagsUndefined|
		checker.TypeFlagsNull|checker.TypeFlagsNever|
		checker.TypeFlagsESSymbolLike|checker.TypeFlagsAny|
		checker.TypeFlagsUnknown) {
		return readonlynessReadonly
	}

	// Enum types
	if utils.IsTypeFlagSet(t, checker.TypeFlagsEnumLike) {
		return readonlynessReadonly
	}

	// Mark this type as seen
	seenTypes[t] = true
	defer delete(seenTypes, t)

	// Check for arrays and tuples
	if arrayResult := isReadonlyArrayOrTuple(program, typeChecker, t, opts, seenTypes); arrayResult != readonlynessUnknown {
		return arrayResult
	}

	// Union types - all members must be readonly
	if flags&checker.TypeFlagsUnion != 0 {
		for _, memberType := range t.Types() {
			if isReadonlyType(program, typeChecker, memberType, opts, seenTypes) == readonlynessMutable {
				return readonlynessMutable
			}
		}
		return readonlynessReadonly
	}

	// Intersection types - at least one member must provide readonly constraint
	if flags&checker.TypeFlagsIntersection != 0 {
		hasReadonly := false
		for _, memberType := range t.Types() {
			result := isReadonlyType(program, typeChecker, memberType, opts, seenTypes)
			if result == readonlynessReadonly {
				hasReadonly = true
			} else if result == readonlynessMutable {
				return readonlynessMutable
			}
		}
		if hasReadonly {
			return readonlynessReadonly
		}
		return readonlynessUnknown
	}

	// Check for objects
	if flags&checker.TypeFlagsObject != 0 {
		return isReadonlyObject(program, typeChecker, t, opts, seenTypes)
	}

	return readonlynessUnknown
}

// isReadonlyArrayOrTuple checks if array/tuple types are readonly
func isReadonlyArrayOrTuple(
	program *compiler.Program,
	typeChecker *checker.Checker,
	t *checker.Type,
	opts PreferReadonlyParameterTypesOptions,
	seenTypes map[*checker.Type]bool,
) readonlynessResult {
	// Check if it's an array type
	if checker.Checker_isArrayType(typeChecker, t) {
		symbol := checker.Type_symbol(t)
		if symbol != nil {
			escapedName := symbol.EscapedName
			// Mutable Array type
			if escapedName == "Array" {
				return readonlynessMutable
			}
			// ReadonlyArray
			if escapedName == "ReadonlyArray" {
				typeArgs := checker.Checker_getTypeArguments(typeChecker, t)
				if len(typeArgs) > 0 {
					// Check element type is also readonly
					for _, typeArg := range typeArgs {
						if isReadonlyType(program, typeChecker, typeArg, opts, seenTypes) == readonlynessMutable {
							return readonlynessMutable
						}
					}
				}
				return readonlynessReadonly
			}
		}
	}

	// Check if it's a tuple type
	if checker.IsTupleType(t) {
		target := t.Target()
		if target != nil && target.AsTupleType() != nil {
			// Check if tuple is readonly
			if !target.AsTupleType().Readonly {
				return readonlynessMutable
			}
			// Check all element types are readonly
			typeArgs := checker.Checker_getTypeArguments(typeChecker, t)
			for _, typeArg := range typeArgs {
				if isReadonlyType(program, typeChecker, typeArg, opts, seenTypes) == readonlynessMutable {
					return readonlynessMutable
				}
			}
			return readonlynessReadonly
		}
	}

	return readonlynessUnknown
}

// isReadonlyObject checks if an object type is readonly
func isReadonlyObject(
	program *compiler.Program,
	typeChecker *checker.Checker,
	t *checker.Type,
	opts PreferReadonlyParameterTypesOptions,
	seenTypes map[*checker.Type]bool,
) readonlynessResult {
	// Check index signatures
	stringIndexType := typeChecker.GetIndexTypeOfType(t, checker.IndexKindString)
	if stringIndexType != nil {
		indexInfo := typeChecker.GetIndexInfoOfType(t, checker.IndexKindString)
		if indexInfo != nil && !indexInfo.IsReadonly {
			return readonlynessMutable
		}
		if stringIndexType != t && !seenTypes[stringIndexType] {
			if isReadonlyType(program, typeChecker, stringIndexType, opts, seenTypes) == readonlynessMutable {
				return readonlynessMutable
			}
		}
	}

	numberIndexType := typeChecker.GetIndexTypeOfType(t, checker.IndexKindNumber)
	if numberIndexType != nil {
		indexInfo := typeChecker.GetIndexInfoOfType(t, checker.IndexKindNumber)
		if indexInfo != nil && !indexInfo.IsReadonly {
			return readonlynessMutable
		}
		if numberIndexType != t && !seenTypes[numberIndexType] {
			if isReadonlyType(program, typeChecker, numberIndexType, opts, seenTypes) == readonlynessMutable {
				return readonlynessMutable
			}
		}
	}

	// Check properties
	properties := t.GetProperties()
	if len(properties) > 0 {
		for _, prop := range properties {
			// Check if property is a method and we should treat methods as readonly
			if opts.TreatMethodsAsReadonly {
				if len(prop.Declarations) > 0 {
					decl := prop.Declarations[len(prop.Declarations)-1]
					if ast.IsMethodDeclaration(decl) || ast.IsMethodSignature(decl) {
						continue
					}
					// Check if it's a function-valued property
					propType := typeChecker.GetTypeOfSymbolAtLocation(prop, decl)
					if propType != nil && utils.IsTypeFlagSet(propType, checker.TypeFlagsObject) {
						callSigs := utils.GetCallSignatures(typeChecker, propType)
						if len(callSigs) > 0 {
							continue
						}
					}
				}
			}

			// Check if this is a private identifier (always readonly)
			if len(prop.Declarations) > 0 {
				decl := prop.Declarations[0]
				if decl.Name() != nil && ast.IsPrivateIdentifier(decl.Name()) {
					continue
				}
			}

			// Check if property is readonly
			if !isPropertyReadonly(t, prop.EscapedName, typeChecker) {
				return readonlynessMutable
			}
		}

		// All properties are readonly, now check their types
		for _, prop := range properties {
			if len(prop.Declarations) > 0 {
				decl := prop.Declarations[0]
				propType := typeChecker.GetTypeOfSymbolAtLocation(prop, decl)
				if propType != nil && !seenTypes[propType] {
					if isReadonlyType(program, typeChecker, propType, opts, seenTypes) == readonlynessMutable {
						return readonlynessMutable
					}
				}
			}
		}

		return readonlynessReadonly
	}

	// Empty interface/object
	return readonlynessReadonly
}

// checkParameter validates a parameter node
func checkParameter(ctx rule.RuleContext, param *ast.Node, opts PreferReadonlyParameterTypesOptions) {
	paramDecl := param.AsParameterDeclaration()
	if paramDecl == nil {
		return
	}

	// Skip if ignoring inferred types and parameter has no explicit type annotation
	if opts.IgnoreInferredTypes && paramDecl.Type == nil {
		return
	}

	// Get the type of the parameter
	paramType := ctx.TypeChecker.GetTypeAtLocation(param)
	if paramType == nil {
		return
	}

	// Create a new seenTypes map for each parameter check
	seenTypes := make(map[*checker.Type]bool)

	// Check if the parameter type is readonly
	result := isReadonlyType(ctx.Program, ctx.TypeChecker, paramType, opts, seenTypes)
	if result == readonlynessMutable {
		ctx.ReportNode(param, buildShouldBeReadonlyMessage())
	}
}

var PreferReadonlyParameterTypesRule = rule.CreateRule(rule.Rule{
	Name: "prefer-readonly-parameter-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		checkParameters := func(node *ast.Node) {
			params := node.Parameters()
			if params == nil {
				return
			}

			for _, param := range params {
				checkParameter(ctx, param, opts)
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: checkParameters,
			ast.KindFunctionExpression:  checkParameters,
			ast.KindArrowFunction:       checkParameters,
			ast.KindMethodDeclaration:   checkParameters,
			ast.KindConstructor: func(node *ast.Node) {
				// For constructors, check parameter properties if enabled
				if !opts.CheckParameterProperties {
					return
				}
				checkParameters(node)
			},
		}
	},
})

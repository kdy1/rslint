package restrict_template_expressions

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type TypeOrValueFromSpecifier struct {
	From string `json:"from"`
	Name string `json:"name"`
}

type RestrictTemplateExpressionsOptions struct {
	AllowNumber  *bool                      `json:"allowNumber"`
	AllowBoolean *bool                      `json:"allowBoolean"`
	AllowAny     *bool                      `json:"allowAny"`
	AllowNullish *bool                      `json:"allowNullish"`
	AllowRegExp  *bool                      `json:"allowRegExp"`
	AllowNever   *bool                      `json:"allowNever"`
	AllowArray   *bool                      `json:"allowArray"`
	Allow        []TypeOrValueFromSpecifier `json:"allow"`
}

func boolPtr(b bool) *bool {
	return &b
}

func getDefaultAllowList() []TypeOrValueFromSpecifier {
	return []TypeOrValueFromSpecifier{
		{From: "lib", Name: "Error"},
		{From: "lib", Name: "URL"},
		{From: "lib", Name: "URLSearchParams"},
	}
}

// RestrictTemplateExpressionsRule implements the restrict-template-expressions rule
// Enforce template literal expressions to be of string type
var RestrictTemplateExpressionsRule = rule.CreateRule(rule.Rule{
	Name: "restrict-template-expressions",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := RestrictTemplateExpressionsOptions{
		AllowNumber:  boolPtr(true),
		AllowBoolean: boolPtr(true),
		AllowAny:     boolPtr(true),
		AllowNullish: boolPtr(true),
		AllowRegExp:  boolPtr(true),
		AllowNever:   boolPtr(false),
		AllowArray:   boolPtr(false),
		Allow:        getDefaultAllowList(),
	}

	// Parse options
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
			if allowNumber, ok := optsMap["allowNumber"].(bool); ok {
				opts.AllowNumber = &allowNumber
			}
			if allowBoolean, ok := optsMap["allowBoolean"].(bool); ok {
				opts.AllowBoolean = &allowBoolean
			}
			if allowAny, ok := optsMap["allowAny"].(bool); ok {
				opts.AllowAny = &allowAny
			}
			if allowNullish, ok := optsMap["allowNullish"].(bool); ok {
				opts.AllowNullish = &allowNullish
			}
			if allowRegExp, ok := optsMap["allowRegExp"].(bool); ok {
				opts.AllowRegExp = &allowRegExp
			}
			if allowNever, ok := optsMap["allowNever"].(bool); ok {
				opts.AllowNever = &allowNever
			}
			if allowArray, ok := optsMap["allowArray"].(bool); ok {
				opts.AllowArray = &allowArray
			}
			if allow, ok := optsMap["allow"].([]interface{}); ok {
				opts.Allow = []TypeOrValueFromSpecifier{}
				for _, item := range allow {
					if itemMap, ok := item.(map[string]interface{}); ok {
						spec := TypeOrValueFromSpecifier{}
						if from, ok := itemMap["from"].(string); ok {
							spec.From = from
						}
						if name, ok := itemMap["name"].(string); ok {
							spec.Name = name
						}
						opts.Allow = append(opts.Allow, spec)
					}
				}
			}
		}
	}

	isAllowedType := func(t *checker.Type) bool {
		if t == nil {
			return false
		}

		typeName := utils.GetTypeName(ctx.TypeChecker, t)

		// Check if type is in the allow list
		for _, allowSpec := range opts.Allow {
			if allowSpec.From == "lib" {
				// For lib types, match by type name
				if typeName == allowSpec.Name {
					return true
				}
			} else if allowSpec.From == "file" {
				// For file types, check if the type or any base type matches
				if isTypeOrBaseType(ctx.TypeChecker, t, allowSpec.Name) {
					return true
				}
			}
		}

		return false
	}

	var isValidTemplateExpressionType func(t *checker.Type) bool
	isValidTemplateExpressionType = func(t *checker.Type) bool {
		if t == nil {
			return false
		}

		// Handle type parameters (generics)
		if utils.IsTypeParameter(t) {
			constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, t)
			if constraint != nil {
				return isValidTemplateExpressionType(constraint)
			}
			// Unconstrained type parameter is not allowed
			return false
		}

		// Check if type is in allow list
		if isAllowedType(t) {
			return true
		}

		// String is always allowed
		if utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike) {
			return true
		}

		// Number and bigint
		if opts.AllowNumber != nil && *opts.AllowNumber && utils.IsTypeFlagSet(t, checker.TypeFlagsNumberLike|checker.TypeFlagsBigIntLike) {
			return true
		}

		// Boolean
		if opts.AllowBoolean != nil && *opts.AllowBoolean && utils.IsTypeFlagSet(t, checker.TypeFlagsBooleanLike) {
			return true
		}

		// Nullish (null | undefined)
		if opts.AllowNullish != nil && *opts.AllowNullish && (utils.IsTypeFlagSet(t, checker.TypeFlagsNull) || utils.IsTypeFlagSet(t, checker.TypeFlagsUndefined) || utils.IsTypeFlagSet(t, checker.TypeFlagsVoid)) {
			return true
		}

		// Never
		if opts.AllowNever != nil && *opts.AllowNever && utils.IsTypeFlagSet(t, checker.TypeFlagsNever) {
			return true
		}

		// Any
		if opts.AllowAny != nil && *opts.AllowAny && utils.IsTypeFlagSet(t, checker.TypeFlagsAny) {
			return true
		}

		// RegExp
		if opts.AllowRegExp != nil && *opts.AllowRegExp {
			typeName := utils.GetTypeName(ctx.TypeChecker, t)
			if typeName == "RegExp" {
				return true
			}
		}

		// Array/Tuple
		if opts.AllowArray != nil && *opts.AllowArray {
			if checker.Checker_isArrayType(ctx.TypeChecker, t) || checker.IsTupleType(t) {
				// For tuples, need to check if all element types are allowed
				if checker.IsTupleType(t) {
					typeArgs := checker.Checker_getTypeArguments(ctx.TypeChecker, t)
					for _, elemType := range typeArgs {
						if !isValidTemplateExpressionType(elemType) {
							return false
						}
					}
					return true
				}
				// For regular arrays, check element type
				elemType := utils.GetNumberIndexType(ctx.TypeChecker, t)
				if elemType != nil {
					return isValidTemplateExpressionType(elemType)
				}
				return true
			}
		}

		// Union types - all parts must be valid
		if utils.IsUnionType(t) {
			parts := utils.UnionTypeParts(t)
			for _, part := range parts {
				if !isValidTemplateExpressionType(part) {
					return false
				}
			}
			return true
		}

		// Intersection types - at least one part must be valid string-like
		if utils.IsIntersectionType(t) {
			parts := utils.IntersectionTypeParts(t)
			for _, part := range parts {
				if isValidTemplateExpressionType(part) {
					return true
				}
			}
			return false
		}

		return false
	}

	checkTemplateExpression := func(node *ast.Node) {
		exprType := ctx.TypeChecker.GetTypeAtLocation(node)

		if !isValidTemplateExpressionType(exprType) {
			typeName := utils.GetTypeName(ctx.TypeChecker, exprType)
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "invalidType",
				Description: "Invalid type \"" + typeName + "\" of template literal expression.",
				Data: map[string]interface{}{
					"type": typeName,
				},
			})
		}
	}

	return rule.RuleListeners{
		ast.KindTemplateExpression: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			// Tagged templates are not checked
			parent := node.Parent
			if parent != nil && ast.IsTaggedTemplateExpression(parent) {
				return
			}

			templateExpr := node.AsTemplateExpression()
			if templateExpr == nil || templateExpr.TemplateSpans == nil {
				return
			}

			// Check each template span's expression
			for _, span := range templateExpr.TemplateSpans.Nodes {
				templateSpan := span.AsTemplateSpan()
				if templateSpan == nil || templateSpan.Expression == nil {
					continue
				}

				checkTemplateExpression(templateSpan.Expression)
			}
		},
	}
}

// isTypeOrBaseType checks if a type matches a name or has a base type that matches
func isTypeOrBaseType(checker *checker.Checker, t *checker.Type, name string) bool {
	if t == nil {
		return false
	}

	// Check if the type itself matches
	typeName := utils.GetTypeName(checker, t)
	if typeName == name {
		return true
	}

	// For object types, check base types (class inheritance and interface extension)
	if utils.IsTypeFlagSet(t, checker.TypeFlagsObject) {
		// Check base types
		baseTypes := checker.Checker_getBaseTypes(t)
		if baseTypes != nil {
			for _, baseType := range baseTypes {
				if isTypeOrBaseType(checker, baseType, name) {
					return true
				}
			}
		}
	}

	// For type aliases, check the alias symbol
	if symbol := t.AliasSymbol; symbol != nil {
		if symbol.Name == name {
			return true
		}
	}

	return false
}

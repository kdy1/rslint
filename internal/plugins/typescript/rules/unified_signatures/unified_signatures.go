package unified_signatures

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type UnifiedSignaturesOptions struct {
	IgnoreDifferentlyNamedParameters  bool `json:"ignoreDifferentlyNamedParameters"`
	IgnoreOverloadsWithDifferentJSDoc bool `json:"ignoreOverloadsWithDifferentJSDoc"`
}

type SignatureInfo struct {
	Node           *ast.Node
	Parameters     []*ast.Node
	TypeParameters []*ast.Node
	ReturnType     *ast.Node
	JSDocComment   string
}

type OverloadKey struct {
	Name      string
	Static    bool
	NameType  utils.MemberNameType
	IsPrivate bool
}

func (k OverloadKey) String() string {
	return fmt.Sprintf("%s:%t:%d:%t", k.Name, k.Static, k.NameType, k.IsPrivate)
}

// getJSDocComment extracts JSDoc comment text from a node
func getJSDocComment(sourceFile *ast.SourceFile, node *ast.Node) string {
	if node == nil {
		return ""
	}

	// For now, we don't check JSDoc since ast.GetJSDocTags is not available
	// This can be improved when the API becomes available
	return ""
}

// getSignatureInfo extracts signature information from a node
func getSignatureInfo(ctx rule.RuleContext, node *ast.Node) *SignatureInfo {
	if node == nil {
		return nil
	}

	var params []*ast.Node
	var typeParams []*ast.Node
	var returnType *ast.Node

	switch node.Kind {
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl == nil {
			return nil
		}
		if funcDecl.Parameters != nil {
			params = funcDecl.Parameters.Nodes
		}
		if funcDecl.TypeParameters != nil {
			typeParams = funcDecl.TypeParameters.Nodes
		}
		returnType = funcDecl.Type

	case ast.KindMethodDeclaration:
		methodDecl := node.AsMethodDeclaration()
		if methodDecl == nil {
			return nil
		}
		if methodDecl.Parameters != nil {
			params = methodDecl.Parameters.Nodes
		}
		if methodDecl.TypeParameters != nil {
			typeParams = methodDecl.TypeParameters.Nodes
		}
		returnType = methodDecl.Type

	case ast.KindMethodSignature:
		methodSig := node.AsMethodSignatureDeclaration()
		if methodSig == nil {
			return nil
		}
		if methodSig.Parameters != nil {
			params = methodSig.Parameters.Nodes
		}
		if methodSig.TypeParameters != nil {
			typeParams = methodSig.TypeParameters.Nodes
		}
		returnType = methodSig.Type

	case ast.KindConstructor:
		constructor := node.AsConstructorDeclaration()
		if constructor == nil {
			return nil
		}
		if constructor.Parameters != nil {
			params = constructor.Parameters.Nodes
		}
		if constructor.TypeParameters != nil {
			typeParams = constructor.TypeParameters.Nodes
		}

	case ast.KindConstructSignature:
		constructSig := node.AsConstructSignatureDeclaration()
		if constructSig == nil {
			return nil
		}
		if constructSig.Parameters != nil {
			params = constructSig.Parameters.Nodes
		}
		if constructSig.TypeParameters != nil {
			typeParams = constructSig.TypeParameters.Nodes
		}
		returnType = constructSig.Type

	case ast.KindCallSignature:
		callSig := node.AsCallSignatureDeclaration()
		if callSig == nil {
			return nil
		}
		if callSig.Parameters != nil {
			params = callSig.Parameters.Nodes
		}
		if callSig.TypeParameters != nil {
			typeParams = callSig.TypeParameters.Nodes
		}
		returnType = callSig.Type

	default:
		return nil
	}

	jsdoc := getJSDocComment(ctx.SourceFile, node)

	return &SignatureInfo{
		Node:           node,
		Parameters:     params,
		TypeParameters: typeParams,
		ReturnType:     returnType,
		JSDocComment:   jsdoc,
	}
}

// getOverloadKey generates a unique key for an overload based on name and properties
func getOverloadKey(ctx rule.RuleContext, node *ast.Node) *OverloadKey {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindFunctionDeclaration:
		funcDecl := node.AsFunctionDeclaration()
		if funcDecl == nil || funcDecl.Name() == nil {
			return nil
		}
		name := funcDecl.Name().Text()
		return &OverloadKey{
			Name:      name,
			Static:    false,
			NameType:  utils.MemberNameTypeNormal,
			IsPrivate: false,
		}

	case ast.KindMethodDeclaration:
		methodDecl := node.AsMethodDeclaration()
		if methodDecl == nil || methodDecl.Name() == nil {
			return nil
		}
		name, nameType := utils.GetNameFromMember(ctx.SourceFile, methodDecl.Name())
		isPrivate := ast.IsPrivateIdentifier(methodDecl.Name())
		return &OverloadKey{
			Name:      name,
			Static:    ast.IsStatic(node),
			NameType:  nameType,
			IsPrivate: isPrivate,
		}

	case ast.KindMethodSignature:
		methodSig := node.AsMethodSignatureDeclaration()
		if methodSig == nil || methodSig.Name() == nil {
			return nil
		}
		name, nameType := utils.GetNameFromMember(ctx.SourceFile, methodSig.Name())
		isPrivate := ast.IsPrivateIdentifier(methodSig.Name())
		return &OverloadKey{
			Name:      name,
			Static:    false,
			NameType:  nameType,
			IsPrivate: isPrivate,
		}

	case ast.KindConstructor:
		return &OverloadKey{
			Name:      "constructor",
			Static:    false,
			NameType:  utils.MemberNameTypeNormal,
			IsPrivate: false,
		}

	case ast.KindConstructSignature:
		return &OverloadKey{
			Name:      "new",
			Static:    false,
			NameType:  utils.MemberNameTypeNormal,
			IsPrivate: false,
		}

	case ast.KindCallSignature:
		return &OverloadKey{
			Name:      "call",
			Static:    false,
			NameType:  utils.MemberNameTypeNormal,
			IsPrivate: false,
		}
	}

	return nil
}

// getParameterName gets the name of a parameter
func getParameterName(sourceFile *ast.SourceFile, param *ast.Node) string {
	if param == nil {
		return ""
	}

	paramDecl := param.AsParameterDeclaration()
	if paramDecl == nil {
		return ""
	}

	if paramDecl.Name() == nil {
		return ""
	}

	nameRange := utils.TrimNodeTextRange(sourceFile, paramDecl.Name())
	return sourceFile.Text()[nameRange.Pos():nameRange.End()]
}

// getTypeText gets the text representation of a type
func getTypeText(sourceFile *ast.SourceFile, typeNode *ast.Node) string{
	if typeNode == nil {
		return ""
	}

	typeRange := utils.TrimNodeTextRange(sourceFile, typeNode)
	return sourceFile.Text()[typeRange.Pos():typeRange.End()]
}

// areTypeParametersEqual checks if two type parameter lists are equal
func areTypeParametersEqual(sourceFile *ast.SourceFile, params1 []*ast.Node, params2 []*ast.Node) bool {
	if len(params1) != len(params2) {
		return false
	}

	for i := 0; i < len(params1); i++ {
		tp1 := params1[i].AsTypeParameter()
		tp2 := params2[i].AsTypeParameter()

		if tp1 == nil || tp2 == nil {
			return false
		}

		// Check name
		if tp1.Name() == nil || tp2.Name() == nil {
			return false
		}
		if tp1.Name().Text() != tp2.Name().Text() {
			return false
		}

		// Check constraint
		constraint1 := getTypeText(sourceFile, tp1.Constraint)
		constraint2 := getTypeText(sourceFile, tp2.Constraint)
		if constraint1 != constraint2 {
			return false
		}
	}

	return true
}

// checkSignaturesCanBeUnified checks if two consecutive signatures can be unified
func checkSignaturesCanBeUnified(ctx rule.RuleContext, opts UnifiedSignaturesOptions, sig1 *SignatureInfo, sig2 *SignatureInfo, totalOverloads int) *rule.RuleMessage {
	if sig1 == nil || sig2 == nil {
		return nil
	}

	// Check if JSDoc comments are different
	if opts.IgnoreOverloadsWithDifferentJSDoc && sig1.JSDocComment != sig2.JSDocComment {
		return nil
	}

	// Check if type parameters are equal
	if !areTypeParametersEqual(ctx.SourceFile, sig1.TypeParameters, sig2.TypeParameters) {
		return nil
	}

	// Check return types
	returnType1 := getTypeText(ctx.SourceFile, sig1.ReturnType)
	returnType2 := getTypeText(ctx.SourceFile, sig2.ReturnType)
	if returnType1 != returnType2 {
		return nil
	}

	params1 := sig1.Parameters
	params2 := sig2.Parameters

	// Filter out 'this' parameters
	params1 = filterThisParameters(params1)
	params2 = filterThisParameters(params2)

	// Check if one has a 'this' parameter and the other doesn't at the same position
	hasThis1 := hasThisParameter(sig1.Parameters)
	hasThis2 := hasThisParameter(sig2.Parameters)

	// If only one has 'this', they might be comparable by their types
	if hasThis1 != hasThis2 {
		// Compare regular params if 'this' differs
		if len(params1) == len(params2) {
			// Continue to parameter comparison
		} else {
			return nil
		}
	} else if hasThis1 && hasThis2 {
		// Both have 'this', check if types are different
		thisParam1 := getThisParameter(sig1.Parameters)
		thisParam2 := getThisParameter(sig2.Parameters)

		thisType1 := getParameterType(ctx.SourceFile, thisParam1)
		thisType2 := getParameterType(ctx.SourceFile, thisParam2)

		if thisType1 != thisType2 {
			// 'this' types differ - could be unified
			if len(params1) == len(params2) {
				allParamsMatch := true
				for i := 0; i < len(params1); i++ {
					type1 := getParameterType(ctx.SourceFile, params1[i])
					type2 := getParameterType(ctx.SourceFile, params2[i])
					if type1 != type2 {
						allParamsMatch = false
						break
					}
				}

				if allParamsMatch {
					// Only 'this' parameter differs
					failureString := "These overloads can be combined into one signature"
					if totalOverloads > 2 {
						// line1 := 0 // TODO: get line number
						failureString = "This overload and the one above can be combined into one signature"
					}

					return &rule.RuleMessage{
						Id:          "singleParameterDifference",
						Description: failureString,
					}
				}
			}
		}
	}

	// Calculate parameter difference
	minLen := len(params1)
	if len(params2) < minLen {
		minLen = len(params2)
	}

	paramDiffs := 0
	diffIndex := -1

	for i := 0; i < minLen; i++ {
		param1 := params1[i]
		param2 := params2[i]

		type1 := getParameterType(ctx.SourceFile, param1)
		type2 := getParameterType(ctx.SourceFile, param2)

		isOptional1 := isParameterOptional(param1)
		isOptional2 := isParameterOptional(param2)

		isRest1 := isParameterRest(param1)
		isRest2 := isParameterRest(param2)

		// Check parameter names if the option is enabled
		if !opts.IgnoreDifferentlyNamedParameters {
			name1 := getParameterName(ctx.SourceFile, param1)
			name2 := getParameterName(ctx.SourceFile, param2)

			if name1 != name2 {
				// Different names - can't unify unless types also differ
				if type1 == type2 && isOptional1 == isOptional2 && isRest1 == isRest2 {
					return nil
				}
			}
		}

		// If types differ
		if type1 != type2 {
			paramDiffs++
			diffIndex = i
		} else if isOptional1 != isOptional2 {
			// Same type but different optionality
			paramDiffs++
			diffIndex = i
		} else if isRest1 != isRest2 {
			// One is rest, the other isn't
			return nil
		}
	}

	// Check for arity differences
	arityDiff := len(params1) - len(params2)
	if arityDiff < 0 {
		arityDiff = -arityDiff
	}

	// Single parameter difference
	if paramDiffs == 1 && arityDiff == 0 {
		param1 := params1[diffIndex]
		param2 := params2[diffIndex]

		type1 := getParameterType(ctx.SourceFile, param1)
		type2 := getParameterType(ctx.SourceFile, param2)

		isOptional1 := isParameterOptional(param1)
		isOptional2 := isParameterOptional(param2)

		isRest1 := isParameterRest(param1)
		isRest2 := isParameterRest(param2)

		// Both are rest or both are not rest
		if isRest1 == isRest2 {
			// If optionality differs but types are the same
			if type1 == type2 && isOptional1 != isOptional2 {
				failureString := "These overloads can be combined into one signature"
				if totalOverloads > 2 {
					// line1 := 0 // TODO: get line number
					failureString = "This overload and the one above can be combined into one signature"
				}

				return &rule.RuleMessage{
					Id:          "singleParameterDifference",
					Description: failureString,
				}
			}

			// Types differ - can use union
			if type1 != type2 {
				failureString := "These overloads can be combined into one signature"
				if totalOverloads > 2 {
					// line1 := 0 // TODO: get line number
					failureString = "This overload and the one above can be combined into one signature"
				}

				return &rule.RuleMessage{
					Id:          "singleParameterDifference",
					Description: failureString,
				}
			}
		}
	}

	// Check for omitting single parameter
	if arityDiff == 1 && paramDiffs == 0 {
		var longerParams []*ast.Node

		if len(params1) > len(params2) {
			longerParams = params1
		} else {
			longerParams = params2
		}

		// The extra parameter should be optional or rest
		extraParam := longerParams[len(longerParams)-1]
		if isParameterOptional(extraParam) || isParameterRest(extraParam) {
			failureString := "These overloads can be combined into one signature"
			if totalOverloads > 2 {
				// line := 0 // TODO: get line number
				failureString = "This overload and the one above can be combined into one signature"
			}

			return &rule.RuleMessage{
				Id:          "omittingSingleParameter",
				Description: failureString,
			}
		}
	}

	// Check for omitting rest parameter
	if arityDiff > 1 {
		var longerParams []*ast.Node
		var shorterLen int

		if len(params1) > len(params2) {
			longerParams = params1
			shorterLen = len(params2)
		} else {
			longerParams = params2
			shorterLen = len(params1)
		}

		// All extra parameters should be optional or rest
		allOptionalOrRest := true
		for i := shorterLen; i < len(longerParams); i++ {
			if !isParameterOptional(longerParams[i]) && !isParameterRest(longerParams[i]) {
				allOptionalOrRest = false
				break
			}
		}

		if allOptionalOrRest {
			// Check if there's a rest parameter
			hasRest := false
			for i := shorterLen; i < len(longerParams); i++ {
				if isParameterRest(longerParams[i]) {
					hasRest = true
					break
				}
			}

			failureString := "These overloads can be combined into one signature"
			if totalOverloads > 2 {
				// line := 0 // TODO: get line number
				failureString = "This overload and the one above can be combined into one signature"
			}

			if hasRest {
				return &rule.RuleMessage{
					Id:          "omittingRestParameter",
					Description: failureString,
				}
			}

			return &rule.RuleMessage{
				Id:          "omittingSingleParameter",
				Description: failureString,
			}
		}
	}

	return nil
}

// Helper functions

func filterThisParameters(params []*ast.Node) []*ast.Node {
	var filtered []*ast.Node
	for _, param := range params {
		if !isThisParameter(param) {
			filtered = append(filtered, param)
		}
	}
	return filtered
}

func hasThisParameter(params []*ast.Node) bool {
	for _, param := range params {
		if isThisParameter(param) {
			return true
		}
	}
	return false
}

func getThisParameter(params []*ast.Node) *ast.Node {
	for _, param := range params {
		if isThisParameter(param) {
			return param
		}
	}
	return nil
}

func isThisParameter(param *ast.Node) bool {
	if param == nil {
		return false
	}

	paramDecl := param.AsParameterDeclaration()
	if paramDecl == nil {
		return false
	}

	if paramDecl.Name() == nil {
		return false
	}

	return paramDecl.Name().Kind == ast.KindIdentifier && paramDecl.Name().Text() == "this"
}

func getParameterType(sourceFile *ast.SourceFile, param *ast.Node) string {
	if param == nil {
		return ""
	}

	paramDecl := param.AsParameterDeclaration()
	if paramDecl == nil {
		return ""
	}

	return getTypeText(sourceFile, paramDecl.Type)
}

func isParameterOptional(param *ast.Node) bool {
	if param == nil {
		return false
	}

	paramDecl := param.AsParameterDeclaration()
	if paramDecl == nil {
		return false
	}

	return paramDecl.QuestionToken != nil
}

func isParameterRest(param *ast.Node) bool {
	if param == nil {
		return false
	}

	paramDecl := param.AsParameterDeclaration()
	if paramDecl == nil {
		return false
	}

	return paramDecl.DotDotDotToken != nil
}

// checkOverloads checks all overloads for a specific key
func checkOverloads(ctx rule.RuleContext, opts UnifiedSignaturesOptions, overloads []*SignatureInfo) {
	if len(overloads) < 2 {
		return
	}

	// Compare consecutive overloads
	for i := 0; i < len(overloads)-1; i++ {
		for j := i + 1; j < len(overloads); j++ {
			msg := checkSignaturesCanBeUnified(ctx, opts, overloads[i], overloads[j], len(overloads))
			if msg != nil {
				// Report on the later signature
				ctx.ReportNode(overloads[j].Node, *msg)
				return
			}
		}
	}
}

// getMembers returns the members of a container node
func getMembers(node *ast.Node) []*ast.Node {
	switch node.Kind {
	case ast.KindClassDeclaration:
		classDecl := node.AsClassDeclaration()
		if classDecl == nil || classDecl.Members == nil {
			return nil
		}
		return classDecl.Members.Nodes
	case ast.KindSourceFile:
		sourceFile := node.AsSourceFile()
		if sourceFile == nil || sourceFile.Statements == nil {
			return nil
		}
		return sourceFile.Statements.Nodes
	case ast.KindModuleBlock:
		moduleBlock := node.AsModuleBlock()
		if moduleBlock == nil || moduleBlock.Statements == nil {
			return nil
		}
		return moduleBlock.Statements.Nodes
	case ast.KindModuleDeclaration:
		moduleDecl := node.AsModuleDeclaration()
		if moduleDecl == nil || moduleDecl.Body == nil {
			return nil
		}
		return getMembers(moduleDecl.Body)
	case ast.KindInterfaceDeclaration:
		interfaceDecl := node.AsInterfaceDeclaration()
		if interfaceDecl == nil || interfaceDecl.Members == nil {
			return nil
		}
		return interfaceDecl.Members.Nodes
	case ast.KindTypeLiteral:
		typeLiteral := node.AsTypeLiteralNode()
		if typeLiteral == nil || typeLiteral.Members == nil {
			return nil
		}
		return typeLiteral.Members.Nodes
	}
	return nil
}

// checkBodyForOverloads checks a container for unified signatures
func checkBodyForOverloads(ctx rule.RuleContext, opts UnifiedSignaturesOptions, node *ast.Node) {
	members := getMembers(node)
	if members == nil {
		return
	}

	// Group overloads by key
	overloadGroups := make(map[string][]*SignatureInfo)

	for _, member := range members {
		key := getOverloadKey(ctx, member)
		if key == nil {
			continue
		}

		sig := getSignatureInfo(ctx, member)
		if sig == nil {
			continue
		}

		keyStr := key.String()
		overloadGroups[keyStr] = append(overloadGroups[keyStr], sig)
	}

	// Check each group
	for _, group := range overloadGroups {
		checkOverloads(ctx, opts, group)
	}
}

var UnifiedSignaturesRule = rule.CreateRule(rule.Rule{
	Name: "unified-signatures",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := UnifiedSignaturesOptions{
			IgnoreDifferentlyNamedParameters:  false,
			IgnoreOverloadsWithDifferentJSDoc: false,
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
				if val, ok := optsMap["ignoreDifferentlyNamedParameters"].(bool); ok {
					opts.IgnoreDifferentlyNamedParameters = val
				}
				if val, ok := optsMap["ignoreOverloadsWithDifferentJSDoc"].(bool); ok {
					opts.IgnoreOverloadsWithDifferentJSDoc = val
				}
			}
		}

		// Check the source file at the beginning
		checkBodyForOverloads(ctx, opts, &ctx.SourceFile.Node)

		return rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				checkBodyForOverloads(ctx, opts, node)
			},
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				checkBodyForOverloads(ctx, opts, node)
			},
			ast.KindTypeLiteral: func(node *ast.Node) {
				checkBodyForOverloads(ctx, opts, node)
			},
			ast.KindModuleBlock: func(node *ast.Node) {
				checkBodyForOverloads(ctx, opts, node)
			},
			ast.KindModuleDeclaration: func(node *ast.Node) {
				checkBodyForOverloads(ctx, opts, node)
			},
		}
	},
})

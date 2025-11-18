package member_ordering

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// MemberType represents the type of class/interface member
type MemberType string

const (
	// Signatures
	MemberTypeSignature          MemberType = "signature"
	MemberTypeCallSignature      MemberType = "call-signature"
	MemberTypeConstructSignature MemberType = "construct-signature"
	MemberTypeIndexSignature     MemberType = "index-signature"

	// Fields
	MemberTypeField         MemberType = "field"
	MemberTypeReadonlyField MemberType = "readonly-field"

	// Constructors
	MemberTypeConstructor MemberType = "constructor"

	// Getters/Setters
	MemberTypeGetAccessor MemberType = "get"
	MemberTypeSetAccessor MemberType = "set"
	MemberTypeAccessor    MemberType = "accessor"

	// Methods
	MemberTypeMethod MemberType = "method"

	// Static initialization
	MemberTypeStaticInitialization MemberType = "static-initialization"
)

// Modifiers
const (
	ModifierPublic     = "public"
	ModifierProtected  = "protected"
	ModifierPrivate    = "private"
	ModifierAbstract   = "abstract"
	ModifierStatic     = "static"
	ModifierInstance   = "instance"
	ModifierReadonly   = "readonly"
	ModifierNonReadonly = "non-readonly"
	ModifierDecorated  = "decorated"
	ModifierOverride   = "override"
	ModifierOptional   = "optional"
	ModifierRequired   = "required"
)

// OrderType represents the ordering strategy
type OrderType string

const (
	OrderTypeAsWritten                     OrderType = "as-written"
	OrderTypeAlphabetically                OrderType = "alphabetically"
	OrderTypeAlphabeticallyCaseInsensitive OrderType = "alphabetically-case-insensitive"
	OrderTypeNatural                       OrderType = "natural"
	OrderTypeNaturalCaseInsensitive        OrderType = "natural-case-insensitive"
)

// OptionalityOrder represents the ordering of optional vs required members
type OptionalityOrder string

const (
	OptionalityOrderOptionalFirst OptionalityOrder = "optional-first"
	OptionalityOrderRequiredFirst OptionalityOrder = "required-first"
)

// Config represents the configuration for a specific construct type
type Config struct {
	MemberTypes      interface{} // can be string "never" or []string
	Order            OrderType
	OptionalityOrder OptionalityOrder
}

// Options represents the rule options
type Options struct {
	Default         *Config
	Classes         *Config
	ClassExpressions *Config
	Interfaces      *Config
	TypeLiterals    *Config
}

// Member represents a class/interface member with its properties
type Member struct {
	Node             *ast.Node
	Name             string
	Type             MemberType
	IsStatic         bool
	IsAbstract       bool
	IsReadonly       bool
	IsDecorated      bool
	IsOverride       bool
	IsOptional       bool
	Accessibility    string
	Index            int
	Rank             int
	NameType         utils.MemberNameType
}

var defaultMemberTypes = []string{
	"signature",
	"field",
	"constructor",
	"method",
}

func parseOptions(options any) Options {
	opts := Options{}

	if options == nil {
		return opts
	}

	var optMap map[string]interface{}

	// Handle array format: [{ option: value }]
	if arr, ok := options.([]interface{}); ok && len(arr) > 0 {
		if m, ok := arr[0].(map[string]interface{}); ok {
			optMap = m
		}
	}

	// Handle direct object format
	if m, ok := options.(map[string]interface{}); ok {
		optMap = m
	}

	if optMap == nil {
		return opts
	}

	// Parse each config type
	if defaultCfg, ok := optMap["default"]; ok {
		opts.Default = parseConfig(defaultCfg)
	}
	if classesCfg, ok := optMap["classes"]; ok {
		opts.Classes = parseConfig(classesCfg)
	}
	if classExprCfg, ok := optMap["classExpressions"]; ok {
		opts.ClassExpressions = parseConfig(classExprCfg)
	}
	if interfacesCfg, ok := optMap["interfaces"]; ok {
		opts.Interfaces = parseConfig(interfacesCfg)
	}
	if typeLiteralsCfg, ok := optMap["typeLiterals"]; ok {
		opts.TypeLiterals = parseConfig(typeLiteralsCfg)
	}

	return opts
}

func parseConfig(cfg interface{}) *Config {
	config := &Config{
		Order: OrderTypeAsWritten,
	}

	// If it's a string "never", member types are disabled
	if str, ok := cfg.(string); ok {
		if str == "never" {
			config.MemberTypes = "never"
		}
		return config
	}

	// If it's an array, it's a simple member type list
	if arr, ok := cfg.([]interface{}); ok {
		memberTypes := make([]string, 0, len(arr))
		for _, item := range arr {
			if str, ok := item.(string); ok {
				memberTypes = append(memberTypes, str)
			}
		}
		config.MemberTypes = memberTypes
		return config
	}

	// If it's an object, parse its properties
	if m, ok := cfg.(map[string]interface{}); ok {
		if memberTypes, exists := m["memberTypes"]; exists {
			if str, ok := memberTypes.(string); ok && str == "never" {
				config.MemberTypes = "never"
			} else if arr, ok := memberTypes.([]interface{}); ok {
				types := make([]string, 0, len(arr))
				for _, item := range arr {
					if str, ok := item.(string); ok {
						types = append(types, str)
					}
				}
				config.MemberTypes = types
			}
		}

		if order, ok := m["order"].(string); ok {
			config.Order = OrderType(order)
		}

		if optOrder, ok := m["optionalityOrder"].(string); ok {
			config.OptionalityOrder = OptionalityOrder(optOrder)
		}
	}

	return config
}

func getMemberType(node *ast.Node) MemberType {
	switch node.Kind {
	case ast.KindPropertySignature:
		sig := node.AsPropertySignature()
		if sig != nil && ast.IsReadonly(node) {
			return MemberTypeReadonlyField
		}
		return MemberTypeField

	case ast.KindPropertyDeclaration:
		if ast.IsReadonly(node) {
			return MemberTypeReadonlyField
		}
		return MemberTypeField

	case ast.KindMethodSignature:
		return MemberTypeMethod

	case ast.KindMethodDeclaration:
		return MemberTypeMethod

	case ast.KindConstructSignature:
		return MemberTypeConstructSignature

	case ast.KindConstructor:
		return MemberTypeConstructor

	case ast.KindCallSignature:
		return MemberTypeCallSignature

	case ast.KindIndexSignature:
		return MemberTypeIndexSignature

	case ast.KindGetAccessor:
		return MemberTypeGetAccessor

	case ast.KindSetAccessor:
		return MemberTypeSetAccessor

	case ast.KindClassStaticBlockDeclaration:
		return MemberTypeStaticInitialization
	}

	return ""
}

func getMemberName(sourceFile *ast.SourceFile, node *ast.Node) (string, utils.MemberNameType) {
	switch node.Kind {
	case ast.KindPropertySignature:
		if sig := node.AsPropertySignature(); sig != nil && sig.Name() != nil {
			return utils.GetNameFromMember(sourceFile, sig.Name())
		}
	case ast.KindPropertyDeclaration:
		if decl := node.AsPropertyDeclaration(); decl != nil && decl.Name() != nil {
			return utils.GetNameFromMember(sourceFile, decl.Name())
		}
	case ast.KindMethodSignature:
		if sig := node.AsMethodSignatureDeclaration(); sig != nil && sig.Name() != nil {
			return utils.GetNameFromMember(sourceFile, sig.Name())
		}
	case ast.KindMethodDeclaration:
		if decl := node.AsMethodDeclaration(); decl != nil && decl.Name() != nil {
			return utils.GetNameFromMember(sourceFile, decl.Name())
		}
	case ast.KindGetAccessor:
		if acc := node.AsGetAccessorDeclaration(); acc != nil && acc.Name() != nil {
			return utils.GetNameFromMember(sourceFile, acc.Name())
		}
	case ast.KindSetAccessor:
		if acc := node.AsSetAccessorDeclaration(); acc != nil && acc.Name() != nil {
			return utils.GetNameFromMember(sourceFile, acc.Name())
		}
	case ast.KindConstructor:
		return "constructor", utils.MemberNameTypeNormal
	case ast.KindConstructSignature:
		return "new", utils.MemberNameTypeNormal
	case ast.KindCallSignature:
		return "call", utils.MemberNameTypeNormal
	case ast.KindIndexSignature:
		return "index", utils.MemberNameTypeNormal
	}
	return "", utils.MemberNameTypeNormal
}

func getAccessibility(node *ast.Node) string {
	if ast.IsPrivate(node) {
		return ModifierPrivate
	}
	if ast.IsProtected(node) {
		return ModifierProtected
	}
	return ModifierPublic
}

func isOptional(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindPropertySignature:
		if sig := node.AsPropertySignature(); sig != nil {
			return sig.QuestionToken() != nil
		}
	case ast.KindPropertyDeclaration:
		if decl := node.AsPropertyDeclaration(); decl != nil {
			return decl.QuestionToken() != nil
		}
	case ast.KindMethodSignature:
		if sig := node.AsMethodSignatureDeclaration(); sig != nil {
			return sig.QuestionToken() != nil
		}
	case ast.KindMethodDeclaration:
		if decl := node.AsMethodDeclaration(); decl != nil {
			return decl.QuestionToken() != nil
		}
	}
	return false
}

func parseMember(sourceFile *ast.SourceFile, node *ast.Node, index int) *Member {
	memberType := getMemberType(node)
	if memberType == "" {
		return nil
	}

	name, nameType := getMemberName(sourceFile, node)

	return &Member{
		Node:          node,
		Name:          name,
		Type:          memberType,
		IsStatic:      ast.IsStatic(node),
		IsAbstract:    ast.IsAbstract(node),
		IsReadonly:    ast.IsReadonly(node),
		IsDecorated:   hasDecorators(node),
		IsOverride:    hasOverrideModifier(node),
		IsOptional:    isOptional(node),
		Accessibility: getAccessibility(node),
		Index:         index,
		NameType:      nameType,
	}
}

func hasDecorators(node *ast.Node) bool {
	// Check if node has decorators
	switch node.Kind {
	case ast.KindPropertyDeclaration:
		if decl := node.AsPropertyDeclaration(); decl != nil {
			return decl.Modifiers() != nil && hasDecoratorModifiers(decl.Modifiers())
		}
	case ast.KindMethodDeclaration:
		if decl := node.AsMethodDeclaration(); decl != nil {
			return decl.Modifiers() != nil && hasDecoratorModifiers(decl.Modifiers())
		}
	case ast.KindGetAccessor:
		if acc := node.AsGetAccessorDeclaration(); acc != nil {
			return acc.Modifiers() != nil && hasDecoratorModifiers(acc.Modifiers())
		}
	case ast.KindSetAccessor:
		if acc := node.AsSetAccessorDeclaration(); acc != nil {
			return acc.Modifiers() != nil && hasDecoratorModifiers(acc.Modifiers())
		}
	}
	return false
}

func hasDecoratorModifiers(modifiers *ast.NodeArray) bool {
	if modifiers == nil || modifiers.Nodes == nil {
		return false
	}
	for _, mod := range modifiers.Nodes {
		if mod.Kind == ast.KindDecorator {
			return true
		}
	}
	return false
}

func hasOverrideModifier(node *ast.Node) bool {
	// The override keyword is part of modifiers
	switch node.Kind {
	case ast.KindPropertyDeclaration:
		if decl := node.AsPropertyDeclaration(); decl != nil {
			return hasModifierKind(decl.Modifiers(), ast.KindOverrideKeyword)
		}
	case ast.KindMethodDeclaration:
		if decl := node.AsMethodDeclaration(); decl != nil {
			return hasModifierKind(decl.Modifiers(), ast.KindOverrideKeyword)
		}
	}
	return false
}

func hasModifierKind(modifiers *ast.NodeArray, kind ast.SyntaxKind) bool {
	if modifiers == nil || modifiers.Nodes == nil {
		return false
	}
	for _, mod := range modifiers.Nodes {
		if mod.Kind == kind {
			return true
		}
	}
	return false
}

func getMemberRank(member *Member, memberTypes []string) int {
	// If memberTypes is "never", return same rank for all
	if len(memberTypes) == 0 {
		return 0
	}

	// Check for exact match first
	for i, mt := range memberTypes {
		if mt == string(member.Type) {
			return i
		}
	}

	// Check for grouped types
	for i, mt := range memberTypes {
		if matchesMemberType(member, mt) {
			return i
		}
	}

	// Default: put at the end
	return len(memberTypes)
}

func matchesMemberType(member *Member, typeStr string) bool {
	parts := strings.Split(typeStr, "-")

	// Parse type specification
	var requiredModifiers []string
	var memberType string

	for i, part := range parts {
		switch part {
		case "public", "protected", "private":
			requiredModifiers = append(requiredModifiers, part)
		case "static", "instance":
			requiredModifiers = append(requiredModifiers, part)
		case "readonly":
			requiredModifiers = append(requiredModifiers, part)
		case "decorated":
			requiredModifiers = append(requiredModifiers, part)
		case "abstract":
			requiredModifiers = append(requiredModifiers, part)
		case "override":
			requiredModifiers = append(requiredModifiers, part)
		default:
			// This is the member type
			memberType = strings.Join(parts[i:], "-")
			break
		}
	}

	// Check if member type matches
	if memberType != "" && memberType != string(member.Type) {
		// Check for group types
		switch memberType {
		case "signature":
			if member.Type != MemberTypeCallSignature &&
				member.Type != MemberTypeConstructSignature &&
				member.Type != MemberTypeIndexSignature {
				return false
			}
		case "field":
			if member.Type != MemberTypeField && member.Type != MemberTypeReadonlyField {
				return false
			}
		case "method":
			if member.Type != MemberTypeMethod {
				return false
			}
		case "constructor":
			if member.Type != MemberTypeConstructor && member.Type != MemberTypeConstructSignature {
				return false
			}
		case "accessor":
			if member.Type != MemberTypeGetAccessor && member.Type != MemberTypeSetAccessor {
				return false
			}
		default:
			return false
		}
	}

	// Check modifiers
	for _, mod := range requiredModifiers {
		switch mod {
		case "public":
			if member.Accessibility != ModifierPublic {
				return false
			}
		case "protected":
			if member.Accessibility != ModifierProtected {
				return false
			}
		case "private":
			if member.Accessibility != ModifierPrivate {
				return false
			}
		case "static":
			if !member.IsStatic {
				return false
			}
		case "instance":
			if member.IsStatic {
				return false
			}
		case "readonly":
			if !member.IsReadonly {
				return false
			}
		case "decorated":
			if !member.IsDecorated {
				return false
			}
		case "abstract":
			if !member.IsAbstract {
				return false
			}
		case "override":
			if !member.IsOverride {
				return false
			}
		}
	}

	return true
}

func compareMembers(a, b *Member, orderType OrderType, optionalityOrder OptionalityOrder) int {
	// First compare by rank
	if a.Rank != b.Rank {
		return a.Rank - b.Rank
	}

	// Then by optionality if specified
	if optionalityOrder != "" {
		if a.IsOptional != b.IsOptional {
			if optionalityOrder == OptionalityOrderOptionalFirst {
				if a.IsOptional {
					return -1
				}
				return 1
			} else if optionalityOrder == OptionalityOrderRequiredFirst {
				if a.IsOptional {
					return 1
				}
				return -1
			}
		}
	}

	// Then by name if order type is specified
	if orderType != OrderTypeAsWritten {
		cmp := compareNames(a.Name, b.Name, orderType)
		if cmp != 0 {
			return cmp
		}
	}

	// Finally by original index
	return a.Index - b.Index
}

func compareNames(a, b string, orderType OrderType) int {
	switch orderType {
	case OrderTypeAlphabetically:
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0

	case OrderTypeAlphabeticallyCaseInsensitive:
		aLower := strings.ToLower(a)
		bLower := strings.ToLower(b)
		if aLower < bLower {
			return -1
		} else if aLower > bLower {
			return 1
		}
		// If equal ignoring case, compare case-sensitive as tiebreaker
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0

	case OrderTypeNatural:
		return naturalCompare(a, b, false)

	case OrderTypeNaturalCaseInsensitive:
		return naturalCompare(a, b, true)
	}

	return 0
}

// naturalCompare implements natural sorting (e.g., "item2" < "item10")
func naturalCompare(a, b string, caseInsensitive bool) int {
	if caseInsensitive {
		a = strings.ToLower(a)
		b = strings.ToLower(b)
	}

	// Split into parts of digits and non-digits
	aParts := splitNatural(a)
	bParts := splitNatural(b)

	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		aPart := aParts[i]
		bPart := bParts[i]

		// If both are numeric, compare as numbers
		if isNumeric(aPart) && isNumeric(bPart) {
			aNum, _ := strconv.Atoi(aPart)
			bNum, _ := strconv.Atoi(bPart)
			if aNum != bNum {
				return aNum - bNum
			}
		} else {
			// Compare as strings
			if aPart < bPart {
				return -1
			} else if aPart > bPart {
				return 1
			}
		}
	}

	// If all parts equal, shorter one comes first
	return len(aParts) - len(bParts)
}

func splitNatural(s string) []string {
	re := regexp.MustCompile(`\d+|\D+`)
	return re.FindAllString(s, -1)
}

func isNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

func checkMembers(ctx rule.RuleContext, members []*ast.Node, config *Config) {
	if config == nil {
		return
	}

	// Parse member types configuration
	var memberTypes []string
	if config.MemberTypes == "never" {
		// When "never", we only enforce name ordering
		memberTypes = []string{}
	} else if types, ok := config.MemberTypes.([]string); ok {
		memberTypes = types
	} else {
		// Use default member types
		memberTypes = defaultMemberTypes
	}

	// Parse members
	parsedMembers := make([]*Member, 0, len(members))
	for i, node := range members {
		if member := parseMember(ctx.SourceFile, node, i); member != nil {
			member.Rank = getMemberRank(member, memberTypes)
			parsedMembers = append(parsedMembers, member)
		}
	}

	// Check ordering
	for i := 1; i < len(parsedMembers); i++ {
		prev := parsedMembers[i-1]
		curr := parsedMembers[i]

		cmp := compareMembers(prev, curr, config.Order, config.OptionalityOrder)
		if cmp > 0 {
			// Members are out of order
			var messageId string
			var data map[string]string

			if config.OptionalityOrder != "" && prev.IsOptional != curr.IsOptional {
				messageId = "incorrectRequiredMembersOrder"
				optOrReq := "required"
				if curr.IsOptional {
					optOrReq = "optional"
				}
				data = map[string]string{
					"member":            curr.Name,
					"beforeMember":      prev.Name,
					"optionalOrRequired": optOrReq,
				}
			} else {
				messageId = "incorrectOrder"
				data = map[string]string{
					"member":       curr.Name,
					"beforeMember": prev.Name,
				}
			}

			ctx.ReportNode(curr.Node, buildMessage(messageId, data))
		}
	}
}

func buildMessage(messageId string, data map[string]string) rule.RuleMessage {
	var description string
	switch messageId {
	case "incorrectOrder":
		description = fmt.Sprintf("Member '%s' should be declared before member '%s'.",
			data["member"], data["beforeMember"])
	case "incorrectRequiredMembersOrder":
		description = fmt.Sprintf("Member '%s' (%s) should be declared before member '%s'.",
			data["member"], data["optionalOrRequired"], data["beforeMember"])
	case "incorrectGroupOrder":
		description = fmt.Sprintf("Member %s should be declared before all %s definitions.",
			data["name"], data["rank"])
	}

	return rule.RuleMessage{
		Id:          messageId,
		Description: description,
	}
}

func getConfig(opts Options, nodeKind ast.SyntaxKind) *Config {
	var config *Config

	switch nodeKind {
	case ast.KindClassDeclaration:
		if opts.Classes != nil {
			config = opts.Classes
		}
	case ast.KindClassExpression:
		if opts.ClassExpressions != nil {
			config = opts.ClassExpressions
		}
	case ast.KindInterfaceDeclaration:
		if opts.Interfaces != nil {
			config = opts.Interfaces
		}
	case ast.KindTypeLiteral:
		if opts.TypeLiterals != nil {
			config = opts.TypeLiterals
		}
	}

	// Fall back to default if no specific config
	if config == nil && opts.Default != nil {
		config = opts.Default
	}

	return config
}

func getMembers(node *ast.Node) []*ast.Node {
	switch node.Kind {
	case ast.KindClassDeclaration:
		if classDecl := node.AsClassDeclaration(); classDecl != nil && classDecl.Members != nil {
			return classDecl.Members.Nodes
		}
	case ast.KindClassExpression:
		if classExpr := node.AsClassExpression(); classExpr != nil && classExpr.Members != nil {
			return classExpr.Members.Nodes
		}
	case ast.KindInterfaceDeclaration:
		if interfaceDecl := node.AsInterfaceDeclaration(); interfaceDecl != nil && interfaceDecl.Members != nil {
			return interfaceDecl.Members.Nodes
		}
	case ast.KindTypeLiteral:
		if typeLiteral := node.AsTypeLiteralNode(); typeLiteral != nil && typeLiteral.Members != nil {
			return typeLiteral.Members.Nodes
		}
	}
	return nil
}

var MemberOrderingRule = rule.CreateRule(rule.Rule{
	Name: "member-ordering",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		checkNode := func(node *ast.Node) {
			members := getMembers(node)
			if members == nil || len(members) == 0 {
				return
			}

			config := getConfig(opts, node.Kind)
			if config == nil {
				return
			}

			checkMembers(ctx, members, config)
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: checkNode,
			ast.KindClassExpression:  checkNode,
			ast.KindInterfaceDeclaration: checkNode,
			ast.KindTypeLiteral: checkNode,
		}
	},
})

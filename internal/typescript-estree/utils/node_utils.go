// Package utils provides utility functions for the typescript-estree module.
package utils

import (
	"github.com/web-infra-dev/rslint/internal/typescript-estree/types"
)

// Node Type Guards

// IsExpression checks if a node is an expression.
func IsExpression(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(types.Expression)
	return ok
}

// IsStatement checks if a node is a statement.
func IsStatement(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(types.Statement)
	return ok
}

// IsDeclaration checks if a node is a declaration.
func IsDeclaration(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(types.Declaration)
	return ok
}

// IsPattern checks if a node is a pattern.
func IsPattern(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(types.Pattern)
	return ok
}

// IsModuleDeclaration checks if a node is a module declaration.
func IsModuleDeclaration(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(types.ModuleDeclaration)
	return ok
}

// IsLiteral checks if a node is a literal.
func IsLiteral(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(types.Literal)
	return ok
}

// Specific Node Type Checks

// IsIdentifier checks if a node is an Identifier.
func IsIdentifier(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.Identifier)
	return ok
}

// IsProgram checks if a node is a Program.
func IsProgram(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.Program)
	return ok
}

// IsFunctionDeclaration checks if a node is a FunctionDeclaration.
func IsFunctionDeclaration(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.FunctionDeclaration)
	return ok
}

// IsFunctionExpression checks if a node is a FunctionExpression.
func IsFunctionExpression(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.FunctionExpression)
	return ok
}

// IsArrowFunctionExpression checks if a node is an ArrowFunctionExpression.
func IsArrowFunctionExpression(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.ArrowFunctionExpression)
	return ok
}

// IsFunction checks if a node is any function type.
func IsFunction(node types.Node) bool {
	return IsFunctionDeclaration(node) || IsFunctionExpression(node) || IsArrowFunctionExpression(node)
}

// IsBlockStatement checks if a node is a BlockStatement.
func IsBlockStatement(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.BlockStatement)
	return ok
}

// IsVariableDeclaration checks if a node is a VariableDeclaration.
func IsVariableDeclaration(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.VariableDeclaration)
	return ok
}

// IsClassDeclaration checks if a node is a ClassDeclaration.
func IsClassDeclaration(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.ClassDeclaration)
	return ok
}

// IsClassExpression checks if a node is a ClassExpression.
func IsClassExpression(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.ClassExpression)
	return ok
}

// IsClass checks if a node is any class type.
func IsClass(node types.Node) bool {
	return IsClassDeclaration(node) || IsClassExpression(node)
}

// IsMemberExpression checks if a node is a MemberExpression.
func IsMemberExpression(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.MemberExpression)
	return ok
}

// IsCallExpression checks if a node is a CallExpression.
func IsCallExpression(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.CallExpression)
	return ok
}

// IsBinaryExpression checks if a node is a BinaryExpression.
func IsBinaryExpression(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.BinaryExpression)
	return ok
}

// IsUnaryExpression checks if a node is a UnaryExpression.
func IsUnaryExpression(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.UnaryExpression)
	return ok
}

// IsLogicalExpression checks if a node is a LogicalExpression.
func IsLogicalExpression(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.LogicalExpression)
	return ok
}

// IsAssignmentExpression checks if a node is an AssignmentExpression.
func IsAssignmentExpression(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.AssignmentExpression)
	return ok
}

// IsConditionalExpression checks if a node is a ConditionalExpression.
func IsConditionalExpression(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.ConditionalExpression)
	return ok
}

// IsObjectExpression checks if a node is an ObjectExpression.
func IsObjectExpression(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.ObjectExpression)
	return ok
}

// IsArrayExpression checks if a node is an ArrayExpression.
func IsArrayExpression(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.ArrayExpression)
	return ok
}

// IsIfStatement checks if a node is an IfStatement.
func IsIfStatement(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.IfStatement)
	return ok
}

// IsForStatement checks if a node is a ForStatement.
func IsForStatement(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.ForStatement)
	return ok
}

// IsWhileStatement checks if a node is a WhileStatement.
func IsWhileStatement(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.WhileStatement)
	return ok
}

// IsReturnStatement checks if a node is a ReturnStatement.
func IsReturnStatement(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.ReturnStatement)
	return ok
}

// IsImportDeclaration checks if a node is an ImportDeclaration.
func IsImportDeclaration(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.ImportDeclaration)
	return ok
}

// IsExportNamedDeclaration checks if a node is an ExportNamedDeclaration.
func IsExportNamedDeclaration(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.ExportNamedDeclaration)
	return ok
}

// IsExportDefaultDeclaration checks if a node is an ExportDefaultDeclaration.
func IsExportDefaultDeclaration(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.ExportDefaultDeclaration)
	return ok
}

// TypeScript-specific type guards

// IsTSTypeAnnotation checks if a node is a TSTypeAnnotation.
func IsTSTypeAnnotation(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.TSTypeAnnotation)
	return ok
}

// IsTSInterfaceDeclaration checks if a node is a TSInterfaceDeclaration.
func IsTSInterfaceDeclaration(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.TSInterfaceDeclaration)
	return ok
}

// IsTSTypeAliasDeclaration checks if a node is a TSTypeAliasDeclaration.
func IsTSTypeAliasDeclaration(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.TSTypeAliasDeclaration)
	return ok
}

// IsTSEnumDeclaration checks if a node is a TSEnumDeclaration.
func IsTSEnumDeclaration(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.TSEnumDeclaration)
	return ok
}

// IsTSModuleDeclaration checks if a node is a TSModuleDeclaration.
func IsTSModuleDeclaration(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.TSModuleDeclaration)
	return ok
}

// IsTSAsExpression checks if a node is a TSAsExpression.
func IsTSAsExpression(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.TSAsExpression)
	return ok
}

// IsTSTypeAssertion checks if a node is a TSTypeAssertion.
func IsTSTypeAssertion(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.TSTypeAssertion)
	return ok
}

// IsTSNonNullExpression checks if a node is a TSNonNullExpression.
func IsTSNonNullExpression(node types.Node) bool {
	if node == nil {
		return false
	}
	_, ok := node.(*types.TSNonNullExpression)
	return ok
}

// Helper Functions

// IsValidNode checks if a node is valid (not nil and has a valid type).
func IsValidNode(node types.Node) bool {
	return node != nil && node.Type() != ""
}

// IsLogicalOperator checks if the given operator is a logical operator.
func IsLogicalOperator(operator string) bool {
	return operator == "&&" || operator == "||" || operator == "??"
}

// IsBinaryOperator checks if the given operator is a binary operator.
func IsBinaryOperator(operator string) bool {
	switch operator {
	case "==", "!=", "===", "!==",
		"<", "<=", ">", ">=",
		"<<", ">>", ">>>",
		"+", "-", "*", "/", "%", "**",
		"|", "^", "&",
		"in", "instanceof":
		return true
	default:
		return false
	}
}

// IsAssignmentOperator checks if the given operator is an assignment operator.
func IsAssignmentOperator(operator string) bool {
	switch operator {
	case "=", "+=", "-=", "*=", "/=", "%=", "**=",
		"<<=", ">>=", ">>>=",
		"|=", "^=", "&=",
		"||=", "&&=", "??=":
		return true
	default:
		return false
	}
}

// IsUpdateOperator checks if the given operator is an update operator.
func IsUpdateOperator(operator string) bool {
	return operator == "++" || operator == "--"
}

// IsUnaryOperator checks if the given operator is a unary operator.
func IsUnaryOperator(operator string) bool {
	switch operator {
	case "-", "+", "!", "~", "typeof", "void", "delete":
		return true
	default:
		return false
	}
}

// GetDeclarationKind returns the kind of a variable declaration ("var", "let", or "const").
func GetDeclarationKind(node *types.VariableDeclaration) string {
	if node == nil {
		return ""
	}
	return node.Kind
}

// IsConstDeclaration checks if a variable declaration is a const declaration.
func IsConstDeclaration(node types.Node) bool {
	if decl, ok := node.(*types.VariableDeclaration); ok {
		return decl.Kind == "const"
	}
	return false
}

// IsLetDeclaration checks if a variable declaration is a let declaration.
func IsLetDeclaration(node types.Node) bool {
	if decl, ok := node.(*types.VariableDeclaration); ok {
		return decl.Kind == "let"
	}
	return false
}

// IsVarDeclaration checks if a variable declaration is a var declaration.
func IsVarDeclaration(node types.Node) bool {
	if decl, ok := node.(*types.VariableDeclaration); ok {
		return decl.Kind == "var"
	}
	return false
}

// IsAsyncFunction checks if a function node is async.
func IsAsyncFunction(node types.Node) bool {
	switch n := node.(type) {
	case *types.FunctionDeclaration:
		return n.Async
	case *types.FunctionExpression:
		return n.Async
	case *types.ArrowFunctionExpression:
		return n.Async
	default:
		return false
	}
}

// IsGeneratorFunction checks if a function node is a generator.
func IsGeneratorFunction(node types.Node) bool {
	switch n := node.(type) {
	case *types.FunctionDeclaration:
		return n.Generator
	case *types.FunctionExpression:
		return n.Generator
	default:
		return false
	}
}

// GetIdentifierName safely extracts the name from an Identifier node.
func GetIdentifierName(node types.Node) string {
	if id, ok := node.(*types.Identifier); ok {
		return id.Name
	}
	return ""
}

// IsComputedMember checks if a member expression is computed.
func IsComputedMember(node types.Node) bool {
	if member, ok := node.(*types.MemberExpression); ok {
		return member.Computed
	}
	return false
}

// IsOptionalMember checks if a member expression is optional.
func IsOptionalMember(node types.Node) bool {
	if member, ok := node.(*types.MemberExpression); ok {
		return member.Optional
	}
	return false
}

// IsOptionalCall checks if a call expression is optional.
func IsOptionalCall(node types.Node) bool {
	if call, ok := node.(*types.CallExpression); ok {
		return call.Optional
	}
	return false
}

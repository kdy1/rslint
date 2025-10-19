package utils

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/typescript-estree/types"
)

func TestIsExpression(t *testing.T) {
	tests := []struct {
		name     string
		node     types.Node
		expected bool
	}{
		{"Identifier", &types.Identifier{BaseNode: types.BaseNode{NodeType: "Identifier"}}, true},
		{"SimpleLiteral", &types.SimpleLiteral{BaseNode: types.BaseNode{NodeType: "Literal"}}, true},
		{"BinaryExpression", &types.BinaryExpression{BaseNode: types.BaseNode{NodeType: "BinaryExpression"}}, true},
		{"BlockStatement", &types.BlockStatement{BaseNode: types.BaseNode{NodeType: "BlockStatement"}}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsExpression(tt.node)
			if result != tt.expected {
				t.Errorf("IsExpression(%s) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestIsStatement(t *testing.T) {
	tests := []struct {
		name     string
		node     types.Node
		expected bool
	}{
		{"BlockStatement", &types.BlockStatement{BaseNode: types.BaseNode{NodeType: "BlockStatement"}}, true},
		{"ExpressionStatement", &types.ExpressionStatement{BaseNode: types.BaseNode{NodeType: "ExpressionStatement"}}, true},
		{"IfStatement", &types.IfStatement{BaseNode: types.BaseNode{NodeType: "IfStatement"}}, true},
		{"Identifier", &types.Identifier{BaseNode: types.BaseNode{NodeType: "Identifier"}}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsStatement(tt.node)
			if result != tt.expected {
				t.Errorf("IsStatement(%s) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestIsDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		node     types.Node
		expected bool
	}{
		{"FunctionDeclaration", &types.FunctionDeclaration{BaseNode: types.BaseNode{NodeType: "FunctionDeclaration"}}, true},
		{"VariableDeclaration", &types.VariableDeclaration{BaseNode: types.BaseNode{NodeType: "VariableDeclaration"}}, true},
		{"ClassDeclaration", &types.ClassDeclaration{BaseNode: types.BaseNode{NodeType: "ClassDeclaration"}}, true},
		{"Identifier", &types.Identifier{BaseNode: types.BaseNode{NodeType: "Identifier"}}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDeclaration(tt.node)
			if result != tt.expected {
				t.Errorf("IsDeclaration(%s) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestIsPattern(t *testing.T) {
	tests := []struct {
		name     string
		node     types.Node
		expected bool
	}{
		{"Identifier", &types.Identifier{BaseNode: types.BaseNode{NodeType: "Identifier"}}, true},
		{"ObjectPattern", &types.ObjectPattern{BaseNode: types.BaseNode{NodeType: "ObjectPattern"}}, true},
		{"ArrayPattern", &types.ArrayPattern{BaseNode: types.BaseNode{NodeType: "ArrayPattern"}}, true},
		{"BlockStatement", &types.BlockStatement{BaseNode: types.BaseNode{NodeType: "BlockStatement"}}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPattern(tt.node)
			if result != tt.expected {
				t.Errorf("IsPattern(%s) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestIsIdentifier(t *testing.T) {
	id := &types.Identifier{BaseNode: types.BaseNode{NodeType: "Identifier"}}
	if !IsIdentifier(id) {
		t.Error("Expected IsIdentifier to return true for Identifier")
	}

	lit := &types.SimpleLiteral{BaseNode: types.BaseNode{NodeType: "Literal"}}
	if IsIdentifier(lit) {
		t.Error("Expected IsIdentifier to return false for Literal")
	}

	if IsIdentifier(nil) {
		t.Error("Expected IsIdentifier to return false for nil")
	}
}

func TestIsFunction(t *testing.T) {
	tests := []struct {
		name     string
		node     types.Node
		expected bool
	}{
		{"FunctionDeclaration", &types.FunctionDeclaration{BaseNode: types.BaseNode{NodeType: "FunctionDeclaration"}}, true},
		{"FunctionExpression", &types.FunctionExpression{BaseNode: types.BaseNode{NodeType: "FunctionExpression"}}, true},
		{"ArrowFunctionExpression", &types.ArrowFunctionExpression{BaseNode: types.BaseNode{NodeType: "ArrowFunctionExpression"}}, true},
		{"CallExpression", &types.CallExpression{BaseNode: types.BaseNode{NodeType: "CallExpression"}}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFunction(tt.node)
			if result != tt.expected {
				t.Errorf("IsFunction(%s) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestIsClass(t *testing.T) {
	tests := []struct {
		name     string
		node     types.Node
		expected bool
	}{
		{"ClassDeclaration", &types.ClassDeclaration{BaseNode: types.BaseNode{NodeType: "ClassDeclaration"}}, true},
		{"ClassExpression", &types.ClassExpression{BaseNode: types.BaseNode{NodeType: "ClassExpression"}}, true},
		{"FunctionDeclaration", &types.FunctionDeclaration{BaseNode: types.BaseNode{NodeType: "FunctionDeclaration"}}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsClass(tt.node)
			if result != tt.expected {
				t.Errorf("IsClass(%s) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestIsValidNode(t *testing.T) {
	validNode := &types.Identifier{BaseNode: types.BaseNode{NodeType: "Identifier"}}
	if !IsValidNode(validNode) {
		t.Error("Expected IsValidNode to return true for valid node")
	}

	invalidNode := &types.Identifier{BaseNode: types.BaseNode{NodeType: ""}}
	if IsValidNode(invalidNode) {
		t.Error("Expected IsValidNode to return false for node with empty type")
	}

	if IsValidNode(nil) {
		t.Error("Expected IsValidNode to return false for nil")
	}
}

func TestIsLogicalOperator(t *testing.T) {
	tests := []struct {
		operator string
		expected bool
	}{
		{"&&", true},
		{"||", true},
		{"??", true},
		{"+", false},
		{"==", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.operator, func(t *testing.T) {
			result := IsLogicalOperator(tt.operator)
			if result != tt.expected {
				t.Errorf("IsLogicalOperator(%s) = %v, want %v", tt.operator, result, tt.expected)
			}
		})
	}
}

func TestIsBinaryOperator(t *testing.T) {
	operators := []string{"==", "!=", "===", "!==", "<", "<=", ">", ">=", "<<", ">>", ">>>", "+", "-", "*", "/", "%", "**", "|", "^", "&", "in", "instanceof"}
	for _, op := range operators {
		if !IsBinaryOperator(op) {
			t.Errorf("Expected %s to be a binary operator", op)
		}
	}

	notOperators := []string{"&&", "||", "=", "++", "typeof"}
	for _, op := range notOperators {
		if IsBinaryOperator(op) {
			t.Errorf("Expected %s not to be a binary operator", op)
		}
	}
}

func TestIsAssignmentOperator(t *testing.T) {
	operators := []string{"=", "+=", "-=", "*=", "/=", "%=", "**=", "<<=", ">>=", ">>>=", "|=", "^=", "&=", "||=", "&&=", "??="}
	for _, op := range operators {
		if !IsAssignmentOperator(op) {
			t.Errorf("Expected %s to be an assignment operator", op)
		}
	}

	notOperators := []string{"+", "-", "==", "&&"}
	for _, op := range notOperators {
		if IsAssignmentOperator(op) {
			t.Errorf("Expected %s not to be an assignment operator", op)
		}
	}
}

func TestIsUpdateOperator(t *testing.T) {
	if !IsUpdateOperator("++") {
		t.Error("Expected ++ to be an update operator")
	}

	if !IsUpdateOperator("--") {
		t.Error("Expected -- to be an update operator")
	}

	if IsUpdateOperator("+") {
		t.Error("Expected + not to be an update operator")
	}
}

func TestIsUnaryOperator(t *testing.T) {
	operators := []string{"-", "+", "!", "~", "typeof", "void", "delete"}
	for _, op := range operators {
		if !IsUnaryOperator(op) {
			t.Errorf("Expected %s to be a unary operator", op)
		}
	}

	notOperators := []string{"++", "--", "==", "&&"}
	for _, op := range notOperators {
		if IsUnaryOperator(op) {
			t.Errorf("Expected %s not to be a unary operator", op)
		}
	}
}

func TestGetDeclarationKind(t *testing.T) {
	tests := []struct {
		kind     string
		expected string
	}{
		{"const", "const"},
		{"let", "let"},
		{"var", "var"},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			decl := &types.VariableDeclaration{
				BaseNode: types.BaseNode{NodeType: "VariableDeclaration"},
				Kind:     tt.kind,
			}
			result := GetDeclarationKind(decl)
			if result != tt.expected {
				t.Errorf("GetDeclarationKind() = %s, want %s", result, tt.expected)
			}
		})
	}

	if GetDeclarationKind(nil) != "" {
		t.Error("Expected empty string for nil declaration")
	}
}

func TestIsConstDeclaration(t *testing.T) {
	constDecl := &types.VariableDeclaration{
		BaseNode: types.BaseNode{NodeType: "VariableDeclaration"},
		Kind:     "const",
	}
	if !IsConstDeclaration(constDecl) {
		t.Error("Expected IsConstDeclaration to return true for const declaration")
	}

	letDecl := &types.VariableDeclaration{
		BaseNode: types.BaseNode{NodeType: "VariableDeclaration"},
		Kind:     "let",
	}
	if IsConstDeclaration(letDecl) {
		t.Error("Expected IsConstDeclaration to return false for let declaration")
	}
}

func TestIsAsyncFunction(t *testing.T) {
	tests := []struct {
		name     string
		node     types.Node
		expected bool
	}{
		{
			"AsyncFunctionDeclaration",
			&types.FunctionDeclaration{
				BaseNode: types.BaseNode{NodeType: "FunctionDeclaration"},
				Async:    true,
			},
			true,
		},
		{
			"SyncFunctionDeclaration",
			&types.FunctionDeclaration{
				BaseNode: types.BaseNode{NodeType: "FunctionDeclaration"},
				Async:    false,
			},
			false,
		},
		{
			"AsyncFunctionExpression",
			&types.FunctionExpression{
				BaseNode: types.BaseNode{NodeType: "FunctionExpression"},
				Async:    true,
			},
			true,
		},
		{
			"AsyncArrowFunction",
			&types.ArrowFunctionExpression{
				BaseNode: types.BaseNode{NodeType: "ArrowFunctionExpression"},
				Async:    true,
			},
			true,
		},
		{
			"NonFunction",
			&types.Identifier{BaseNode: types.BaseNode{NodeType: "Identifier"}},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAsyncFunction(tt.node)
			if result != tt.expected {
				t.Errorf("IsAsyncFunction() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsGeneratorFunction(t *testing.T) {
	tests := []struct {
		name     string
		node     types.Node
		expected bool
	}{
		{
			"GeneratorFunctionDeclaration",
			&types.FunctionDeclaration{
				BaseNode:  types.BaseNode{NodeType: "FunctionDeclaration"},
				Generator: true,
			},
			true,
		},
		{
			"NormalFunctionDeclaration",
			&types.FunctionDeclaration{
				BaseNode:  types.BaseNode{NodeType: "FunctionDeclaration"},
				Generator: false,
			},
			false,
		},
		{
			"GeneratorFunctionExpression",
			&types.FunctionExpression{
				BaseNode:  types.BaseNode{NodeType: "FunctionExpression"},
				Generator: true,
			},
			true,
		},
		{
			"ArrowFunction",
			&types.ArrowFunctionExpression{
				BaseNode: types.BaseNode{NodeType: "ArrowFunctionExpression"},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGeneratorFunction(tt.node)
			if result != tt.expected {
				t.Errorf("IsGeneratorFunction() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetIdentifierName(t *testing.T) {
	id := &types.Identifier{
		BaseNode: types.BaseNode{NodeType: "Identifier"},
		Name:     "myVar",
	}

	name := GetIdentifierName(id)
	if name != "myVar" {
		t.Errorf("Expected name 'myVar', got '%s'", name)
	}

	lit := &types.SimpleLiteral{BaseNode: types.BaseNode{NodeType: "Literal"}}
	name = GetIdentifierName(lit)
	if name != "" {
		t.Errorf("Expected empty string for non-identifier, got '%s'", name)
	}
}

func TestIsComputedMember(t *testing.T) {
	computed := &types.MemberExpression{
		BaseNode: types.BaseNode{NodeType: "MemberExpression"},
		Computed: true,
	}
	if !IsComputedMember(computed) {
		t.Error("Expected IsComputedMember to return true for computed member")
	}

	notComputed := &types.MemberExpression{
		BaseNode: types.BaseNode{NodeType: "MemberExpression"},
		Computed: false,
	}
	if IsComputedMember(notComputed) {
		t.Error("Expected IsComputedMember to return false for non-computed member")
	}
}

func TestIsOptionalMember(t *testing.T) {
	optional := &types.MemberExpression{
		BaseNode: types.BaseNode{NodeType: "MemberExpression"},
		Optional: true,
	}
	if !IsOptionalMember(optional) {
		t.Error("Expected IsOptionalMember to return true for optional member")
	}

	notOptional := &types.MemberExpression{
		BaseNode: types.BaseNode{NodeType: "MemberExpression"},
		Optional: false,
	}
	if IsOptionalMember(notOptional) {
		t.Error("Expected IsOptionalMember to return false for non-optional member")
	}
}

func TestIsOptionalCall(t *testing.T) {
	optional := &types.CallExpression{
		BaseNode: types.BaseNode{NodeType: "CallExpression"},
		Optional: true,
	}
	if !IsOptionalCall(optional) {
		t.Error("Expected IsOptionalCall to return true for optional call")
	}

	notOptional := &types.CallExpression{
		BaseNode: types.BaseNode{NodeType: "CallExpression"},
		Optional: false,
	}
	if IsOptionalCall(notOptional) {
		t.Error("Expected IsOptionalCall to return false for non-optional call")
	}
}

func TestTypeScriptTypeGuards(t *testing.T) {
	tests := []struct {
		name     string
		node     types.Node
		checker  func(types.Node) bool
		expected bool
	}{
		{
			"TSInterfaceDeclaration",
			&types.TSInterfaceDeclaration{BaseNode: types.BaseNode{NodeType: "TSInterfaceDeclaration"}},
			IsTSInterfaceDeclaration,
			true,
		},
		{
			"TSTypeAliasDeclaration",
			&types.TSTypeAliasDeclaration{BaseNode: types.BaseNode{NodeType: "TSTypeAliasDeclaration"}},
			IsTSTypeAliasDeclaration,
			true,
		},
		{
			"TSEnumDeclaration",
			&types.TSEnumDeclaration{BaseNode: types.BaseNode{NodeType: "TSEnumDeclaration"}},
			IsTSEnumDeclaration,
			true,
		},
		{
			"TSAsExpression",
			&types.TSAsExpression{BaseNode: types.BaseNode{NodeType: "TSAsExpression"}},
			IsTSAsExpression,
			true,
		},
		{
			"TSNonNullExpression",
			&types.TSNonNullExpression{BaseNode: types.BaseNode{NodeType: "TSNonNullExpression"}},
			IsTSNonNullExpression,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checker(tt.node)
			if result != tt.expected {
				t.Errorf("Type guard for %s = %v, want %v", tt.name, result, tt.expected)
			}

			// Also test with nil
			if tt.checker(nil) {
				t.Errorf("Type guard for %s should return false for nil", tt.name)
			}
		})
	}
}

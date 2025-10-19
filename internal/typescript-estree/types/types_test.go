package types_test

import (
	"encoding/json"
	"testing"

	"github.com/web-infra-dev/rslint/internal/typescript-estree/types"
)

func TestBaseNode(t *testing.T) {
	t.Parallel()

	node := &types.BaseNode{
		NodeType: "Identifier",
		Location: &types.SourceLocation{
			Start: types.Position{Line: 1, Column: 0},
			End:   types.Position{Line: 1, Column: 3},
		},
		Span: types.Range{0, 3},
	}

	if node.Type() != "Identifier" {
		t.Errorf("Expected type 'Identifier', got '%s'", node.Type())
	}

	if node.Loc() == nil {
		t.Error("Expected non-nil location")
	}

	if node.GetRange()[0] != 0 || node.GetRange()[1] != 3 {
		t.Errorf("Expected range [0, 3], got %v", node.GetRange())
	}
}

func TestSourceLocation(t *testing.T) {
	t.Parallel()

	loc := &types.SourceLocation{
		Start: types.Position{Line: 1, Column: 0},
		End:   types.Position{Line: 2, Column: 5},
	}

	if loc.Start.Line != 1 {
		t.Errorf("Expected start line 1, got %d", loc.Start.Line)
	}

	if loc.End.Line != 2 {
		t.Errorf("Expected end line 2, got %d", loc.End.Line)
	}
}

func TestIdentifier(t *testing.T) {
	t.Parallel()

	id := &types.Identifier{
		BaseNode: types.BaseNode{
			NodeType: "Identifier",
			Location: &types.SourceLocation{
				Start: types.Position{Line: 1, Column: 0},
				End:   types.Position{Line: 1, Column: 3},
			},
			Span: types.Range{0, 3},
		},
		Name: "foo",
	}

	if id.Name != "foo" {
		t.Errorf("Expected name 'foo', got '%s'", id.Name)
	}

	if id.Type() != "Identifier" {
		t.Errorf("Expected type 'Identifier', got '%s'", id.Type())
	}
}

func TestProgram(t *testing.T) {
	t.Parallel()

	program := &types.Program{
		BaseNode: types.BaseNode{
			NodeType: "Program",
		},
		SourceType: "module",
		Body:       []types.Statement{},
	}

	if program.Type() != "Program" {
		t.Errorf("Expected type 'Program', got '%s'", program.Type())
	}

	if program.SourceType != "module" {
		t.Errorf("Expected sourceType 'module', got '%s'", program.SourceType)
	}
}

func TestSimpleLiteralJSON(t *testing.T) {
	t.Parallel()

	literal := &types.SimpleLiteral{
		BaseNode: types.BaseNode{
			NodeType: "Literal",
			Span:     types.Range{0, 5},
		},
		Value: "hello",
		Raw:   "\"hello\"",
	}

	data, err := json.Marshal(literal)
	if err != nil {
		t.Fatalf("Failed to marshal literal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal literal: %v", err)
	}

	if result["type"] != "Literal" {
		t.Errorf("Expected type 'Literal', got '%v'", result["type"])
	}

	if result["value"] != "hello" {
		t.Errorf("Expected value 'hello', got '%v'", result["value"])
	}

	if result["raw"] != "\"hello\"" {
		t.Errorf("Expected raw '\"hello\"', got '%v'", result["raw"])
	}
}

func TestBinaryExpression(t *testing.T) {
	t.Parallel()

	left := &types.Identifier{
		BaseNode: types.BaseNode{NodeType: "Identifier"},
		Name:     "x",
	}

	right := &types.SimpleLiteral{
		BaseNode: types.BaseNode{NodeType: "Literal"},
		Value:    float64(42),
		Raw:      "42",
	}

	binExpr := &types.BinaryExpression{
		BaseNode: types.BaseNode{NodeType: "BinaryExpression"},
		Operator: "+",
		Left:     left,
		Right:    right,
	}

	if binExpr.Operator != "+" {
		t.Errorf("Expected operator '+', got '%s'", binExpr.Operator)
	}
}

func TestFunctionDeclaration(t *testing.T) {
	t.Parallel()

	funcDecl := &types.FunctionDeclaration{
		BaseNode: types.BaseNode{NodeType: "FunctionDeclaration"},
		ID: &types.Identifier{
			BaseNode: types.BaseNode{NodeType: "Identifier"},
			Name:     "test",
		},
		Params: []types.Pattern{},
		Body: &types.BlockStatement{
			BaseNode: types.BaseNode{NodeType: "BlockStatement"},
			Body:     []types.Statement{},
		},
		Generator: false,
		Async:     false,
	}

	if funcDecl.ID.Name != "test" {
		t.Errorf("Expected function name 'test', got '%s'", funcDecl.ID.Name)
	}

	if funcDecl.Generator {
		t.Error("Expected non-generator function")
	}
}

func TestVariableDeclaration(t *testing.T) {
	t.Parallel()

	varDecl := &types.VariableDeclaration{
		BaseNode: types.BaseNode{NodeType: "VariableDeclaration"},
		Declarations: []types.VariableDeclarator{
			{
				BaseNode: types.BaseNode{NodeType: "VariableDeclarator"},
				ID: &types.Identifier{
					BaseNode: types.BaseNode{NodeType: "Identifier"},
					Name:     "x",
				},
			},
		},
		Kind: "const",
	}

	if varDecl.Kind != "const" {
		t.Errorf("Expected kind 'const', got '%s'", varDecl.Kind)
	}

	if len(varDecl.Declarations) != 1 {
		t.Errorf("Expected 1 declaration, got %d", len(varDecl.Declarations))
	}
}

func TestArrowFunctionExpression(t *testing.T) {
	t.Parallel()

	arrowFunc := &types.ArrowFunctionExpression{
		BaseNode:   types.BaseNode{NodeType: "ArrowFunctionExpression"},
		Params:     []types.Pattern{},
		Expression: true,
		Async:      false,
	}

	if !arrowFunc.Expression {
		t.Error("Expected expression arrow function")
	}
}

func TestArrayPattern(t *testing.T) {
	t.Parallel()

	arrayPattern := &types.ArrayPattern{
		BaseNode: types.BaseNode{NodeType: "ArrayPattern"},
		Elements: []types.Pattern{
			&types.Identifier{
				BaseNode: types.BaseNode{NodeType: "Identifier"},
				Name:     "a",
			},
			&types.Identifier{
				BaseNode: types.BaseNode{NodeType: "Identifier"},
				Name:     "b",
			},
		},
	}

	if len(arrayPattern.Elements) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(arrayPattern.Elements))
	}
}

func TestObjectPattern(t *testing.T) {
	t.Parallel()

	objectPattern := &types.ObjectPattern{
		BaseNode:   types.BaseNode{NodeType: "ObjectPattern"},
		Properties: []types.Node{},
	}

	if objectPattern.Type() != "ObjectPattern" {
		t.Errorf("Expected type 'ObjectPattern', got '%s'", objectPattern.Type())
	}
}

func TestTSTypeAnnotation(t *testing.T) {
	t.Parallel()

	typeAnnotation := &types.TSTypeAnnotation{
		BaseNode: types.BaseNode{NodeType: "TSTypeAnnotation"},
		TypeAnnotation: &types.TSStringKeyword{
			BaseNode: types.BaseNode{NodeType: "TSStringKeyword"},
		},
	}

	if typeAnnotation.Type() != "TSTypeAnnotation" {
		t.Errorf("Expected type 'TSTypeAnnotation', got '%s'", typeAnnotation.Type())
	}
}

func TestTSInterfaceDeclaration(t *testing.T) {
	t.Parallel()

	interfaceDecl := &types.TSInterfaceDeclaration{
		BaseNode: types.BaseNode{NodeType: "TSInterfaceDeclaration"},
		ID: &types.Identifier{
			BaseNode: types.BaseNode{NodeType: "Identifier"},
			Name:     "MyInterface",
		},
		Body: &types.TSInterfaceBody{
			BaseNode: types.BaseNode{NodeType: "TSInterfaceBody"},
			Body:     []types.Node{},
		},
		Extends: []types.TSInterfaceHeritage{},
	}

	if interfaceDecl.ID.Name != "MyInterface" {
		t.Errorf("Expected interface name 'MyInterface', got '%s'", interfaceDecl.ID.Name)
	}
}

func TestTSTypeAliasDeclaration(t *testing.T) {
	t.Parallel()

	typeAlias := &types.TSTypeAliasDeclaration{
		BaseNode: types.BaseNode{NodeType: "TSTypeAliasDeclaration"},
		ID: &types.Identifier{
			BaseNode: types.BaseNode{NodeType: "Identifier"},
			Name:     "MyType",
		},
		TypeAnnotation: &types.TSNumberKeyword{
			BaseNode: types.BaseNode{NodeType: "TSNumberKeyword"},
		},
	}

	if typeAlias.ID.Name != "MyType" {
		t.Errorf("Expected type alias name 'MyType', got '%s'", typeAlias.ID.Name)
	}
}

func TestTSEnumDeclaration(t *testing.T) {
	t.Parallel()

	enumDecl := &types.TSEnumDeclaration{
		BaseNode: types.BaseNode{NodeType: "TSEnumDeclaration"},
		ID: &types.Identifier{
			BaseNode: types.BaseNode{NodeType: "Identifier"},
			Name:     "MyEnum",
		},
		Members: []types.TSEnumMember{},
		Const:   false,
	}

	if enumDecl.ID.Name != "MyEnum" {
		t.Errorf("Expected enum name 'MyEnum', got '%s'", enumDecl.ID.Name)
	}

	if enumDecl.Const {
		t.Error("Expected non-const enum")
	}
}

func TestTokenTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tokenType types.TokenType
		expected  string
	}{
		{"Boolean", types.TokenBoolean, "Boolean"},
		{"Null", types.TokenNull, "Null"},
		{"Numeric", types.TokenNumeric, "Numeric"},
		{"String", types.TokenString, "String"},
		{"Identifier", types.TokenIdentifier, "Identifier"},
		{"Keyword", types.TokenKeyword, "Keyword"},
		{"Punctuator", types.TokenPunctuator, "Punctuator"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.tokenType) != tt.expected {
				t.Errorf("Expected token type '%s', got '%s'", tt.expected, string(tt.tokenType))
			}
		})
	}
}

func TestCommentTypes(t *testing.T) {
	t.Parallel()

	lineComment := types.Comment{
		Type:  string(types.CommentLine),
		Value: "This is a comment",
	}

	if lineComment.Type != "Line" {
		t.Errorf("Expected comment type 'Line', got '%s'", lineComment.Type)
	}

	blockComment := types.Comment{
		Type:  string(types.CommentBlock),
		Value: "This is a block comment",
	}

	if blockComment.Type != "Block" {
		t.Errorf("Expected comment type 'Block', got '%s'", blockComment.Type)
	}
}

func TestJSONSerialization(t *testing.T) {
	t.Parallel()

	program := &types.Program{
		BaseNode: types.BaseNode{
			NodeType: "Program",
			Location: &types.SourceLocation{
				Start: types.Position{Line: 1, Column: 0},
				End:   types.Position{Line: 1, Column: 10},
			},
			Span: types.Range{0, 10},
		},
		SourceType: "module",
		Body:       []types.Statement{},
		Comments:   []types.Comment{},
		Tokens:     []types.Token{},
	}

	data, err := json.Marshal(program)
	if err != nil {
		t.Fatalf("Failed to marshal program: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal program: %v", err)
	}

	if result["type"] != "Program" {
		t.Errorf("Expected type 'Program', got '%v'", result["type"])
	}

	if result["sourceType"] != "module" {
		t.Errorf("Expected sourceType 'module', got '%v'", result["sourceType"])
	}
}

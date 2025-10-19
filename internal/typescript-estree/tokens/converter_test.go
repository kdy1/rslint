package tokens

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/typescript-estree/types"
)

func TestGetTokenType(t *testing.T) {
	tests := []struct {
		name     string
		kind     ast.Kind
		expected types.TokenType
	}{
		{
			name:     "null keyword",
			kind:     ast.KindNullKeyword,
			expected: types.TokenNull,
		},
		{
			name:     "true keyword",
			kind:     ast.KindTrueKeyword,
			expected: types.TokenBoolean,
		},
		{
			name:     "false keyword",
			kind:     ast.KindFalseKeyword,
			expected: types.TokenBoolean,
		},
		{
			name:     "identifier",
			kind:     ast.KindIdentifier,
			expected: types.TokenIdentifier,
		},
		{
			name:     "private identifier",
			kind:     ast.KindPrivateIdentifier,
			expected: types.TokenIdentifier,
		},
		{
			name:     "numeric literal",
			kind:     ast.KindNumericLiteral,
			expected: types.TokenNumeric,
		},
		{
			name:     "bigint literal",
			kind:     ast.KindBigIntLiteral,
			expected: types.TokenNumeric,
		},
		{
			name:     "string literal",
			kind:     ast.KindStringLiteral,
			expected: types.TokenString,
		},
		{
			name:     "no substitution template",
			kind:     ast.KindNoSubstitutionTemplateLiteral,
			expected: types.TokenString,
		},
		{
			name:     "regular expression",
			kind:     ast.KindRegularExpressionLiteral,
			expected: types.TokenRegularExpression,
		},
		{
			name:     "template head",
			kind:     ast.KindTemplateHead,
			expected: types.TokenTemplate,
		},
		{
			name:     "template middle",
			kind:     ast.KindTemplateMiddle,
			expected: types.TokenTemplate,
		},
		{
			name:     "template tail",
			kind:     ast.KindTemplateTail,
			expected: types.TokenTemplate,
		},
		{
			name:     "jsx text",
			kind:     ast.KindJsxText,
			expected: types.TokenJSXText,
		},
		{
			name:     "const keyword",
			kind:     ast.KindConstKeyword,
			expected: types.TokenKeyword,
		},
		{
			name:     "let keyword",
			kind:     ast.KindLetKeyword,
			expected: types.TokenKeyword,
		},
		{
			name:     "function keyword",
			kind:     ast.KindFunctionKeyword,
			expected: types.TokenKeyword,
		},
		{
			name:     "plus token",
			kind:     ast.KindPlusToken,
			expected: types.TokenPunctuator,
		},
		{
			name:     "minus token",
			kind:     ast.KindMinusToken,
			expected: types.TokenPunctuator,
		},
		{
			name:     "equals token",
			kind:     ast.KindEqualsToken,
			expected: types.TokenPunctuator,
		},
		{
			name:     "semicolon token",
			kind:     ast.KindSemicolonToken,
			expected: types.TokenPunctuator,
		},
		{
			name:     "open brace",
			kind:     ast.KindOpenBraceToken,
			expected: types.TokenPunctuator,
		},
		{
			name:     "close brace",
			kind:     ast.KindCloseBraceToken,
			expected: types.TokenPunctuator,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ast.Node{
				Kind: tt.kind,
			}

			result := GetTokenType(node)

			if result != tt.expected {
				t.Errorf("GetTokenType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsKeyword(t *testing.T) {
	tests := []struct {
		name     string
		kind     ast.Kind
		expected bool
	}{
		{
			name:     "const is keyword",
			kind:     ast.KindConstKeyword,
			expected: true,
		},
		{
			name:     "let is keyword",
			kind:     ast.KindLetKeyword,
			expected: true,
		},
		{
			name:     "function is keyword",
			kind:     ast.KindFunctionKeyword,
			expected: true,
		},
		{
			name:     "if is keyword",
			kind:     ast.KindIfKeyword,
			expected: true,
		},
		{
			name:     "identifier is not keyword",
			kind:     ast.KindIdentifier,
			expected: false,
		},
		{
			name:     "plus token is not keyword",
			kind:     ast.KindPlusToken,
			expected: false,
		},
		{
			name:     "numeric literal is not keyword",
			kind:     ast.KindNumericLiteral,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isKeyword(tt.kind)

			if result != tt.expected {
				t.Errorf("isKeyword() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsPunctuator(t *testing.T) {
	tests := []struct {
		name     string
		kind     ast.Kind
		expected bool
	}{
		{
			name:     "open brace is punctuator",
			kind:     ast.KindOpenBraceToken,
			expected: true,
		},
		{
			name:     "close brace is punctuator",
			kind:     ast.KindCloseBraceToken,
			expected: true,
		},
		{
			name:     "semicolon is punctuator",
			kind:     ast.KindSemicolonToken,
			expected: true,
		},
		{
			name:     "plus is punctuator",
			kind:     ast.KindPlusToken,
			expected: true,
		},
		{
			name:     "equals is punctuator",
			kind:     ast.KindEqualsToken,
			expected: true,
		},
		{
			name:     "arrow is punctuator",
			kind:     ast.KindEqualsGreaterThanToken,
			expected: true,
		},
		{
			name:     "identifier is not punctuator",
			kind:     ast.KindIdentifier,
			expected: false,
		},
		{
			name:     "keyword is not punctuator",
			kind:     ast.KindConstKeyword,
			expected: false,
		},
		{
			name:     "numeric literal is not punctuator",
			kind:     ast.KindNumericLiteral,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPunctuator(tt.kind)

			if result != tt.expected {
				t.Errorf("isPunctuator() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsToken(t *testing.T) {
	tests := []struct {
		name     string
		kind     ast.Kind
		expected bool
	}{
		{
			name:     "identifier is token",
			kind:     ast.KindIdentifier,
			expected: true,
		},
		{
			name:     "numeric literal is token",
			kind:     ast.KindNumericLiteral,
			expected: true,
		},
		{
			name:     "string literal is token",
			kind:     ast.KindStringLiteral,
			expected: true,
		},
		{
			name:     "keyword is token",
			kind:     ast.KindConstKeyword,
			expected: true,
		},
		{
			name:     "punctuator is token",
			kind:     ast.KindPlusToken,
			expected: true,
		},
		{
			name:     "binary expression is not token",
			kind:     ast.KindBinaryExpression,
			expected: false,
		},
		{
			name:     "variable declaration is not token",
			kind:     ast.KindVariableDeclaration,
			expected: false,
		},
		{
			name:     "function declaration is not token",
			kind:     ast.KindFunctionDeclaration,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ast.Node{
				Kind: tt.kind,
			}

			result := isToken(node)

			if result != tt.expected {
				t.Errorf("isToken() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestTokenTypeConstants verifies that our token type constants are correctly defined.
func TestTokenTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		token    types.TokenType
		expected string
	}{
		{"Boolean", types.TokenBoolean, "Boolean"},
		{"Null", types.TokenNull, "Null"},
		{"Numeric", types.TokenNumeric, "Numeric"},
		{"String", types.TokenString, "String"},
		{"RegularExpression", types.TokenRegularExpression, "RegularExpression"},
		{"Template", types.TokenTemplate, "Template"},
		{"Identifier", types.TokenIdentifier, "Identifier"},
		{"Keyword", types.TokenKeyword, "Keyword"},
		{"Punctuator", types.TokenPunctuator, "Punctuator"},
		{"JSXIdentifier", types.TokenJSXIdentifier, "JSXIdentifier"},
		{"JSXText", types.TokenJSXText, "JSXText"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.token) != tt.expected {
				t.Errorf("Token constant %s = %v, want %v", tt.name, tt.token, tt.expected)
			}
		})
	}
}

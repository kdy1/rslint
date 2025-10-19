// Package tokens provides utilities for converting TypeScript tokens to ESTree-compatible tokens.
package tokens

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/typescript-estree/types"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// GetTokenType maps a TypeScript token kind to an ESTree token type.
// Based on typescript-eslint's getTokenType function.
func GetTokenType(token *ast.Node) types.TokenType {
	kind := token.Kind

	// Handle null keyword
	if kind == ast.KindNullKeyword {
		return types.TokenNull
	}

	// Handle boolean keywords
	if kind == ast.KindTrueKeyword || kind == ast.KindFalseKeyword {
		return types.TokenBoolean
	}

	// Handle identifiers
	if kind == ast.KindIdentifier {
		return types.TokenIdentifier
	}

	// Handle private identifiers
	if kind == ast.KindPrivateIdentifier {
		return types.TokenIdentifier
	}

	// Handle numeric literals
	if kind == ast.KindNumericLiteral {
		return types.TokenNumeric
	}

	// Handle BigInt literals
	if kind == ast.KindBigIntLiteral {
		return types.TokenNumeric
	}

	// Handle string literals
	if kind == ast.KindStringLiteral || kind == ast.KindNoSubstitutionTemplateLiteral {
		return types.TokenString
	}

	// Handle regular expression literals
	if kind == ast.KindRegularExpressionLiteral {
		return types.TokenRegularExpression
	}

	// Handle template literals
	if kind == ast.KindTemplateHead ||
		kind == ast.KindTemplateMiddle ||
		kind == ast.KindTemplateTail {
		return types.TokenTemplate
	}

	// Handle JSX tokens
	if kind == ast.KindJsxText {
		return types.TokenJSXText
	}

	if kind == ast.KindJsxTextAllWhiteSpaces {
		return types.TokenJSXText
	}

	// Check if it's a keyword
	if isKeyword(kind) {
		return types.TokenKeyword
	}

	// Everything else is a punctuator
	return types.TokenPunctuator
}

// isKeyword checks if a token kind is a keyword.
func isKeyword(kind ast.Kind) bool {
	return kind >= ast.KindFirstKeyword && kind <= ast.KindLastKeyword
}

// ConvertToken converts a TypeScript token to an ESTree token.
// Based on typescript-eslint's convertToken function.
func ConvertToken(token *ast.Node, sourceFile *ast.SourceFile) *types.Token {
	start := token.Pos()
	end := token.End()

	// Get the token text from the source
	tokenText := scanner.GetSourceTextOfNodeFromSourceFile(sourceFile, token, false)

	tokenType := GetTokenType(token)
	kind := token.Kind

	// Handle special cases for token values
	value := tokenText

	// Handle private identifiers - remove the leading '#'
	if kind == ast.KindPrivateIdentifier {
		value = strings.TrimPrefix(tokenText, "#")
	}

	// Create the ESTree token
	estreeToken := &types.Token{
		Type:  string(tokenType),
		Value: value,
		Range: types.Range{int(start), int(end)},
	}

	// Calculate source location
	startLine, startColumn := scanner.GetLineAndCharacterOfPosition(sourceFile, int(start))
	endLine, endColumn := scanner.GetLineAndCharacterOfPosition(sourceFile, int(end))

	estreeToken.Loc = &types.SourceLocation{
		Start: types.Position{
			Line:   startLine + 1, // ESTree uses 1-based line numbers
			Column: startColumn,
		},
		End: types.Position{
			Line:   endLine + 1, // ESTree uses 1-based line numbers
			Column: endColumn,
		},
	}

	return estreeToken
}

// ConvertTokens extracts and converts all tokens from a TypeScript SourceFile.
// Based on typescript-eslint's convertTokens function.
// Note: This is a simplified implementation. A full implementation would require
// walking the AST tree to extract all tokens. For now, this provides the structure
// and can be extended when integrated with the actual AST traversal logic.
func ConvertTokens(sourceFile *ast.SourceFile) []*types.Token {
	// TODO: Implement full AST traversal to extract tokens
	// This will require understanding the complete Children iteration API
	// For now, return empty slice as placeholder
	return []*types.Token{}
}

// isToken checks if a node is a token (leaf node) rather than a composite node.
func isToken(node *ast.Node) bool {
	kind := node.Kind

	// Tokens are generally leaf nodes with specific kinds
	// Identifiers, literals, keywords, and punctuators are tokens
	if kind == ast.KindIdentifier ||
		kind == ast.KindPrivateIdentifier ||
		kind == ast.KindNumericLiteral ||
		kind == ast.KindBigIntLiteral ||
		kind == ast.KindStringLiteral ||
		kind == ast.KindNoSubstitutionTemplateLiteral ||
		kind == ast.KindRegularExpressionLiteral ||
		kind == ast.KindTemplateHead ||
		kind == ast.KindTemplateMiddle ||
		kind == ast.KindTemplateTail ||
		kind == ast.KindJsxText ||
		kind == ast.KindJsxTextAllWhiteSpaces {
		return true
	}

	// Check if it's a keyword
	if isKeyword(kind) {
		return true
	}

	// Check if it's a punctuator token
	return isPunctuator(kind)
}

// isPunctuator checks if a token kind is a punctuator.
func isPunctuator(kind ast.Kind) bool {
	switch kind {
	case ast.KindOpenBraceToken,
		ast.KindCloseBraceToken,
		ast.KindOpenParenToken,
		ast.KindCloseParenToken,
		ast.KindOpenBracketToken,
		ast.KindCloseBracketToken,
		ast.KindDotToken,
		ast.KindDotDotDotToken,
		ast.KindSemicolonToken,
		ast.KindCommaToken,
		ast.KindQuestionDotToken,
		ast.KindLessThanToken,
		ast.KindLessThanSlashToken,
		ast.KindGreaterThanToken,
		ast.KindLessThanEqualsToken,
		ast.KindGreaterThanEqualsToken,
		ast.KindEqualsEqualsToken,
		ast.KindExclamationEqualsToken,
		ast.KindEqualsEqualsEqualsToken,
		ast.KindExclamationEqualsEqualsToken,
		ast.KindEqualsGreaterThanToken,
		ast.KindPlusToken,
		ast.KindMinusToken,
		ast.KindAsteriskToken,
		ast.KindAsteriskAsteriskToken,
		ast.KindSlashToken,
		ast.KindPercentToken,
		ast.KindPlusPlusToken,
		ast.KindMinusMinusToken,
		ast.KindLessThanLessThanToken,
		ast.KindGreaterThanGreaterThanToken,
		ast.KindGreaterThanGreaterThanGreaterThanToken,
		ast.KindAmpersandToken,
		ast.KindBarToken,
		ast.KindCaretToken,
		ast.KindExclamationToken,
		ast.KindTildeToken,
		ast.KindAmpersandAmpersandToken,
		ast.KindBarBarToken,
		ast.KindQuestionToken,
		ast.KindColonToken,
		ast.KindAtToken,
		ast.KindQuestionQuestionToken,
		ast.KindBacktickToken,
		ast.KindHashToken,
		ast.KindEqualsToken,
		ast.KindPlusEqualsToken,
		ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken,
		ast.KindAsteriskAsteriskEqualsToken,
		ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken,
		ast.KindBarBarEqualsToken,
		ast.KindAmpersandAmpersandEqualsToken,
		ast.KindQuestionQuestionEqualsToken,
		ast.KindCaretEqualsToken:
		return true
	default:
		return false
	}
}

// Package comments provides functionality for extracting and converting
// TypeScript comments to ESTree comment format.
package comments

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/typescript-estree/types"
)

// ConvertComments extracts all comments from a TypeScript SourceFile and converts
// them to ESTree comment format.
//
// This function iterates through both leading and trailing comments in the source
// file and converts them to the ESTree Comment type with proper location information.
//
// Parameters:
//   - sourceFile: The TypeScript SourceFile to extract comments from
//   - code: The original source code text
//
// Returns:
//   - A slice of ESTree Comment objects
func ConvertComments(sourceFile *ast.SourceFile, code string) []types.Comment {
	var comments []types.Comment

	// Extract all comment ranges from the source file
	commentRanges := extractCommentRanges(sourceFile, code)

	// Convert each comment range to ESTree comment format
	for _, cr := range commentRanges {
		comment := convertCommentRange(cr, code, sourceFile)
		comments = append(comments, comment)
	}

	return comments
}

// extractCommentRanges extracts all comment ranges from the TypeScript source file.
//
// TODO: Implement proper comment extraction using TypeScript's scanner.
// The current implementation needs to properly traverse the AST or use
// a scanner-based approach similar to TypeScript-ESTree's forEachComment.
//
// Reference: https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/typescript-estree/src/convert-comments.ts
func extractCommentRanges(sourceFile *ast.SourceFile, code string) []ast.CommentRange {
	var ranges []ast.CommentRange

	// TODO: Implement comment extraction
	// Options:
	// 1. Use scanner.GetLeadingCommentRanges and GetTrailingCommentRanges properly
	// 2. Manually scan the source text for comment patterns
	// 3. Use TypeScript's forEachComment equivalent if available

	return ranges
}

// convertCommentRange converts a TypeScript CommentRange to an ESTree Comment.
func convertCommentRange(cr ast.CommentRange, code string, sourceFile *ast.SourceFile) types.Comment {
	// Determine comment type based on TypeScript syntax kind
	commentType := getCommentType(cr.Kind())

	// Extract the comment text
	value := extractCommentValue(cr, code, commentType)

	// Calculate position range
	start := cr.Pos()
	end := cr.End()
	commentRange := types.Range{start, end}

	// Calculate source location
	loc := calculateLocation(start, end, sourceFile)

	return types.Comment{
		Type:  string(commentType),
		Value: value,
		Range: commentRange,
		Loc:   loc,
	}
}

// getCommentType determines the ESTree comment type from TypeScript syntax kind.
func getCommentType(kind ast.Kind) types.CommentType {
	if kind == ast.KindSingleLineCommentTrivia {
		return types.CommentLine
	}
	return types.CommentBlock
}

// extractCommentValue extracts the comment text without the comment delimiters.
//
// For line comments (//), it removes the leading "//"
// For block comments (/* */), it removes the leading "/*" and trailing "*/"
func extractCommentValue(cr ast.CommentRange, code string, commentType types.CommentType) string {
	start := cr.Pos()
	end := cr.End()

	if commentType == types.CommentLine {
		// Skip the "//" prefix (2 characters)
		textStart := start + 2
		if textStart > end {
			return ""
		}
		return code[textStart:end]
	}

	// Block comment - skip "/*" prefix and "*/" suffix (2 characters each)
	textStart := start + 2
	textEnd := end - 2
	if textStart > textEnd {
		return ""
	}
	return code[textStart:textEnd]
}

// calculateLocation calculates the ESTree SourceLocation from start and end positions.
func calculateLocation(start, end int, sourceFile *ast.SourceFile) *types.SourceLocation {
	startLine, startCol := scanner.GetLineAndCharacterOfPosition(sourceFile, start)
	endLine, endCol := scanner.GetLineAndCharacterOfPosition(sourceFile, end)

	return &types.SourceLocation{
		Start: types.Position{
			Line:   startLine + 1, // ESTree uses 1-based line numbers
			Column: startCol,
		},
		End: types.Position{
			Line:   endLine + 1, // ESTree uses 1-based line numbers
			Column: endCol,
		},
	}
}

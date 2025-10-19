// Package types provides ESTree-compliant type definitions for TypeScript AST nodes.
package types

// TokenType represents the type of a token.
type TokenType string

// Token types based on ESTree and TypeScript-ESTree specifications.
const (
	// Boolean literal
	TokenBoolean TokenType = "Boolean"

	// Null literal
	TokenNull TokenType = "Null"

	// Numeric literal
	TokenNumeric TokenType = "Numeric"

	// String literal
	TokenString TokenType = "String"

	// Regular expression literal
	TokenRegularExpression TokenType = "RegularExpression"

	// Template literal
	TokenTemplate TokenType = "Template"

	// Identifiers
	TokenIdentifier TokenType = "Identifier"

	// Keywords
	TokenKeyword TokenType = "Keyword"

	// Punctuators
	TokenPunctuator TokenType = "Punctuator"

	// JSX tokens
	TokenJSXIdentifier TokenType = "JSXIdentifier"
	TokenJSXText       TokenType = "JSXText"
)

// Keyword types
const (
	KeywordBreak      = "break"
	KeywordCase       = "case"
	KeywordCatch      = "catch"
	KeywordClass      = "class"
	KeywordConst      = "const"
	KeywordContinue   = "continue"
	KeywordDebugger   = "debugger"
	KeywordDefault    = "default"
	KeywordDelete     = "delete"
	KeywordDo         = "do"
	KeywordElse       = "else"
	KeywordEnum       = "enum"
	KeywordExport     = "export"
	KeywordExtends    = "extends"
	KeywordFalse      = "false"
	KeywordFinally    = "finally"
	KeywordFor        = "for"
	KeywordFunction   = "function"
	KeywordIf         = "if"
	KeywordImport     = "import"
	KeywordIn         = "in"
	KeywordInstanceof = "instanceof"
	KeywordLet        = "let"
	KeywordNew        = "new"
	KeywordNull       = "null"
	KeywordReturn     = "return"
	KeywordSuper      = "super"
	KeywordSwitch     = "switch"
	KeywordThis       = "this"
	KeywordThrow      = "throw"
	KeywordTrue       = "true"
	KeywordTry        = "try"
	KeywordTypeof     = "typeof"
	KeywordVar        = "var"
	KeywordVoid       = "void"
	KeywordWhile      = "while"
	KeywordWith       = "with"
	KeywordYield      = "yield"

	// Strict mode reserved words
	KeywordImplements = "implements"
	KeywordInterface  = "interface"
	KeywordPackage    = "package"
	KeywordPrivate    = "private"
	KeywordProtected  = "protected"
	KeywordPublic     = "public"
	KeywordStatic     = "static"

	// Contextual keywords
	KeywordAs        = "as"
	KeywordAsync     = "async"
	KeywordAwait     = "await"
	KeywordFrom      = "from"
	KeywordGet       = "get"
	KeywordOf        = "of"
	KeywordSet       = "set"
	KeywordTarget    = "target"

	// TypeScript keywords
	KeywordAbstract   = "abstract"
	KeywordAny        = "any"
	KeywordAsserts    = "asserts"
	KeywordBigint     = "bigint"
	KeywordBoolean    = "boolean"
	KeywordConstructor = "constructor"
	KeywordDeclare    = "declare"
	KeywordInfer      = "infer"
	KeywordIs         = "is"
	KeywordKeyof      = "keyof"
	KeywordModule     = "module"
	KeywordNamespace  = "namespace"
	KeywordNever      = "never"
	KeywordNumber     = "number"
	KeywordObject     = "object"
	KeywordReadonly   = "readonly"
	KeywordRequire    = "require"
	KeywordString     = "string"
	KeywordSymbol     = "symbol"
	KeywordType       = "type"
	KeywordUndefined  = "undefined"
	KeywordUnique     = "unique"
	KeywordUnknown    = "unknown"
	KeywordGlobal     = "global"
	KeywordOverride   = "override"
	KeywordSatisfies  = "satisfies"
)

// Punctuator types
const (
	PunctuatorBraceL        = "{"
	PunctuatorBraceR        = "}"
	PunctuatorParenL        = "("
	PunctuatorParenR        = ")"
	PunctuatorBracketL      = "["
	PunctuatorBracketR      = "]"
	PunctuatorDot           = "."
	PunctuatorSemi          = ";"
	PunctuatorComma         = ","
	PunctuatorLt            = "<"
	PunctuatorGt            = ">"
	PunctuatorLtEq          = "<="
	PunctuatorGtEq          = ">="
	PunctuatorEq            = "=="
	PunctuatorNotEq         = "!="
	PunctuatorStrictEq      = "==="
	PunctuatorStrictNotEq   = "!=="
	PunctuatorPlus          = "+"
	PunctuatorMinus         = "-"
	PunctuatorStar          = "*"
	PunctuatorSlash         = "/"
	PunctuatorPercent       = "%"
	PunctuatorPlusPlus      = "++"
	PunctuatorMinusMinus    = "--"
	PunctuatorLtLt          = "<<"
	PunctuatorGtGt          = ">>"
	PunctuatorGtGtGt        = ">>>"
	PunctuatorAmp           = "&"
	PunctuatorPipe          = "|"
	PunctuatorCaret         = "^"
	PunctuatorBang          = "!"
	PunctuatorTilde         = "~"
	PunctuatorAmpAmp        = "&&"
	PunctuatorPipePipe      = "||"
	PunctuatorQuestion      = "?"
	PunctuatorColon         = ":"
	PunctuatorAssign        = "="
	PunctuatorPlusAssign    = "+="
	PunctuatorMinusAssign   = "-="
	PunctuatorStarAssign    = "*="
	PunctuatorSlashAssign   = "/="
	PunctuatorPercentAssign = "%="
	PunctuatorLtLtAssign    = "<<="
	PunctuatorGtGtAssign    = ">>="
	PunctuatorGtGtGtAssign  = ">>>="
	PunctuatorAmpAssign     = "&="
	PunctuatorPipeAssign    = "|="
	PunctuatorCaretAssign   = "^="
	PunctuatorArrow         = "=>"
	PunctuatorSpread        = "..."
	PunctuatorStarStar      = "**"
	PunctuatorStarStarAssign = "**="
	PunctuatorQuestionQuestion = "??"
	PunctuatorQuestionQuestionAssign = "??="
	PunctuatorAmpAmpAssign  = "&&="
	PunctuatorPipePipeAssign = "||="
	PunctuatorQuestionDot   = "?."
	PunctuatorHash          = "#"
)

// CommentType represents the type of a comment.
type CommentType string

const (
	// Line comment (//)
	CommentLine CommentType = "Line"

	// Block comment (/* */)
	CommentBlock CommentType = "Block"
)

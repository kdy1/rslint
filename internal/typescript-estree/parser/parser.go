// Package parser provides the main parsing functionality for converting TypeScript
// source code into ESTree-compliant AST nodes.
//
// This package implements the core parsing functions from typescript-estree,
// delegating to the typescript-go compiler and converting the results to ESTree format.
package parser

import (
	"context"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/typescript-estree/converter"
	"github.com/web-infra-dev/rslint/internal/typescript-estree/types"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// JSDocParsingMode controls how JSDoc comments are parsed.
type JSDocParsingMode string

const (
	// JSDocParsingModeAll parses all JSDoc comments
	JSDocParsingModeAll JSDocParsingMode = "all"
	// JSDocParsingModeNone skips JSDoc parsing
	JSDocParsingModeNone JSDocParsingMode = "none"
	// JSDocParsingModeTypeInfo parses JSDoc only when type info is available
	JSDocParsingModeTypeInfo JSDocParsingMode = "type-info"
)

// ParseOptions contains configuration options for parsing TypeScript/JavaScript code.
// These options match the typescript-estree API for compatibility.
type ParseOptions struct {
	// SourceType specifies whether to parse as "script" or "module".
	// Default is "script" unless import/export is detected.
	SourceType string

	// AllowInvalidAST prevents throwing errors on invalid ASTs.
	// When true, the parser will do its best to produce an AST even for invalid code.
	AllowInvalidAST bool

	// Comment creates a top-level comments array containing all comments.
	Comment bool

	// SuppressDeprecatedPropertyWarnings skips warnings for deprecated AST properties.
	SuppressDeprecatedPropertyWarnings bool

	// DebugLevel controls debugging output for specific modules.
	// Can be true to enable all debug output, or an array of module names.
	DebugLevel interface{} // bool or []string

	// ErrorOnUnknownASTType throws an error when an unknown AST node type is encountered.
	ErrorOnUnknownASTType bool

	// FilePath is the absolute or relative path to the file being parsed.
	// Used for error messages and resolving imports.
	FilePath string

	// JSDocParsingMode controls how JSDoc comments are parsed.
	// Options: "all", "none", "type-info"
	JSDocParsingMode JSDocParsingMode

	// JSX enables parsing of JSX syntax.
	JSX bool

	// Loc includes location information (line/column) for each node.
	Loc bool

	// LoggerFn overrides the default logging function.
	// Set to false to disable logging, or provide a custom function.
	LoggerFn interface{} // func(string) or bool

	// Range includes [start, end] byte offsets for each node.
	Range bool

	// Tokens creates a top-level array of tokens from the file.
	Tokens bool

	// Project specifies the path to a tsconfig.json file or a directory containing one.
	// Required for parseAndGenerateServices when you need type information.
	Project string

	// TsconfigRootDir specifies the root directory for relative tsconfig paths.
	TsconfigRootDir string

	// Programs provides pre-created TypeScript programs to use for parsing.
	// This can improve performance when parsing multiple files.
	Programs []*compiler.Program
}

// ParseSettings holds the internal configuration derived from ParseOptions.
// This is used internally to configure the TypeScript compiler.
type ParseSettings struct {
	Code                string
	FilePath            string
	SourceType          string
	JSX                 bool
	Loc                 bool
	Range               bool
	Tokens              bool
	Comment             bool
	Project             string
	TsconfigRootDir     string
	AllowInvalidAST     bool
	JSDocParsingMode    JSDocParsingMode
	Programs            []*compiler.Program
	ErrorOnUnknownAST   bool
	SuppressDeprecations bool
}

// ParserServices provides additional services for working with the parsed AST.
// This includes type information and bidirectional mappings between ESTree and TypeScript AST nodes.
type ParserServices struct {
	// Program is the TypeScript compiler program, providing type checking services
	Program *compiler.Program

	// ESTreeNodeToTSNodeMap maps ESTree nodes to their corresponding TypeScript AST nodes
	ESTreeNodeToTSNodeMap map[types.Node]*ast.Node

	// TSNodeToESTreeNodeMap maps TypeScript AST nodes to their corresponding ESTree nodes
	TSNodeToESTreeNodeMap map[*ast.Node]types.Node
}

// ParseResult contains the result of parsing with services.
type ParseResult struct {
	AST      *types.Program
	Services *ParserServices
}

// Parse parses TypeScript/JavaScript source code and returns an ESTree-compliant AST.
// This is the basic parsing function that does not include type information.
//
// Example:
//
//	ast, err := parser.Parse("const x = 42;", &parser.ParseOptions{
//	    FilePath: "example.ts",
//	    Loc: true,
//	    Range: true,
//	})
func Parse(source string, options *ParseOptions) (*types.Program, error) {
	if options == nil {
		options = &ParseOptions{}
	}

	settings := buildParseSettings(source, options)

	// Parse using TypeScript compiler
	sourceFile, program, err := parseWithTypeScript(settings, false)
	if err != nil {
		return nil, err
	}

	// Convert TypeScript AST to ESTree format
	convertOptions := &converter.ConvertOptions{
		FilePath:   settings.FilePath,
		SourceType: settings.SourceType,
		Loc:        settings.Loc,
		Range:      settings.Range,
		Tokens:     settings.Tokens,
		Comment:    settings.Comment,
	}

	estree, err := converter.ConvertProgram(sourceFile, program, convertOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to convert AST: %w", err)
	}

	return estree, nil
}

// ParseAndGenerateServices parses TypeScript/JavaScript source code and returns both
// the ESTree AST and parser services that provide access to type information.
//
// This function is required when you need TypeScript type checking capabilities.
// You must provide a tsconfig.json via the Project option.
//
// Example:
//
//	result, err := parser.ParseAndGenerateServices("const x: number = 42;", &parser.ParseOptions{
//	    FilePath: "example.ts",
//	    Project: "./tsconfig.json",
//	    Loc: true,
//	    Range: true,
//	})
//	// Access type information via result.Services.Program
func ParseAndGenerateServices(source string, options *ParseOptions) (*ParseResult, error) {
	if options == nil {
		options = &ParseOptions{}
	}

	settings := buildParseSettings(source, options)

	// Validate that we have a project configuration for type information
	if settings.Project == "" && len(settings.Programs) == 0 {
		return nil, fmt.Errorf("parseAndGenerateServices requires either 'project' or 'programs' option to be specified")
	}

	// Parse using TypeScript compiler with type information
	sourceFile, program, err := parseWithTypeScript(settings, true)
	if err != nil {
		return nil, err
	}

	// Convert TypeScript AST to ESTree format with node mappings
	convertOptions := &converter.ConvertOptions{
		FilePath:          settings.FilePath,
		SourceType:        settings.SourceType,
		Loc:               settings.Loc,
		Range:             settings.Range,
		Tokens:            settings.Tokens,
		Comment:           settings.Comment,
		PreserveNodeMaps:  true, // Enable bidirectional node mapping
	}

	estree, err := converter.ConvertProgram(sourceFile, program, convertOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to convert AST: %w", err)
	}

	// Create parser services with type information
	services := &ParserServices{
		Program: program,
		// TODO: Get actual node maps from converter
		ESTreeNodeToTSNodeMap: make(map[types.Node]*ast.Node),
		TSNodeToESTreeNodeMap: make(map[*ast.Node]types.Node),
	}

	return &ParseResult{
		AST:      estree,
		Services: services,
	}, nil
}

// buildParseSettings creates ParseSettings from ParseOptions.
func buildParseSettings(source string, options *ParseOptions) *ParseSettings {
	settings := &ParseSettings{
		Code:                 source,
		FilePath:             options.FilePath,
		SourceType:           options.SourceType,
		JSX:                  options.JSX,
		Loc:                  options.Loc,
		Range:                options.Range,
		Tokens:               options.Tokens,
		Comment:              options.Comment,
		Project:              options.Project,
		TsconfigRootDir:      options.TsconfigRootDir,
		AllowInvalidAST:      options.AllowInvalidAST,
		JSDocParsingMode:     options.JSDocParsingMode,
		Programs:             options.Programs,
		ErrorOnUnknownAST:    options.ErrorOnUnknownASTType,
		SuppressDeprecations: options.SuppressDeprecatedPropertyWarnings,
	}

	// Default source type to "script"
	if settings.SourceType == "" {
		settings.SourceType = "script"
	}

	// Default JSDoc parsing mode
	if settings.JSDocParsingMode == "" {
		settings.JSDocParsingMode = JSDocParsingModeAll
	}

	return settings
}

// parseWithTypeScript uses the TypeScript compiler to parse source code.
func parseWithTypeScript(settings *ParseSettings, needsProgram bool) (*ast.SourceFile, *compiler.Program, error) {
	_ = context.Background() // TODO: Use context for cancellation

	// Determine file path
	filePath := settings.FilePath
	if filePath == "" {
		filePath = "file.ts"
		if settings.JSX {
			filePath = "file.tsx"
		}
	}

	// Use existing program if provided
	var program *compiler.Program
	if len(settings.Programs) > 0 {
		program = settings.Programs[0]
	} else if needsProgram && settings.Project != "" {
		// Create program from tsconfig
		cwd := settings.TsconfigRootDir
		if cwd == "" {
			cwd = "."
		}

		fs := osvfs.FS()
		host := utils.CreateCompilerHost(cwd, fs)

		var err error
		program, err = utils.CreateProgram(true, fs, cwd, settings.Project, host)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create program: %w", err)
		}
	}

	var sourceFile *ast.SourceFile

	if program != nil {
		// Get or update source file in existing program
		path := tspath.ResolvePath(".", filePath)
		sourceFile = program.GetSourceFile(path)

		if sourceFile == nil {
			return nil, nil, fmt.Errorf("source file not found in program: %s", filePath)
		}
	} else {
		// For standalone parsing without a program, we create a placeholder source file
		// The actual TypeScript parsing will be done by the converter
		// TODO: Implement standalone parsing via TypeScript API
		// For now, return a minimal structure that indicates parsing mode
		return nil, nil, fmt.Errorf("standalone parsing without program not yet fully implemented - please provide a program or project config")
	}

	return sourceFile, program, nil
}

// GetSupportedTypeScriptVersion returns the version range of TypeScript supported by this parser.
func GetSupportedTypeScriptVersion() string {
	// This should match the version of typescript-go being used
	return ">=4.7.0 <6.0.0"
}

// ValidateTypeScriptVersion checks if the given TypeScript version is supported.
func ValidateTypeScriptVersion(version string) error {
	// TODO: Implement proper semver validation
	// For now, we accept any version
	return nil
}

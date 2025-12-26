package triple_slash_reference

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type TripleSlashReferenceOptions struct {
	Lib   string `json:"lib"`   // "always" | "never"
	Path  string `json:"path"`  // "always" | "never" | "prefer-import"
	Types string `json:"types"` // "always" | "never" | "prefer-import"
}

var tripleSlashRegex = regexp.MustCompile(`^///\s*<reference\s+(path|types|lib)\s*=\s*["']([^"']+)["']`)

// TripleSlashReferenceRule implements the triple-slash-reference rule
// Disallow certain triple slash directives
var TripleSlashReferenceRule = rule.CreateRule(rule.Rule{
	Name: "triple-slash-reference",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	// Default options match TypeScript-ESLint defaults
	opts := TripleSlashReferenceOptions{
		Lib:   "always",
		Path:  "never",
		Types: "prefer-import",
	}

	// Parse options
	if options != nil {
		var optsMap map[string]interface{}
		var ok bool

		// Handle array format: [{ option: value }]
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			optsMap, ok = optArray[0].(map[string]interface{})
		} else {
			// Handle direct object format: { option: value }
			optsMap, ok = options.(map[string]interface{})
		}

		if ok {
			if lib, ok := optsMap["lib"].(string); ok {
				opts.Lib = lib
			}
			if path, ok := optsMap["path"].(string); ok {
				opts.Path = path
			}
			if types, ok := optsMap["types"].(string); ok {
				opts.Types = types
			}
		}
	}

	// Get the full text of the source file
	text := ctx.SourceFile.Text()

	// Process comments to find triple-slash references
	processTripleSlashReferences(ctx, text, opts)

	return rule.RuleListeners{}
}

// processTripleSlashReferences scans the source text for triple-slash references
func processTripleSlashReferences(ctx rule.RuleContext, text string, opts TripleSlashReferenceOptions) {
	pos := 0
	length := len(text)

	// Collect all imported module names for prefer-import mode
	importedModules := getImportedModules(ctx.SourceFile)

	for pos < length {
		// Skip to next potential comment
		if pos+1 < length {
			if text[pos] == '/' && text[pos+1] == '/' {
				// Check for triple-slash (///)
				if pos+2 < length && text[pos+2] == '/' {
					// Triple-slash comment - potential reference directive
					commentStart := pos
					pos += 3
					lineEnd := pos
					for lineEnd < length && text[lineEnd] != '\n' && text[lineEnd] != '\r' {
						lineEnd++
					}
					commentText := text[commentStart:lineEnd]
					checkTripleSlashReference(ctx, commentText, commentStart, opts, importedModules)
					pos = lineEnd
				} else {
					// Regular single-line comment - skip
					pos += 2
					for pos < length && text[pos] != '\n' && text[pos] != '\r' {
						pos++
					}
				}
			} else if text[pos] == '/' && text[pos+1] == '*' {
				// Multi-line comment - skip entirely (references inside are ignored)
				pos += 2
				for pos+1 < length {
					if text[pos] == '*' && text[pos+1] == '/' {
						pos += 2
						break
					}
					pos++
				}
			} else {
				pos++
			}
		} else {
			pos++
		}
	}
}

// checkTripleSlashReference checks a single triple-slash comment for reference directives
func checkTripleSlashReference(ctx rule.RuleContext, commentText string, commentStart int, opts TripleSlashReferenceOptions, importedModules map[string]bool) {
	matches := tripleSlashRegex.FindStringSubmatch(commentText)
	if matches == nil {
		return
	}

	refType := matches[1]      // "path", "types", or "lib"
	refValue := matches[2]     // the value in quotes

	// Check if this reference should be reported
	shouldReport := false
	switch refType {
	case "path":
		shouldReport = opts.Path == "never"
	case "types":
		if opts.Types == "never" {
			shouldReport = true
		} else if opts.Types == "prefer-import" {
			// Only report if there's an import for this specific module
			shouldReport = importedModules[refValue]
		}
	case "lib":
		shouldReport = opts.Lib == "never"
	}

	if shouldReport {
		ctx.ReportRange(
			core.NewTextRange(commentStart, commentStart+len(commentText)),
			rule.RuleMessage{
				Id:          "tripleSlashReference",
				Description: "Do not use a triple slash reference for " + refType + ", use `import` style instead.",
			},
		)
	}
}

// getImportedModules returns a map of module names that are imported in the source file
func getImportedModules(sourceFile *ast.SourceFile) map[string]bool {
	modules := make(map[string]bool)

	if sourceFile.Statements == nil {
		return modules
	}

	for _, stmt := range sourceFile.Statements.Nodes {
		switch stmt.Kind {
		case ast.KindImportDeclaration:
			// ES6 import: import * as foo from 'module-name'
			importDecl := stmt.AsImportDeclaration()
			if importDecl != nil && importDecl.ModuleSpecifier != nil {
				if strLit := importDecl.ModuleSpecifier.AsStringLiteral(); strLit != nil {
					moduleName := strLit.Text
					modules[moduleName] = true
				}
			}
		case ast.KindImportEqualsDeclaration:
			// TypeScript import: import foo = require('module-name')
			importEq := stmt.AsImportEqualsDeclaration()
			if importEq != nil && importEq.ModuleReference != nil {
				// Check if it's an external module reference
				if importEq.ModuleReference.Kind == ast.KindExternalModuleReference {
					extRef := importEq.ModuleReference.AsExternalModuleReference()
					if extRef != nil && extRef.Expression != nil {
						if strLit := extRef.Expression.AsStringLiteral(); strLit != nil {
							moduleName := strLit.Text
							modules[moduleName] = true
						}
					}
				}
			}
		}
	}

	return modules
}

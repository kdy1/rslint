package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/microsoft/typescript-go/shim/tspath"
	importPlugin "github.com/web-infra-dev/rslint/internal/plugins/import"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_self_import"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_webpack_loader_syntax"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/adjacent_overload_signatures"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/array_type"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/await_thenable"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/class_literal_property_style"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_array_delete"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_base_to_string"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_confusing_void_expression"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_duplicate_type_constituents"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_empty_function"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_empty_interface"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_explicit_any"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_floating_promises"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_for_in_array"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_implied_eval"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_meaningless_void_operator"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_misused_promises"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_misused_spread"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_mixed_enums"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_namespace"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_redundant_type_constituents"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_require_imports"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unnecessary_boolean_literal_compare"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unnecessary_template_expression"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unnecessary_type_arguments"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unnecessary_type_assertion"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_argument"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_assignment"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_call"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_enum_comparison"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_member_access"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_return"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_type_assertion"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unsafe_unary_minus"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_unused_vars"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_useless_empty_export"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/no_var_requires"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/non_nullable_type_assertion_style"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/only_throw_error"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_as_const"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_promise_reject_errors"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_reduce_type_parameter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_return_this_type"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/promise_function_async"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/related_getter_setter_pairs"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/require_array_sort_compare"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/require_await"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/restrict_plus_operands"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/restrict_template_expressions"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/return_await"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/switch_exhaustiveness_check"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/unbound_method"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/use_unknown_in_catch_callback_variable"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules/accessor_pairs"
	"github.com/web-infra-dev/rslint/internal/rules/block_scoped_var"
	"github.com/web-infra-dev/rslint/internal/rules/camelcase"
	"github.com/web-infra-dev/rslint/internal/rules/capitalized_comments"
	"github.com/web-infra-dev/rslint/internal/rules/class_methods_use_this"
	"github.com/web-infra-dev/rslint/internal/rules/complexity"
	"github.com/web-infra-dev/rslint/internal/rules/consistent_return"
	"github.com/web-infra-dev/rslint/internal/rules/consistent_this"
	"github.com/web-infra-dev/rslint/internal/rules/default_case"
	"github.com/web-infra-dev/rslint/internal/rules/default_case_last"
	"github.com/web-infra-dev/rslint/internal/rules/default_param_last"
	"github.com/web-infra-dev/rslint/internal/rules/dot_notation"
	"github.com/web-infra-dev/rslint/internal/rules/func_name_matching"
	"github.com/web-infra-dev/rslint/internal/rules/func_names"
	"github.com/web-infra-dev/rslint/internal/rules/func_style"
	"github.com/web-infra-dev/rslint/internal/rules/grouped_accessor_pairs"
	"github.com/web-infra-dev/rslint/internal/rules/guard_for_in"
	"github.com/web-infra-dev/rslint/internal/rules/id_denylist"
	"github.com/web-infra-dev/rslint/internal/rules/id_length"
	"github.com/web-infra-dev/rslint/internal/rules/id_match"
	"github.com/web-infra-dev/rslint/internal/rules/init_declarations"
	"github.com/web-infra-dev/rslint/internal/rules/logical_assignment_operators"
	"github.com/web-infra-dev/rslint/internal/rules/max_classes_per_file"
	"github.com/web-infra-dev/rslint/internal/rules/max_depth"
	"github.com/web-infra-dev/rslint/internal/rules/max_lines"
	"github.com/web-infra-dev/rslint/internal/rules/max_lines_per_function"
	"github.com/web-infra-dev/rslint/internal/rules/max_nested_callbacks"
	"github.com/web-infra-dev/rslint/internal/rules/max_params"
	"github.com/web-infra-dev/rslint/internal/rules/max_statements"
	"github.com/web-infra-dev/rslint/internal/rules/new_cap"
	"github.com/web-infra-dev/rslint/internal/rules/no_array_constructor"
	"github.com/web-infra-dev/rslint/internal/rules/no_bitwise"
	"github.com/web-infra-dev/rslint/internal/rules/no_caller"
	"github.com/web-infra-dev/rslint/internal/rules/no_case_declarations"
	"github.com/web-infra-dev/rslint/internal/rules/no_continue"
	"github.com/web-infra-dev/rslint/internal/rules/no_delete_var"
	"github.com/web-infra-dev/rslint/internal/rules/no_div_regex"
	"github.com/web-infra-dev/rslint/internal/rules/no_eq_null"
	"github.com/web-infra-dev/rslint/internal/rules/no_extend_native"
	"github.com/web-infra-dev/rslint/internal/rules/no_global_assign"
	"github.com/web-infra-dev/rslint/internal/rules/no_implicit_coercion"
	"github.com/web-infra-dev/rslint/internal/rules/no_implicit_globals"
	"github.com/web-infra-dev/rslint/internal/rules/no_inline_comments"
	"github.com/web-infra-dev/rslint/internal/rules/no_invalid_this"
	"github.com/web-infra-dev/rslint/internal/rules/no_iterator"
	"github.com/web-infra-dev/rslint/internal/rules/no_label_var"
	"github.com/web-infra-dev/rslint/internal/rules/no_labels"
	"github.com/web-infra-dev/rslint/internal/rules/no_lone_blocks"
	"github.com/web-infra-dev/rslint/internal/rules/no_magic_numbers"
	"github.com/web-infra-dev/rslint/internal/rules/no_multi_assign"
	"github.com/web-infra-dev/rslint/internal/rules/no_multi_str"
	"github.com/web-infra-dev/rslint/internal/rules/no_negated_condition"
	"github.com/web-infra-dev/rslint/internal/rules/no_new"
	"github.com/web-infra-dev/rslint/internal/rules/no_new_func"
	"github.com/web-infra-dev/rslint/internal/rules/no_new_wrappers"
	"github.com/web-infra-dev/rslint/internal/rules/no_nonoctal_decimal_escape"
	"github.com/web-infra-dev/rslint/internal/rules/no_object_constructor"
	"github.com/web-infra-dev/rslint/internal/rules/no_octal"
	"github.com/web-infra-dev/rslint/internal/rules/no_octal_escape"
	"github.com/web-infra-dev/rslint/internal/rules/no_param_reassign"
	"github.com/web-infra-dev/rslint/internal/rules/no_plusplus"
	"github.com/web-infra-dev/rslint/internal/rules/no_proto"
	"github.com/web-infra-dev/rslint/internal/rules/no_redeclare"
	"github.com/web-infra-dev/rslint/internal/rules/no_regex_spaces"
	"github.com/web-infra-dev/rslint/internal/rules/no_restricted_exports"
	"github.com/web-infra-dev/rslint/internal/rules/no_restricted_globals"
	"github.com/web-infra-dev/rslint/internal/rules/no_restricted_imports"
	"github.com/web-infra-dev/rslint/internal/rules/no_restricted_properties"
	"github.com/web-infra-dev/rslint/internal/rules/no_restricted_syntax"
	"github.com/web-infra-dev/rslint/internal/rules/no_script_url"
	"github.com/web-infra-dev/rslint/internal/rules/no_shadow_restricted_names"
	"github.com/web-infra-dev/rslint/internal/rules/no_ternary"
	"github.com/web-infra-dev/rslint/internal/rules/no_undef_init"
	"github.com/web-infra-dev/rslint/internal/rules/no_undefined"
	"github.com/web-infra-dev/rslint/internal/rules/no_underscore_dangle"
	"github.com/web-infra-dev/rslint/internal/rules/no_unneeded_ternary"
	"github.com/web-infra-dev/rslint/internal/rules/no_useless_computed_key"
	"github.com/web-infra-dev/rslint/internal/rules/no_void"
	"github.com/web-infra-dev/rslint/internal/rules/no_warning_comments"
	"github.com/web-infra-dev/rslint/internal/rules/one_var"
	"github.com/web-infra-dev/rslint/internal/rules/operator_assignment"
	"github.com/web-infra-dev/rslint/internal/rules/prefer_exponentiation_operator"
	"github.com/web-infra-dev/rslint/internal/rules/prefer_named_capture_group"
	"github.com/web-infra-dev/rslint/internal/rules/prefer_numeric_literals"
	"github.com/web-infra-dev/rslint/internal/rules/prefer_object_has_own"
	"github.com/web-infra-dev/rslint/internal/rules/prefer_object_spread"
	"github.com/web-infra-dev/rslint/internal/rules/prefer_regex_literals"
	"github.com/web-infra-dev/rslint/internal/rules/preserve_caught_error"
	"github.com/web-infra-dev/rslint/internal/rules/radix"
	"github.com/web-infra-dev/rslint/internal/rules/require_unicode_regexp"
	"github.com/web-infra-dev/rslint/internal/rules/require_yield"
	"github.com/web-infra-dev/rslint/internal/rules/sort_imports"
	"github.com/web-infra-dev/rslint/internal/rules/sort_keys"
	"github.com/web-infra-dev/rslint/internal/rules/sort_vars"
	"github.com/web-infra-dev/rslint/internal/rules/strict"
	"github.com/web-infra-dev/rslint/internal/rules/symbol_description"
	"github.com/web-infra-dev/rslint/internal/rules/vars_on_top"
)

// RslintConfig represents the top-level configuration array
type RslintConfig []ConfigEntry

// ConfigEntry represents a single configuration entry in the rslint.json array
type ConfigEntry struct {
	Language        string           `json:"language"`
	Files           []string         `json:"files"`
	Ignores         []string         `json:"ignores,omitempty"` // List of file patterns to ignore
	LanguageOptions *LanguageOptions `json:"languageOptions,omitempty"`
	Rules           Rules            `json:"rules"`
	Plugins         []string         `json:"plugins,omitempty"` // List of plugin names
}

// LanguageOptions contains language-specific configuration options
type LanguageOptions struct {
	ParserOptions *ParserOptions `json:"parserOptions,omitempty"`
}

// ProjectPaths represents project paths that can be either a single string or an array of strings
type ProjectPaths []string

// UnmarshalJSON implements custom JSON unmarshaling to support both string and string[] formats
func (p *ProjectPaths) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var singlePath string
	if err := json.Unmarshal(data, &singlePath); err == nil {
		*p = []string{singlePath}
		return nil
	}

	// If that fails, try to unmarshal as array of strings
	var paths []string
	if err := json.Unmarshal(data, &paths); err != nil {
		return err
	}
	*p = paths
	return nil
}

// ParserOptions contains parser-specific configuration
type ParserOptions struct {
	ProjectService bool         `json:"projectService"`
	Project        ProjectPaths `json:"project,omitempty"`
}

// Rules represents the rules configuration
// This can be extended to include specific rule configurations
type Rules map[string]interface{}

// Alternative: If you want type-safe rule configurations
type TypedRules struct {
	// Example rule configurations - extend as needed
	AdjacentOverloadSignatures         *RuleConfig `json:"@typescript-eslint/adjacent-overload-signatures,omitempty"`
	ArrayType                          *RuleConfig `json:"@typescript-eslint/array-type,omitempty"`
	ClassLiteralPropertyStyle          *RuleConfig `json:"@typescript-eslint/class-literal-property-style,omitempty"`
	NoArrayDelete                      *RuleConfig `json:"@typescript-eslint/no-array-delete,omitempty"`
	NoBaseToString                     *RuleConfig `json:"@typescript-eslint/no-base-to-string,omitempty"`
	NoForInArray                       *RuleConfig `json:"@typescript-eslint/no-for-in-array,omitempty"`
	NoImpliedEval                      *RuleConfig `json:"@typescript-eslint/no-implied-eval,omitempty"`
	OnlyThrowError                     *RuleConfig `json:"@typescript-eslint/only-throw-error,omitempty"`
	AwaitThenable                      *RuleConfig `json:"@typescript-eslint/await-thenable,omitempty"`
	NoConfusingVoidExpression          *RuleConfig `json:"@typescript-eslint/no-confusing-void-expression,omitempty"`
	NoDuplicateTypeConstituents        *RuleConfig `json:"@typescript-eslint/no-duplicate-type-constituents,omitempty"`
	NoFloatingPromises                 *RuleConfig `json:"@typescript-eslint/no-floating-promises,omitempty"`
	NoMeaninglessVoidOperator          *RuleConfig `json:"@typescript-eslint/no-meaningless-void-operator,omitempty"`
	NoMisusedPromises                  *RuleConfig `json:"@typescript-eslint/no-misused-promises,omitempty"`
	NoMisusedSpread                    *RuleConfig `json:"@typescript-eslint/no-misused-spread,omitempty"`
	NoMixedEnums                       *RuleConfig `json:"@typescript-eslint/no-mixed-enums,omitempty"`
	NoRedundantTypeConstituents        *RuleConfig `json:"@typescript-eslint/no-redundant-type-constituents,omitempty"`
	NoUnnecessaryBooleanLiteralCompare *RuleConfig `json:"@typescript-eslint/no-unnecessary-boolean-literal-compare,omitempty"`
	NoUnnecessaryTemplateExpression    *RuleConfig `json:"@typescript-eslint/no-unnecessary-template-expression,omitempty"`
	NoUnnecessaryTypeArguments         *RuleConfig `json:"@typescript-eslint/no-unnecessary-type-arguments,omitempty"`
	NoUnnecessaryTypeAssertion         *RuleConfig `json:"@typescript-eslint/no-unnecessary-type-assertion,omitempty"`
	NoUnsafeArgument                   *RuleConfig `json:"@typescript-eslint/no-unsafe-argument,omitempty"`
	NoUnsafeAssignment                 *RuleConfig `json:"@typescript-eslint/no-unsafe-assignment,omitempty"`
	NoUnsafeCall                       *RuleConfig `json:"@typescript-eslint/no-unsafe-call,omitempty"`
	NoUnsafeEnumComparison             *RuleConfig `json:"@typescript-eslint/no-unsafe-enum-comparison,omitempty"`
	NoUnsafeMemberAccess               *RuleConfig `json:"@typescript-eslint/no-unsafe-member-access,omitempty"`
	NoUnsafeReturn                     *RuleConfig `json:"@typescript-eslint/no-unsafe-return,omitempty"`
	NoUnsafeTypeAssertion              *RuleConfig `json:"@typescript-eslint/no-unsafe-type-assertion,omitempty"`
	NoUnsafeUnaryMinus                 *RuleConfig `json:"@typescript-eslint/no-unsafe-unary-minus,omitempty"`
}

// RuleConfig represents individual rule configuration
type RuleConfig struct {
	Level   string                 `json:"level,omitempty"`   // "error", "warn", "off"
	Options map[string]interface{} `json:"options,omitempty"` // Rule-specific options
}

const defaultJsonc = `
[
  {
    // ignore files and folders for linting
    "ignores": [],
    "languageOptions": {
      "parserOptions": {
        // Rslint will lint all files included in your typescript projects defined here
        // support lint multi packages in monorepo
        "project": ["./tsconfig.json"]
      }
    },
    // same configuration as https://typescript-eslint.io/rules/
    "rules": {
      "@typescript-eslint/require-await": "off",
      "@typescript-eslint/no-unnecessary-type-assertion": "warn",
      "@typescript-eslint/array-type": ["warn", { "default": "array-simple" }]
    },
    "plugins": [
      "@typescript-eslint" // will enable all implemented @typescript-eslint rules by default
    ]
  }
]
`

// IsEnabled returns true if the rule is enabled (not "off")
func (rc *RuleConfig) IsEnabled() bool {
	if rc == nil {
		return false
	}
	return rc.Level != "off" && rc.Level != ""
}

// GetLevel returns the rule level, defaulting to "error" if not specified
func (rc *RuleConfig) GetLevel() string {
	if rc == nil || rc.Level == "" {
		return "error"
	}
	return rc.Level
}

// GetOptions returns the rule options, ensuring we return a usable value
func (rc *RuleConfig) GetOptions() map[string]interface{} {
	if rc == nil || rc.Options == nil {
		return make(map[string]interface{})
	}
	return rc.Options
}

// SetOptions sets the rule options
func (rc *RuleConfig) SetOptions(options map[string]interface{}) {
	if rc != nil {
		rc.Options = options
	}
}

// GetSeverity returns the diagnostic severity for this rule configuration
func (rc *RuleConfig) GetSeverity() rule.DiagnosticSeverity {
	if rc == nil {
		return rule.SeverityError
	}
	return rule.ParseSeverity(rc.Level)
}
func GetAllRulesForPlugin(plugin string) []rule.Rule {
	switch plugin {
	case "@typescript-eslint":
		return getAllTypeScriptEslintPluginRules()
	case "eslint-plugin-import":
		return importPlugin.GetAllRules()
	case "eslint-plugin-import/recommended":
		return importPlugin.GetRecommendedRules()
	default:
		return []rule.Rule{} // Return empty slice for unsupported plugins
	}
}

// parseArrayRuleConfig parses array-style rule configuration like ["error", {...options}]
// Supports ESLint-compatible formats:
// - ["off"] -> disabled rule
// - ["error"] -> enabled rule with error severity
// - ["warn"] -> enabled rule with warning severity
// - ["error", {...options}] -> enabled rule with error severity and options
// - ["warn", {...options}] -> enabled rule with warning severity and options
func parseArrayRuleConfig(ruleArray []interface{}) *RuleConfig {
	if len(ruleArray) == 0 {
		return nil
	}

	// First element should always be the severity level
	level, ok := ruleArray[0].(string)
	if !ok {
		return nil
	}

	ruleConfig := &RuleConfig{Level: level}

	// Second element (if present) should be the options object
	if len(ruleArray) > 1 {
		switch opts := ruleArray[1].(type) {
		case map[string]interface{}:
			ruleConfig.Options = opts
		case nil:
			// Explicitly null/nil options are valid
			ruleConfig.Options = make(map[string]interface{})
		default:
			// Invalid options type, but still create the rule config with just the level
			ruleConfig.Options = make(map[string]interface{})
		}
	}

	// Additional elements are ignored (following ESLint behavior)
	return ruleConfig
}

// GetRulesForFile returns enabled rules for a given file based on the configuration
func (config RslintConfig) GetRulesForFile(filePath string) map[string]*RuleConfig {
	enabledRules := make(map[string]*RuleConfig)

	for _, entry := range config {
		// First check if the file should be ignored
		if isFileIgnored(filePath, entry.Ignores) {
			continue // Skip this config entry for ignored files
		}

		// Check if the file matches the files pattern
		matches := true

		if matches {

			/// Merge rules from plugin
			for _, plugin := range entry.Plugins {

				for _, rule := range GetAllRulesForPlugin(plugin) {
					enabledRules[rule.Name] = &RuleConfig{Level: "error"} // Default level for plugin rules
				}
			}
			// Merge rules from this entry
			for ruleName, ruleValue := range entry.Rules {

				switch v := ruleValue.(type) {
				case string:
					// Handle simple string values like "error", "warn", "off"
					enabledRules[ruleName] = &RuleConfig{Level: v}
				case map[string]interface{}:
					// Handle object configuration
					ruleConfig := &RuleConfig{}
					if level, ok := v["level"].(string); ok {
						ruleConfig.Level = level
					}
					if options, ok := v["options"].(map[string]interface{}); ok {
						ruleConfig.Options = options
					}
					if ruleConfig.IsEnabled() {
						enabledRules[ruleName] = ruleConfig
					}
				case []interface{}:
					// Handle array format like ["error", {...options}] or ["warn"] or ["off"]
					ruleConfig := parseArrayRuleConfig(v)
					if ruleConfig != nil && ruleConfig.IsEnabled() {
						enabledRules[ruleName] = ruleConfig
					}
				}
			}
		}
	}
	return enabledRules
}

func RegisterAllRules() {
	registerAllTypeScriptEslintPluginRules()
	registerAllEslintImportPluginRules()
	GlobalRuleRegistry.Register("accessor-pairs", accessor_pairs.AccessorPairsRule)
	GlobalRuleRegistry.Register("block-scoped-var", block_scoped_var.BlockScopedVarRule)
	GlobalRuleRegistry.Register("camelcase", camelcase.CamelcaseRule)
	GlobalRuleRegistry.Register("capitalized-comments", capitalized_comments.CapitalizedCommentsRule)
	GlobalRuleRegistry.Register("class-methods-use-this", class_methods_use_this.ClassMethodsUseThisRule)
	GlobalRuleRegistry.Register("complexity", complexity.ComplexityRule)
	GlobalRuleRegistry.Register("consistent-return", consistent_return.ConsistentReturnRule)
	GlobalRuleRegistry.Register("consistent-this", consistent_this.ConsistentThisRule)
	GlobalRuleRegistry.Register("default-case", default_case.DefaultCaseRule)
	GlobalRuleRegistry.Register("default-case-last", default_case_last.DefaultCaseLastRule)
	GlobalRuleRegistry.Register("default-param-last", default_param_last.DefaultParamLastRule)
	GlobalRuleRegistry.Register("func-name-matching", func_name_matching.FuncNameMatchingRule)
	GlobalRuleRegistry.Register("func-names", func_names.FuncNamesRule)
	GlobalRuleRegistry.Register("func-style", func_style.FuncStyleRule)
	GlobalRuleRegistry.Register("grouped-accessor-pairs", grouped_accessor_pairs.GroupedAccessorPairsRule)
	GlobalRuleRegistry.Register("guard-for-in", guard_for_in.GuardForInRule)
	GlobalRuleRegistry.Register("id-denylist", id_denylist.IdDenylistRule)
	GlobalRuleRegistry.Register("id-length", id_length.IdLengthRule)
	GlobalRuleRegistry.Register("id-match", id_match.IdMatchRule)
	GlobalRuleRegistry.Register("init-declarations", init_declarations.InitDeclarationsRule)
	GlobalRuleRegistry.Register("logical-assignment-operators", logical_assignment_operators.LogicalAssignmentOperatorsRule)
	GlobalRuleRegistry.Register("max-classes-per-file", max_classes_per_file.MaxClassesPerFileRule)
	GlobalRuleRegistry.Register("max-depth", max_depth.MaxDepthRule)
	GlobalRuleRegistry.Register("max-lines", max_lines.MaxLinesRule)
	GlobalRuleRegistry.Register("max-lines-per-function", max_lines_per_function.MaxLinesPerFunctionRule)
	GlobalRuleRegistry.Register("max-nested-callbacks", max_nested_callbacks.MaxNestedCallbacksRule)
	GlobalRuleRegistry.Register("max-params", max_params.MaxParamsRule)
	GlobalRuleRegistry.Register("max-statements", max_statements.MaxStatementsRule)
	GlobalRuleRegistry.Register("new-cap", new_cap.NewCapRule)
	GlobalRuleRegistry.Register("no-array-constructor", no_array_constructor.NoArrayConstructorRule)
	GlobalRuleRegistry.Register("no-bitwise", no_bitwise.NoBitwiseRule)
	GlobalRuleRegistry.Register("no-caller", no_caller.NoCallerRule)
	GlobalRuleRegistry.Register("no-case-declarations", no_case_declarations.NoCaseDeclarationsRule)
	GlobalRuleRegistry.Register("no-continue", no_continue.NoContinueRule)
	GlobalRuleRegistry.Register("no-delete-var", no_delete_var.NoDeleteVarRule)
	GlobalRuleRegistry.Register("no-div-regex", no_div_regex.NoDivRegexRule)
	GlobalRuleRegistry.Register("no-eq-null", no_eq_null.NoEqNullRule)
	GlobalRuleRegistry.Register("no-extend-native", no_extend_native.NoExtendNativeRule)
	GlobalRuleRegistry.Register("no-global-assign", no_global_assign.NoGlobalAssignRule)
	GlobalRuleRegistry.Register("no-implicit-coercion", no_implicit_coercion.NoImplicitCoercionRule)
	GlobalRuleRegistry.Register("no-implicit-globals", no_implicit_globals.NoImplicitGlobalsRule)
	GlobalRuleRegistry.Register("no-inline-comments", no_inline_comments.NoInlineCommentsRule)
	GlobalRuleRegistry.Register("no-invalid-this", no_invalid_this.NoInvalidThisRule)
	GlobalRuleRegistry.Register("no-iterator", no_iterator.NoIteratorRule)
	GlobalRuleRegistry.Register("no-label-var", no_label_var.NoLabelVarRule)
	GlobalRuleRegistry.Register("no-labels", no_labels.NoLabelsRule)
	GlobalRuleRegistry.Register("no-lone-blocks", no_lone_blocks.NoLoneBlocksRule)
	GlobalRuleRegistry.Register("no-magic-numbers", no_magic_numbers.NoMagicNumbersRule)
	GlobalRuleRegistry.Register("no-multi-assign", no_multi_assign.NoMultiAssignRule)
	GlobalRuleRegistry.Register("no-multi-str", no_multi_str.NoMultiStrRule)
	GlobalRuleRegistry.Register("no-negated-condition", no_negated_condition.NoNegatedConditionRule)
	GlobalRuleRegistry.Register("no-new", no_new.NoNewRule)
	GlobalRuleRegistry.Register("no-new-func", no_new_func.NoNewFuncRule)
	GlobalRuleRegistry.Register("no-new-wrappers", no_new_wrappers.NoNewWrappersRule)
	GlobalRuleRegistry.Register("no-nonoctal-decimal-escape", no_nonoctal_decimal_escape.NoNonoctalDecimalEscapeRule)
	GlobalRuleRegistry.Register("no-object-constructor", no_object_constructor.NoObjectConstructorRule)
	GlobalRuleRegistry.Register("no-octal", no_octal.NoOctalRule)
	GlobalRuleRegistry.Register("no-octal-escape", no_octal_escape.NoOctalEscapeRule)
	GlobalRuleRegistry.Register("no-param-reassign", no_param_reassign.NoParamReassignRule)
	GlobalRuleRegistry.Register("no-plusplus", no_plusplus.NoPlusplusRule)
	GlobalRuleRegistry.Register("no-proto", no_proto.NoProtoRule)
	GlobalRuleRegistry.Register("no-redeclare", no_redeclare.NoRedeclareRule)
	GlobalRuleRegistry.Register("no-regex-spaces", no_regex_spaces.NoRegexSpacesRule)
	GlobalRuleRegistry.Register("no-restricted-exports", no_restricted_exports.NoRestrictedExportsRule)
	GlobalRuleRegistry.Register("no-restricted-globals", no_restricted_globals.NoRestrictedGlobalsRule)
	GlobalRuleRegistry.Register("no-restricted-imports", no_restricted_imports.NoRestrictedImportsRule)
	GlobalRuleRegistry.Register("no-restricted-properties", no_restricted_properties.NoRestrictedPropertiesRule)
	GlobalRuleRegistry.Register("no-restricted-syntax", no_restricted_syntax.NoRestrictedSyntaxRule)
	GlobalRuleRegistry.Register("no-script-url", no_script_url.NoScriptUrlRule)
	GlobalRuleRegistry.Register("no-shadow-restricted-names", no_shadow_restricted_names.NoShadowRestrictedNamesRule)
	GlobalRuleRegistry.Register("no-ternary", no_ternary.NoTernaryRule)
	GlobalRuleRegistry.Register("no-undef-init", no_undef_init.NoUndefInitRule)
	GlobalRuleRegistry.Register("no-undefined", no_undefined.NoUndefinedRule)
	GlobalRuleRegistry.Register("no-underscore-dangle", no_underscore_dangle.NoUnderscoreDangleRule)
	GlobalRuleRegistry.Register("no-unneeded-ternary", no_unneeded_ternary.NoUnneededTernaryRule)
	GlobalRuleRegistry.Register("no-useless-computed-key", no_useless_computed_key.NoUselessComputedKeyRule)
	GlobalRuleRegistry.Register("no-void", no_void.NoVoidRule)
	GlobalRuleRegistry.Register("no-warning-comments", no_warning_comments.NoWarningCommentsRule)
	GlobalRuleRegistry.Register("one-var", one_var.OneVarRule)
	GlobalRuleRegistry.Register("operator-assignment", operator_assignment.OperatorAssignmentRule)
	GlobalRuleRegistry.Register("prefer-exponentiation-operator", prefer_exponentiation_operator.PreferExponentiationOperatorRule)
	GlobalRuleRegistry.Register("prefer-named-capture-group", prefer_named_capture_group.PreferNamedCaptureGroupRule)
	GlobalRuleRegistry.Register("prefer-numeric-literals", prefer_numeric_literals.PreferNumericLiteralsRule)
	GlobalRuleRegistry.Register("prefer-object-has-own", prefer_object_has_own.PreferObjectHasOwnRule)
	GlobalRuleRegistry.Register("prefer-object-spread", prefer_object_spread.PreferObjectSpreadRule)
	GlobalRuleRegistry.Register("prefer-regex-literals", prefer_regex_literals.PreferRegexLiteralsRule)
	GlobalRuleRegistry.Register("preserve-caught-error", preserve_caught_error.PreserveCaughtErrorRule)
	GlobalRuleRegistry.Register("radix", radix.RadixRule)
	GlobalRuleRegistry.Register("require-unicode-regexp", require_unicode_regexp.RequireUnicodeRegexpRule)
	GlobalRuleRegistry.Register("require-yield", require_yield.RequireYieldRule)
	GlobalRuleRegistry.Register("sort-imports", sort_imports.SortImportsRule)
	GlobalRuleRegistry.Register("sort-keys", sort_keys.SortKeysRule)
	GlobalRuleRegistry.Register("sort-vars", sort_vars.SortVarsRule)
	GlobalRuleRegistry.Register("strict", strict.StrictRule)
	GlobalRuleRegistry.Register("symbol-description", symbol_description.SymbolDescriptionRule)
	GlobalRuleRegistry.Register("vars-on-top", vars_on_top.VarsOnTopRule)
}

// registerAllTypeScriptEslintPluginRules registers all available rules in the global registry
func registerAllTypeScriptEslintPluginRules() {
	GlobalRuleRegistry.Register("@typescript-eslint/adjacent-overload-signatures", adjacent_overload_signatures.AdjacentOverloadSignaturesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/array-type", array_type.ArrayTypeRule)
	GlobalRuleRegistry.Register("@typescript-eslint/await-thenable", await_thenable.AwaitThenableRule)
	GlobalRuleRegistry.Register("@typescript-eslint/class-literal-property-style", class_literal_property_style.ClassLiteralPropertyStyleRule)
	GlobalRuleRegistry.Register("@typescript-eslint/dot-notation", dot_notation.DotNotationRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-array-delete", no_array_delete.NoArrayDeleteRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-base-to-string", no_base_to_string.NoBaseToStringRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-confusing-void-expression", no_confusing_void_expression.NoConfusingVoidExpressionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-duplicate-type-constituents", no_duplicate_type_constituents.NoDuplicateTypeConstituentsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-explicit-any", no_explicit_any.NoExplicitAnyRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-empty-function", no_empty_function.NoEmptyFunctionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-empty-interface", no_empty_interface.NoEmptyInterfaceRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-floating-promises", no_floating_promises.NoFloatingPromisesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-for-in-array", no_for_in_array.NoForInArrayRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-implied-eval", no_implied_eval.NoImpliedEvalRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-meaningless-void-operator", no_meaningless_void_operator.NoMeaninglessVoidOperatorRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-misused-promises", no_misused_promises.NoMisusedPromisesRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-misused-spread", no_misused_spread.NoMisusedSpreadRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-mixed-enums", no_mixed_enums.NoMixedEnumsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-namespace", no_namespace.NoNamespaceRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-redundant-type-constituents", no_redundant_type_constituents.NoRedundantTypeConstituentsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-require-imports", no_require_imports.NoRequireImportsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-boolean-literal-compare", no_unnecessary_boolean_literal_compare.NoUnnecessaryBooleanLiteralCompareRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-template-expression", no_unnecessary_template_expression.NoUnnecessaryTemplateExpressionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-type-arguments", no_unnecessary_type_arguments.NoUnnecessaryTypeArgumentsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unnecessary-type-assertion", no_unnecessary_type_assertion.NoUnnecessaryTypeAssertionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-argument", no_unsafe_argument.NoUnsafeArgumentRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-assignment", no_unsafe_assignment.NoUnsafeAssignmentRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-call", no_unsafe_call.NoUnsafeCallRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-enum-comparison", no_unsafe_enum_comparison.NoUnsafeEnumComparisonRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-member-access", no_unsafe_member_access.NoUnsafeMemberAccessRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-return", no_unsafe_return.NoUnsafeReturnRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-type-assertion", no_unsafe_type_assertion.NoUnsafeTypeAssertionRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unsafe-unary-minus", no_unsafe_unary_minus.NoUnsafeUnaryMinusRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-unused-vars", no_unused_vars.NoUnusedVarsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-useless-empty-export", no_useless_empty_export.NoUselessEmptyExportRule)
	GlobalRuleRegistry.Register("@typescript-eslint/no-var-requires", no_var_requires.NoVarRequiresRule)
	GlobalRuleRegistry.Register("@typescript-eslint/non-nullable-type-assertion-style", non_nullable_type_assertion_style.NonNullableTypeAssertionStyleRule)
	GlobalRuleRegistry.Register("@typescript-eslint/only-throw-error", only_throw_error.OnlyThrowErrorRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-as-const", prefer_as_const.PreferAsConstRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-promise-reject-errors", prefer_promise_reject_errors.PreferPromiseRejectErrorsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-reduce-type-parameter", prefer_reduce_type_parameter.PreferReduceTypeParameterRule)
	GlobalRuleRegistry.Register("@typescript-eslint/prefer-return-this-type", prefer_return_this_type.PreferReturnThisTypeRule)
	GlobalRuleRegistry.Register("@typescript-eslint/promise-function-async", promise_function_async.PromiseFunctionAsyncRule)
	GlobalRuleRegistry.Register("@typescript-eslint/related-getter-setter-pairs", related_getter_setter_pairs.RelatedGetterSetterPairsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/require-array-sort-compare", require_array_sort_compare.RequireArraySortCompareRule)
	GlobalRuleRegistry.Register("@typescript-eslint/require-await", require_await.RequireAwaitRule)
	GlobalRuleRegistry.Register("@typescript-eslint/restrict-plus-operands", restrict_plus_operands.RestrictPlusOperandsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/restrict-template-expressions", restrict_template_expressions.RestrictTemplateExpressionsRule)
	GlobalRuleRegistry.Register("@typescript-eslint/return-await", return_await.ReturnAwaitRule)
	GlobalRuleRegistry.Register("@typescript-eslint/switch-exhaustiveness-check", switch_exhaustiveness_check.SwitchExhaustivenessCheckRule)
	GlobalRuleRegistry.Register("@typescript-eslint/unbound-method", unbound_method.UnboundMethodRule)
	GlobalRuleRegistry.Register("@typescript-eslint/use-unknown-in-catch-callback-variable", use_unknown_in_catch_callback_variable.UseUnknownInCatchCallbackVariableRule)
}

func registerAllEslintImportPluginRules() {
	for _, rule := range importPlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
		GlobalRuleRegistry.Register("import/no-self-import", no_self_import.NoSelfImportRule)
		GlobalRuleRegistry.Register("import/no-webpack-loader-syntax", no_webpack_loader_syntax.NoWebpackLoaderSyntax)
	}
}

// getAllTypeScriptEslintPluginRules returns all registered rules (for backward compatibility when no config is provided)
func getAllTypeScriptEslintPluginRules() []rule.Rule {
	allRules := GlobalRuleRegistry.GetAllRules()
	var rules []rule.Rule
	for _, rule := range allRules {
		rules = append(rules, rule)
	}
	return rules
}

// isFileIgnored checks if a file should be ignored based on ignore patterns
func isFileIgnored(filePath string, ignorePatterns []string) bool {
	// Get current working directory for relative path resolution
	cwd, err := os.Getwd()
	if err != nil {
		// If we can't get cwd, fall back to simple matching
		return isFileIgnoredSimple(filePath, ignorePatterns)
	}

	// Normalize the file path relative to cwd
	normalizedPath := normalizePath(filePath, cwd)

	for _, pattern := range ignorePatterns {
		// Try matching against normalized path
		if matched, err := doublestar.Match(pattern, normalizedPath); err == nil && matched {
			return true
		}

		// Also try matching against original path for absolute patterns
		if normalizedPath != filePath {
			if matched, err := doublestar.Match(pattern, filePath); err == nil && matched {
				return true
			}
		}

		// Try Unix-style path for cross-platform compatibility
		unixPath := strings.ReplaceAll(normalizedPath, "\\", "/")
		if unixPath != normalizedPath {
			if matched, err := doublestar.Match(pattern, unixPath); err == nil && matched {
				return true
			}
		}
	}
	return false
}

// normalizePath converts file path to be relative to cwd for consistent matching
func normalizePath(filePath, cwd string) string {
	return tspath.NormalizePath(tspath.ConvertToRelativePath(filePath, tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: true,
		CurrentDirectory:          cwd,
	}))
}

// isFileIgnoredSimple provides fallback matching when cwd is unavailable
func isFileIgnoredSimple(filePath string, ignorePatterns []string) bool {
	for _, pattern := range ignorePatterns {
		if matched, err := doublestar.Match(pattern, filePath); err == nil && matched {
			return true
		}
	}
	return false
}

// initialize a default config in the directory
func InitDefaultConfig(directory string) error {
	configPath := filepath.Join(directory, "rslint.jsonc")

	// if the config exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("rslint.json already exists in %s", directory)
	}

	// write file content
	err := os.WriteFile(configPath, []byte(defaultJsonc), 0644)
	if err != nil {
		return fmt.Errorf("failed to create rslint.json: %w", err)
	}

	return nil
}

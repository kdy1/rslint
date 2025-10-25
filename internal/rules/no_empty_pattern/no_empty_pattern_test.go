package no_empty_pattern

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoEmptyPatternRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyPatternRule,
		[]rule_tester.ValidTestCase{
			// Object with default value
			{Code: `var {a = {}} = foo;`},

			// Multiple properties with default
			{Code: `var {a, b = {}} = foo;`},

			// Array default value
			{Code: `var {a = []} = foo;`},

			// Function parameter with default
			{Code: `function foo({a = {}}) {}`},

			// Function parameter with array default
			{Code: `function foo({a = []}) {}`},

			// Array destructuring with element
			{Code: `var [a] = foo`},

			// Empty object parameter with option
			{
				Code:    `function foo({}) {}`,
				Options: map[string]interface{}{"allowObjectPatternsAsParameters": true},
			},

			// Function expression with empty object and option
			{
				Code:    `var foo = function({}) {}`,
				Options: map[string]interface{}{"allowObjectPatternsAsParameters": true},
			},

			// Arrow function with empty object and option
			{
				Code:    `var foo = ({}) => {}`,
				Options: map[string]interface{}{"allowObjectPatternsAsParameters": true},
			},

			// Empty object with default value
			{Code: `function foo({} = {}) {}`},

			// Function expression with empty object default
			{Code: `var foo = function({} = {}) {}`},

			// Arrow function with empty object default
			{Code: `var foo = ({} = {}) => {}`},
		},
		[]rule_tester.InvalidTestCase{
			// Empty object pattern
			{
				Code: `var {} = foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Empty array pattern
			{
				Code: `var [] = foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Nested empty object pattern
			{
				Code: `var {a: {}} = foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Multiple properties with nested empty object
			{
				Code: `var {a, b: {}} = foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Nested empty array pattern
			{
				Code: `var {a: []} = foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Function with empty object parameter
			{
				Code: `function foo({}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Function with empty array parameter
			{
				Code: `function foo([]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Function with nested empty object parameter
			{
				Code: `function foo({a: {}}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Function with nested empty array parameter
			{
				Code: `function foo({a: []}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Arrow function with nested empty object (with option)
			{
				Code:    `var foo = ({a: {}}) => {}`,
				Options: map[string]interface{}{"allowObjectPatternsAsParameters": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Arrow function with empty object and default (with option)
			{
				Code:    `var foo = ({} = bar) => {}`,
				Options: map[string]interface{}{"allowObjectPatternsAsParameters": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Arrow function with empty object and complex default (with option)
			{
				Code:    `var foo = ({} = { bar: 1 }) => {}`,
				Options: map[string]interface{}{"allowObjectPatternsAsParameters": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Arrow function with empty array (with option)
			{
				Code:    `var foo = ([]) => {}`,
				Options: map[string]interface{}{"allowObjectPatternsAsParameters": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
		},
	)
}

package no_unexpected_multiline

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoUnexpectedMultilineRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnexpectedMultilineRule,
		[]rule_tester.ValidTestCase{
			// Basic valid cases
			{Code: `(x || y).aFunction()`},
			{Code: `[a, b, c].forEach(doSomething)`},
			{Code: `var a = b;
(x || y).doSomething()`},
			{Code: `var a = b
;(x || y).doSomething()`},
			{Code: `var a = b
void (x || y).doSomething()`},
			{Code: `var a = b;
[1, 2, 3].forEach(console.log)`},
			{Code: `var a = b
void [1, 2, 3].forEach(console.log)`},
			{Code: `var a = (
(123)
)`},
			{Code: `f(
(x)
)`},
			{Code: `(
function () {}
)[1]`},

			// Template literal cases
			{Code: `let x = function() {};
   \`hello\``},
			{Code: `let x = function() {}
x \`hello\``},
			{Code: `String.raw \`Hi
${2+3}!\`;`},
			{Code: `x
.y
z \`Valid Test Case\``},

			// Optional chaining cases
			{Code: `var a = b
  ?.(x || y).doSomething()`},
			{Code: `var a = b
  ?.[a, b, c].forEach(doSomething)`},

			// Class field cases
			{Code: `class C { field1
[field2]; }`},
			{Code: `class C { field1
*gen() {} }`},
			{Code: `class C { field1 = () => {}
[field2]; }`},
			{Code: `class C { field1 = () => {}
*gen() {} }`},
		},
		[]rule_tester.InvalidTestCase{
			// Function call errors
			{
				Code: `var a = b
(x || y).doSomething()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 2, Column: 1},
				},
			},
			{
				Code: `var a = (a || b)
(x || y).doSomething()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 2, Column: 1},
				},
			},
			{
				Code: `var a = (a || b)
(x).doSomething()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 2, Column: 1},
				},
			},

			// Property access errors
			{
				Code: `var a = b
[a, b, c].forEach(doSomething)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "property", Line: 2, Column: 1},
				},
			},
			{
				Code: `var a = b
    (x || y).doSomething()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 2, Column: 5},
				},
			},
			{
				Code: `var a = b
  [a, b, c].forEach(doSomething)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "property", Line: 2, Column: 3},
				},
			},

			// Tagged template errors
			{
				Code: `let x = function() {}
 \`hello\``,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "taggedTemplate", Line: 2, Column: 2},
				},
			},
			{
				Code: `let x = function() {}
x
\`hello\``,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "taggedTemplate", Line: 3, Column: 1},
				},
			},
			{
				Code: `x
.y
z
\`Invalid Test Case\``,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "taggedTemplate", Line: 4, Column: 1},
				},
			},

			// Class field errors
			{
				Code: `class C { field1 = obj
[field2]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "property", Line: 2, Column: 1},
				},
			},
			{
				Code: `class C { field1 = function() {}
[field2]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "property", Line: 2, Column: 1},
				},
			},
		},
	)
}

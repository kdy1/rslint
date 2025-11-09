package no_unnecessary_type_constraint

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryTypeConstraintRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryTypeConstraintRule,
		[]rule_tester.ValidTestCase{
			{Code: `function data() {}`},
			{Code: `function data<T>() {}`},
			{Code: `function data<T, U>() {}`},
			{Code: `function data<T extends number>() {}`},
			{Code: `function data<T extends number | string>() {}`},
			{Code: `function data<T extends any | number>() {}`},
			{Code: `type TODO = any; function data<T extends TODO>() {}`},
			{Code: `const data = () => {};`},
			{Code: `const data = <T>() => {};`},
			{Code: `const data = <T, U>() => {};`},
			{Code: `const data = <T extends number>() => {};`},
			{Code: `const data = <T extends number | string>() => {};`},
			{Code: `const data = <T extends any | number>() => {};`},
			{Code: `type TODO = any; const data = <T extends TODO>() => {};`},
			{Code: `class Data<T> {}`},
			{Code: `class Data<T, U> {}`},
			{Code: `class Data<T extends number> {}`},
			{Code: `interface Data<T> {}`},
			{Code: `interface Data<T, U> {}`},
			{Code: `interface Data<T extends number> {}`},
			{Code: `type Data<T> = {};`},
			{Code: `type Data<T, U> = {};`},
			{Code: `type Data<T extends number> = {};`},
		},
		[]rule_tester.InvalidTestCase{
			// Function declarations with 'any' constraint
			{
				Code: `function data<T extends any>() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    24,
					},
				},
				Output: `function data<T>() {}`,
			},
			{
				Code: `function data<T extends any, U>() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    24,
					},
				},
				Output: `function data<T, U>() {}`,
			},
			{
				Code: `function data<T, U extends any>() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    27,
					},
				},
				Output: `function data<T, U>() {}`,
			},
			{
				Code: `function data<T extends any, U extends T>() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    24,
					},
				},
				Output: `function data<T, U extends T>() {}`,
			},
			// Arrow functions with 'any' constraint
			{
				Code: `const data = <T extends any>() => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    24,
					},
				},
				Output: `const data = <T>() => {};`,
			},
			{
				Code: `const data = <T extends any,>() => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    24,
					},
				},
				Output: `const data = <T>() => {};`,
			},
			{
				Code: `const data = <T extends any, >() => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    24,
					},
				},
				Output: `const data = <T>() => {};`,
			},
			{
				Code: `const data = <T extends any = unknown>() => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    24,
					},
				},
				Output: `const data = <T = unknown>() => {};`,
			},
			{
				Code: `const data = <T extends any, U extends any>() => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    24,
					},
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    39,
					},
				},
				Output: `const data = <T, U>() => {};`,
			},
			// Function declarations with 'unknown' constraint
			{
				Code: `function data<T extends unknown>() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    24,
					},
				},
				Output: `function data<T>() {}`,
			},
			// Arrow functions with 'unknown' constraint
			{
				Code: `const data = <T extends unknown>() => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    24,
					},
				},
				Output: `const data = <T>() => {};`,
			},
			// Class declarations with 'unknown' constraint
			{
				Code: `class Data<T extends unknown> {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    21,
					},
				},
				Output: `class Data<T> {}`,
			},
			// Class expressions with 'unknown' constraint
			{
				Code: `const Data = class<T extends unknown> {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    29,
					},
				},
				Output: `const Data = class<T> {};`,
			},
			// Class methods with 'unknown' constraint
			{
				Code: `class Data { method<T extends unknown>() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    30,
					},
				},
				Output: `class Data { method<T>() {} }`,
			},
			// Interface declarations with 'unknown' constraint
			{
				Code: `interface Data<T extends unknown> {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    25,
					},
				},
				Output: `interface Data<T> {}`,
			},
			// Type alias declarations with 'unknown' constraint
			{
				Code: `type Data<T extends unknown> = {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryConstraint",
						Line:      1,
						Column:    20,
					},
				},
				Output: `type Data<T> = {};`,
			},
		},
	)
}

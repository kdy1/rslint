package no_useless_constructor

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoUselessConstructorRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessConstructorRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Empty classes
			{Code: `class A { }`},
			// Constructors with logic
			{Code: `class A { constructor() { doSomething(); } }`},
			// Extended classes with empty constructors
			{Code: `class A extends B { }`},
			{Code: `class A extends B { constructor() { } }`},
			// Constructors calling super with different arguments
			{Code: `class A extends B { constructor(value) { super(value); doSomething(); } }`},
			{Code: `class A extends B { constructor(value) { super(); } }`},
			{Code: `class A extends B { constructor(value) { super(value, 1); } }`},
			{Code: `class A extends B { constructor(...args) { super(...args, 1); } }`},
			{Code: `class A extends B { constructor(value) { super(1, value); } }`},
			// Constructors with additional logic after super
			{Code: `class A extends B { constructor() { super(); doSomething(); } }`},
			{Code: `class A extends B { constructor(...args) { super(...args); doSomething(); } }`},
			// Non-constructor methods
			{Code: `class A { method() { } }`},
			// Nested property access in super
			{Code: `class A extends B.C { constructor() { super(); } }`},
			// Destructured or default parameters
			{Code: `class A extends B { constructor({ a }) { super({ a }); } }`},
			{Code: `class A extends B { constructor(a = 1) { super(a); } }`},
			// Parameter reordering in super
			{Code: `class A extends B { constructor(a, b) { super(b, a); } }`},
			// TypeScript parameter modifiers (parameter properties)
			{Code: `class A { constructor(private a) { } }`},
			{Code: `class A { constructor(public a) { } }`},
			{Code: `class A { constructor(protected a) { } }`},
			{Code: `class A { constructor(readonly a) { } }`},
			{Code: `class A extends B { constructor(private a) { super(); } }`},
			{Code: `class A extends B { constructor(public a) { super(); } }`},
			{Code: `class A extends B { constructor(protected a) { super(); } }`},
			{Code: `class A extends B { constructor(readonly a) { super(); } }`},
			// Access modifiers on constructors
			{Code: `class A { private constructor() { } }`},
			{Code: `class A { protected constructor() { } }`},
			{Code: `class A extends B { public constructor() { super(); } }`},
			// Type definitions and overloads
			{Code: `declare class A { constructor(a: string); }`},
			// Parameter decorators
			{Code: `class A { constructor(@Foo a) { } }`},
			{Code: `class A { constructor(@Foo() a) { } }`},
			{Code: `class A extends B { constructor(@Foo a) { super(); } }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Empty constructor in non-extended class
			{
				Code: `class A { constructor() { } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUselessConstructor"},
				},
			},
			// Constructor only calling super()
			{
				Code: `class A extends B { constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUselessConstructor"},
				},
			},
			// Constructor forwarding single parameter
			{
				Code: `class A extends B { constructor(foo) { super(foo); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUselessConstructor"},
				},
			},
			// Constructor forwarding multiple parameters
			{
				Code: `class A extends B { constructor(foo, bar) { super(foo, bar); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUselessConstructor"},
				},
			},
			// Constructor forwarding spread arguments
			{
				Code: `class A extends B { constructor(...args) { super(...args); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUselessConstructor"},
				},
			},
			// Constructor using spread with arguments
			{
				Code: `class A extends B { constructor() { super(...arguments); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUselessConstructor"},
				},
			},
			// Public constructor only calling super
			{
				Code: `class A extends B { public constructor() { super(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUselessConstructor"},
				},
			},
			// Empty constructor (alternative formatting)
			{
				Code: `class A {
  constructor() {
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUselessConstructor"},
				},
			},
			// Multiple parameters passed through
			{
				Code: `class A extends B {
  constructor(a, b, c) {
    super(a, b, c);
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUselessConstructor"},
				},
			},
		},
	)
}

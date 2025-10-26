package no_inner_declarations

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInnerDeclarationsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInnerDeclarationsRule,
		[]rule_tester.ValidTestCase{
			// Valid: Top-level function declarations
			{Code: `function doSomething() { }`},

			// Valid: Nested functions in function bodies
			{Code: `function doSomething() { function somethingElse() { } }`},

			// Valid: Functions in IIFEs
			{Code: `(function() { function doSomething() { } }());`},

			// Valid: Variable declarations in blocks (default mode only checks functions)
			{Code: `if (test) { var fn = function() { }; }`},

			// Valid: Named function expressions
			{Code: `if (test) { var fn = function expr() { }; }`},

			// Valid: Function expressions in function declarations
			{Code: `function decl() { var fn = function expr() { }; }`},

			// Valid: Function assignments in conditionals
			{Code: `function decl(arg) { var fn; if (arg) { fn = function() { }; } }`},

			// Valid: Object methods
			{Code: `var x = {doSomething() {function doSomethingElse() {}}}`},

			// Valid: Function expressions in function with arg
			{Code: `function decl(arg) { var fn; if (arg) { fn = function expr() { }; } }`},

			// Valid: Variable declarations at top level
			{Code: `if (test) { var foo; }`},

			// Valid: Function declaration in while loop body (inside function)
			{Code: `function doSomething() { while (test) { var foo; } }`},

			// Valid: Arrow functions
			{Code: `foo(() => { function bar() { } });`},

			// Valid: exports.foo assignment
			{Code: `exports.foo = () => {}`},
			{Code: `exports.foo = function(){}`},
			{Code: `module.exports = function foo(){}`},

			// Valid: Class methods
			{Code: `class C { method() { function foo() {} } }`},
			{Code: `class C { method() { var x; } }`},

			// Valid: Class static blocks
			{Code: `class C { static { function foo() {} } }`},
			{Code: `class C { static { var x; } }`},

			// Valid: Strict mode with blockScopedFunctions: "allow"
			{Code: `'use strict'
if (test) { function doSomething() { } }`, Options: []interface{}{"functions", map[string]interface{}{"blockScopedFunctions": "allow"}}},

			// Valid: Strict mode function declarations (default allows block-scoped in strict mode)
			{Code: `'use strict'
if (test) { function doSomething() { } }`},

			// Valid: Nested strict mode with blockScopedFunctions: "allow"
			{Code: `function foo() {'use strict'
if (test) { function doSomething() { } } }`, Options: []interface{}{"functions", map[string]interface{}{"blockScopedFunctions": "allow"}}},

			// Valid: Block in strict mode module with blockScopedFunctions: "allow"
			{Code: `function foo() { { function bar() { } } }`, Options: []interface{}{"functions", map[string]interface{}{"blockScopedFunctions": "allow"}}},

			// Valid: Class method in strict mode with blockScopedFunctions: "allow"
			{Code: `class C { method() { if(test) { function somethingElse() { } } } }`, Options: []interface{}{"functions", map[string]interface{}{"blockScopedFunctions": "allow"}}},

			// Valid: Class expression method with blockScopedFunctions: "allow"
			{Code: `const C = class { method() { if(test) { function somethingElse() { } } } }`, Options: []interface{}{"functions", map[string]interface{}{"blockScopedFunctions": "allow"}}},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid: Function in if statement with "both" option
			{
				Code:    `if (test) { function doSomething() { } }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Var in if statement with "both" option
			{
				Code:    `if (foo) var a;`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Var in if with comments with "both" option
			{
				Code:    `if (foo) /* some comments */ var a;`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Nested function and var declarations
			{
				Code:    `if (foo){ function f(){ if(bar){ var a; } } }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Function in if with nested var
			{
				Code:    `if (foo) function f(){ if(bar) var a; }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Var function expression in if with "both" option
			{
				Code:    `if (foo) { var fn = function(){} }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Function in if (default mode)
			{
				Code: `if (foo) function f(){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Function in function body's if statement
			{
				Code:    `function bar() { if (foo) function f(){}; }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Var in function body's if statement
			{
				Code:    `function bar() { if (foo) var a; }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Var in if block
			{
				Code:    `if (foo) { var a; }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Function in do-while loop
			{
				Code: `function doSomething() { do { function somethingElse() { } } while (test); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Function in IIFE's if statement
			{
				Code: `(function() { if (test) { function doSomething() { } } }());`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Var in while loop
			{
				Code:    `while (test) { var foo; }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Var in function's if statement
			{
				Code:    `function doSomething() { if (test) { var foo = 42; } }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Var in IIFE's if statement
			{
				Code:    `(function() { if (test) { var foo; } }());`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Var in arrow function's if statement
			{
				Code:    `const doSomething = () => { if (test) { var foo = 42; } }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Var in class method's if statement
			{
				Code:    `class C { method() { if(test) { var foo; } } }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Var in class static block's if statement
			{
				Code:    `class C { static { if (test) { var foo; } } }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Function in class static block with blockScopedFunctions: "disallow"
			{
				Code:    `class C { static { if (test) { function foo() {} } } }`,
				Options: []interface{}{"both", map[string]interface{}{"blockScopedFunctions": "disallow"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Nested var in class static block
			{
				Code:    `class C { static { if (test) { if (anotherTest) { var foo; } } } }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Function in if with blockScopedFunctions: "disallow"
			{
				Code:    `if (test) { function doSomething() { } }`,
				Options: []interface{}{"both", map[string]interface{}{"blockScopedFunctions": "disallow"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Strict mode with blockScopedFunctions: "disallow"
			{
				Code: `'use strict'
if (test) { function doSomething() { } }`,
				Options: []interface{}{"both", map[string]interface{}{"blockScopedFunctions": "disallow"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Function in strict mode function body with blockScopedFunctions: "disallow"
			{
				Code: `function foo() {'use strict'
{ function bar() { } } }`,
				Options: []interface{}{"both", map[string]interface{}{"blockScopedFunctions": "disallow"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Function in do-while with strict mode and blockScopedFunctions: "disallow"
			{
				Code: `function doSomething() { 'use strict'
do { function somethingElse() { } } while (test); }`,
				Options: []interface{}{"both", map[string]interface{}{"blockScopedFunctions": "disallow"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},

			// Invalid: Block-scoped function at program level
			{
				Code: `{ function foo () {'use strict'
console.log('foo called'); } }`,
				Options: []interface{}{"both"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "moveDeclToRoot"},
				},
			},
		},
	)
}

func TestNoInnerDeclarationsRuleBothMode(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInnerDeclarationsRule,
		[]rule_tester.ValidTestCase{
			// Valid: let/const in blocks (not checked by this rule)
			{
				Code:    `if (test) { let x = 1; }`,
				Options: []interface{}{"both"},
			},
			{
				Code:    `if (test) { const x = 1; }`,
				Options: []interface{}{"both"},
			},

			// Valid: Top-level var
			{
				Code:    `var foo;`,
				Options: []interface{}{"both"},
			},
			{
				Code:    `var foo = 42;`,
				Options: []interface{}{"both"},
			},

			// Valid: Var in function body
			{
				Code:    `function doSomething() { var foo; }`,
				Options: []interface{}{"both"},
			},
			{
				Code:    `(function() { var foo; }());`,
				Options: []interface{}{"both"},
			},

			// Valid: Var in arrow function
			{
				Code:    `var fn = () => {var foo;}`,
				Options: []interface{}{"both"},
			},

			// Valid: Var in object method
			{
				Code:    `var x = {doSomething() {var foo;}}`,
				Options: []interface{}{"both"},
			},

			// Valid: Export statements
			{
				Code:    `export var foo;`,
				Options: []interface{}{"both"},
			},
			{
				Code:    `export function bar() {}`,
				Options: []interface{}{"both"},
			},
			{
				Code:    `export default function baz() {}`,
				Options: []interface{}{"both"},
			},
		},
		[]rule_tester.InvalidTestCase{},
	)
}

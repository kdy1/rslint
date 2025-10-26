package no_unsafe_finally

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoUnsafeFinallyRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnsafeFinallyRule,
		[]rule_tester.ValidTestCase{
			{Code: `var foo = function() {
 try {
 return 1;
 } catch(err) {
 return 2;
 } finally {
 console.log('hola!')
 }
 }`},
			{Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { console.log('hola!') } }`},
			{Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { function a(x) { return x } } }`},
			{Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { var a = function(x) { if(!x) { throw new Error() } } } }`},
			{Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { var a = function(x) { while(true) { if(x) { break } else { continue } } } } }`},
			{Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { var a = function(x) { label: while(true) { if(x) { break label; } else { continue } } } } }`},
			{Code: `var foo = function() { try {} finally { while (true) break; } }`},
			{Code: `var foo = function() { try {} finally { while (true) continue; } }`},
			{Code: `var foo = function() { try {} finally { switch (true) { case true: break; } } }`},
			{Code: `var foo = function() { try {} finally { do { break; } while (true) } }`},
			{Code: `var foo = function() { try { return 1; } catch(err) { return 2; } finally { var bar = () => { throw new Error(); }; } };`},
			{Code: `var foo = function() { try { return 1; } catch(err) { return 2 } finally { (x) => x } }`},
			{Code: `var foo = function() { try { return 1; } finally { class bar { constructor() {} static ehm() { return 'Hola!'; } } } };`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `var foo = function() {
 try {
 return 1;
 } catch(err) {
 return 2;
 } finally {
 return 3;
 }
 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 7, Column: 2},
				},
			},
			{
				Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { if(true) { return 3 } else { return 2 } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 86},
					{MessageId: "unsafeUsage", Line: 1, Column: 104},
				},
			},
			{
				Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { return 3 } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 75},
				},
			},
			{
				Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { return function(x) { return y } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 75},
				},
			},
			{
				Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { return { x: function(c) { return c } } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 75},
				},
			},
			{
				Code: `var foo = function() { try { return 1 } catch(err) { return 2 } finally { throw new Error() } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 75},
				},
			},
			{
				Code: `var foo = function() { try { foo(); } finally { try { bar(); } finally { return; } } };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 74},
				},
			},
			{
				Code: `var foo = function() { label: try { return 0; } finally { break label; } return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 59},
				},
			},
			{
				Code: `var foo = function() {
 a: try {
 return 1;
 } catch(err) {
 return 2;
 } finally {
 break a;
 }
 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 7, Column: 2},
				},
			},
			{
				Code: `var foo = function() { while (true) try {} finally { break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 54},
				},
			},
			{
				Code: `var foo = function() { while (true) try {} finally { continue; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 54},
				},
			},
			{
				Code: `var foo = function() { switch (true) { case true: try {} finally { break; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 68},
				},
			},
			{
				Code: `var foo = function() { a: while (true) try {} finally { switch (true) { case true: break a; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 84},
				},
			},
			{
				Code: `var foo = function() { a: while (true) try {} finally { switch (true) { case true: continue; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 84},
				},
			},
			{
				Code: `var foo = function() { a: do {} while (true); try {} finally { break a; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unsafeUsage", Line: 1, Column: 64},
				},
			},
		},
	)
}

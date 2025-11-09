package member_ordering

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestMemberOrderingRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&MemberOrderingRule,
		[]rule_tester.ValidTestCase{
			// Default configuration - interface with correct order
			{
				Code: `
interface Foo {
  [Z: string]: any;
  A: string;
  B: string;
  new ();
  G(): void;
  H(): void;
}`,
			},
			// "never" configuration - no enforcement
			{
				Code: `
interface Foo {
  A: string;
  J(): void;
  K(): void;
  D: string;
  [Z: string]: any;
}`,
				Options: []interface{}{map[string]interface{}{"default": "never"}},
			},
			// Specific member type order
			{
				Code: `
interface Foo {
  [Z: string]: any;
  A: string;
  B: string;
  new ();
  G(): void;
  H(): void;
}`,
				Options: []interface{}{map[string]interface{}{
					"default": []interface{}{"signature", "field", "constructor", "method"},
				}},
			},
			// Class with correct order
			{
				Code: `
class Foo {
  public static A: string;
  protected static B: string = "";
  private static C: string = "";

  public D: string = "";
  protected E: string = "";
  private F: string = "";

  constructor() {}

  public static G(): void {}
  protected static H(): void {}
  private static I(): void {}

  public J(): void {}
  protected K(): void {}
  private L(): void {}
}`,
			},
			// Alphabetical ordering
			{
				Code: `
interface Foo {
  a: string;
  b: string;
  c: string;
}`,
				Options: []interface{}{map[string]interface{}{
					"default": map[string]interface{}{
						"order": "alphabetically",
					},
				}},
			},
			// Case-insensitive alphabetical ordering
			{
				Code: `
interface Foo {
  a: string;
  B: string;
  c: string;
}`,
				Options: []interface{}{map[string]interface{}{
					"default": map[string]interface{}{
						"order": "alphabetically-case-insensitive",
					},
				}},
			},
			// Natural ordering
			{
				Code: `
interface Foo {
  a1: number;
  a5: number;
  a10: number;
}`,
				Options: []interface{}{map[string]interface{}{
					"default": map[string]interface{}{
						"order": "natural",
					},
				}},
			},
			// Optional members
			{
				Code: `
interface Foo {
  a?: string;
  b?: string;
  c: string;
  d: string;
}`,
				Options: []interface{}{map[string]interface{}{
					"default": map[string]interface{}{
						"optionalityOrder": "optional-first",
					},
				}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Interface with wrong order
			{
				Code: `
interface Foo {
  G(): void;
  A: string;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "incorrectOrder",
						Line:      4,
						Column:    3,
					},
				},
			},
			// Wrong alphabetical order
			{
				Code: `
interface Foo {
  c: string;
  b: string;
  a: string;
}`,
				Options: []interface{}{map[string]interface{}{
					"default": map[string]interface{}{
						"order": "alphabetically",
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "incorrectOrder",
						Line:      4,
						Column:    3,
					},
					{
						MessageId: "incorrectOrder",
						Line:      5,
						Column:    3,
					},
				},
			},
			// Wrong natural order
			{
				Code: `
interface Foo {
  a10: number;
  a5: number;
}`,
				Options: []interface{}{map[string]interface{}{
					"default": map[string]interface{}{
						"order": "natural",
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "incorrectOrder",
						Line:      4,
						Column:    3,
					},
				},
			},
			// Wrong optionality order
			{
				Code: `
interface Foo {
  a: string;
  b?: string;
}`,
				Options: []interface{}{map[string]interface{}{
					"default": map[string]interface{}{
						"optionalityOrder": "optional-first",
					},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "incorrectRequiredMembersOrder",
						Line:      4,
						Column:    3,
					},
				},
			},
		},
	)
}

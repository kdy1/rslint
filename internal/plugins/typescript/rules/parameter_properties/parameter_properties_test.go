package parameter_properties

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestParameterPropertiesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ParameterPropertiesRule,
		[]rule_tester.ValidTestCase{
			{Code: `class Foo { constructor(name: string) {} }`},
			{Code: `class Foo { constructor(...name: string[]) {} }`},
			{Code: `class Foo { constructor(readonly name: string) {} }`, Options: []interface{}{map[string]interface{}{"allow": []interface{}{"readonly"}}}},
			{Code: `class Foo { constructor(private name: string) {} }`, Options: []interface{}{map[string]interface{}{"allow": []interface{}{"private"}}}},
			{Code: `class Foo { constructor(protected name: string) {} }`, Options: []interface{}{map[string]interface{}{"allow": []interface{}{"protected"}}}},
			{Code: `class Foo { constructor(public name: string) {} }`, Options: []interface{}{map[string]interface{}{"allow": []interface{}{"public"}}}},
			{Code: `class Foo { constructor(private readonly name: string) {} }`, Options: []interface{}{map[string]interface{}{"allow": []interface{}{"private readonly"}}}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `class Foo { constructor(readonly name: string) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferClassProperty",
					Line:      1,
					Column:    25,
				}},
			},
			{
				Code: `class Foo { constructor(private name: string) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferClassProperty",
					Line:      1,
					Column:    25,
				}},
			},
			{
				Code: `class Foo { constructor(protected name: string) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferClassProperty",
					Line:      1,
					Column:    25,
				}},
			},
			{
				Code: `class Foo { constructor(public name: string) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferClassProperty",
					Line:      1,
					Column:    25,
				}},
			},
		},
	)
}

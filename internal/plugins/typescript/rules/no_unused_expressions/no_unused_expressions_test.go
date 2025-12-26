package no_unused_expressions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/test"
)

func TestNoUnusedExpressions(t *testing.T) {
	ruleTester := test.NewRuleTester(t)
	ruleTester.Run(NoUnusedExpressionsRule, test.RuleTesterOptions{
		Valid: []test.ValidTestCase{
			{
				Code: `test.age?.toLocaleString();`,
			},
			{
				Code: `let a = (a?.b).c;`,
			},
			{
				Code: `let b = a?.['b'];`,
			},
			{
				Code: `let c = one[2]?.[3][4];`,
			},
			{
				Code: `one[2]?.[3][4]?.();`,
			},
			{
				Code: `a?.['b']?.c();`,
			},
			{
				Code: `
module Foo {
  'use strict';
}
				`,
			},
			{
				Code: `
namespace Foo {
  'use strict';

  export class Foo {}
  export class Bar {}
}
				`,
			},
			{
				Code: `
function foo() {
  'use strict';

  return null;
}
				`,
			},
			{
				Code: `import('./foo');`,
			},
			{
				Code: `import('./foo').then(() => {});`,
			},
			{
				Code: `
class Foo<T> {}
new Foo<string>();
				`,
			},
			{
				Code: `foo && foo?.();`,
				Options: map[string]interface{}{
					"allowShortCircuit": true,
				},
			},
			{
				Code: `foo && import('./foo');`,
				Options: map[string]interface{}{
					"allowShortCircuit": true,
				},
			},
			{
				Code: `foo ? import('./foo') : import('./bar');`,
				Options: map[string]interface{}{
					"allowTernary": true,
				},
			},
		},
		Invalid: []test.InvalidTestCase{
			{
				Code: `if (0) 0;`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `f(0), {};`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `a, b();`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `
a() &&
  function namedFunctionInExpressionContext() {
    f();
  };
				`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `a?.b;`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `(a?.b).c;`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `a?.['b'];`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `(a?.['b']).c;`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `a?.b()?.c;`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `(a?.b()).c;`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `one[2]?.[3][4];`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `one.two?.three.four;`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `
module Foo {
  const foo = true;
  'use strict';
}
				`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `
namespace Foo {
  export class Foo {}
  export class Bar {}

  'use strict';
}
				`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `
function foo() {
  const foo = true;

  ('use strict');
}
				`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `foo && foo?.bar;`,
				Options: map[string]interface{}{
					"allowShortCircuit": true,
				},
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `foo ? foo?.bar : bar.baz;`,
				Options: map[string]interface{}{
					"allowTernary": true,
				},
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `
class Foo<T> {}
Foo<string>;
				`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `Map<string, string>;`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `
declare const foo: number | undefined;
foo;
				`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `
declare const foo: number | undefined;
foo as any;
				`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `
declare const foo: number | undefined;
<any>foo;
				`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
			{
				Code: `
declare const foo: number | undefined;
foo!;
				`,
				Errors: []rule.RuleMessage{
					{
						Id:          "unusedExpression",
						Description: "Expected an assignment or function call and instead saw an expression.",
					},
				},
			},
		},
	})
}

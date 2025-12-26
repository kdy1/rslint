import { noFormat, RuleTester } from '@typescript-eslint/rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('restrict-plus-operands', {
  valid: [
    'let x = 5;',
    "let y = '10';",
    'let foo = 5 + 10;',
    "let foo = '5.5' + '10';",
    'let foo = 1n + 1n;',
    `
function test(s: string, n: number): number {
  return 2;
}
let foo = test('5.5', 10) + 10;
    `,
    `
let x = 5;
let z = 8.2;
let foo = x + z;
    `,
    'let foo = 1 + 1;',
    "let foo = '1' + '1';",
    `
function foo<T extends string>(a: T) {
  return a + '';
}
    `,
    `
function foo<T extends number>(a: T) {
  return a + 1;
}
    `,
    `
declare const a: {} & string;
declare const b: string;
const x = a + b;
    `,
    `
declare const a: {} & number;
declare const b: number;
const x = a + b;
    `,
    `
declare const a: {} & bigint;
declare const b: bigint;
const x = a + b;
    `,
    {
      code: `
        declare const a: RegExp;
        declare const b: string;
        const x = a + b;
      `,
      options: [
        {
          allowAny: false,
          allowBoolean: false,
          allowNullish: false,
          allowNumberAndString: false,
          allowRegExp: true,
        },
      ],
    },
    {
      code: `
let foo: string | undefined;
foo = foo + 'some data';
      `,
      options: [
        {
          allowNullish: true,
        },
      ],
    },
    {
      code: `
let foo = '';
foo += 0;
      `,
      options: [
        {
          allowAny: false,
          allowBoolean: false,
          allowNullish: false,
          allowNumberAndString: false,
          allowRegExp: false,
          skipCompoundAssignments: true,
        },
      ],
    },
    {
      code: `
const f = (a: any, b: any) => a + b;
      `,
      options: [
        {
          allowAny: true,
        },
      ],
    },
    {
      code: `
const f = (a: string, b: string | number) => a + b;
      `,
      options: [
        {
          allowAny: true,
          allowBoolean: true,
          allowNullish: true,
          allowNumberAndString: true,
          allowRegExp: true,
        },
      ],
    },
  ],
  invalid: [
    {
      code: "let foo = '1' + 1;",
      errors: [
        {
          column: 11,
          line: 1,
          messageId: 'mismatched',
        },
      ],
      options: [{ allowNumberAndString: false }],
    },
    {
      code: 'let foo = [] + {};',
      errors: [
        {
          column: 11,
          endColumn: 13,
          line: 1,
          messageId: 'invalid',
        },
        {
          column: 16,
          endColumn: 18,
          line: 1,
          messageId: 'invalid',
        },
      ],
    },
    {
      code: "let foo = 5 + '10';",
      errors: [
        {
          column: 11,
          line: 1,
          messageId: 'mismatched',
        },
      ],
      options: [
        {
          allowAny: false,
          allowBoolean: false,
          allowNullish: false,
          allowNumberAndString: false,
          allowRegExp: false,
        },
      ],
    },
    {
      code: 'let foo = [] + 5;',
      errors: [
        {
          column: 11,
          endColumn: 13,
          line: 1,
          messageId: 'invalid',
        },
      ],
    },
    {
      code: 'let foo = [] + [];',
      errors: [
        {
          column: 11,
          endColumn: 13,
          line: 1,
          messageId: 'invalid',
        },
        {
          column: 16,
          endColumn: 18,
          line: 1,
          messageId: 'invalid',
        },
      ],
    },
    {
      code: "let foo = 5 + {};",
      errors: [
        {
          column: 15,
          endColumn: 17,
          line: 1,
          messageId: 'invalid',
        },
      ],
    },
    {
      code: "let foo = '5.5' + {};",
      errors: [
        {
          column: 19,
          endColumn: 21,
          line: 1,
          messageId: 'invalid',
        },
      ],
    },
    {
      code: 'let foo = 5.5 + [];',
      errors: [
        {
          column: 17,
          endColumn: 19,
          line: 1,
          messageId: 'invalid',
        },
      ],
    },
    {
      code: `
let x = 5;
let y = [];
let foo = x + y;
      `,
      errors: [
        {
          column: 15,
          endColumn: 16,
          line: 4,
          messageId: 'invalid',
        },
      ],
    },
    {
      code: 'let foo = 1n + 1;',
      errors: [
        {
          column: 11,
          line: 1,
          messageId: 'bigintAndNumber',
        },
      ],
    },
    {
      code: 'let foo = 1 + 1n;',
      errors: [
        {
          column: 11,
          line: 1,
          messageId: 'bigintAndNumber',
        },
      ],
    },
    {
      code: `
let foo = 1n;
foo + 1;
      `,
      errors: [
        {
          column: 1,
          line: 3,
          messageId: 'bigintAndNumber',
        },
      ],
    },
    {
      code: `
function test(s: string, n: never) {
  return s + n;
}
      `,
      errors: [
        {
          column: 12,
          endColumn: 13,
          line: 3,
          messageId: 'invalid',
        },
      ],
    },
    {
      code: `
let foo: boolean = true;
foo + 'a';
      `,
      errors: [
        {
          column: 1,
          endColumn: 4,
          line: 3,
          messageId: 'invalid',
        },
      ],
      options: [{ allowBoolean: false }],
    },
    {
      code: `
let foo = 0;
foo += 'some data';
      `,
      errors: [
        {
          column: 1,
          line: 3,
          messageId: 'mismatched',
        },
      ],
      options: [
        {
          allowAny: false,
          allowBoolean: false,
          allowNullish: false,
          allowNumberAndString: false,
          allowRegExp: false,
        },
      ],
    },
  ],
});

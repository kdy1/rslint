import { RuleTester } from '@typescript-eslint/rule-tester';

import { getFixturesRootDir } from '../RuleTester';

const rootDir = getFixturesRootDir();
const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootDir,
    },
  },
});

ruleTester.run('prefer-includes', {
  valid: [
    'foo.indexOf(bar)',
    'foo.indexOf(bar) + 1',
    'foo.indexOf(bar) - 1',
    `
      declare const val: string | number;
      val.indexOf('foo') === -1;
    `,
    `
      type UserDefinedType = { indexOf(x: string): number };
      declare const foo: UserDefinedType;
      foo.indexOf('x') === -1;
    `,
    `
      type WithDifferentIncludes = {
        indexOf(x: string): number;
        includes(x: string, fromIndex: number): boolean;
      };
      declare const foo: WithDifferentIncludes;
      foo.indexOf('x') === -1;
    `,
    `
      type WithBooleanIncludes = {
        indexOf(x: string): number;
        includes: boolean;
      };
      declare const foo: WithBooleanIncludes;
      foo.indexOf('x') === -1;
    `,
    '/bar/i.test(foo)',
    '/ba[rz]/.test(foo)',
    '/foo|bar/.test(foo)',
    'pattern.test()',
    `
      type WithTest = { test(x: string): boolean };
      declare const obj: WithTest;
      obj.test(foo);
    `,
    `
      const pattern = /foo/;
      const regex = pattern;
      regex.test(str);
    `,
  ],
  invalid: [
    {
      code: 'foo.indexOf(bar) !== -1',
      errors: [{ messageId: 'preferIncludes' }],
      output: 'foo.includes(bar)',
    },
    {
      code: 'foo.indexOf(bar) != -1',
      errors: [{ messageId: 'preferIncludes' }],
      output: 'foo.includes(bar)',
    },
    {
      code: 'foo.indexOf(bar) > -1',
      errors: [{ messageId: 'preferIncludes' }],
      output: 'foo.includes(bar)',
    },
    {
      code: 'foo.indexOf(bar) >= 0',
      errors: [{ messageId: 'preferIncludes' }],
      output: 'foo.includes(bar)',
    },
    {
      code: 'foo.indexOf(bar) === -1',
      errors: [{ messageId: 'preferIncludes' }],
      output: '!foo.includes(bar)',
    },
    {
      code: 'foo.indexOf(bar) == -1',
      errors: [{ messageId: 'preferIncludes' }],
      output: '!foo.includes(bar)',
    },
    {
      code: 'foo.indexOf(bar) < 0',
      errors: [{ messageId: 'preferIncludes' }],
      output: '!foo.includes(bar)',
    },
    {
      code: 'foo.indexOf(bar) <= -1',
      errors: [{ messageId: 'preferIncludes' }],
      output: '!foo.includes(bar)',
    },
    {
      code: '-1 !== foo.indexOf(bar)',
      errors: [{ messageId: 'preferIncludes' }],
      output: 'foo.includes(bar)',
    },
    {
      code: '0 <= foo.indexOf(bar)',
      errors: [{ messageId: 'preferIncludes' }],
      output: 'foo.includes(bar)',
    },
    {
      code: '-1 === foo.indexOf(bar)',
      errors: [{ messageId: 'preferIncludes' }],
      output: '!foo.includes(bar)',
    },
    {
      code: '0 > foo.indexOf(bar)',
      errors: [{ messageId: 'preferIncludes' }],
      output: '!foo.includes(bar)',
    },
    {
      code: 'a?.indexOf(b) === -1',
      errors: [{ messageId: 'preferIncludes' }],
      output: null,
    },
    {
      code: 'a?.indexOf(b) !== -1',
      errors: [{ messageId: 'preferIncludes' }],
      output: null,
    },
    {
      code: '/bar/.test(foo)',
      errors: [{ messageId: 'preferStringIncludes' }],
      output: 'foo.includes("bar")',
    },
    {
      code: '/bar/.test((1 + 1, foo))',
      errors: [{ messageId: 'preferStringIncludes' }],
      output: '(1 + 1, foo).includes("bar")',
    },
    {
      code: String.raw`/\0'\\n\r\v\t\f/.test(foo)`,
      errors: [{ messageId: 'preferStringIncludes' }],
      output: String.raw`foo.includes("\0'\\n\r\v\t\f")`,
    },
    {
      code: "new RegExp('bar').test(foo)",
      errors: [{ messageId: 'preferStringIncludes' }],
      output: 'foo.includes("bar")',
    },
    {
      code: `
        const pattern = 'bar';
        new RegExp(pattern).test(foo + bar);
      `,
      errors: [{ messageId: 'preferStringIncludes' }],
      output: `
        const pattern = 'bar';
        (foo + bar).includes("bar");
      `,
    },
    {
      code: `
        declare const arr: any[];
        arr.indexOf(x) !== -1;
      `,
      errors: [{ messageId: 'preferIncludes' }],
      output: `
        declare const arr: any[];
        arr.includes(x);
      `,
    },
    {
      code: `
        declare const arr: ReadonlyArray<any>;
        arr.indexOf(x) !== -1;
      `,
      errors: [{ messageId: 'preferIncludes' }],
      output: `
        declare const arr: ReadonlyArray<any>;
        arr.includes(x);
      `,
    },
    {
      code: `
        declare const arr: Int8Array;
        arr.indexOf(x) !== -1;
      `,
      errors: [{ messageId: 'preferIncludes' }],
      output: `
        declare const arr: Int8Array;
        arr.includes(x);
      `,
    },
    {
      code: `
        declare const arr: Uint8Array;
        arr.indexOf(x) !== -1;
      `,
      errors: [{ messageId: 'preferIncludes' }],
      output: `
        declare const arr: Uint8Array;
        arr.includes(x);
      `,
    },
    {
      code: `
        declare const arr: Float32Array;
        arr.indexOf(x) !== -1;
      `,
      errors: [{ messageId: 'preferIncludes' }],
      output: `
        declare const arr: Float32Array;
        arr.includes(x);
      `,
    },
    {
      code: `
        function fn<T extends string>(x: T) {
          return x.indexOf('a') !== -1;
        }
      `,
      errors: [{ messageId: 'preferIncludes' }],
      output: `
        function fn<T extends string>(x: T) {
          return x.includes('a');
        }
      `,
    },
    {
      code: `
        declare const arr: Readonly<any[]>;
        arr.indexOf(x) !== -1;
      `,
      errors: [{ messageId: 'preferIncludes' }],
      output: `
        declare const arr: Readonly<any[]>;
        arr.includes(x);
      `,
    },
    {
      code: `
        type WithBothMethods = {
          indexOf(x: string): number;
          includes(x: string): boolean;
        };
        declare const foo: WithBothMethods;
        foo.indexOf('x') !== -1;
      `,
      errors: [{ messageId: 'preferIncludes' }],
      output: `
        type WithBothMethods = {
          indexOf(x: string): number;
          includes(x: string): boolean;
        };
        declare const foo: WithBothMethods;
        foo.includes('x');
      `,
    },
  ],
});

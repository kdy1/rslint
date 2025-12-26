import { noFormat, RuleTester } from '@typescript-eslint/rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-function-type', {
  valid: [
    `
interface Foo {
  (): void;
  bar: number;
}
    `,
    `
type Foo = {
  (): void;
  bar: number;
};
    `,
    `
function foo(bar: { (): string; baz: number }): string {
  return bar();
}
    `,
    `
interface Foo {
  bar: string;
}
interface Bar extends Foo {
  (): void;
}
    `,
    `
interface Foo {
  bar: string;
}
interface Bar extends Function, Foo {
  (): void;
}
    `,
  ],

  invalid: [
    {
      code: `
interface Foo {
  (): string;
}
      `,
      errors: [
        {
          messageId: 'functionTypeOverCallableType',
        },
      ],
      output: `
type Foo = () => string;
      `,
    },
    // https://github.com/typescript-eslint/typescript-eslint/issues/3004
    {
      code: `
export default interface Foo {
  /** comment */
  (): string;
}
      `,
      errors: [
        {          messageId: 'functionTypeOverCallableType',        },
      ],
      output: null,
    },
    {
      code: `
interface Foo {
  // comment
  (): string;
}
      `,
      errors: [
        {          messageId: 'functionTypeOverCallableType',        },
      ],
      output: `
// comment
type Foo = () => string;
      `,
    },
    {
      code: `
export interface Foo {
  /** comment */
  (): string;
}
      `,
      errors: [
        {          messageId: 'functionTypeOverCallableType',        },
      ],
      output: `
/** comment */
export type Foo = () => string;
      `,
    },
    {
      code: `
export interface Foo {
  // comment
  (): string;
}
      `,
      errors: [
        {          messageId: 'functionTypeOverCallableType',        },
      ],
      output: `
// comment
export type Foo = () => string;
      `,
    },
    {
      code: `
function foo(bar: { /* comment */ (s: string): number } | undefined): number {
  return bar('hello');
}
      `,
      errors: [
        {          messageId: 'functionTypeOverCallableType',        },
      ],
      output: `
function foo(bar: /* comment */ ((s: string) => number) | undefined): number {
  return bar('hello');
}
      `,
    },
    {
      code: `
type Foo = {
  (): string;
};
      `,
      errors: [
        {          messageId: 'functionTypeOverCallableType',        },
      ],
      output: `
type Foo = () => string;
      `,
    },
    {
      code: `
function foo(bar: { (s: string): number }): number {
  return bar('hello');
}
      `,
      errors: [
        {          messageId: 'functionTypeOverCallableType',        },
      ],
      output: `
function foo(bar: (s: string) => number): number {
  return bar('hello');
}
      `,
    },
    {
      code: `
function foo(bar: { (s: string): number } | undefined): number {
  return bar('hello');
}
      `,
      errors: [
        {          messageId: 'functionTypeOverCallableType',        },
      ],
      output: `
function foo(bar: ((s: string) => number) | undefined): number {
  return bar('hello');
}
      `,
    },
    {
      code: `
interface Foo extends Function {
  (): void;
}
      `,
      errors: [
        {          messageId: 'functionTypeOverCallableType',        },
      ],
      output: `
type Foo = () => void;
      `,
    },
    {
      code: `
interface Foo<T> {
  (bar: T): string;
}
      `,
      errors: [
        {          messageId: 'functionTypeOverCallableType',        },
      ],
      output: `
type Foo<T> = (bar: T) => string;
      `,
    },
    {
      code: `
interface Foo<T> {
  (this: T): void;
}
      `,
      errors: [
        {          messageId: 'functionTypeOverCallableType',        },
      ],
      output: `
type Foo<T> = (this: T) => void;
      `,
    },
    {
      code: `
type Foo<T> = { (this: string): T };
      `,
      errors: [
        {          messageId: 'functionTypeOverCallableType',        },
      ],
      output: `
type Foo<T> = (this: string) => T;
      `,
    },
    {
      code: `
interface Foo {
  (arg: this): void;
}
      `,
      errors: [
        {          messageId: 'unexpectedThisOnFunctionOnlyInterface',        },
      ],
      output: null,
    },
    {
      code: `
interface Foo {
  (arg: number): this | undefined;
}
      `,
      errors: [
        {          messageId: 'unexpectedThisOnFunctionOnlyInterface',        },
      ],
      output: null,
    },
    {
      code: `
// isn't actually valid ts but want to not give message saying it refers to Foo.
interface Foo {
  (): {
    a: {
      nested: this;
    };
    between: this;
    b: {
      nested: string;
    };
  };
}
      `,
      errors: [
        {          messageId: 'functionTypeOverCallableType',        },
      ],
      output: `
// isn't actually valid ts but want to not give message saying it refers to Foo.
type Foo = () => {
    a: {
      nested: this;
    };
    between: this;
    b: {
      nested: string;
    };
  };
      `,
    },
    {
      code: noFormat`
type X = {} | { (): void; }
      `,
      errors: [
        {          messageId: 'functionTypeOverCallableType',        },
      ],
      output: `
type X = {} | (() => void)
      `,
    },
    {
      code: noFormat`
type X = {} & { (): void; };
      `,
      errors: [
        {          messageId: 'functionTypeOverCallableType',        },
      ],
      output: `
type X = {} & (() => void);
      `,
    },
  ],
});

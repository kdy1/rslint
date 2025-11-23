import { RuleTester } from '@typescript-eslint/rule-tester';



const ruleTester = new RuleTester();

ruleTester.run('no-loss-of-precision', {
  valid: [
    'const x = 12345;',
    'const x = 123.456;',
    'const x = -123.456;',
    'const x = 123_456;',
    'const x = 123_00_000_000_000_000_000_000_000;',
    'const x = 123.000_000_000_000_000_000_000_0;',
    'const x = 0;',
    'const x = 0.0;',
    'const x = 0.000000000000001;',
    'const x = -0.000000000000001;',
    'const x = 1234567890;',
    'const x = 9007199254740991;', // MAX_SAFE_INTEGER
    'const x = -9007199254740991;', // -MAX_SAFE_INTEGER
    'const x = 0x1fffffffffffff;', // MAX_SAFE_INTEGER in hex
    'const x = 0b11111111111111111111111111111111111111111111111111111;', // MAX_SAFE_INTEGER in binary
    'const x = 0o377777777777777777;', // MAX_SAFE_INTEGER in octal
    'const x = 123e34;',
    'const x = 123.456e78;',
  ],
  invalid: [
    {
      code: 'const x = 9007199254740993;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 9_007_199_254_740_993;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 9_007_199_254_740.993e3;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 0b100_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_001;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 5123000000000000000000000000001;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = -9007199254740993;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 0x20000000000001;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 0X20000000000001;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 0o400000000000000001;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 0O400000000000000001;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 0b100000000000000000000000000000000000000000000000000001;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 0B100000000000000000000000000000000000000000000000000001;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 1.230000000000000000000000000000000000000000000000000000001;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 1e999;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
  ],
});

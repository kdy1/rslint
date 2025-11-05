import { createTestCases } from './createTestCases';

createTestCases([
  {
    code: [
      'function foo(%) {}',
      '(function (%) {});',
      'declare function foo(%);',
      'function foo({%}) {}',
      'function foo(...%) {}',
      'function foo({% = 1}) {}',
      'function foo({...%}) {}',
      'function foo([%]) {}',
      'function foo([% = 1]) {}',
      'function foo([...%]) {}',
    ],
    options: {
      selector: 'parameter',
    },
  },
]);

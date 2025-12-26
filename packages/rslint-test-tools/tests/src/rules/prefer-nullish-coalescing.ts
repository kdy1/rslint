export type MessageIds =
  | 'preferNullishOverOr'
  | 'preferNullishOverTernary'
  | 'noStrictNullCheck'
  | 'suggestNullish';

export type IgnorePrimitivesOption =
  | {
      string?: boolean;
      number?: boolean;
      bigint?: boolean;
      boolean?: boolean;
    }
  | true;

export type Options = [
  {
    ignoreTernaryTests?: boolean;
    ignoreConditionalTests?: boolean;
    ignoreMixedLogicalExpressions?: boolean;
    allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing?: boolean;
    ignorePrimitives?: IgnorePrimitivesOption;
    ignoreBooleanCoercion?: boolean;
  }?,
];

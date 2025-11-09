package no_non_null_asserted_optional_chain

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestNoNonNullAssertedOptionalChain(t *testing.T) {
	rule.RunRuleTest(t, NoNonNullAssertedOptionalChainRule)
}

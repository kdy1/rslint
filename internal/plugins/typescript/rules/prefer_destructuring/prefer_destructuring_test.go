package prefer_destructuring

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferDestructuring(t *testing.T) {
	rule_tester.RunTestsFromTypeScriptEslint(t, "prefer-destructuring", PreferDestructuringRule)
}

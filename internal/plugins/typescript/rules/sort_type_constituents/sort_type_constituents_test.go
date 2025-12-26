package sort_type_constituents

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestSortTypeConstituentsRule(t *testing.T) {
	rule.RunRuleTestWrapper(t, SortTypeConstituentsRule)
}

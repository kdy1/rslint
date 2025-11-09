package no_unnecessary_qualifier

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestNoUnnecessaryQualifierRule(t *testing.T) {
	rule.TestRule(t, NoUnnecessaryQualifierRule)
}

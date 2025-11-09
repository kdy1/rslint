package no_loss_of_precision

import (
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoLossOfPrecisionRule implements the no-loss-of-precision rule
// Disallows number literals that lose precision at runtime
var NoLossOfPrecisionRule = rule.CreateRule(rule.Rule{
	Name: "no-loss-of-precision",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindNumericLiteral: func(node *ast.Node) {
			numLiteral := node.AsNumericLiteral()
			if numLiteral == nil {
				return
			}

			text := numLiteral.Text
			if text == "" {
				return
			}

			// Check if this numeric literal loses precision
			if lossesPrecision(text) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noLossOfPrecision",
					Description: "This number literal will lose precision at runtime.",
				})
			}
		},
	}
}

// lossesPrecision checks if a numeric literal string loses precision
func lossesPrecision(text string) bool {
	// Normalize the text (remove underscores used as separators)
	normalized := strings.ReplaceAll(text, "_", "")

	// Check different number formats
	if strings.HasPrefix(normalized, "0b") || strings.HasPrefix(normalized, "0B") {
		return checkBinaryPrecision(normalized[2:])
	}

	if strings.HasPrefix(normalized, "0o") || strings.HasPrefix(normalized, "0O") {
		return checkOctalPrecision(normalized[2:])
	}

	if strings.HasPrefix(normalized, "0x") || strings.HasPrefix(normalized, "0X") {
		return checkHexPrecision(normalized[2:])
	}

	// Check decimal numbers (including scientific notation)
	return checkDecimalPrecision(normalized)
}

// checkBinaryPrecision checks if a binary number loses precision
func checkBinaryPrecision(binary string) bool {
	// Binary literals are integers, check if they exceed safe integer range
	// JavaScript safe integer range is -(2^53 - 1) to (2^53 - 1)
	// For binary, we need to check if the number of bits exceeds 53

	// Remove leading zeros
	binary = strings.TrimLeft(binary, "0")
	if binary == "" {
		return false // Zero doesn't lose precision
	}

	// If more than 53 bits, it will lose precision
	if len(binary) > 53 {
		return true
	}

	// If exactly 53 bits, check if the value exceeds Number.MAX_SAFE_INTEGER
	if len(binary) == 53 {
		// Convert to integer and check
		val := new(big.Int)
		val.SetString(binary, 2)
		maxSafeInt := new(big.Int)
		maxSafeInt.SetInt64(9007199254740991) // 2^53 - 1
		return val.Cmp(maxSafeInt) > 0
	}

	return false
}

// checkOctalPrecision checks if an octal number loses precision
func checkOctalPrecision(octal string) bool {
	// Convert to big.Int and check against safe integer range
	val := new(big.Int)
	_, success := val.SetString(octal, 8)
	if !success {
		return false
	}

	maxSafeInt := new(big.Int)
	maxSafeInt.SetInt64(9007199254740991) // 2^53 - 1
	minSafeInt := new(big.Int)
	minSafeInt.SetInt64(-9007199254740991) // -(2^53 - 1)

	return val.Cmp(maxSafeInt) > 0 || val.Cmp(minSafeInt) < 0
}

// checkHexPrecision checks if a hexadecimal number loses precision
func checkHexPrecision(hex string) bool {
	// Convert to big.Int and check against safe integer range
	val := new(big.Int)
	_, success := val.SetString(hex, 16)
	if !success {
		return false
	}

	maxSafeInt := new(big.Int)
	maxSafeInt.SetInt64(9007199254740991) // 2^53 - 1
	minSafeInt := new(big.Int)
	minSafeInt.SetInt64(-9007199254740991) // -(2^53 - 1)

	return val.Cmp(maxSafeInt) > 0 || val.Cmp(minSafeInt) < 0
}

// checkDecimalPrecision checks if a decimal number (including scientific notation) loses precision
func checkDecimalPrecision(text string) bool {
	// Handle scientific notation
	scientificRegex := regexp.MustCompile(`^[+-]?(\d+\.?\d*|\.\d+)[eE][+-]?\d+$`)
	if scientificRegex.MatchString(text) {
		// Parse as float and check if conversion loses precision
		return checkFloatPrecision(text)
	}

	// Handle regular decimal numbers
	decimalRegex := regexp.MustCompile(`^[+-]?(\d+\.?\d*|\.\d+)$`)
	if !decimalRegex.MatchString(text) {
		return false
	}

	// Check if it's an integer
	if !strings.Contains(text, ".") && !strings.Contains(text, "e") && !strings.Contains(text, "E") {
		// It's an integer, check against safe integer range
		val := new(big.Int)
		_, success := val.SetString(text, 10)
		if !success {
			return false
		}

		maxSafeInt := new(big.Int)
		maxSafeInt.SetInt64(9007199254740991) // 2^53 - 1
		minSafeInt := new(big.Int)
		minSafeInt.SetInt64(-9007199254740991) // -(2^53 - 1)

		return val.Cmp(maxSafeInt) > 0 || val.Cmp(minSafeInt) < 0
	}

	// For floating point, check if conversion loses precision
	return checkFloatPrecision(text)
}

// checkFloatPrecision checks if a float loses precision when converted
func checkFloatPrecision(text string) bool {
	// Parse the number as a float64
	floatVal, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return false
	}

	// Check for infinity
	if math.IsInf(floatVal, 0) {
		return true
	}

	// Convert the float back to string and compare with original
	// If they don't match exactly, precision was lost

	// Use big.Float for higher precision comparison
	original := new(big.Float)
	original.SetPrec(1000) // High precision
	_, success := original.SetString(text)
	if !success {
		return false
	}

	// Convert float64 back to big.Float
	converted := new(big.Float)
	converted.SetPrec(1000)
	converted.SetFloat64(floatVal)

	// Compare the two
	return original.Cmp(converted) != 0
}

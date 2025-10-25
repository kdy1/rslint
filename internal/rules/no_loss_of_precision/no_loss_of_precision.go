package no_loss_of_precision

import (
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoLossOfPrecisionRule implements the no-loss-of-precision rule
// Disallow number literals that lose precision
var NoLossOfPrecisionRule = rule.Rule{
	Name: "no-loss-of-precision",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindNumericLiteral: func(node *ast.Node) {
			numLiteral := node.AsNumericLiteral()
			if numLiteral == nil {
				return
			}

			// Get the raw text of the numeric literal
			rng := utils.TrimNodeTextRange(ctx.SourceFile, node)
			rawText := ctx.SourceFile.Text()[rng.Pos():rng.End()]

			// Remove numeric separators (underscores) if present
			cleanText := strings.ReplaceAll(rawText, "_", "")

			// Check if this literal loses precision
			if losesPrecision(cleanText) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noLossOfPrecision",
					Description: "This number literal will lose precision at runtime.",
				})
			}
		},
	}
}

// losesPrecision checks if a numeric literal loses precision when converted to float64
func losesPrecision(text string) bool {
	// Handle different number formats
	if strings.HasPrefix(text, "0x") || strings.HasPrefix(text, "0X") {
		return checkHexPrecision(text)
	}
	if strings.HasPrefix(text, "0b") || strings.HasPrefix(text, "0B") {
		return checkBinaryPrecision(text)
	}
	if strings.HasPrefix(text, "0o") || strings.HasPrefix(text, "0O") {
		return checkOctalPrecision(text)
	}

	// Handle decimal numbers (including scientific notation)
	return checkDecimalPrecision(text)
}

// checkDecimalPrecision checks decimal numbers for precision loss
func checkDecimalPrecision(text string) bool {
	// Parse as float64
	floatVal, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return false
	}

	// Special case: infinity means the number is too large
	if math.IsInf(floatVal, 0) {
		return true
	}

	// Use big.Float for arbitrary precision comparison
	bigFloat := new(big.Float).SetPrec(1000) // High precision
	_, _, err = bigFloat.Parse(text, 10)
	if err != nil {
		return false
	}

	// Convert the float64 back to big.Float
	roundTrip := big.NewFloat(floatVal).SetPrec(1000)

	// Compare: if they're not equal, precision was lost
	return bigFloat.Cmp(roundTrip) != 0
}

// checkHexPrecision checks hexadecimal numbers for precision loss
func checkHexPrecision(text string) bool {
	// Remove 0x prefix
	hexStr := text[2:]

	// Parse as big.Int
	bigInt := new(big.Int)
	_, success := bigInt.SetString(hexStr, 16)
	if !success {
		return false
	}

	// Check if it fits in safe integer range (2^53 - 1)
	maxSafeInt := new(big.Int).SetInt64(9007199254740991) // 2^53 - 1
	if bigInt.Cmp(maxSafeInt) > 0 {
		return true
	}

	return false
}

// checkBinaryPrecision checks binary numbers for precision loss
func checkBinaryPrecision(text string) bool {
	// Remove 0b prefix
	binStr := text[2:]

	// Parse as big.Int
	bigInt := new(big.Int)
	_, success := bigInt.SetString(binStr, 2)
	if !success {
		return false
	}

	// Check if it fits in safe integer range
	maxSafeInt := new(big.Int).SetInt64(9007199254740991)
	if bigInt.Cmp(maxSafeInt) > 0 {
		return true
	}

	return false
}

// checkOctalPrecision checks octal numbers for precision loss
func checkOctalPrecision(text string) bool {
	// Remove 0o prefix
	octStr := text[2:]

	// Parse as big.Int
	bigInt := new(big.Int)
	_, success := bigInt.SetString(octStr, 8)
	if !success {
		return false
	}

	// Check if it fits in safe integer range
	maxSafeInt := new(big.Int).SetInt64(9007199254740991)
	if bigInt.Cmp(maxSafeInt) > 0 {
		return true
	}

	return false
}

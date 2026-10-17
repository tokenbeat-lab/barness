package ai

import (
	"encoding/json"
	"math/big"
	"strconv"
	"strings"
)

// Exact upper endpoints, with lexical sign/nonzero checks for tiny exponents.
// The common float fast path never expands decimal exponents; only a number
// rounding exactly to the small upper endpoint needs bounded big.Rat work.
func miniMaxSamplingNumber(raw any, maximum int64, zeroAllowed bool) bool {
	number, ok := raw.(json.Number)
	if !ok {
		return false
	}
	s := string(number)
	mantissa, _, _ := strings.Cut(strings.ToLower(s), "e")
	nonzero := strings.ContainsAny(mantissa, "123456789")
	if !nonzero {
		return zeroAllowed
	}
	if strings.HasPrefix(s, "-") {
		return false
	}
	approximate, e := strconv.ParseFloat(s, 64)
	if e != nil || approximate > float64(maximum) {
		return false
	}
	if approximate < float64(maximum) {
		return true
	}
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		exp, e := strconv.Atoi(s[i+1:])
		bound := len(s) + 16
		if e != nil || exp > bound || exp < -bound {
			return false
		}
	}
	value, ok := new(big.Rat).SetString(s)
	return ok && value.Sign() > 0 && value.Cmp(new(big.Rat).SetInt64(maximum)) <= 0
}

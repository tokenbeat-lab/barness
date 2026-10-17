package ai

import (
	"encoding/json"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// glmOutputBudgetNumber validates the exact final JSON number against an
// integer schema. float64 alone rounds fractional values such as
// 1.00000000000000001 to an integer. Chat and Responses share this constraint,
// but their field names and model limits remain at each protocol boundary.
func glmOutputBudgetNumber(raw json.RawMessage, limit int) bool {
	s := strings.TrimSpace(string(raw))
	if s == "" || (s[0] != '-' && (s[0] < '0' || s[0] > '9')) || !json.Valid(raw) {
		return false
	}
	approx, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsInf(approx, 0) || approx < 1 || approx > float64(limit) {
		return false
	}
	// Bound big.Rat's decimal expansion by the request's own digit count.
	// No positive value in our small output range needs an exponent larger
	// than its mantissa plus the limit's digits. This also rejects hostile
	// giant exponent spellings without allocating their expanded value.
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		exp, err := strconv.Atoi(s[i+1:])
		bound := len(s) + len(strconv.Itoa(limit))
		if err != nil || exp > bound || exp < -bound {
			return false
		}
	}
	exact, ok := new(big.Rat).SetString(s)
	return ok && exact.IsInt() && exact.Sign() > 0 && exact.Cmp(new(big.Rat).SetInt64(int64(limit))) <= 0
}

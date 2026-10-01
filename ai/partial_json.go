package ai

import (
	"errors"
	"math"
	"strings"
)

// parsePartialJSON ports the npm package partial-json 0.1.7 `parse` with
// Allow.ALL, which pi-ai's parseStreamingJson uses to show tool arguments
// while they stream. It recovers what it can from incomplete or malformed
// JSON: unterminated strings, objects, arrays and literals, number prefixes.
// Its quirks are kept, because they decide what pi shows and stores for a
// call cut off mid-arguments.
//
// One deliberate difference: the JavaScript version can loop forever on some
// malformed arrays (for example `[1e5}`), where re-parsing a number makes no
// progress. Here an element that consumes no input ends the array instead.
// Nesting is bounded like parseJSON.
func parsePartialJSON(s string) (any, error) {
	s = strings.TrimFunc(s, isJSSpace)
	if s == "" {
		return nil, errors.New("ai: empty partial JSON")
	}
	p := partialParser{s: s}
	return p.parseAny(0)
}

var (
	errPartialJSON   = errors.New("ai: partial JSON")
	errMalformedJSON = errors.New("ai: malformed JSON")
)

type partialParser struct {
	s string
	i int
}

// at is the byte at i, or 0 past the end (JavaScript's undefined).
func (p *partialParser) at(i int) byte {
	if i < len(p.s) {
		return p.s[i]
	}
	return 0
}

func (p *partialParser) skipBlank() {
	for p.i < len(p.s) && strings.IndexByte(" \n\r\t", p.s[p.i]) >= 0 {
		p.i++
	}
}

// literal matches a keyword, or a prefix of it that ends the input.
func (p *partialParser) literal(word string, minRest int) bool {
	rest := p.s[p.i:]
	if strings.HasPrefix(rest, word) || (len(rest) < len(word) && len(rest) > minRest && strings.HasPrefix(word, rest)) {
		p.i += len(word)
		return true
	}
	return false
}

func (p *partialParser) parseAny(depth int) (any, error) {
	if depth > maxJSONDepth {
		return nil, errJSONDepth
	}
	p.skipBlank()
	if p.i >= len(p.s) {
		return nil, errPartialJSON
	}
	switch p.s[p.i] {
	case '"':
		return p.parseStr()
	case '{':
		return p.parseObj(depth)
	case '[':
		return p.parseArr(depth)
	}
	switch {
	case p.literal("null", 0):
		return nil, nil
	case p.literal("true", 0):
		return true, nil
	case p.literal("false", 0):
		return false, nil
	case p.literal("Infinity", 0):
		return math.Inf(1), nil
	case p.literal("-Infinity", 1):
		return math.Inf(-1), nil
	case p.literal("NaN", 0):
		return math.NaN(), nil
	}
	return p.parseNum()
}

func (p *partialParser) parseStr() (any, error) {
	start := p.i
	escape := false
	p.i++ // opening quote
	for p.i < len(p.s) && (p.s[p.i] != '"' || (escape && p.s[p.i-1] == '\\')) {
		escape = p.s[p.i] == '\\' && !escape
		p.i++
	}
	esc := 0
	if escape {
		esc = 1
	}
	if p.at(p.i) == '"' {
		p.i++
		v, err := parseJSON(p.s[start : p.i-esc])
		if err != nil {
			return nil, errMalformedJSON
		}
		return v, nil
	}
	// Unterminated: close it, dropping a dangling escape.
	if v, err := parseJSON(p.s[start:p.i-esc] + `"`); err == nil {
		return v, nil
	}
	return parseJSON(jsSubstring(p.s, start, strings.LastIndexByte(p.s, '\\')) + `"`)
}

func (p *partialParser) parseObj(depth int) (any, error) {
	p.i++ // opening brace
	p.skipBlank()
	obj := newJSONObject()
	for p.at(p.i) != '}' {
		p.skipBlank()
		if p.i >= len(p.s) {
			return obj, nil
		}
		key, err := p.parseStr()
		if err != nil {
			return obj, nil
		}
		name, ok := key.(string)
		if !ok {
			return obj, nil
		}
		p.skipBlank()
		p.i++ // colon
		v, err := p.parseAny(depth + 1)
		if err != nil {
			return obj, nil
		}
		obj.set(name, v)
		p.skipBlank()
		if p.at(p.i) == ',' {
			p.i++
		}
	}
	p.i++ // closing brace
	return obj, nil
}

func (p *partialParser) parseArr(depth int) (any, error) {
	p.i++ // opening bracket
	arr := []any{}
	for p.at(p.i) != ']' {
		before := p.i
		v, err := p.parseAny(depth + 1)
		if err != nil {
			return arr, nil
		}
		arr = append(arr, v)
		p.skipBlank()
		if p.at(p.i) == ',' {
			p.i++
		}
		if p.i == before {
			return arr, nil // no progress; see parsePartialJSON
		}
	}
	p.i++ // closing bracket
	return arr, nil
}

func (p *partialParser) parseNum() (any, error) {
	if p.i == 0 {
		if p.s == "-" {
			return nil, errMalformedJSON
		}
		if v, err := parseJSON(p.s); err == nil {
			return v, nil
		}
		if v, err := parseJSON(jsSubstring(p.s, 0, strings.LastIndexByte(p.s, 'e'))); err == nil {
			return v, nil
		}
		return nil, errMalformedJSON
	}
	start := p.i
	if p.s[p.i] == '-' {
		p.i++
	}
	for p.i < len(p.s) && strings.IndexByte(",]}", p.s[p.i]) < 0 {
		p.i++
	}
	if v, err := parseJSON(p.s[start:p.i]); err == nil {
		return v, nil
	}
	if p.s[start:p.i] == "-" {
		return nil, errPartialJSON
	}
	if v, err := parseJSON(jsSubstring(p.s, start, strings.LastIndexByte(p.s, 'e'))); err == nil {
		return v, nil
	}
	return nil, errMalformedJSON
}

// jsSubstring is JavaScript's String#substring: negative bounds are 0, bounds
// past the end are the length, and swapped bounds are swapped back.
func jsSubstring(s string, a, b int) string {
	clamp := func(n int) int { return max(0, min(n, len(s))) }
	a, b = clamp(a), clamp(b)
	if a > b {
		a, b = b, a
	}
	return s[a:b]
}

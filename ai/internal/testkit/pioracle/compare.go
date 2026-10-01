package pioracle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"math/big"
	"regexp"
	"sort"
	"strings"
)

// Observation is what one implementation did for a case, in pi-ai's public
// JSON shape: the request the local controlled Provider received, every event
// in order, and the final assistant message.
type Observation struct {
	Request Request           `json:"request"`
	Events  []json.RawMessage `json:"events"`
	Result  json.RawMessage   `json:"result"`
}

// Request is the captured inference request. Auth is the alias of the
// credential the Provider received (the authentication source); Headers are
// redacted at capture, so they never hold key material.
type Request struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   string            `json:"query"`
	Auth    string            `json:"auth"`
	Headers map[string]string `json:"headers"`
	Body    json.RawMessage   `json:"body"`
}

// DiffKind says how the two sides differ at a path.
type DiffKind string

const (
	Changed     DiffKind = "changed"
	OnlyPi      DiffKind = "only_pi"
	OnlyBarness DiffKind = "only_barness"
)

// Diff is one difference. Path is a JSON path such as
// "request.body.input[0].role" or "events[3].delta"; a missing or extra event
// is reported at its position, e.g. "events[5]".
type Diff struct {
	Path    string          `json:"path"`
	Kind    DiffKind        `json:"kind"`
	Pi      json.RawMessage `json:"pi,omitempty"`
	Barness json.RawMessage `json:"barness,omitempty"`
}

// timePaths are the generated-time fields mapped one-to-one before
// comparison: pi stamps every assistant message (final, terminal event, live
// partial) with its creation time. Each distinct value becomes an ordinal token
// in first-appearance order, so equal times stay equal (references are kept)
// while the clock itself does not matter. Only these explicit paths are
// mapped; any other field named timestamp is compared as is, and generated IDs
// join this list explicitly when a case first produces one client-side.
var timePaths = regexp.MustCompile(`^(result|events\[\d+\]\.(message|error|partial))\.timestamp$`)

// Compare reports every difference between pi's and barness's observations in
// a fixed order: request (method, path, query, auth, headers, body), events,
// result. Only JSON object keys are put in semantic order; arrays and events
// keep theirs. null, zero, empty arrays, unknown fields and IDs are compared
// as they are. Numbers compare by exact decimal value.
func Compare(pi, barness Observation) ([]Diff, error) {
	ps, err := sections(pi)
	if err != nil {
		return nil, fmt.Errorf("pi observation: %w", err)
	}
	bs, err := sections(barness)
	if err != nil {
		return nil, fmt.Errorf("barness observation: %w", err)
	}
	c := comparer{diffs: []Diff{}}
	for i := range ps {
		c.walk(ps[i].path, ps[i].value, bs[i].value, true, true)
	}
	return c.diffs, nil
}

type section struct {
	path  string
	value any
}

func sections(o Observation) ([]section, error) {
	body, err := decode(o.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("request body: %w", err)
	}
	headers := map[string]any{}
	for k, v := range o.Request.Headers {
		headers[k] = v
	}
	events := make([]any, len(o.Events))
	for i, raw := range o.Events {
		if events[i], err = decode(raw); err != nil {
			return nil, fmt.Errorf("event %d: %w", i, err)
		}
	}
	result, err := decode(o.Result)
	if err != nil {
		return nil, fmt.Errorf("result: %w", err)
	}
	out := []section{
		{"request.method", o.Request.Method},
		{"request.path", o.Request.Path},
		{"request.query", o.Request.Query},
		{"request.auth", o.Request.Auth},
		{"request.headers", headers},
		{"request.body", body},
		{"events", events},
		{"result", result},
	}
	times := map[string]int{}
	for i := range out {
		out[i].value = mapTimes(out[i].path, out[i].value, times)
	}
	return out, nil
}

func decode(raw json.RawMessage) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if dec.More() {
		return nil, fmt.Errorf("trailing data after JSON value")
	}
	return v, nil
}

// timeToken is a mapped time value; it is a distinct type so it can never
// equal a real string.
type timeToken int

func (t timeToken) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("<time#%d>", int(t)))
}

func mapTimes(path string, v any, seen map[string]int) any {
	switch v := v.(type) {
	case map[string]any:
		for _, k := range sortedKeys(v) {
			v[k] = mapTimes(path+"."+k, v[k], seen)
		}
	case []any:
		for i := range v {
			v[i] = mapTimes(fmt.Sprintf("%s[%d]", path, i), v[i], seen)
		}
	case json.Number:
		if !timePaths.MatchString(path) {
			return v
		}
		key := canonicalNumber(v)
		if _, ok := seen[key]; !ok {
			seen[key] = len(seen) + 1
		}
		return timeToken(seen[key])
	}
	return v
}

type comparer struct {
	diffs []Diff
}

func (c *comparer) walk(path string, p, b any, hasP, hasB bool) {
	switch {
	case hasP && !hasB:
		c.add(path, OnlyPi, p, nil)
		return
	case !hasP && hasB:
		c.add(path, OnlyBarness, nil, b)
		return
	}
	switch pv := p.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok {
			c.add(path, Changed, p, b)
			return
		}
		union := maps.Clone(pv)
		maps.Copy(union, bv)
		for _, k := range sortedKeys(union) {
			x, okP := pv[k]
			y, okB := bv[k]
			c.walk(path+"."+k, x, y, okP, okB)
		}
	case []any:
		bv, ok := b.([]any)
		if !ok {
			c.add(path, Changed, p, b)
			return
		}
		for i := 0; i < max(len(pv), len(bv)); i++ {
			var x, y any
			if i < len(pv) {
				x = pv[i]
			}
			if i < len(bv) {
				y = bv[i]
			}
			c.walk(fmt.Sprintf("%s[%d]", path, i), x, y, i < len(pv), i < len(bv))
		}
	default:
		if !scalarEqual(p, b) {
			c.add(path, Changed, p, b)
		}
	}
}

func scalarEqual(p, b any) bool {
	pn, okP := p.(json.Number)
	bn, okB := b.(json.Number)
	if okP || okB {
		return okP && okB && canonicalNumber(pn) == canonicalNumber(bn)
	}
	// Remaining scalars are strings, bools, nil and time tokens; maps and
	// slices never reach here on the pi side, and a container on the barness
	// side is not comparable to a scalar.
	switch b.(type) {
	case map[string]any, []any:
		return false
	}
	return p == b
}

// canonicalNumber renders a JSON number as an exact rational so 1, 1.0 and
// 1e0 agree while 2^53+1 and 2^53 do not.
func canonicalNumber(n json.Number) string {
	r, ok := new(big.Rat).SetString(string(n))
	if !ok {
		return "invalid:" + string(n)
	}
	return r.RatString()
}

func (c *comparer) add(path string, kind DiffKind, p, b any) {
	d := Diff{Path: path, Kind: kind}
	if kind != OnlyBarness {
		d.Pi = mustJSON(p)
	}
	if kind != OnlyPi {
		d.Barness = mustJSON(b)
	}
	c.diffs = append(c.diffs, d)
}

func mustJSON(v any) json.RawMessage {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return json.RawMessage(`"<unencodable>"`)
	}
	return json.RawMessage(strings.TrimSuffix(buf.String(), "\n"))
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

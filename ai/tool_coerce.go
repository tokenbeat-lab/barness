package ai

import (
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

// This file ports the schema conversion of pi-ai's validateToolArguments
// (utils/validation.ts) for plain JSON schemas: normalizeOptionalNulls, then
// coerceWithJsonSchema. pi first also runs TypeBox's generic Value.Convert;
// barness does not reproduce it, and relies on pi's own JSON-schema rules,
// which pi applies to every non-TypeBox schema (the only kind a Go host can
// pass). Schemas here are decoded JSON (map[string]any, []any, json.Number);
// values are the JSON values of tool_json.go and are converted in place.

// schemaTypes is getSchemaTypes: the schema's "type" as a list.
func schemaTypes(schema map[string]any) []string {
	switch t := schema["type"].(type) {
	case string:
		return []string{t}
	case []any:
		var out []string
		for _, e := range t {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func matchesJSONType(v any, typ string) bool {
	switch typ {
	case "number":
		_, ok := v.(float64)
		return ok
	case "integer":
		f, ok := v.(float64)
		return ok && f == math.Trunc(f) && !math.IsInf(f, 0)
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "string":
		_, ok := v.(string)
		return ok
	case "null":
		return v == nil
	case "array":
		_, ok := v.([]any)
		return ok
	case "object":
		_, ok := v.(*jsonObject)
		return ok
	}
	return false
}

// coercePrimitive is coercePrimitiveByType. It returns v itself when it does
// not apply.
func coercePrimitive(v any, typ string) any {
	switch typ {
	case "number", "integer":
		switch v := v.(type) {
		case nil:
			return 0.0
		case string:
			if strings.TrimFunc(v, isJSSpace) != "" {
				if f, ok := jsStringToNumber(v); ok && !math.IsInf(f, 0) && !math.IsNaN(f) && (typ == "number" || f == math.Trunc(f)) {
					return f
				}
			}
		case bool:
			if v {
				return 1.0
			}
			return 0.0
		}
	case "boolean":
		switch v := v.(type) {
		case nil:
			return false
		case string:
			if v == "true" || v == "false" {
				return v == "true"
			}
		case float64:
			if v == 1 || v == 0 {
				return v == 1
			}
		}
	case "string":
		switch v := v.(type) {
		case nil:
			return ""
		case float64:
			return jsNumberString(v)
		case bool:
			return strconv.FormatBool(v)
		}
	case "null":
		if v == "" || v == 0.0 || v == false {
			return nil
		}
	}
	return v
}

// primitiveReplaced is JavaScript's `before !== after` for coercePrimitive's
// result: composites are never replaced by it, so only primitives can differ.
func primitiveReplaced(before, after any) bool {
	switch before.(type) {
	case []any, *jsonObject:
		return false
	}
	switch after.(type) {
	case []any, *jsonObject:
		return true
	}
	return before != after
}

// coerceWithSchema is coerceWithJsonSchema.
func (s *toolSchema) coerceWithSchema(v any, schemaNode any) any {
	schema, ok := schemaNode.(map[string]any)
	if !ok {
		return v
	}
	next := v
	if all, ok := schema["allOf"].([]any); ok {
		for _, nested := range all {
			next = s.coerceWithSchema(next, nested)
		}
	}
	if anyOf, ok := schema["anyOf"].([]any); ok {
		next = s.coerceUnion(next, anyOf)
	}
	if oneOf, ok := schema["oneOf"].([]any); ok {
		next = s.coerceUnion(next, oneOf)
	}
	types := schemaTypes(schema)
	matchesMember := len(types) > 1 && slices.ContainsFunc(types, func(t string) bool { return matchesJSONType(next, t) })
	if len(types) > 0 && !matchesMember {
		for _, t := range types {
			if candidate := coercePrimitive(next, t); primitiveReplaced(next, candidate) {
				next = candidate
				break
			}
		}
	}
	if obj, ok := next.(*jsonObject); ok && slices.Contains(types, "object") {
		s.coerceObject(obj, schema)
	}
	if arr, ok := next.([]any); ok && slices.Contains(types, "array") {
		s.coerceArray(arr, schema)
	}
	return next
}

func (s *toolSchema) coerceObject(obj *jsonObject, schema map[string]any) {
	props, _ := schema["properties"].(map[string]any)
	for key, prop := range props {
		if v, ok := obj.get(key); ok {
			obj.set(key, s.coerceWithSchema(v, prop))
		}
	}
	if extra, ok := schema["additionalProperties"].(map[string]any); ok {
		for _, key := range slices.Clone(obj.keys) {
			if _, defined := props[key]; !defined {
				obj.set(key, s.coerceWithSchema(obj.values[key], extra))
			}
		}
	}
}

func (s *toolSchema) coerceArray(arr []any, schema map[string]any) {
	switch items := schema["items"].(type) {
	case []any:
		for i := range arr {
			if i < len(items) {
				arr[i] = s.coerceWithSchema(arr[i], items[i])
			}
		}
	case map[string]any:
		for i := range arr {
			arr[i] = s.coerceWithSchema(arr[i], items)
		}
	}
}

// coerceUnion is coerceWithUnionSchema: keep a value an arm accepts, else
// take the first arm whose conversion it then satisfies.
func (s *toolSchema) coerceUnion(v any, arms []any) any {
	for _, arm := range arms {
		if s.accepts(arm, v) == schemaAccepts {
			return v
		}
	}
	for _, arm := range arms {
		coerced := s.coerceWithSchema(cloneJSON(v), arm)
		if s.accepts(arm, coerced) == schemaAccepts {
			return coerced
		}
	}
	return v
}

// normalizeOptionalNulls drops a null given for an optional property whose
// own schema rejects null, as pi treats it as an omission.
func (s *toolSchema) normalizeOptionalNulls(v any, schemaNode any) {
	schema, ok := schemaNode.(map[string]any)
	if !ok {
		return
	}
	if arr, ok := v.([]any); ok {
		switch items := schema["items"].(type) {
		case []any:
			for i := range arr {
				if i < len(items) {
					s.normalizeOptionalNulls(arr[i], items[i])
				}
			}
		case map[string]any:
			for _, e := range arr {
				s.normalizeOptionalNulls(e, items)
			}
		}
		return
	}
	obj, ok := v.(*jsonObject)
	props, hasProps := schema["properties"].(map[string]any)
	if !ok || !hasProps {
		return
	}
	required := map[string]bool{}
	if req, ok := schema["required"].([]any); ok {
		for _, r := range req {
			if name, ok := r.(string); ok {
				required[name] = true
			}
		}
	}
	for key, prop := range props {
		val, present := obj.get(key)
		if !present {
			continue
		}
		propSchema, _ := prop.(map[string]any)
		_, isRef := propSchema["$ref"].(string)
		if val == nil && !required[key] && !isRef && s.accepts(prop, nil) == schemaRejects {
			obj.remove(key)
		} else {
			s.normalizeOptionalNulls(val, prop)
		}
	}
}

func cloneJSON(v any) any {
	switch v := v.(type) {
	case []any:
		out := make([]any, len(v))
		for i, e := range v {
			out[i] = cloneJSON(e)
		}
		return out
	case *jsonObject:
		out := newJSONObject()
		for _, k := range v.keys {
			out.set(k, cloneJSON(v.values[k]))
		}
		return out
	}
	return v
}

// jsDecimal is JavaScript's StrDecimalLiteral without Infinity.
var jsDecimal = regexp.MustCompile(`^[+-]?(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?$`)

// jsStringToNumber is JavaScript's Number(string) for a non-blank string.
func jsStringToNumber(s string) (float64, bool) {
	s = strings.TrimFunc(s, isJSSpace)
	switch {
	case s == "Infinity" || s == "+Infinity":
		return math.Inf(1), true
	case s == "-Infinity":
		return math.Inf(-1), true
	case len(s) > 2 && s[0] == '0' && strings.ContainsRune("xXoObB", rune(s[1])):
		base := map[byte]int{'x': 16, 'o': 8, 'b': 2}[s[1]|0x20]
		n, err := strconv.ParseUint(s[2:], base, 64)
		if err != nil {
			return 0, false
		}
		return float64(n), true
	case jsDecimal.MatchString(s):
		f, err := jsNumberLiteral(s)
		return f, err == nil
	}
	return 0, false
}

// jsNumberString is JavaScript's String(number).
func jsNumberString(f float64) string {
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	}
	var b strings.Builder
	writeJSNumber(&b, f)
	return b.String()
}

// isJSSpace is the whitespace JavaScript trims: Unicode spaces and the BOM.
func isJSSpace(r rune) bool { return unicode.IsSpace(r) || r == 0xFEFF }

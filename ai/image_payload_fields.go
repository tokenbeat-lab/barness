package ai

import (
	"reflect"
	"strings"
	"unicode"
)

type imageJSONField struct {
	name              string
	index             []int
	options           string
	tagged, ambiguous bool
}

// Resolve ordinary JSON fields once per struct inspection. Resource fields
// and content traversal share encoding/json's promotion/dominance rules;
// ambiguous fields are omitted, never interpreted as image occurrences.
func imageJSONFields(t reflect.Type) []imageJSONField {
	selected := map[string]imageJSONField{}
	visiting := map[reflect.Type]bool{}
	var walk func(reflect.Type, []int)
	walk = func(t reflect.Type, path []int) {
		if visiting[t] {
			return
		}
		visiting[t] = true
		defer delete(visiting, t)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			name, options, _ := strings.Cut(field.Tag.Get("json"), ",")
			if name == "-" {
				continue
			}
			if !imageJSONTagName(name) {
				name = ""
			}
			embedded := field.Type
			if embedded.Kind() == reflect.Pointer {
				embedded = embedded.Elem()
			}
			if field.PkgPath != "" && (!field.Anonymous || embedded.Kind() != reflect.Struct) {
				continue
			}
			index := append(append([]int(nil), path...), i)
			if field.Anonymous && name == "" && embedded.Kind() == reflect.Struct {
				walk(embedded, index)
				continue
			}
			tagged := name != ""
			if name == "" {
				name = field.Name
			}
			candidate := imageJSONField{name: name, index: index, options: options, tagged: tagged}
			previous, exists := selected[name]
			if !exists || len(index) < len(previous.index) || (len(index) == len(previous.index) && tagged && !previous.tagged) {
				selected[name] = candidate
			} else if len(index) == len(previous.index) && tagged == previous.tagged {
				previous.ambiguous = true
				selected[name] = previous
			}
		}
	}
	walk(t, nil)
	out := make([]imageJSONField, 0, len(selected))
	for _, field := range selected {
		if !field.ambiguous {
			out = append(out, field)
		}
	}
	return out
}

// Match encoding/json's allowed tag-name characters; invalid tags use the
// exported Go name. In particular, quotes and backslashes are not JSON names.
func imageJSONTagName(name string) bool {
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", r) {
			return false
		}
	}
	return name != ""
}

func (f imageJSONField) value(v reflect.Value) reflect.Value {
	field, err := v.FieldByIndexErr(f.index)
	if err != nil {
		return reflect.Value{}
	} // Nil anonymous pointer: omitted.
	for _, option := range strings.Split(f.options, ",") {
		if option == "omitzero" {
			zeroer := reflect.TypeFor[interface{ IsZero() bool }]()
			if field.Type().Implements(zeroer) || reflect.PointerTo(field.Type()).Implements(zeroer) {
				// Host methods execute only at the encoder boundary. Their field
				// presence is unknown until the independent final native guard.
				return reflect.Value{}
			}
			if field.IsZero() {
				return reflect.Value{}
			}
		}
		if option == "omitempty" {
			switch field.Kind() {
			case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
				if field.Len() == 0 {
					return reflect.Value{}
				}
			case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Interface, reflect.Pointer:
				if field.IsZero() {
					return reflect.Value{}
				}
			}
		}
	}
	return field
}

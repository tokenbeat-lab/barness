package livegrant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"syscall"
)

// Admission decoding stays independent of native report packages, which depend
// on this package. Reject ambiguous spelling, duplicate fields and null bounds
// before encoding/json can erase those distinctions.
func grantJSON(raw []byte) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	fields := func(t reflect.Type) map[string]bool {
		out := map[string]bool{}
		for i := 0; i < t.NumField(); i++ {
			out[strings.Split(t.Field(i).Tag.Get("json"), ",")[0]] = true
		}
		return out
	}
	var object func(map[string]bool, bool) error
	object = func(allowed map[string]bool, scope bool) error {
		t, e := d.Token()
		if e != nil || t != json.Delim('{') {
			return fmt.Errorf("grant object required")
		}
		seen := map[string]bool{}
		for d.More() {
			k, e := d.Token()
			if e != nil {
				return e
			}
			name, ok := k.(string)
			if !ok || seen[name] || !allowed[name] {
				return fmt.Errorf("duplicate or noncanonical grant field")
			}
			seen[name] = true
			if !scope && name == "scope" {
				if e = object(fields(reflect.TypeFor[Scope]()), true); e != nil {
					return e
				}
				continue
			}
			var value json.RawMessage
			if e = d.Decode(&value); e != nil {
				return e
			}
			if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				return fmt.Errorf("null grant field forbidden")
			}
		}
		_, e = d.Token()
		return e
	}
	if e := object(fields(reflect.TypeFor[Grant]()), false); e != nil {
		return e
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return fmt.Errorf("trailing grant input")
	}
	return nil
}

func ReadFile(path string) ([]byte, error) {
	f, e := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	s, e := f.Stat()
	if e != nil || !s.Mode().IsRegular() || s.Size() > 1<<20 {
		return nil, fmt.Errorf("grant file must be regular and bounded")
	}
	raw, e := io.ReadAll(io.LimitReader(f, 1<<20+1))
	if e != nil || len(raw) > 1<<20 {
		return nil, fmt.Errorf("grant read exceeds bound")
	}
	return raw, nil
}

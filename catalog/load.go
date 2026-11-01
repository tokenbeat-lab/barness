package catalog

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

// LoadFile reads exactly the selected file. A failed read or load never falls
// back to built-in data. Host code can explicitly make that separate choice.
func LoadFile(path string) (ai.Catalog, Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ai.Catalog{}, Report{}, &FileError{FileReadFailed, "cannot read selected local catalog"}
	}
	return Load(raw)
}

// Load checks complete source integrity before isolating model definitions.
// Source bytes, report bytes and the returned catalog have independent storage.
// A model failure is reported without a fatal error, including when none load.
func Load(raw []byte) (ai.Catalog, Report, error) {
	body, sourceHash, err := envelope(raw)
	if err != nil {
		return ai.Catalog{}, Report{}, err
	}
	var source struct {
		Version          string            `json:"version"`
		Models           []json.RawMessage `json:"models"`
		ImageModels      []json.RawMessage `json:"imageModels"`
		ClassifierModels []json.RawMessage `json:"classifierModels"`
	}
	if err := strictDecode(body, &source); err != nil {
		return ai.Catalog{}, Report{}, &FileError{FileInvalidStructure, err.Error()}
	}
	fields, _ := objectFields(body)
	for _, collection := range []string{"models", "imageModels", "classifierModels"} {
		value, present := fields[collection]
		if !present && collection != "models" {
			continue
		}
		if !present || len(bytes.TrimSpace(value)) == 0 || bytes.TrimSpace(value)[0] != '[' {
			return ai.Catalog{}, Report{}, &FileError{FileInvalidStructure, "catalog collections must be arrays; models is required"}
		}
	}
	if strings.TrimSpace(source.Version) == "" || source.Version != strings.TrimSpace(source.Version) {
		return ai.Catalog{}, Report{}, &FileError{FileInvalidVersion, "catalog version must be a nonempty string without surrounding whitespace"}
	}
	effective := ai.Catalog{Version: source.Version, Models: []ai.Model{}}
	report := Report{SourceVersion: source.Version, SourceHash: sourceHash, Models: []ModelStatus{}}
	groups := []struct {
		name    string
		op      ai.Operation
		entries []json.RawMessage
	}{
		{"models", ai.OperationChat, source.Models},
		{"imageModels", ai.OperationImage, source.ImageModels},
		{"classifierModels", ai.OperationClassifier, source.ClassifierModels},
	}
	identities := map[Identity]int{}
	for _, group := range groups {
		for _, entry := range group.entries {
			id := partialIdentity(group.op, entry)
			if id.Provider != "" && id.API != "" && id.ID != "" {
				identities[id]++
			}
		}
	}
	for _, group := range groups {
		for i, entry := range group.entries {
			status := ModelStatus{Identity: partialIdentity(group.op, entry), Collection: group.name, Index: i, Status: Loaded}
			var single ai.Catalog
			duplicate := identities[status.Identity] > 1
			switch group.op {
			case ai.OperationChat:
				var model ai.Model
				err = strictModel(entry, &model)
				single.Models = []ai.Model{model}
			case ai.OperationImage:
				var model ai.ImageModel
				err = strictModel(entry, &model)
				single.ImageModels = []ai.ImageModel{model}
			case ai.OperationClassifier:
				var model ai.ClassifierModel
				err = strictModel(entry, &model)
				single.ClassifierModels = []ai.ClassifierModel{model}
			}
			if err == nil {
				err = single.ValidateDefinitions()
			}
			if duplicate {
				err = fmt.Errorf("all definitions sharing this complete identity are rejected")
			}
			if err != nil {
				status.Status, status.Reason = Failed, InvalidDefinition
				var unknown *unknownFieldError
				if errors.As(err, &unknown) {
					status.Reason = UnknownField
				}
				var unsupported *ai.UnsupportedDefinitionError
				if errors.As(err, &unsupported) {
					status.Reason = UnsupportedDefinition
				}
				if duplicate {
					status.Reason = DuplicateIdentity
				}
				status.Diagnostic = err.Error()
				status.Suggestion = "Repair this model definition and reload the catalog."
				if status.Reason == UnknownField || status.Reason == UnsupportedDefinition {
					status.Suggestion = "Check the catalog constraints or try a program version that understands this field."
				}
				status.Raw = bytes.Clone(entry)
				report.Failed++
			} else {
				effective.Models = append(effective.Models, single.Models...)
				effective.ImageModels = append(effective.ImageModels, single.ImageModels...)
				effective.ClassifierModels = append(effective.ClassifierModels, single.ClassifierModels...)
				report.Loaded++
			}
			report.Models = append(report.Models, status)
		}
	}
	report.EffectiveHash, err = effective.Hash()
	if err != nil {
		return ai.Catalog{}, Report{}, &FileError{FileInvalidStructure, "effective catalog could not be assembled"}
	}
	return effective, report, nil
}

func envelope(raw []byte) ([]byte, string, error) {
	if !json.Valid(raw) {
		return nil, "", &FileError{FileInvalidJSON, "source must contain exactly one valid JSON value"}
	}
	var wrapper struct {
		Hash    string          `json:"hash"`
		Catalog json.RawMessage `json:"catalog"`
	}
	if len(bytes.TrimSpace(raw)) == 0 || bytes.TrimSpace(raw)[0] != '{' {
		return nil, "", &FileError{FileInvalidStructure, "source root must be an object"}
	}
	if err := strictDecode(raw, &wrapper); err != nil {
		return nil, "", &FileError{FileInvalidStructure, err.Error()}
	}
	if len(wrapper.Hash) != len("sha256:")+64 || !strings.HasPrefix(wrapper.Hash, "sha256:") {
		return nil, "", &FileError{FileInvalidHash, "hash must be a sha256 digest"}
	}
	if _, err := hex.DecodeString(strings.TrimPrefix(wrapper.Hash, "sha256:")); err != nil {
		return nil, "", &FileError{FileInvalidHash, "hash must be a sha256 digest"}
	}
	if len(wrapper.Catalog) == 0 || bytes.TrimSpace(wrapper.Catalog)[0] != '{' {
		return nil, "", &FileError{FileInvalidStructure, "catalog must be an object"}
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, wrapper.Catalog); err != nil {
		return nil, "", &FileError{FileInvalidJSON, "catalog is invalid JSON"}
	}
	sum := sha256.Sum256(compact.Bytes())
	if wrapper.Hash != "sha256:"+hex.EncodeToString(sum[:]) {
		return nil, "", &FileError{FileHashMismatch, "complete source catalog does not match hash"}
	}
	return wrapper.Catalog, wrapper.Hash, nil
}

func strictDecode(raw []byte, value any) error {
	if err := jsonShape(raw, reflect.TypeOf(value).Elem()); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(value)
}
func strictModel(raw []byte, value any) error {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.TrimSpace(raw)[0] != '{' {
		return fmt.Errorf("model definition must be an object")
	}
	return strictDecode(raw, value)
}
func partialIdentity(op ai.Operation, raw []byte) Identity {
	identity := Identity{Operation: op}
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) == nil {
		_ = json.Unmarshal(fields["provider"], &identity.Provider)
		_ = json.Unmarshal(fields["api"], &identity.API)
		_ = json.Unmarshal(fields["id"], &identity.ID)
	}
	return identity
}

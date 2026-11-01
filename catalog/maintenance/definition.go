package maintenance

import (
	"reflect"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/catalog/internal/jsonshape"
)

// Validate raw fields before mapping/typed canonicalization can turn invalid
// nulls into zero values. This deliberately does not impose protocol support;
// unsupported routes still become inspectable code-gap candidates at check.
func validateDefinitionJSON(raw []byte, operation ai.Operation) error {
	var typ reflect.Type
	switch operation {
	case ai.OperationChat:
		typ = reflect.TypeFor[ai.Model]()
	case ai.OperationImage:
		typ = reflect.TypeFor[ai.ImageModel]()
	case ai.OperationClassifier:
		typ = reflect.TypeFor[ai.ClassifierModel]()
	default:
		return failure("unsupported_definition", "unknown model operation")
	}
	return jsonshape.Validate(raw, typ)
}

func DecodeDecisions(raw []byte) (map[string]FieldDecision, error) {
	if err := jsonshape.Validate(raw, reflect.TypeFor[map[string]FieldDecision]()); err != nil {
		return nil, failure("invalid_definition", "field decisions: %v", err)
	}
	var decisions map[string]FieldDecision
	if err := decode(raw, &decisions); err != nil || decisions == nil {
		return nil, failure("invalid_definition", "field decisions must be an object")
	}
	return decisions, nil
}

package release

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// The registry is documentation for independent extensions, never an oracle
// exemption. Pin its normalized JSON so case/path deletion cannot widen scope.
const approvedClassifierExtensions = "ef8d021542131d0c3e1c900c7685b4397e32bd22c5465c054e667ab4a27d72ac"

func CheckClassifierExtensions(path string) (Snapshot, []byte) {
	s := Snapshot{Name: "classifier-extension-plan", Detail: "finite case/path registry missing or changed"}
	raw, err := readArtifact(path, 64<<10)
	if err != nil {
		return s, nil
	}
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return s, nil
	}
	normalized, err := json.Marshal(value)
	if err != nil || pioracle.SHA256(normalized) != approvedClassifierExtensions {
		return s, nil
	}
	s.Current, s.Detail = true, "finite extension case/path registry matches approved Pi 1.1.0 contract; never differential PASS"
	return s, raw
}

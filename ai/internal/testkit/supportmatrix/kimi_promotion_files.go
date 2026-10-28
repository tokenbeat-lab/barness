package supportmatrix

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func kimiPromotionPath(root, path string, inputs []string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	base := filepath.Join(root, ".evidence/barness-ai-kimi/promotion")
	within := func(a, b string) bool {
		rel, e := filepath.Rel(a, b)
		return e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
	}
	if path == base || !within(base, path) {
		return "", refuse("Kimi receipt requires a new Kimi promotion .evidence location")
	}
	for _, input := range inputs {
		input, e := filepath.Abs(input)
		if e != nil || within(input, path) {
			return "", refuse("Kimi receipt cannot mutate a candidate")
		}
	}
	for at := filepath.Dir(path); ; at = filepath.Dir(at) {
		st, e := os.Lstat(at)
		if e != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
			return "", refuse("Kimi receipt parent must be an existing real directory")
		}
		if at == filepath.Dir(at) {
			break
		}
	}
	return path, nil
}

func WriteKimiControlledProposal(dirs []string, path string) error {
	root, err := minimaxlive.Root()
	if err != nil {
		return err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		return err
	}
	path = filepath.Join(parent, filepath.Base(path))
	for _, forbidden := range append([]string{filepath.Join(root, "ai")}, dirs...) {
		base, err := filepath.EvalSymlinks(forbidden)
		if err != nil {
			return refuse("Kimi candidate proof path unavailable")
		}
		rel, err := filepath.Rel(base, path)
		if err != nil || rel == "." || rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return refuse("Kimi controlled proposal cannot mutate production or originals")
		}
	}
	proposal, err := kimiPromotionProposal(dirs, "controlled")
	if err != nil {
		return err
	}
	if err := kimiPromotionProtectOriginals(path, proposal); err != nil {
		return err
	}
	raw, err := Encode(proposal)
	if err != nil {
		return err
	}
	return regionalPromotionReceipt(path, raw)
}

// Compositions reference originals outside the descriptor directory. A new
// receipt in a native/producer tree would invalidate its inventory; the trusted
// authority directory also must never be used as an output destination.
func kimiPromotionProtectOriginals(path string, proof KimiPromotion) error {
	root, err := kimiRepositoryRoot()
	if err != nil {
		return err
	}
	for _, candidate := range proof.Candidates {
		store, err := regionalReservationStore(candidate.Combo)
		if err != nil {
			return err
		}
		forbidden := []string{store}
		for _, measurement := range candidate.Qualification.Measurements {
			forbidden = append(forbidden, filepath.Join(root, measurement.Evidence.Bundle), filepath.Dir(filepath.Join(root, measurement.Evidence.Producer.Path)))
		}
		for _, base := range forbidden {
			rel, err := filepath.Rel(base, path)
			if err != nil || rel == "." || rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return refuse("Kimi promotion output cannot mutate transitive originals or authority")
			}
		}
	}
	return nil
}

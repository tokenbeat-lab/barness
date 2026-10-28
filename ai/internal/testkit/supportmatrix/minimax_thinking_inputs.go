package supportmatrix

import (
	"path/filepath"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// Close over both WHOLE packets, including excluded failures, original producer
// input provenance and authority chosen by trusted common Git, never descriptor.
func MiniMaxThinkingCompositionInputs(dir string) (map[string]string, error) {
	q, e := VerifyMiniMaxThinkingComposition(dir, false)
	if e != nil {
		return nil, e
	}
	root, e := minimaxlive.Root()
	if e != nil {
		return nil, e
	}
	raw, e := minimaxlive.ReadFile(filepath.Join(dir, MiniMaxThinkingCompositionFile), 1<<20)
	if e != nil {
		return nil, e
	}
	out := map[string]string{MiniMaxThinkingCompositionFile: minimaxlive.Hash(raw)}
	for i, m := range q.Measurements {
		native, e := minimaxlive.LocalEvidencePath(root, m.Evidence.Bundle)
		if e != nil {
			return nil, e
		}
		files, e := minimaxlive.Inventory(native)
		if e != nil {
			return nil, e
		}
		raw, e := minimaxlive.ReadFile(filepath.Join(native, minimaxlive.SealFile), 4<<20)
		if e != nil {
			return nil, e
		}
		files[minimaxlive.SealFile] = minimaxlive.Hash(raw)
		for n, h := range files {
			out["@minimax/"+filepath.ToSlash(filepath.Join(m.Evidence.Bundle, n))] = h
		}
		producer, e := minimaxlive.VerifyThinkingProducer(native, m.Evidence.Producer, m.Original, i == 0)
		if e != nil {
			return nil, e
		}
		for n, h := range producer {
			out["@minimax/"+n] = h
		}
		authority, e := livegrant.EvidenceInputs(native)
		if e != nil {
			return nil, e
		}
		for n, h := range authority {
			out["@minimax-authority/"+n] = h
		}
	}
	return out, nil
}

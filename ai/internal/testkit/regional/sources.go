package regional

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
)

// VerifySources binds complete candidate data, official originals and retrieval
// receipts to a reviewed immutable snapshot. It does not claim the remote
// documents remain unchanged or that any account has live eligibility.
func VerifySources() error {
	raw, err := files.ReadFile("snapshot.json")
	if err != nil || hash(raw) != snapshotPin {
		return errors.New("regional: snapshot manifest changed")
	}
	var snapshot struct {
		Version string            `json:"version"`
		Files   map[string]string `json:"files"`
	}
	if json.Unmarshal(raw, &snapshot) != nil || len(snapshot.Files) != 3 {
		return errors.New("regional: invalid snapshot manifest")
	}
	for name, pin := range snapshot.Files {
		raw, err = files.ReadFile(name)
		if err != nil || hash(raw) != pin {
			return errors.New("regional: candidate/source manifest changed")
		}
	}
	raw, err = files.ReadFile("source-files.json")
	if err != nil {
		return err
	}
	var sources []struct {
		ID, File, SHA256, URL, Status, Encoding string
		OriginalSHA256                          string `json:"originalSha256"`
		OriginalBytes                           int    `json:"originalBytes"`
		ReadAt                                  string `json:"readAt"`
	}
	if json.Unmarshal(raw, &sources) != nil {
		return errors.New("regional: invalid source receipt")
	}
	seen := map[string]bool{}
	for _, source := range sources {
		if source.Status == "FETCH_FAILED" {
			continue
		}
		if seen[source.ID] || source.ReadAt == "" || source.URL == "" || source.Encoding != "gzip" || source.OriginalBytes > 1<<20 {
			return errors.New("regional: incomplete source identity")
		}
		seen[source.ID] = true
		raw, err = files.ReadFile(source.File)
		if err != nil || hash(raw) != source.SHA256 {
			return errors.New("regional: official source changed")
		}
		reader, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			return err
		}
		original, readErr := io.ReadAll(io.LimitReader(reader, 1<<20+1))
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil || len(original) != source.OriginalBytes || hash(original) != source.OriginalSHA256 {
			return errors.New("regional: official original changed")
		}
	}
	for _, id := range []string{"zai-price", "zai-glm-53", "zai-glm-52", "zai-chat", "zai-keys", "minimax-cn-price", "minimax-cn-models", "minimax-cn-chat", "minimax-cn-messages", "minimax-cn-responses", "minimax-cn-schema-chat", "fx-2026-10-16"} {
		if !seen[id] {
			return errors.New("regional: required source absent")
		}
	}
	return nil
}

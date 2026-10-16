package ai

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ClassifierImageInput identifies one final native image occurrence. SHA256 is
// the lowercase hexadecimal SHA-256 of the decoded file, not its Base64 or URL.
// Indices belong to this result only; Hooks can change the original input order
// and repeated Decisions blocks retain separate entries with identical hashes.
// This is result content for the calling tenant, never an access credential.
type ClassifierImageInput struct {
	MimeType string `json:"mimeType"`
	SHA256   string `json:"sha256"`
}

func (i ClassifierImageInput) valid() bool {
	if !classifierImageMIME(i.MimeType) || len(i.SHA256) != sha256.Size*2 {
		return false
	}
	for _, ch := range i.SHA256 {
		if !(ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f') {
			return false
		}
	}
	return true
}

func (i ClassifierImageInput) MarshalJSON() ([]byte, error) {
	if !i.valid() {
		return nil, fmt.Errorf("invalid classifier image identity")
	}
	type record ClassifierImageInput
	return json.Marshal(record(i))
}

func (i *ClassifierImageInput) UnmarshalJSON(raw []byte) error {
	var wire struct {
		MimeType *string `json:"mimeType"`
		SHA256   *string `json:"sha256"`
	}
	if err := decodeClassifierJSON(raw, &wire); err != nil {
		return err
	}
	if wire.MimeType == nil || wire.SHA256 == nil {
		return fmt.Errorf("classifier image identity requires mimeType and sha256")
	}
	out := ClassifierImageInput{MimeType: *wire.MimeType, SHA256: *wire.SHA256}
	if !out.valid() {
		return fmt.Errorf("invalid classifier image identity")
	}
	*i = out
	return nil
}

// Images have already passed the final native authority and size checks.
// Hash one decoded stream at a time with fixed memory and cancellation points;
// no decoded files or cross-call cache are retained.
func classifierImageInputs(ctx context.Context, images []Image) ([]ClassifierImageInput, *Error) {
	var inputs []ClassifierImageInput
	var buffer [4096]byte
	for _, img := range images {
		digest := sha256.New()
		reader := base64.NewDecoder(base64.StdEncoding.Strict(), strings.NewReader(img.Data))
		for {
			if failure := contextError(ctx, PhaseRequest); failure != nil {
				return nil, failure
			}
			n, err := reader.Read(buffer[:])
			_, _ = digest.Write(buffer[:n]) // hash.Hash.Write always succeeds.
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, decisionsSchemaFailure()
			}
		}
		inputs = append(inputs, ClassifierImageInput{MimeType: img.MimeType, SHA256: hex.EncodeToString(digest.Sum(nil))})
	}
	return inputs, contextError(ctx, PhaseRequest)
}

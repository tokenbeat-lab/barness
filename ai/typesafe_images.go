package ai

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"maps"
	"strings"
)

const typeSafeImagesHeader = "Typesafe-Danger-Zone-Images"

type typeSafeAsset struct {
	Type   string `json:"type"`
	Format string `json:"format"`
	Data   string `json:"data_base64"`
}

type typeSafeContext struct {
	State  json.RawMessage `json:"state"`
	Images []string        `json:"images"`
}

func typeSafeImagesSelected(opts Options) bool {
	o, ok := opts.(TypeSafeOptions)
	return ok && o.ExperimentalImages
}

func typeSafeBusinessState(req typeSafeRequest) json.RawMessage {
	if len(req.Assets) == 0 {
		return req.State
	}
	var state typeSafeContext
	if decodeClassifierJSON(req.State, &state) != nil {
		return nil
	}
	return state.State
}

func (c *Client) prepareTypeSafeImages(ctx context.Context, model ClassifierModel, req ClassifierRequest, plan classifierImagePlan) ([]byte, *Error) {
	if failure := checkTypeSafeUploadSize(req); failure != nil {
		return nil, failure
	}
	if failure := checkTypeSafeImageFiles(ctx, req.Images, PhaseCapability); failure != nil {
		return nil, failure
	}
	if len(req.Images) > typeSafeAllKeyBytes/34 {
		return nil, limitFailure(PhaseCapability, "Jev.MaxAssetKeyBytes", typeSafeAllKeyBytes)
	}
	body, err := encodeTypeSafeRequest(model.ID, req)
	if err != nil {
		return nil, newError(CodeInvalidRequest, PhaseRequest, "classifier request could not be encoded")
	}
	var native typeSafeRequest
	if decodeClassifierJSON(body, &native) != nil {
		return nil, typeSafeRequestFailure()
	}
	native.Assets = make(map[string]typeSafeAsset, len(req.Images))
	keys := make([]string, len(req.Images))
	// Search decoded original keys and values, including escaped strings. Eight
	// bounded selections prevent accidental vendor substring substitutions.
	var original any
	d := json.NewDecoder(bytes.NewReader(body))
	d.UseNumber()
	if d.Decode(&original) != nil {
		return nil, typeSafeRequestFailure()
	}
	keyRemaining := typeSafeAllKeyBytes
	for i, img := range req.Images {
		for attempt := 0; attempt < 8; attempt++ {
			if failure := contextError(ctx, PhaseRequest); failure != nil {
				return nil, failure
			}
			key := "<image:" + rand.Text() + ">"
			if typeSafeContainsKey(original, key) {
				continue
			}
			conflict := false
			for previous := range native.Assets {
				if strings.Contains(key, previous) || strings.Contains(previous, key) {
					conflict = true
					break
				}
			}
			if conflict {
				continue
			}
			if len(key) > typeSafeKeyBytes || len(key) > keyRemaining {
				return nil, limitFailure(PhaseCapability, "Jev.MaxAssetKeyBytes", typeSafeAllKeyBytes)
			}
			keyRemaining -= len(key)
			keys[i] = key
			format := "png"
			if img.MimeType == "image/jpeg" {
				format = "jpeg"
			}
			native.Assets[key] = typeSafeAsset{Type: "image", Format: format, Data: img.Data}
			break
		}
		if keys[i] == "" {
			return nil, newError(CodeInvalidRequest, PhaseRequest, "could not allocate unambiguous image references")
		}
	}
	state := typeSafeContext{State: req.State}
	for _, ref := range plan {
		state.Images = append(state.Images, keys[ref.ImageIndex])
	}
	native.State, err = json.Marshal(state)
	if err != nil {
		return nil, typeSafeRequestFailure()
	}
	body, err = json.Marshal(native)
	if err != nil {
		return nil, typeSafeRequestFailure()
	}
	if int64(len(body)) > c.policy.MaxRequestBytes {
		return nil, limitFailure(PhaseScope, "MaxRequestBytes", c.policy.MaxRequestBytes)
	}
	if int64(len(body)) > typeSafeBodyBytes {
		return nil, limitFailure(PhaseCapability, "Jev.MaxRequestBytes", typeSafeBodyBytes)
	}
	return body, nil
}

func typeSafeContainsKey(value any, key string) bool {
	switch v := value.(type) {
	case string:
		return strings.Contains(v, key)
	case []any:
		for _, item := range v {
			if typeSafeContainsKey(item, key) {
				return true
			}
		}
	case map[string]any:
		for name, item := range v {
			if strings.Contains(name, key) || typeSafeContainsKey(item, key) {
				return true
			}
		}
	}
	return false
}

func (c *Client) checkTypeSafePayload(body map[string]any, limit int64) *Error {
	if _, ok := body["assets"]; !ok {
		return c.policy.Classifier.checkPayload(body, limit)
	}
	// Context framing is not business State. Its independent state subtree is
	// checked on final native bytes; bound the entire borrowed body before encode.
	if failure := c.checkTypeSafeAssetPayload(body["assets"]); failure != nil {
		return failure
	}
	view := maps.Clone(body)
	view["state"] = nil
	if failure := c.policy.Classifier.checkPayload(view, limit); failure != nil {
		return failure
	}
	if _, ok := jsonPayloadSize(body, typeSafeBodyBytes); !ok {
		return limitFailure(PhaseRequest, "Jev.MaxRequestBytes", typeSafeBodyBytes)
	}
	if _, ok := jsonPayloadSize(body, limit); !ok {
		return limitFailure(PhaseRequest, "MaxRequestBytes", limit)
	}
	return nil
}

// Check the known raw upload lower bound before building an encoded body.
// Reused references upload once; expanded decoded bytes have a separate budget.
func checkTypeSafeUploadSize(req ClassifierRequest) *Error {
	remaining := typeSafeBodyBytes
	take := func(n int64) bool {
		if n > remaining {
			return false
		}
		remaining -= n
		return true
	}
	if !take(int64(len(req.State))) {
		return limitFailure(PhaseCapability, "Jev.MaxRequestBytes", typeSafeBodyBytes)
	}
	for key, q := range req.Questions {
		size, ok := classifierQuestionSize(q, remaining)
		if !ok || !take(size) || !take(int64(len(key))) {
			return limitFailure(PhaseCapability, "Jev.MaxRequestBytes", typeSafeBodyBytes)
		}
	}
	for _, img := range req.Images {
		if !take(int64(len(img.Data))) {
			return limitFailure(PhaseCapability, "Jev.MaxRequestBytes", typeSafeBodyBytes)
		}
	}
	return nil
}

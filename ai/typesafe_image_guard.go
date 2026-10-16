package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

func (a *typeSafeAsset) UnmarshalJSON(raw []byte) error {
	var wire struct {
		Type   json.RawMessage `json:"type"`
		Format string          `json:"format"`
		Data   string          `json:"data_base64"`
	}
	if err := decodeClassifierJSON(raw, &wire); err != nil {
		return err
	}
	kind := "image"
	if len(wire.Type) > 0 {
		if err := json.Unmarshal(wire.Type, &kind); err != nil {
			return err
		}
		if kind != "image" || bytes.Equal(bytes.TrimSpace(wire.Type), []byte("null")) {
			return fmt.Errorf("invalid Jev asset type")
		}
	}
	*a = typeSafeAsset{Type: kind, Format: wire.Format, Data: wire.Data}
	return nil
}

func (c *Client) freezeTypeSafeImages(ctx context.Context, req typeSafeRequest, model ClassifierModel, allowed, selected bool) ([]Image, *Error) {
	if len(req.Assets) == 0 {
		return nil, nil
	}
	if !allowed || !selected || c.policy.Classifier.Images == nil || !slices.Contains(model.Input, ModalityImage) || !slices.Contains(model.Capabilities.ImageLocations, ClassifierImageLocationContext) {
		return nil, newError(CodeTenantDenied, PhaseRequest, "Jev images are not authorized")
	}
	var state typeSafeContext
	if decodeClassifierJSON(req.State, &state) != nil || len(state.Images) == 0 {
		return nil, typeSafeRequestFailure()
	}
	limits := c.policy.Classifier.Images.referenceLimits()
	if len(state.Images) > limits.MaxReferences {
		return nil, limitFailure(PhaseRequest, "Classifier.Images.References.MaxReferences", int64(limits.MaxReferences))
	}
	if len(req.Assets) > model.Capabilities.MaxInputImages {
		return nil, limitFailure(PhaseRequest, "ClassifierModel.MaxInputImages", int64(model.Capabilities.MaxInputImages))
	}
	if len(req.Assets) > c.policy.Classifier.Images.MaxImages {
		return nil, limitFailure(PhaseRequest, "Classifier.Images.MaxImages", int64(c.policy.Classifier.Images.MaxImages))
	}
	if failure := typeSafeImageKeys(ctx, req, state.State); failure != nil {
		return nil, failure
	}
	var images []Image
	seen := map[string]bool{}
	remaining := limits.MaxExpandedImageBytes
	for _, key := range state.Images {
		if failure := contextError(ctx, PhaseRequest); failure != nil {
			return nil, failure
		}
		asset, ok := req.Assets[key]
		if !ok || asset.Type != "image" {
			return nil, typeSafeRequestFailure()
		}
		mime := "image/png"
		if asset.Format == "jpeg" || asset.Format == "jpg" {
			mime = "image/jpeg"
		} else if asset.Format != "png" {
			return nil, typeSafeRequestFailure()
		}
		size := base64Size(asset.Data)
		if size > remaining {
			return nil, limitFailure(PhaseRequest, "Classifier.Images.References.MaxExpandedImageBytes", limits.MaxExpandedImageBytes)
		}
		remaining -= size
		if !seen[key] {
			images = append(images, Image{Data: asset.Data, MimeType: mime})
			seen[key] = true
		}
	}
	if len(seen) != len(req.Assets) {
		return nil, typeSafeRequestFailure()
	}
	if problem := classifierImageCapability(images, model); problem != "" {
		return nil, newError(CodeTenantDenied, PhaseRequest, "Jev images exceed authorized capabilities")
	}
	if failure := c.policy.checkClassifierImageSizes(images, PhaseRequest); failure != nil {
		return nil, failure
	}
	if failure := checkClassifierImageEncoding(ctx, images, PhaseRequest); failure != nil {
		return nil, failure
	}
	if failure := checkTypeSafeImageFiles(ctx, images, PhaseRequest); failure != nil {
		return nil, failure
	}
	return images, nil
}

// This ticket supports only the projected context array. No asset key may
// occur in business State, question content, object keys or identity fields.
// The final native data owns this validation; no public-request reconstruction.
func typeSafeImageKeys(ctx context.Context, req typeSafeRequest, business json.RawMessage) *Error {
	plain := req
	plain.State = business
	plain.Assets = nil
	raw, err := json.Marshal(plain)
	if err != nil {
		return typeSafeRequestFailure()
	}
	var content any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if d.Decode(&content) != nil {
		return typeSafeRequestFailure()
	}
	keys := make([]string, 0, len(req.Assets))
	remaining := typeSafeAllKeyBytes
	for key := range req.Assets {
		if failure := contextError(ctx, PhaseRequest); failure != nil {
			return failure
		}
		if strings.TrimSpace(key) == "" {
			return typeSafeRequestFailure()
		}
		if len(key) > typeSafeKeyBytes {
			return limitFailure(PhaseRequest, "Jev.MaxAssetKeyBytes", typeSafeKeyBytes)
		}
		if len(key) > remaining {
			return limitFailure(PhaseRequest, "Jev.MaxTotalAssetKeyBytes", typeSafeAllKeyBytes)
		}
		remaining -= len(key)
		for _, previous := range keys {
			if strings.Contains(key, previous) || strings.Contains(previous, key) {
				return typeSafeRequestFailure()
			}
		}
		if typeSafeContainsKey(content, key) {
			return typeSafeRequestFailure()
		}
		keys = append(keys, key)
	}
	return nil
}

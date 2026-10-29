package e2e

import (
	"encoding/base64"
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Keep every self-reported copy internally consistent. The consumer must still
// compare images and charges against independent frozen measurement facts.
func imageQualificationReplaceImage(t *testing.T, f imageQualificationFixture) {
	t.Helper()
	source := readXAINativeObject(t, filepath.Join(f.bundle, "XAIIMG-live-generation-auto", "result.json"))["content"].([]any)[0]
	dir := filepath.Join(f.bundle, "XAIIMG-live-generation-low")
	imageQualificationMutate(t, filepath.Join(dir, "result.json"), func(m map[string]any) { m["content"] = []any{source} })
	var rows []map[string]any
	raw, err := os.ReadFile(filepath.Join(dir, "http.json"))
	if err != nil {
		t.Fatal(err)
	}
	mustUnmarshal(t, raw, &rows)
	var response map[string]any
	mustUnmarshal(t, []byte(rows[0]["responseBody"].(string)), &response)
	image := source.(map[string]any)
	response["data"] = []any{map[string]any{"b64_json": image["data"], "mime_type": image["mimeType"]}}
	body, _ := json.Marshal(response)
	rows[0]["responseBody"] = string(body)
	writeXAINativeJSON(t, filepath.Join(dir, "http.json"), rows)
	imageBytes, err := base64.StdEncoding.Strict().DecodeString(image["data"].(string))
	if err != nil {
		t.Fatal(err)
	}
	mustWriteXAINative(t, filepath.Join(dir, "output-0.image"), imageBytes)
	imageQualificationMutate(t, filepath.Join(f.bundle, "manifest.json"), func(m map[string]any) {
		for _, v := range m["cases"].([]any) {
			c := v.(map[string]any)
			if c["id"] == "XAIIMG-live-generation-low" {
				c["fixtures"].(map[string]any)["output-0.image"] = imageQualificationHash(t, filepath.Join(dir, "output-0.image"))
			}
		}
	})
	imageQualificationRefresh(t, f)
}
func imageQualificationSetTicks(t *testing.T, f imageQualificationFixture, id string, ticks int64) {
	t.Helper()
	dir := filepath.Join(f.bundle, "XAIIMG-live-"+id)
	var rows []map[string]any
	raw, err := os.ReadFile(filepath.Join(dir, "http.json"))
	if err != nil {
		t.Fatal(err)
	}
	mustUnmarshal(t, raw, &rows)
	var response map[string]any
	mustUnmarshal(t, []byte(rows[0]["responseBody"].(string)), &response)
	response["usage"].(map[string]any)["cost_in_usd_ticks"] = ticks
	body, _ := json.Marshal(response)
	rows[0]["responseBody"] = string(body)
	writeXAINativeJSON(t, filepath.Join(dir, "http.json"), rows)
	usd := float64(ticks) / 1e10
	imageQualificationMutate(t, filepath.Join(dir, "result.json"), func(m map[string]any) {
		m["metadata"].(map[string]any)["attempts"].([]any)[0].(map[string]any)["reportedCostUsd"] = usd
	})
	var observations []map[string]any
	raw, err = os.ReadFile(filepath.Join(dir, "observations.json"))
	if err != nil {
		t.Fatal(err)
	}
	mustUnmarshal(t, raw, &observations)
	observations[2]["attempt"].(map[string]any)["reportedCostUsd"] = usd
	observations[3]["call"].(map[string]any)["attempts"].([]any)[0].(map[string]any)["reportedCostUsd"] = usd
	writeXAINativeJSON(t, filepath.Join(dir, "observations.json"), observations)
}

func imageQualificationRefreshProducer(t *testing.T, f imageQualificationFixture) {
	imageQualificationMutate(t, filepath.Join(f.attempt, "receipt.json"), func(m map[string]any) {
		m["bundles"].([]any)[0].(map[string]any)["manifestSHA256"] = imageQualificationHash(t, filepath.Join(f.bundle, "manifest.json"))
	})
}
func imageQualificationProxy(t *testing.T, f imageQualificationFixture, proxy string) {
	path := filepath.Join(f.attempt, "authority", "proxy-selection.json")
	imageQualificationMutate(t, path, func(m map[string]any) { m["proxy"] = proxy; m["explicitUserSelection"] = true })
	rel, _ := filepath.Rel(f.root, path)
	imageQualificationMutate(t, filepath.Join(f.attempt, "receipt.json"), func(m map[string]any) {
		m["selectedInputs"].(map[string]any)[filepath.ToSlash(rel)] = imageQualificationHash(t, path)
	})
}
func imageQualificationRemoveAuthority(t *testing.T, f imageQualificationFixture, name string) {
	inputs, err := livegrant.ClaimEvidenceInputs(f.bundle)
	if err != nil {
		t.Fatal(err)
	}
	prefix := strings.TrimPrefix(name, "authority-") + "-"
	for key := range inputs {
		if strings.HasPrefix(key, prefix) {
			if err := os.Remove(filepath.Join(os.Getenv("GIT_DIR"), "barness-live-process-grants", key)); err != nil {
				t.Fatal(err)
			}
		}
	}
}

// Mutate the JSON resource inside requestBody, then refresh the fixture's
// enclosing hashes. Outer HTTP JSON schema checks cannot inspect this string.
func imageQualificationResource(t *testing.T, f imageQualificationFixture, name string) {
	t.Helper()
	shape, mutation, _ := strings.Cut(strings.TrimPrefix(name, "resource-"), "-")
	scene := "single-edit"
	if shape == "multi" {
		scene = "ordered-edit"
	}
	path := filepath.Join(f.bundle, "XAIIMG-live-"+scene, "http.json")
	var rows []map[string]any
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mustUnmarshal(t, raw, &rows)
	var request map[string]any
	mustUnmarshal(t, []byte(rows[0]["requestBody"].(string)), &request)
	var resource map[string]any
	if shape == "single" {
		resource = request["image"].(map[string]any)
	} else {
		resource = request["images"].([]any)[1].(map[string]any)
	}
	switch mutation {
	case "file-id":
		resource["file_id"] = "synthetic-file-id"
	case "null-file-id":
		resource["file_id"] = nil
	case "type-alias":
		resource["TYPE"] = resource["type"]
		delete(resource, "type")
	case "url-alias":
		resource["URL"] = resource["url"]
		delete(resource, "url")
	case "null-url":
		resource["url"] = nil
	}
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if mutation == "duplicate-url" {
		body = []byte(strings.Replace(string(body), `"url":`, `"url":"synthetic-duplicate","url":`, 1))
	}
	rows[0]["requestBody"] = string(body)
	writeXAINativeJSON(t, path, rows)
	imageQualificationRefresh(t, f)
}

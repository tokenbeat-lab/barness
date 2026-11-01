package maintenance

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

const ModelsDevURL = "https://models.dev/catalog.json?type=all"
const maxSourceBytes = 64 << 20

// AcquireSource pins the actual bytes used by this invocation. Local replay
// requires its original fetch time, rather than inventing one from file mtime.
// Network acquisition belongs to the maintenance tool, never to Client/load.
func AcquireSource(ctx context.Context, path, url, fetchedAt string) (FrozenSource, error) {
	if url == "" {
		url = ModelsDevURL
	}
	var reader io.ReadCloser
	if path != "" {
		if _, err := time.Parse(time.RFC3339, fetchedAt); err != nil {
			return FrozenSource{}, failure("invalid_argument", "local source replay requires --source-time in RFC3339")
		}
		file, err := os.Open(path)
		if err != nil {
			return FrozenSource{}, failure("source_unavailable", "open source input: %v", err)
		}
		reader = file
	} else {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return FrozenSource{}, failure("source_unavailable", "invalid source URL")
		}
		client := &http.Client{Timeout: 20 * time.Second}
		response, err := client.Do(request)
		if err != nil {
			return FrozenSource{}, failure("source_unavailable", "source request failed")
		}
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			return FrozenSource{}, failure("source_unavailable", "source returned HTTP %d", response.StatusCode)
		}
		reader = response.Body
		fetchedAt = time.Now().UTC().Format(time.RFC3339Nano)
		url = response.Request.URL.String()
	}
	defer reader.Close()
	raw, err := io.ReadAll(io.LimitReader(reader, maxSourceBytes+1))
	if err != nil || len(raw) > maxSourceBytes {
		return FrozenSource{}, failure("source_unavailable", "source input could not be read within the size limit")
	}
	return FrozenSource{URL: url, FetchedAt: fetchedAt, Digest: digest(raw), Bytes: raw}, nil
}

type sourceCatalog struct {
	Providers map[string]struct {
		ID     string                     `json:"id"`
		Models map[string]json.RawMessage `json:"models"`
	} `json:"providers"`
	Models map[string]json.RawMessage `json:"models"`
}

var sourceProviders = map[string]string{
	"openai": "openai", "anthropic": "anthropic", "google": "google", "deepseek": "deepseek", "xai": "xai", "zai": "zai", "zhipuai": "zhipu", "minimax": "minimax", "minimax-cn": "minimax-cn", "moonshotai": "moonshot", "moonshotai-cn": "moonshot-cn", "typesafe": "typesafe",
}

func sourceOffering(request ChangeRequest) (map[string]json.RawMessage, string, error) {
	source, err := sourceDocument(request)
	if err != nil {
		return nil, "", err
	}
	return offeringFrom(source, request)
}

func sourceDocument(request ChangeRequest) (sourceCatalog, error) {
	if request.Source == nil {
		return sourceCatalog{}, failure("source_unavailable", "a pinned source is required")
	}
	if digest(request.Source.Bytes) != request.Source.Digest || strings.TrimSpace(request.Source.URL) == "" {
		return sourceCatalog{}, failure("source_changed", "source bytes or provenance do not match")
	}
	if _, err := time.Parse(time.RFC3339Nano, request.Source.FetchedAt); err != nil {
		return sourceCatalog{}, failure("invalid_argument", "source fetch time must be RFC3339")
	}
	var source sourceCatalog
	if err := json.Unmarshal(request.Source.Bytes, &source); err != nil || source.Providers == nil {
		return sourceCatalog{}, failure("invalid_source", "expected the models.dev providers/models projection")
	}
	return source, nil
}

func offeringFrom(source sourceCatalog, request ChangeRequest) (map[string]json.RawMessage, string, error) {
	provider, ok := source.Providers[request.SourceProvider]
	if !ok || (provider.ID != "" && provider.ID != request.SourceProvider) {
		return nil, "", failure("source_unavailable", "serving Provider is missing or mismatched")
	}
	model := request.SourceModel
	if model == "" {
		model = request.Identity.ID
	}
	raw, ok := provider.Models[model]
	if !ok {
		return nil, "", failure("source_unavailable", "serving model is missing; canonical metadata cannot replace it")
	}
	var offering map[string]json.RawMessage
	if err := json.Unmarshal(raw, &offering); err != nil || offering == nil {
		return nil, "", failure("invalid_source", "serving offering must be an object")
	}
	for _, field := range []string{"type", "reasoning"} {
		if raw, provided := offering[field]; provided {
			if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				return nil, "", failure("invalid_source", "source %s cannot be null", field)
			}
			var err error
			if field == "type" {
				var value string
				err = json.Unmarshal(raw, &value)
			} else {
				var value bool
				err = json.Unmarshal(raw, &value)
			}
			if err != nil {
				return nil, "", failure("invalid_source", "source %s has an invalid type", field)
			}
		}
	}
	canonical := ""
	if value, ok := offering["canonical_model_id"]; ok {
		if err := json.Unmarshal(value, &canonical); err != nil {
			return nil, "", failure("invalid_source", "canonical model ID must be a string")
		}
	}
	return offering, canonical, nil
}

// Offering operation is independent of API selection. Missing legacy type means
// chat only; the same wire ID never authorizes borrowing another operation.
func offeringOperation(offering map[string]json.RawMessage) (ai.Operation, string, error) {
	var kind string
	if raw, exists := offering["type"]; exists {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &kind) != nil {
			return "", "", failure("invalid_source", "source type must be a string")
		}
	}
	switch kind {
	case "", "chat":
		return ai.OperationChat, kind, nil
	case "image":
		return ai.OperationImage, kind, nil
	case "decision":
		return ai.OperationClassifier, kind, nil
	default:
		return "", kind, failure("unsupported_definition", "source type %q is not implemented", kind)
	}
}

func validateOfferingOperation(offering map[string]json.RawMessage, operation ai.Operation) error {
	actual, _, err := offeringOperation(offering)
	if err != nil {
		return err
	}
	if actual != operation {
		return failure("mapping_conflict", "source type and maintained operation differ")
	}
	return nil
}

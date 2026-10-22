package e2e

import (
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"os"
	"path/filepath"
	"testing"
)

// Public artifact-consumer fixtures evolve independently of release selection.
func writeReleaseJSON(t *testing.T, dir, name string, value any) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
		t.Fatal(err)
	}
}

func mustMarshal(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// Fixture goldens include an explicit content discriminator.
func fixtureContent(t *testing.T, c ai.AssistantContent) map[string]any {
	block := jsonObject(t, c)
	switch c.(type) {
	case ai.Text:
		block["type"] = "text"
	case ai.Thinking:
		block["type"] = "thinking"
	case ai.ToolCall:
		block["type"] = "toolCall"
	default:
		block["type"] = fmt.Sprintf("barness-unmapped:%T", c)
	}
	return block
}

// jsonObject is v's JSON serialization as a generic object.
func jsonObject(t *testing.T, v any) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(mustMarshal(t, v), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

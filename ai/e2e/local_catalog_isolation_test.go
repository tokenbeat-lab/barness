package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/catalog"
)

func TestLocalCatalogModelFailures(t *testing.T) {
	failures := []struct {
		name, entry string
		reason      catalog.Reason
	}{
		{"unknown-field", strings.Replace(localChatJSON, `"reasoning":false`, `"reasoning":false,"newFlag":true`, 1), catalog.UnknownField},
		{"wrong-type", strings.Replace(localChatJSON, `"maxTokens":32768`, `"maxTokens":"large"`, 1), catalog.InvalidDefinition},
		{"fractional-limit", strings.Replace(localChatJSON, `"maxTokens":32768`, `"maxTokens":1.25`, 1), catalog.InvalidDefinition},
		{"null-limit", strings.Replace(localChatJSON, `"maxTokens":32768`, `"maxTokens":null`, 1), catalog.InvalidDefinition},
		{"negative-limit", strings.Replace(localChatJSON, `"contextWindow":1047576`, `"contextWindow":-1`, 1), catalog.InvalidDefinition},
		{"missing-identity", strings.Replace(localChatJSON, `"id":"gpt-4.1-mini",`, ``, 1), catalog.InvalidDefinition},
		{"null-identity", strings.Replace(localChatJSON, `"provider":"openai"`, `"provider":null`, 1), catalog.InvalidDefinition},
		{"nested-thinking-unknown", strings.Replace(localChatJSON, `"reasoning":false`, `"reasoning":false,"thinkingLevelMap":{"low":"low","future":"x"}`, 1), catalog.UnknownField},
		{"nested-nullable-wrong-type", strings.Replace(localChatJSON, `"reasoning":false`, `"reasoning":false,"thinkingLevelMap":{"low":{"future":"x"}}`, 1), catalog.InvalidDefinition},
		{"unknown-api", strings.Replace(localChatJSON, `openai-responses`, `future-api`, 1), catalog.UnsupportedDefinition},
		{"unknown-modality", strings.Replace(localChatJSON, `"text","image"`, `"audio"`, 1), catalog.UnsupportedDefinition},
		{"not-object", `42`, catalog.InvalidDefinition},
		{"duplicate-field", strings.Replace(localChatJSON, `"maxTokens":32768`, `"maxTokens":32768,"maxTokens":1`, 1), catalog.InvalidDefinition},
	}
	for _, failure := range failures {
		t.Run(failure.name, func(t *testing.T) {
			ev := run.Case(t, "CAT-T08-model-"+failure.name)
			raw := localCatalogFile(`{"version":"failures","models":[` + failure.entry + `]}`)
			ev.Record("source", json.RawMessage(raw))
			effective, report, err := catalog.Load(raw)
			ev.Record("report", report)
			ev.Check("only this definition fails", err == nil && report.Loaded == 0 && report.Failed == 1 && len(effective.Models) == 0 && len(report.Models) == 1, "err=%v report=%+v", err, report)
			if len(report.Models) != 1 {
				t.FailNow()
			}
			ev.Check("failure classified and original retained", report.Models[0].Reason == failure.reason && string(report.Models[0].Raw) == failure.entry, "status=%+v", report.Models[0])
			ev.Check("empty effective catalog keeps source version", effective.Version == "failures" && report.EffectiveHash != "", "catalog=%+v", effective)
		})
	}
	t.Run("duplicates-all-rejected", func(t *testing.T) {
		ev := run.Case(t, "CAT-T08-duplicate-group-all-rejected")
		entry := strings.Replace(localChatJSON, `"maxTokens":32768`, `"maxTokens":"invalid"`, 1)
		raw := localCatalogFile(`{"version":"duplicates","models":[` + localChatJSON + `,` + entry + `,` + localChatJSON + `]}`)
		effective, report, err := catalog.Load(raw)
		ev.Record("source", json.RawMessage(raw))
		ev.Record("report", report)
		ev.Check("no duplicate wins even if another is invalid", err == nil && len(effective.Models) == 0 && report.Failed == 3 && report.Loaded == 0, "err=%v report=%+v", err, report)
		for _, status := range report.Models {
			ev.Check("every group member fails for duplicate identity", status.Reason == catalog.DuplicateIdentity, "status=%+v", status)
		}
	})
	t.Run("known-provider-constraints", func(t *testing.T) {
		ev := run.Case(t, "CAT-T08-fixed-provider-subset")
		entry := strings.Replace(localChatJSON, `"provider":"openai"`, `"provider":"moonshot"`, 1)
		effective, report, err := catalog.Load(localCatalogFile(`{"version":"fixed","models":[` + entry + `]}`))
		ev.Record("report", report)
		ev.Check("new Kimi identity cannot bypass existing constraints", err == nil && len(effective.Models) == 0 && report.Models[0].Reason == catalog.UnsupportedDefinition, "err=%v report=%+v", err, report)
	})
}

func TestLocalCatalogPartialIdentities(t *testing.T) {
	ev := run.Case(t, "CAT-T08-partial-identities-are-not-duplicates")
	for _, blank := range []string{"", " "} {
		entry := strings.Replace(localChatJSON, `"id":"gpt-4.1-mini"`, `"id":"`+blank+`"`, 1)
		_, report, err := catalog.Load(localCatalogFile(`{"version":"partial","models":[` + entry + `,` + entry + `]}`))
		ev.Record("report-blank-"+strings.ReplaceAll(blank, " ", "space"), report)
		ev.Check("incomplete identity is located by position, never grouped", err == nil && report.Failed == 2 && report.Loaded == 0 && len(report.Models) == 2 && report.Models[0].Index == 0 && report.Models[1].Index == 1 && report.Models[0].Reason == catalog.InvalidDefinition && report.Models[1].Reason == catalog.InvalidDefinition, "err=%v report=%+v", err, report)
	}
}

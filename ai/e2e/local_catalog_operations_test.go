package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/catalog"
)

const localImageJSON = `{"provider":"openai","api":"openai-images","id":"host-image","name":"Synthetic image","input":["text","image"],"output":["image"],"capabilities":{"maxReferenceImages":1,"maxOutputImages":1,"sizes":["1024x1024"],"qualities":["low","medium"]},"pricing":{"inputText":2,"inputImage":3,"outputImage":5,"source":"synthetic fixture rates"}}`
const localClassifierJSON = `{"provider":"typesafe","api":"typesafe-system-one","id":"host-jev","name":"Synthetic classifier","contextWindow":64000,"capabilities":{"kinds":["choice","score","bool"],"maxChoices":255,"minScoreLevels":2,"maxScoreLevels":10},"cost":{"input":1,"output":0,"cacheRead":0,"cacheWrite":0}}`

func TestLocalCatalogOperationIsolation(t *testing.T) {
	ev := run.Case(t, "CAT-T08-three-collections-isolated")
	groups := []struct{ name, valid, typeFrom, typeTo, valueFrom, valueTo, nestedFrom, nestedTo string }{
		{"models", localChatJSON, `"maxTokens":32768`, `"maxTokens":"bad"`, `"maxTokens":32768`, `"maxTokens":0`, `"supportsStrictMode":true`, `"supportsStrictMode":true,"future":true`},
		{"imageModels", localImageJSON, `"maxOutputImages":1`, `"maxOutputImages":"bad"`, `"inputText":2`, `"inputText":-2`, `"maxOutputImages":1`, `"maxOutputImages":1,"future":true`},
		{"classifierModels", localClassifierJSON, `"contextWindow":64000`, `"contextWindow":"bad"`, `"input":1`, `"input":-1`, `"maxChoices":255`, `"maxChoices":255,"future":true`},
	}
	body := `{"version":"three-collections"`
	expected := []catalog.Reason{catalog.UnknownField, catalog.InvalidDefinition, catalog.InvalidDefinition, catalog.UnsupportedDefinition, catalog.InvalidDefinition, catalog.InvalidDefinition, catalog.DuplicateIdentity, catalog.DuplicateIdentity}
	for _, group := range groups {
		// Distinct identities keep each isolated failure independent. Only the
		// final two entries deliberately share a complete identity.
		badBase := strings.Replace(group.valid, `"id":"`, `"id":"bad-`, 1)
		badType := strings.Replace(strings.Replace(badBase, `bad-`, `type-`, 1), group.typeFrom, group.typeTo, 1)
		badValue := strings.Replace(strings.Replace(badBase, `bad-`, `value-`, 1), group.valueFrom, group.valueTo, 1)
		nested := strings.Replace(badBase, group.nestedFrom, group.nestedTo, 1)
		unknownAPI := strings.Replace(strings.Replace(badBase, `bad-`, `api-`, 1), `"api":"`, `"api":"future-`, 1)
		missing := strings.Replace(badBase, `"provider":`, `"provider":null,"ignoredProvider":`, 1)
		// Missing provider is independent of unknown fields, so remove it.
		var fields map[string]json.RawMessage
		mustUnmarshal(t, []byte(missing), &fields)
		delete(fields, "provider")
		delete(fields, "ignoredProvider")
		missing = string(mustMarshal(t, fields))
		duplicate := strings.Replace(badBase, `bad-`, `duplicate-`, 1)
		entries := []string{group.valid, nested, badType, badValue, unknownAPI, missing, `42`, duplicate, duplicate}
		body += `,"` + group.name + `":[` + strings.Join(entries, ",") + `]`
	}
	body += `}`
	raw := localCatalogFile(body)
	effective, report, err := catalog.Load(raw)
	ev.Record("source", json.RawMessage(raw))
	ev.Record("report", report)
	ev.Check("one valid entry survives in each collection", err == nil && report.Loaded == 3 && report.Failed == 24 && len(effective.Models) == 1 && len(effective.ImageModels) == 1 && len(effective.ClassifierModels) == 1, "err=%v report=%+v", err, report)
	if len(report.Models) != 27 {
		t.FailNow()
	}
	for g, group := range groups {
		for i, want := range expected {
			got := report.Models[g*9+i+1]
			ev.Check("failure position and classification "+group.name, got.Collection == group.name && got.Index == i+1 && got.Reason == want, "status=%+v want=%v", got, want)
		}
	}
	t.Run("image-public-call", func(t *testing.T) {
		ev := run.Case(t, "CAT-T08-loaded-image-public-call")
		fixture, bytes := loadFixture(t, imagesProtocol, "generation.json")
		ev.Fixture("generation.json", bytes)
		sc := fixture.Scenarios[0]
		w := newWorldWith(t, func(c *ai.Config) { configureImages(c); c.Catalog = &effective }, tenantA)
		installImagesBinding(w, tenantA)
		enqueue(ev, w, sc.replies(t)...)
		invokeImagesScenario(t, ev, w, sc)
	})
	t.Run("classifier-public-call", func(t *testing.T) {
		ev := run.Case(t, "CAT-T08-loaded-classifier-public-call")
		fixture, bytes := loadFixture(t, classifierProtocol, "choice.json")
		ev.Fixture("choice.json", bytes)
		sc := fixture.Scenarios[0]
		w := newWorldWith(t, func(c *ai.Config) { configureClassifier(c); c.Catalog = &effective }, tenantA)
		installClassifierBinding(w, tenantA)
		enqueue(ev, w, sc.replies(t)...)
		invokeClassifierScenario(t, ev, w, sc)
	})
}

func TestLocalCatalogUnsupportedCapabilities(t *testing.T) {
	ev := run.Case(t, "CAT-T08-explicitly-unsupported-classifier-capability")
	bad := strings.Replace(localClassifierJSON, `"choice","score","bool"`, `"future-question"`, 1)
	effective, report, err := catalog.Load(localCatalogFile(`{"version":"unsupported","models":[],"classifierModels":[` + bad + `]}`))
	ev.Record("report", report)
	ev.Check("unsupported question kind never becomes usable", err == nil && len(effective.ClassifierModels) == 0 && report.Failed == 1 && report.Models[0].Reason == catalog.UnsupportedDefinition, "err=%v report=%+v", err, report)
}

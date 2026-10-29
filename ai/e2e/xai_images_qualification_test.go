package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestXAIImagesNativeQualificationProcess(t *testing.T) {
	f := newImageQualificationFixture(t)
	for _, name := range []string{"complete", "missing-scene", "controlled", "fake-pass", "source", "catalog", "ticks", "image", "extra-request", "missing-claim", "producer-binding", "missing-observer", "audit-null", "result-null", "consistent-image", "consistent-ticks", "manifest-case", "assertion-case", "proxy-remote", "authority-id", "authority-scope", "producer-source", "missing-approval", "claim-case", "claim-duplicate"} {
		t.Run(name, func(t *testing.T) {
			p := f.packet(t)
			switch name {
			case "claim-case":
				imageQualificationMutate(t, filepath.Join(p.bundle, "process-authorization-receipt.json"), func(m map[string]any) { m["Session"] = m["session"]; delete(m, "session") })
			case "claim-duplicate":
				path := filepath.Join(p.bundle, "process-authorization-receipt.json")
				raw, _ := os.ReadFile(path)
				raw = []byte(strings.Replace(string(raw), `"schema": 1,`, `"schema": 1, "schema": 1,`, 1))
				mustWriteXAINative(t, path, raw)
			case "manifest-case":
				imageQualificationMutate(t, filepath.Join(p.bundle, "manifest.json"), func(m map[string]any) { m["Module"] = m["module"]; delete(m, "module") })
				imageQualificationRefreshProducer(t, p)
			case "assertion-case":
				imageQualificationMutate(t, filepath.Join(p.bundle, "XAIIMG-live-generation-low", "assertions.json"), func(m map[string]any) { m["Status"] = m["status"]; delete(m, "status") })
				imageQualificationRefresh(t, p)
			case "proxy-remote":
				imageQualificationMutate(t, filepath.Join(p.bundle, "xai-images-candidate-report.json"), func(m map[string]any) { m["proxy"] = "http://example.invalid:7890" })
				imageQualificationProxy(t, p, "http://example.invalid:7890")
			case "authority-id", "authority-scope":
				imageQualificationRemoveAuthority(t, p, name)
			case "producer-source":
				imageQualificationMutate(t, filepath.Join(p.attempt, "receipt.json"), func(m map[string]any) { m["source"].(map[string]any)["files"].(map[string]any)["ai/images.go"] = "bad" })
			case "missing-approval":
				imageQualificationMutate(t, filepath.Join(p.attempt, "receipt.json"), func(m map[string]any) {
					for k := range m["selectedInputs"].(map[string]any) {
						if filepath.Base(k) == "user-approval.md" {
							delete(m["selectedInputs"].(map[string]any), k)
						}
					}
				})
			case "consistent-image":
				imageQualificationReplaceImage(t, p)
			case "consistent-ticks":
				imageQualificationSetTicks(t, p, "generation-auto", 610000000)
				imageQualificationSetTicks(t, p, "five-reference-edit", 1290000000)
				imageQualificationRefresh(t, p)
			case "missing-scene":
				imageQualificationMutate(t, filepath.Join(p.bundle, "xai-images-candidate-report.json"), func(m map[string]any) { m["scenarios"] = m["scenarios"].([]any)[:5] })
			case "controlled":
				imageQualificationMutate(t, filepath.Join(p.bundle, "xai-images-candidate-report.json"), func(m map[string]any) { m["mode"] = "controlled" })
			case "fake-pass":
				imageQualificationMutate(t, filepath.Join(p.bundle, "XAIIMG-live-generation-low", "assertions.json"), func(m map[string]any) { m["assertions"].([]any)[0].(map[string]any)["pass"] = false })
				imageQualificationRefresh(t, p)
			case "source":
				imageQualificationMutate(t, filepath.Join(p.bundle, "execution-source.json"), func(m map[string]any) { delete(m, "ai/images.go") })
			case "catalog":
				imageQualificationMutate(t, filepath.Join(p.bundle, "model-catalog.json"), func(m map[string]any) {
					m["imageModels"].([]any)[0].(map[string]any)["capabilities"].(map[string]any)["maxOutputImages"] = 2
				})
			case "ticks":
				imageQualificationMutate(t, filepath.Join(p.bundle, "XAIIMG-live-generation-low", "result.json"), func(m map[string]any) {
					m["metadata"].(map[string]any)["attempts"].([]any)[0].(map[string]any)["reportedCostUsd"] = .05
				})
				imageQualificationRefresh(t, p)
			case "image":
				os.WriteFile(filepath.Join(p.bundle, "XAIIMG-live-generation-low", "output-0.image"), []byte("broken"), 0600)
				imageQualificationRefresh(t, p)
			case "extra-request":
				path := filepath.Join(p.bundle, "XAIIMG-live-generation-low", "http.json")
				var exchanges []any
				raw, _ := os.ReadFile(path)
				mustUnmarshal(t, raw, &exchanges)
				writeXAINativeJSON(t, path, append(exchanges, exchanges[0]))
				imageQualificationRefresh(t, p)
			case "missing-claim":
				os.Remove(filepath.Join(p.bundle, "process-authorization-receipt.json"))
			case "producer-binding":
				imageQualificationMutate(t, filepath.Join(p.attempt, "receipt.json"), func(m map[string]any) { m["bundles"].([]any)[0].(map[string]any)["manifestSHA256"] = "bad" })
			case "audit-null":
				imageQualificationMutate(t, filepath.Join(p.bundle, "audit.json"), func(m map[string]any) { m["findings"] = nil })
				imageQualificationMutate(t, filepath.Join(p.attempt, "receipt.json"), func(m map[string]any) {
					m["bundles"].([]any)[0].(map[string]any)["auditSHA256"] = imageQualificationHash(t, filepath.Join(p.bundle, "audit.json"))
				})
			case "result-null":
				imageQualificationMutate(t, filepath.Join(p.bundle, "XAIIMG-live-generation-low", "result.json"), func(m map[string]any) { m["usage"] = nil })
				imageQualificationRefresh(t, p)
			case "missing-observer":
				os.Remove(filepath.Join(p.bundle, "XAIIMG-live-generation-low", "observations.json"))
				imageQualificationRefresh(t, p)
			}
			proof, code := p.consume(t, name)
			ev := run.Case(t, "XAIIMG-qualification-assert-"+name)
			want := name == "complete"
			if ev.Check("only complete independent contract qualifies", (code == 0) == want && (proof["pass"] == true) == want, "exit=%d pass=%v", code, proof["pass"]) && want {
				imageQualificationOuter(t, p, proof)
			}
		})
	}
}

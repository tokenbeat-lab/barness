package minimaxlive_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func recovery5Adversarial(t *testing.T, f *finalRenewalFixture, mode string) bool {
	t.Helper()
	mutate := func(name, key string, value any) {
		path := filepath.Join(f.root, name)
		var v map[string]any
		json.Unmarshal(mustReadFixture(t, path), &v)
		v[key] = value
		b, _ := json.Marshal(v)
		os.WriteFile(path, b, 0600)
	}
	human := filepath.ToSlash(strings.TrimPrefix(filepath.Join(f.base, "human-approval.json"), f.root+string(filepath.Separator)))
	review := filepath.ToSlash(strings.TrimPrefix(filepath.Join(f.base, "source-review.json"), f.root+string(filepath.Separator)))
	actual := ".evidence/provider-sites/final-acceptance-20261020/recovery5-authorization.json"
	parent := ".evidence/minimax-launches/" + recovery5ID + "-formal-zai-responses.json"
	consumer := fifthOut + "formal-zai-responses-consume-command.json"
	report := fifthOut + "formal-zai-responses/fifth-first/regional-report.json"
	reseal := func() {
		stopPath := filepath.Join(f.root, fourthStop)
		var stop map[string]any
		json.Unmarshal(mustReadFixture(t, stopPath), &stop)
		files := stop["files"].(map[string]any)
		for name := range files {
			if b, e := os.ReadFile(filepath.Join(f.root, name)); e == nil {
				files[name] = minimaxlive.Hash(b)
			}
		}
		total := int64(0)
		for name := range files {
			if v, e := os.Stat(filepath.Join(f.root, name)); e == nil {
				total += v.Size()
			}
		}
		stop["fileBytes"] = total
		b, _ := json.Marshal(stop)
		os.WriteFile(stopPath, b, 0600)
		mutate(fourthClosure, "stopIndexSHA256", minimaxlive.Hash(b))
	}
	switch mode {
	case "first-proof-null-exit":
		mutate(consumer, "exitCode", nil)
	case "first-proof-no-exit":
		path := filepath.Join(f.root, consumer)
		var v map[string]any
		json.Unmarshal(mustReadFixture(t, path), &v)
		delete(v, "exitCode")
		b, _ := json.Marshal(v)
		os.WriteFile(path, b, 0600)
	case "first-proof-missing":
		delete(f.grant, "firstRouteProof")
	case "first-proof-parent":
		mutate(parent, "exitCode", 1)
	case "first-proof-report":
		mutate(report, "gitCommit", strings.Repeat("a", 40))
	case "first-proof-consumer":
		mutate(consumer, "exitCode", 1)
	case "first-proof-argv":
		mutate(consumer, "argv", []string{"echo", "PASS"})
	case "first-proof-case":
		mutate(report, "expected", []string{"selected-subset"})
	case "first-proof-path":
		f.grant["firstRouteProof"].(map[string]string)[".evidence/../.env"] = strings.Repeat("a", 64)
	case "first-proof-source":
		mutate(fifthOut+"formal-zai-responses/fifth-first/regional-execution-source.json", "ai/runtime.go", strings.Repeat("a", 64))
	case "first-proof-binary":
		os.WriteFile(filepath.Join(f.root, fifthOut, "regionalverify"), []byte("changed"), 0600)
	case "guard-source":
		f.grant["spendingGuardSourceSha256"].(map[string]any)["ai/internal/testkit/finalbudget/estimate.go"] = strings.Repeat("a", 64)
	case "guard-human":
		mutate(human, "spendingGuardSourceSHA256", map[string]string{})
	case "guard-review":
		mutate(review, "spendingGuardSourceSHA256", map[string]string{})
	case "guard-actual":
		mutate(actual, "spendingGuardSourceSHA256", map[string]string{})
	case "campaign-budget":
		mutate(human, "campaignMaximumMicrousd", 36000001)
	case "review-reference":
		mutate(review, "requestReference", "synthetic-private-new4-no-paid")
	case "grant-reference":
		f.grant["requestReference"] = "synthetic-private-new4-no-paid"
	case "ledger-chain":
		p := ".evidence/minimax-renewals/" + recovery4ID + "/formal.jsonl"
		b := mustReadFixture(t, filepath.Join(f.root, p))
		b[20] = 'x'
		os.WriteFile(filepath.Join(f.root, p), b, 0600)
		reseal()
	case "duplicate-nonce":
		p := ".evidence/minimax-renewals/" + recovery4ID + "/"
		var c map[string]any
		json.Unmarshal(mustReadFixture(t, filepath.Join(f.root, p, "chat.json")), &c)
		mutate(p+"responses.json", "processNonce", c["processNonce"])
		reseal()
	case "history-extra":
		mutate(fourthStop, "inventedCurrentPass", true)
		reseal()
	case "history-case-alias":
		mutate(fourthStop, "SuccessfulProcesses", 18)
		reseal()
	case "history-path":
		var stop map[string]any
		json.Unmarshal(mustReadFixture(t, filepath.Join(f.root, fourthStop)), &stop)
		files := stop["files"].(map[string]any)
		for name := range files {
			if strings.Contains(name, "/simulated-") {
				delete(files, name)
				break
			}
		}
		files[".evidence/../.env"] = strings.Repeat("a", 64)
		f.writeFifthFixture(t, fourthStop, stop)
		reseal()
	case "history-parent-missing":
		os.Remove(filepath.Join(f.root, ".evidence/minimax-launches/"+recovery4ID+"-openai-chat.json"))
		reseal()
	case "failed-guard":
		mutate(fourthOut+"openai-decisions-usd.json", "attempts", 1)
		reseal()
	case "failed-recorder":
		mutate(fourthStop, "failedNativeRecordedHTTP", 0)
		reseal()
	case "unstarted-six":
		for _, n := range strings.Fields("typesafe-classifier protocol114 text-jev text-luna visual-jev visual-decisions") {
			os.WriteFile(filepath.Join(f.root, ".evidence/minimax-launches/"+recovery4ID+"-"+n+".json"), []byte("{}"), 0600)
		}
	default:
		return false
	}
	if strings.HasPrefix(mode, "first-proof-") && mode != "first-proof-missing" && mode != "first-proof-path" {
		proof := f.grant["firstRouteProof"].(map[string]string)
		for name := range proof {
			proof[name] = minimaxlive.Hash(mustReadFixture(t, filepath.Join(f.root, name)))
		}
		mutate(parent, "reportedFileSha256", proof[report])
		mutate(parent, "nativeConsumerReceiptSha256", proof[consumer])
		proof[parent] = minimaxlive.Hash(mustReadFixture(t, filepath.Join(f.root, parent)))
	}
	// Rebind all outer pins: semantic checks must reject forged contents even if
	// an operator recomputes the surrounding hashes. Historical originals stay in
	// a wholly synthetic owner and are never amended in the real repository.
	if mode == "guard-actual" {
		mutate(human, "actualHumanApprovalSHA256", minimaxlive.Hash(mustReadFixture(t, filepath.Join(f.root, actual))))
	}
	f.grant["humanApprovalSha256"] = minimaxlive.Hash(mustReadFixture(t, filepath.Join(f.root, human)))
	f.grant["sourceReviewSha256"] = minimaxlive.Hash(mustReadFixture(t, filepath.Join(f.root, review)))
	f.write(t, "grant.json", f.grant)
	return true
}

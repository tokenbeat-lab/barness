package e2e

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// This is a read-only source inventory, not a custom pi inference route or an
// additional oracle. The existing oracle verifies its complete dist/source/
// lock provenance first. Scan every data file; the runner's smaller APIS table
// alone cannot establish that a native provider route does not exist.
func TestGLMFrozenRouteIdentity(t *testing.T) {
	ev := run.Case(t, "GLM10-frozen-route-identity")
	ev.ReplayEnv(pioracle.EnableEnv + "=1")
	oracle := openOracle(t)
	_, err := oracle.Models(ctxFor(t), []pioracle.ModelKey{})
	if !ev.Check("sole frozen oracle verifies unchanged pins", err == nil, "%v", err) {
		return
	}
	dir := "../internal/testkit/pioracle/node/node_modules/@earendil-works/pi-ai/dist/providers/data"
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	type model struct{ Provider, API, ID, Type, BaseURL string }
	var targets, china []model
	var inventory strings.Builder
	var standard []model
	for _, f := range files {
		raw, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&inventory, "%x  %s\n", sha256.Sum256(raw), f.Name())
		if f.Name() == ".manifest.json" {
			continue
		}
		var data map[string]map[string]model
		mustUnmarshal(t, raw, &data)
		for _, models := range data {
			for _, m := range models {
				if m.Provider == "zhipu" || m.BaseURL == "https://open.bigmodel.cn/api/paas/v4" || m.BaseURL == "https://open.bigmodel.cn/api/v1" || m.BaseURL == "https://open.bigmodel.cn/api/anthropic" {
					standard = append(standard, m)
				}
				if strings.Contains(m.BaseURL, "open.bigmodel.cn") {
					china = append(china, m)
				}
				if (m.Provider == "zai" || m.Provider == "zai-coding-cn") && (m.ID == "glm-5.3" || m.ID == "glm-5.2") {
					targets = append(targets, m)
				}
			}
		}
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].Provider+targets[i].ID < targets[j].Provider+targets[j].ID })
	sort.Slice(china, func(i, j int) bool { return china[i].Provider+china[i].ID < china[j].Provider+china[j].ID })
	ev.Record("data-inventory", inventory.String())
	ev.Record("native-targets", targets)
	ev.Record("all-China-host-models", china)
	ev.Record("standard-route-matches", standard)
	ev.Check("complete 43-file frozen data SHA is unchanged", len(files) == 43 && fmt.Sprintf("%x", sha256.Sum256([]byte(inventory.String()))) == "8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8", "files=%d", len(files))
	ev.Check("no standard China Provider or endpoint exists in entire data", len(standard) == 0, "%+v", standard)
	ev.Check("only three target model Coding identities exist", len(targets) == 3, "%+v", targets)
	for _, m := range targets {
		base := "https://api.z.ai/api/coding/paas/v4"
		if m.Provider == "zai-coding-cn" {
			base = "https://open.bigmodel.cn/api/coding/paas/v4"
		}
		ev.Check("target retains Coding Provider/API/model/base", m.Type == "chat" && m.API == "openai-completions" && m.BaseURL == base && (m.Provider != "zai-coding-cn" || m.ID == "glm-5.3"), "%+v", m)
	}
	ev.Check("China host inventory has exactly four Coding models", len(china) == 4, "%+v", china)
	for _, m := range china {
		ev.Check("China host is always Coding Chat", m.Provider == "zai-coding-cn" && m.API == "openai-completions" && m.BaseURL == "https://open.bigmodel.cn/api/coding/paas/v4", "%+v", m)
	}
}

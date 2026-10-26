package e2e

import (
	"encoding/json"
	"os"
	"testing"
)

type xaiNativeReviewCase struct {
	name string
	pass bool
	edit func(string)
}

func runXAINativeReview(t *testing.T, group string, cases []xaiNativeReviewCase) {
	t.Helper()
	fixture := newXAIFormalNativeFixture(t)
	for _, sc := range cases {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E17-v5-"+group+"-"+sc.name)
			dir := fixture.copy(t, group+"-"+sc.name)
			if sc.edit != nil {
				sc.edit(dir)
			}
			resealXAINativeManifest(t, dir)
			code, raw := fixture.consume(t, ev, dir)
			var got struct {
				Pass  bool   `json:"pass"`
				Error string `json:"error"`
			}
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatal(err)
			}
			ev.Check("independent original evidence eligibility", (code == 0) == sc.pass && got.Pass == sc.pass, "exit=%d pass=%t want=%t error=%s", code, got.Pass, sc.pass, got.Error)
		})
	}
}

func readXAIReviewArray(t *testing.T, path string) []any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var v []any
	mustUnmarshal(t, raw, &v)
	return v
}

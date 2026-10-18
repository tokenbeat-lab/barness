//go:build live

package live

import (
	"archive/zip"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
)

const glmCaptureIndexPin = "67d90ba4473914c3390677d536b7cc0275879a214a3e0aa184204d55ae367c05"

type glmCaptureRecord struct{ Archive, ArchiveSHA256, ReportPath, ReportSHA256 string }
type glmCaptureIndex struct {
	Schema                                         int
	OriginalLiveSourceCommit, SourceManifestSHA256 string
	Routes                                         map[string]glmCaptureRecord
}
type glmReplayCall struct {
	exchange      exchange
	result        map[string]any
	observations  []map[string]any
	cancelContent any
}
type glmCaptureReplay struct {
	wireMatched map[string]bool
	mu          sync.Mutex
	calls       map[string]glmReplayCall
	order       []string
	next        int
	ack         chan struct{}
}

var glmReplayState *glmCaptureReplay

func TestGLMCaptureReplay(t *testing.T) {
	if glmReportState == nil || glmReportState.Mode != "replay" {
		t.Skip("explicit capture replay not selected")
	}
	driveGLMCandidate(t)
}
func glmFormalSelected() bool { return os.Getenv("BARNESS_GLM_FORMAL") == "1" }
func glmCaptureIndexPath() string {
	if glmFormalSelected() {
		return "testdata/glm-formal/live-captures.json"
	}
	return "testdata/glm-candidate/live-captures.json"
}
func glmOriginalKind() string {
	if glmFormalSelected() {
		return "glm-formal"
	}
	return "glm-candidate-only"
}
func glmCaptureIndexData() ([]byte, error) {
	raw, e := readDecisionsFile(glmCaptureIndexPath(), 16384)
	if e != nil {
		return nil, e
	}
	pin := glmCaptureIndexPin
	if glmFormalSelected() {
		data, e := readDecisionsFile("testdata/glm-formal/live-captures.sha256", 128)
		if e != nil {
			return nil, e
		}
		pin = strings.TrimSpace(string(data))
		decoded, e := hex.DecodeString(pin)
		if e != nil || len(decoded) != 32 || len(pin) != 64 {
			return nil, fmt.Errorf("formal capture index pin invalid")
		}
	}
	if sha256Text(raw) != pin {
		return nil, fmt.Errorf("fixed capture index missing or changed")
	}
	return raw, nil
}
func glmReplayArchiveRelative() string {
	raw, _ := glmCaptureIndexData()
	var index glmCaptureIndex
	json.Unmarshal(raw, &index)
	return index.Routes[glmReportState.Route].Archive
}
func installGLMCaptureReplay(rec *recorder) error {
	raw, e := glmCaptureIndexData()
	var index glmCaptureIndex
	if e != nil || json.Unmarshal(raw, &index) != nil || index.Schema != 1 || index.SourceManifestSHA256 != glmSourcePin {
		return fmt.Errorf("fixed capture index missing or changed")
	}
	rawIndex := append([]byte(nil), raw...)
	entry, ok := index.Routes[glmReportState.Route]
	if !ok {
		return fmt.Errorf("unknown capture route")
	}
	raw, e = readDecisionsFile(os.Getenv("BARNESS_GLM_REPLAY_ARCHIVE"), 32<<20)
	if e != nil || sha256Text(raw) != entry.ArchiveSHA256 {
		return fmt.Errorf("immutable archive missing or changed")
	}
	archive, e := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if e != nil {
		return fmt.Errorf("invalid capture archive")
	}
	files := map[string][]byte{}
	needed := map[string]bool{entry.ReportPath: true}
	prefix := path.Dir(entry.ReportPath)
	for _, sc := range glmFixedPlan.Routes[glmReportState.Route] {
		base := prefix + "/GLM11-" + strings.ReplaceAll(sc.ID, "/", "-") + "/"
		for _, name := range []string{"attempt-1-exchanges.json", "attempt-1-result.json", "observations-1.json", "request.json", "response.sse"} {
			needed[base+name] = true
		}
	}
	total := uint64(0)
	for _, f := range archive.File {
		if !needed[f.Name] {
			continue
		}
		if path.IsAbs(f.Name) || path.Clean(f.Name) != f.Name || strings.HasPrefix(f.Name, "../") || f.UncompressedSize64 > 2<<20 || total+f.UncompressedSize64 > 64<<20 {
			return fmt.Errorf("capture archive exceeds safe path/read bounds")
		}
		total += f.UncompressedSize64
		r, e := f.Open()
		if e != nil {
			return fmt.Errorf("capture entry unavailable")
		}
		data, e := io.ReadAll(io.LimitReader(r, 2<<20+1))
		r.Close()
		if e != nil || len(data) > 2<<20 {
			return fmt.Errorf("capture entry exceeds bound")
		}
		files[f.Name] = data
	}
	reportRaw := files[entry.ReportPath]
	var original glmReport
	if sha256Text(reportRaw) != entry.ReportSHA256 || json.Unmarshal(reportRaw, &original) != nil || original.Mode != "live" || original.Kind != glmOriginalKind() || !original.Complete || !original.AuditPassed || original.SourceManifestHash != glmSourcePin || original.Outcome != "PASS" || original.Route != glmReportState.Route || original.GitCommit != index.OriginalLiveSourceCommit || original.CatalogHash != glmReportState.CatalogHash || original.CatalogVersion != glmReportState.CatalogVersion || len(original.Scenarios) != 22 {
		return fmt.Errorf("original live authority does not match replay")
	}
	state := &glmCaptureReplay{calls: map[string]glmReplayCall{}, wireMatched: map[string]bool{}, ack: make(chan struct{}, 1)}
	prefix = path.Dir(entry.ReportPath)
	for _, sc := range glmFixedPlan.Routes[glmReportState.Route] {
		base := prefix + "/GLM11-" + strings.ReplaceAll(sc.ID, "/", "-") + "/"
		var exchanges []exchange
		var result map[string]any
		var observations []map[string]any
		if json.Unmarshal(files[base+"attempt-1-exchanges.json"], &exchanges) != nil || len(exchanges) != 1 || exchanges[0].Truncated || json.Unmarshal(files[base+"attempt-1-result.json"], &result) != nil || json.Unmarshal(files[base+"observations-1.json"], &observations) != nil {
			return fmt.Errorf("own complete case capture missing")
		}
		if exchanges[0].RequestBody != string(files[base+"request.json"]) || exchanges[0].ResponseBody != string(files[base+"response.sse"]) {
			return fmt.Errorf("capture fixture differs from recorded exchange")
		}
		res, _ := result["result"].(map[string]any)
		msg, _ := res["message"].(map[string]any)
		state.calls[sc.ID] = glmReplayCall{exchange: exchanges[0], result: result, observations: observations, cancelContent: msg["content"]}
		state.order = append(state.order, sc.ID)
	}
	rec.next.(*http.Transport).CloseIdleConnections()
	rec.next = state
	glmReplayState = state
	glmReportState.OriginalLiveGitCommit = original.GitCommit
	glmReportState.OriginalLiveReportSHA256 = entry.ReportSHA256
	glmReportState.CaptureArchiveSHA256 = entry.ArchiveSHA256
	pinPath := "ai/live/glm_replay_test.go:glmCaptureIndexPin"
	if glmFormalSelected() {
		pinPath = "ai/live/testdata/glm-formal/live-captures.sha256"
	}
	return run.Record("capture-replay-source", map[string]any{"captureIndexPinPath": pinPath, "captureIndexPinGitCommit": glmReportState.GitCommit, "captureIndexSchema": index.Schema, "captureIndexSHA256": sha256Text(rawIndex), "archiveSHA256": entry.ArchiveSHA256, "originalLiveReportSHA256": entry.ReportSHA256, "originalLiveGitCommit": original.GitCommit, "networkRequests": 0, "normalization": "only exact Host request/attempt ID and local clock/duration paths; all upstream IDs/signatures/payload data preserved"})
}

func (s *glmCaptureReplay) RoundTrip(req *http.Request) (*http.Response, error) {
	s.mu.Lock()
	if s.next >= len(s.order) {
		s.mu.Unlock()
		return nil, fmt.Errorf("unexpected replay attempt")
	}
	id := s.order[s.next]
	s.next++
	capture := s.calls[id]
	s.mu.Unlock()
	body, e := io.ReadAll(req.Body)
	req.Body.Close()
	if e != nil {
		return nil, e
	}
	var got, want any
	if req.Method != capture.exchange.Method || req.URL.String() != capture.exchange.URL || json.Unmarshal(body, &got) != nil || json.Unmarshal([]byte(capture.exchange.RequestBody), &want) != nil || !glmJSONEqual(got, want) {
		return nil, fmt.Errorf("replay final request differs from exact captured payload")
	}
	s.mu.Lock()
	s.wireMatched[id] = true
	s.mu.Unlock()
	headers := http.Header{}
	for name, value := range capture.exchange.ResponseHeaders {
		headers.Set(name, value)
	}
	var responseBody io.ReadCloser = io.NopCloser(strings.NewReader(capture.exchange.ResponseBody))
	if strings.HasSuffix(id, "/cancel") {
		responseBody = s.cancellationBody(req, capture.exchange.ResponseBody)
	}
	return &http.Response{StatusCode: capture.exchange.Status, Header: headers, Body: responseBody, Request: req}, nil
}

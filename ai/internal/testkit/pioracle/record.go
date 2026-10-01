package pioracle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"maps"
	"regexp"
	"strconv"
)

// Gate is a protocol's differential gate outcome for one case.
type Gate string

const (
	GatePass Gate = "PASS"
	GateFail Gate = "FAIL"
)

// Sides are the per-case facts the oracle cannot see itself: the barness side's
// versions and catalog, and both captured requests plus the shared script.
type Sides struct {
	// BarnessVersions are the Go toolchain and SDK versions; NodeVersion is
	// the runtime the pi runner reported.
	BarnessVersions  map[string]string
	NodeVersion      string
	ModelCatalogHash string
	// PiRequest and BarnessRequest are each side's full request capture
	// (method, path, query, redacted headers, body) as JSON.
	PiRequest      json.RawMessage
	BarnessRequest json.RawMessage
	// Frames is the response script both sides received.
	Frames []byte
}

// Record is one case's differential record, written to the evidence bundle
// (spec Testing Decisions §6).
type Record struct {
	CaseID    string `json:"case_id"`
	Protocol  string `json:"protocol"`
	PiCommit  string `json:"pi_commit"`
	PiVersion string `json:"pi_version"`
	// SDKVersions covers both sides: the pi-side npm packages and Node.js,
	// and the barness-side Go SDKs and toolchain.
	SDKVersions map[string]string `json:"sdk_versions"`
	// ModelCatalogHash is barness-ai's built-in catalog hash;
	// PiModelDataSHA256 is the frozen pi copy's generated model data.
	ModelCatalogHash  string `json:"model_catalog_hash"`
	PiModelDataSHA256 string `json:"pi_model_data_sha256"`
	// RequestSHA256 hashes each side's full request capture; FrameSHA256
	// hashes the response script both sides received.
	RequestSHA256 map[string]string `json:"request_sha256"`
	FrameSHA256   string            `json:"frame_sha256"`
	// FirstDiff is the earliest difference in request → events → result
	// order, whatever its classification; FirstPending is the earliest one
	// that blocks the gate; FirstEventIndex is the event position of the
	// earliest difference inside the event sequence.
	FirstDiff       *Diff `json:"first_diff"`
	FirstPending    *Diff `json:"first_pending"`
	FirstEventIndex *int  `json:"first_event_index"`
	Verdict
	Gate Gate `json:"gate"`
}

var eventIndex = regexp.MustCompile(`^events\[(\d+)\]`)

// Record builds a case record from its verdict and the case's sides.
func (o *Oracle) Record(caseID, protocol string, v Verdict, s Sides) Record {
	versions := maps.Clone(o.sdks)
	if versions == nil {
		versions = map[string]string{}
	}
	maps.Copy(versions, s.BarnessVersions)
	if s.NodeVersion != "" {
		versions["node"] = s.NodeVersion
	}
	r := Record{
		CaseID:            caseID,
		Protocol:          protocol,
		PiCommit:          o.prov.PiCommit,
		PiVersion:         o.prov.PiVersion,
		SDKVersions:       versions,
		ModelCatalogHash:  s.ModelCatalogHash,
		PiModelDataSHA256: o.prov.ModelDataSHA256,
		RequestSHA256:     map[string]string{"pi": SHA256(s.PiRequest), "barness": SHA256(s.BarnessRequest)},
		FrameSHA256:       SHA256(s.Frames),
		Verdict:           v,
		Gate:              GateFail,
	}
	if v.Pass {
		r.Gate = GatePass
	}
	for _, f := range v.Findings {
		d := f.Diff
		if r.FirstDiff == nil {
			r.FirstDiff = &d
		}
		if r.FirstPending == nil && f.Classification == Pending {
			r.FirstPending = &d
		}
		if m := eventIndex.FindStringSubmatch(d.Path); m != nil && r.FirstEventIndex == nil {
			// The pattern guarantees digits.
			i, _ := strconv.Atoi(m[1])
			r.FirstEventIndex = &i
		}
	}
	return r
}

// SHA256 is the hex SHA-256 of data.
func SHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

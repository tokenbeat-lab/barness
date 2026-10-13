package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"slices"
	"strings"
	"sync"
)

// These traces retain shared immutable fixture slices, never a second copy of
// a design-load body. Only the first divergent window is kept. Diagnostics are
// synthetic-only and go through the same evidence redaction audit as results.
type pressureIntegrity struct {
	mu            sync.Mutex
	chunks        [][]byte
	chunk, within int
	hash          hash.Hash
	report        pressureIntegrityReport
}

type pressureIntegrityReport struct {
	RequestID       string              `json:"requestId"`
	ExpectedBytes   int64               `json:"expectedBytes"`
	ReadBytes       int64               `json:"readBytes"`
	ReadCalls       int                 `json:"readCalls"`
	SHA256          string              `json:"sha256"`
	Complete        bool                `json:"complete"`
	NextFrame       int                 `json:"nextFrame"`
	NextFrameOffset int                 `json:"nextFrameOffset"`
	FirstDifference *pressureDifference `json:"firstDifference,omitempty"`
}

type pressureDifference struct {
	Offset      int64  `json:"offset"`
	ReadCall    int    `json:"readCall"`
	Chunk       int    `json:"chunk"`
	ChunkOffset int    `json:"chunkOffset"`
	Frame       int    `json:"frame"`
	FrameOffset int    `json:"frameOffset"`
	ExpectedHex string `json:"expectedHex"`
	ActualHex   string `json:"actualHex"`
}

func (f *pressureNetworkErrors) expect(id string, chunks [][]byte) {
	trace := &pressureIntegrity{chunks: chunks, hash: sha256.New(), report: pressureIntegrityReport{RequestID: id}}
	for _, c := range chunks {
		trace.report.ExpectedBytes += int64(len(c))
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.streams == nil {
		f.streams = map[string]*pressureIntegrity{}
	}
	f.streams[id] = trace
}

func (b *pressureIntegrity) read(p []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.report.ReadCalls++
	_, _ = b.hash.Write(p)
	for len(p) > 0 {
		var want []byte
		if b.chunk < len(b.chunks) {
			want = b.chunks[b.chunk][b.within:]
		}
		n := min(len(p), len(want))
		if b.report.FirstDifference == nil && (n == 0 || !bytes.Equal(p[:n], want[:n])) {
			i := 0
			for i < n && p[i] == want[i] {
				i++
			}
			frame, frameOffset := b.frameAt(b.within + i)
			b.report.FirstDifference = &pressureDifference{Offset: b.report.ReadBytes + int64(i), ReadCall: b.report.ReadCalls,
				Chunk: b.chunk, ChunkOffset: b.within + i, Frame: frame, FrameOffset: frameOffset,
				ExpectedHex: hex.EncodeToString(want[i:min(len(want), i+64)]), ActualHex: hex.EncodeToString(p[i:min(len(p), i+64)])}
		}
		if n == 0 {
			b.report.ReadBytes += int64(len(p))
			break
		}
		b.report.ReadBytes += int64(n)
		b.within += n
		p = p[n:]
		if b.within == len(b.chunks[b.chunk]) {
			b.chunk++
			b.within = 0
		}
	}
}

// Pressure fixtures use LF frames and coalesce only at frame boundaries.
// Locate the earliest byte mismatch in the original SSE sequence, independently
// of HTTP body Read boundaries and the SDK's decoder buffering.
func (b *pressureIntegrity) frameAt(within int) (int, int) {
	frame := 0
	for _, c := range b.chunks[:b.chunk] {
		frame += bytes.Count(c, []byte("\n\n"))
	}
	if b.chunk == len(b.chunks) {
		return frame, 0
	}
	prefix := b.chunks[b.chunk][:within]
	frame += bytes.Count(prefix, []byte("\n\n"))
	start := bytes.LastIndex(prefix, []byte("\n\n"))
	if start < 0 {
		return frame, within
	}
	return frame, within - start - 2
}

func (f *pressureNetworkErrors) integrity() []pressureIntegrityReport {
	f.mu.Lock()
	defer f.mu.Unlock()
	var reports []pressureIntegrityReport
	for _, trace := range f.streams {
		trace.mu.Lock()
		r := trace.report
		r.SHA256 = hex.EncodeToString(trace.hash.Sum(nil))
		r.Complete = r.FirstDifference == nil && r.ReadBytes == r.ExpectedBytes
		r.NextFrame, r.NextFrameOffset = trace.frameAt(trace.within)
		trace.mu.Unlock()
		reports = append(reports, r)
	}
	slices.SortFunc(reports, func(a, b pressureIntegrityReport) int { return strings.Compare(a.RequestID, b.RequestID) })
	return reports
}

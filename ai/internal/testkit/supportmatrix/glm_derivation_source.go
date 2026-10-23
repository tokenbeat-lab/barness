package supportmatrix

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type glmRawIndex struct {
	Schema                                         int
	OriginalLiveSourceCommit, SourceManifestSHA256 string
	Routes                                         map[string]struct{ Archive, ArchiveSHA256, ReportPath, ReportSHA256 string }
}
type glmRawCapture struct {
	indexHash, archiveHash, reportHash, supportHash, commit string
	files                                                   map[string][]byte
}

func glmRepositoryRoot() (string, error) {
	root, e := os.Getwd()
	if e != nil {
		return "", e
	}
	for {
		if _, e = os.Stat(filepath.Join(root, "go.mod")); e == nil {
			return root, nil
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", glmBad("source root missing")
		}
		root = parent
	}
}

// The importer has its own identity; it is distinct from the immutable model
// execution source. Consumer fixes can re-derive views without a vendor call.
func glmImporterIdentity() (string, string, error) {
	root, e := glmRepositoryRoot()
	if e != nil {
		return "", "", e
	}
	dir := filepath.Join(root, "ai/internal/testkit/supportmatrix")
	entries, e := os.ReadDir(dir)
	if e != nil {
		return "", "", e
	}
	files := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		raw, e := glmRead(filepath.Join(dir, entry.Name()))
		if e != nil {
			return "", "", e
		}
		files[entry.Name()] = glmHash(raw)
	}
	raw, _ := json.Marshal(files)
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = root
	commit, e := cmd.Output()
	if e != nil {
		return "", "", e
	}
	return strings.TrimSpace(string(commit)), glmHash(raw), nil
}

// The raw index pin is versioned outside any supplied archive or support view.
// Its original ZIP hash authenticates every original byte, not editable claims.
func loadGLMRawCapture(route string) (glmRawCapture, error) {
	var out glmRawCapture
	root, e := glmRepositoryRoot()
	if e != nil {
		return out, e
	}
	base := filepath.Join(root, "ai/live/testdata/glm-formal")
	raw, e := glmRead(filepath.Join(base, "live-captures.json"))
	pin, e2 := glmRead(filepath.Join(base, "live-captures.sha256"))
	if e != nil || e2 != nil || len(strings.TrimSpace(string(pin))) != 64 || glmHash(raw) != strings.TrimSpace(string(pin)) {
		return out, glmBad("trusted raw capture index changed")
	}
	var idx glmRawIndex
	if json.Unmarshal(raw, &idx) != nil || idx.Schema != 1 || idx.SourceManifestSHA256 != GLMSourcePin {
		return out, glmBad("raw index identity invalid")
	}
	entry, ok := idx.Routes[route]
	if !ok || path.IsAbs(entry.Archive) || path.Clean(entry.Archive) != entry.Archive || strings.HasPrefix(entry.Archive, "../") {
		return out, glmBad("raw route/archive identity invalid")
	}
	// Archive is the immutable capture-time locator. Resolve the same pinned
	// bytes from tracked fixtures, never from a host's historical .evidence.
	digest, digestErr := hex.DecodeString(entry.ArchiveSHA256)
	if digestErr != nil || len(digest) != 32 {
		return out, glmBad("raw archive digest invalid")
	}
	zipRaw, e := glmRead(filepath.Join(base, "archives", entry.ArchiveSHA256+".zip"))
	if e != nil || glmHash(zipRaw) != entry.ArchiveSHA256 {
		return out, glmBad("immutable original archive missing or changed")
	}
	z, e := zip.NewReader(bytes.NewReader(zipRaw), int64(len(zipRaw)))
	if e != nil {
		return out, e
	}
	prefix := path.Dir(entry.ReportPath) + "/"
	files := map[string][]byte{}
	total := uint64(0)
	for _, file := range z.File {
		if file.FileInfo().IsDir() || !strings.HasPrefix(file.Name, prefix) {
			continue
		}
		name := strings.TrimPrefix(file.Name, prefix)
		if path.IsAbs(name) || path.Clean(name) != name || strings.HasPrefix(name, "../") || files[name] != nil || file.UncompressedSize64 > 16<<20 || total+file.UncompressedSize64 > 128<<20 {
			return out, glmBad("unsafe or duplicate original capture entry")
		}
		total += file.UncompressedSize64
		r, e := file.Open()
		if e != nil {
			return out, e
		}
		// This archive is hash-authenticated and its declared sizes are bounded
		// above. Read exactly once instead of repeatedly growing ReadAll buffers;
		// verify EOF so metadata never weakens the actual byte ceiling.
		data := make([]byte, file.UncompressedSize64)
		_, e = io.ReadFull(r, data)
		var extra [1]byte
		n, end := io.ReadFull(r, extra[:])
		r.Close()
		if e != nil || n != 0 || end != io.EOF {
			return out, glmBad("original capture exceeds bound")
		}
		files[name] = data
	}
	report := files[path.Base(entry.ReportPath)]
	if len(files) < 300 || glmHash(report) != entry.ReportSHA256 || len(files[ReportFile]) == 0 {
		return out, glmBad("original formal/support report missing")
	}
	out = glmRawCapture{indexHash: glmHash(raw), archiveHash: entry.ArchiveSHA256, reportHash: entry.ReportSHA256, supportHash: glmHash(files[ReportFile]), commit: idx.OriginalLiveSourceCommit, files: files}
	return out, nil
}

func glmOriginalMatches(dir string, raw glmRawCapture) error {
	// The owned complete path set was bounded before payload reads. Iterate
	// only trusted archive names here; a second unbounded walk would reopen
	// the same denial-of-service path during independent byte comparison.
	for name, expected := range raw.files {
		data, e := glmRead(filepath.Join(dir, filepath.FromSlash(name)))
		if e != nil || glmHash(data) != glmHash(expected) {
			return glmBad("original capture bytes differ from immutable archive")
		}
	}
	return nil
}

// Match the claimed importer commit's complete source set, including additions
// and deletions. A valid-looking unrelated commit cannot certify these bytes.
func glmImporterCommitHash(commit string) (string, error) {
	root, e := glmRepositoryRoot()
	if e != nil {
		return "", e
	}
	cmd := exec.Command("git", "archive", commit, "ai/internal/testkit/supportmatrix")
	cmd.Dir = root
	pipe, e := cmd.StdoutPipe()
	if e != nil {
		return "", e
	}
	if e = cmd.Start(); e != nil {
		return "", e
	}
	raw, e := io.ReadAll(io.LimitReader(pipe, 16<<20+1))
	if e != nil || len(raw) > 16<<20 {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return "", glmBad("importer source archive exceeds bound")
	}
	if e = cmd.Wait(); e != nil {
		return "", glmBad("importer commit unavailable")
	}
	files := map[string]string{}
	reader := tar.NewReader(bytes.NewReader(raw))
	for {
		header, e := reader.Next()
		if e == io.EOF {
			break
		}
		if e != nil {
			return "", e
		}
		name := path.Base(header.Name)
		if header.Typeflag != tar.TypeReg || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, e := io.ReadAll(reader)
		if e != nil {
			return "", e
		}
		files[name] = glmHash(data)
	}
	encoded, _ := json.Marshal(files)
	return glmHash(encoded), nil
}

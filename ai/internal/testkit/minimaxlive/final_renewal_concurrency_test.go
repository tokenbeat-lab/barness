package minimaxlive_test

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func TestFinalRenewalContenderChild(t *testing.T) {
	session := os.Getenv("BARNESS_FINAL_RENEWAL_CONTENDER")
	if session == "" {
		t.Skip("private synthetic process only")
	}
	fmt.Println("CONTENDER_READY")
	if _, e := bufio.NewReader(os.Stdin).ReadString('\n'); e != nil {
		t.Fatal(e)
	}
	path, e := minimaxlive.StageLedgerPath("formal")
	if e != nil {
		t.Fatal(e)
	}
	id, _ := minimaxlive.Select("chat", minimaxlive.M3, "standard")
	l := minimaxlive.Ledger{Path: path, Stage: "formal", Session: session, Identity: id}
	caseID := os.Getenv("BARNESS_FINAL_RENEWAL_CASE")
	if caseID == "" {
		caseID = "full-stream"
	}
	if _, e = l.Reserve(caseID, 1024, 2048); e == nil {
		fmt.Println("CONTENDER_WON")
	} else {
		fmt.Println("CONTENDER_REFUSED", e)
	}
}

func testFinalRenewalContenders(t *testing.T, root, mode string) {
	t.Helper()
	linked := mode == "linked-worktree"
	dir := root
	if linked {
		dir = filepath.Join(t.TempDir(), "linked")
		c := exec.Command("git", "worktree", "add", "-q", "--detach", dir, "HEAD")
		if raw, e := c.CombinedOutput(); e != nil {
			t.Fatalf("linked fixture %v %s", e, raw)
		}
		t.Cleanup(func() {
			c := exec.Command("git", "worktree", "remove", "--force", dir)
			if raw, e := c.CombinedOutput(); e != nil {
				t.Errorf("remove private linked fixture %v %s", e, raw)
			}
		})
	}
	type contender struct {
		cmd   *exec.Cmd
		input io.WriteCloser
		out   *bufio.Reader
	}
	var children []contender
	n := 2
	if mode != "concurrent" {
		n = 1
	}
	exe, _ := os.Executable()
	for i := 0; i < n; i++ {
		c := exec.Command(exe, "-test.run=^TestFinalRenewalContenderChild$", "-test.v")
		c.Dir = dir
		c.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_MINIMAX_FORMAL_RENEWAL=" + finalRenewalID, fmt.Sprintf("BARNESS_FINAL_RENEWAL_CONTENDER=contender-%d", i)}
		if mode == "same-session-process" {
			c.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_MINIMAX_FORMAL_RENEWAL=" + finalRenewalID, "BARNESS_FINAL_RENEWAL_CONTENDER=final-chat", "BARNESS_FINAL_RENEWAL_CASE=full-complete"}
		}
		input, e := c.StdinPipe()
		if e != nil {
			t.Fatal(e)
		}
		output, e := c.StdoutPipe()
		if e != nil {
			t.Fatal(e)
		}
		c.Stderr = c.Stdout
		if e = c.Start(); e != nil {
			t.Fatal(e)
		}
		r := bufio.NewReader(output)
		for {
			line, e := r.ReadString('\n')
			if e != nil {
				t.Fatalf("contender startup %v %s", e, line)
			}
			if strings.Contains(line, "CONTENDER_READY") {
				break
			}
		}
		children = append(children, contender{c, input, r})
	}
	for _, c := range children {
		if _, e := io.WriteString(c.input, "start\n"); e != nil {
			t.Fatal(e)
		}
		c.input.Close()
	}
	won, refused := 0, 0
	for _, c := range children {
		raw, e := io.ReadAll(c.out)
		if e != nil {
			t.Fatal(e)
		}
		if e = c.cmd.Wait(); e != nil {
			t.Fatalf("contender process %v %s", e, raw)
		}
		if strings.Contains(string(raw), "CONTENDER_WON") {
			won++
		}
		if strings.Contains(string(raw), "CONTENDER_REFUSED") {
			refused++
		}
	}
	if mode != "concurrent" && (won != 0 || refused != 1) || mode == "concurrent" && (won != 1 || refused != 1) {
		t.Fatalf("route claim/worktree isolation violated: won=%d refused=%d", won, refused)
	}
	fmt.Printf("PASS contenders won=%d refused=%d linked=%v\n", won, refused, linked)
}

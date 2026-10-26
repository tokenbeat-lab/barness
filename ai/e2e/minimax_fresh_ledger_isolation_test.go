package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func TestMiniMaxFreshLedgerIsolationProcess(t *testing.T) {
	for _, name := range []string{"parent-git-dir", "parent-git-overrides", "independent-fixtures"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "MM19-private-ledger-"+name)
			parent := miniMaxPrivateSourceNamespace(t)
			store := filepath.Join(parent.dir, "barness-live-process-grants")
			if err := os.Mkdir(store, 0700); err != nil {
				t.Fatal(err)
			}
			sentinel := []byte("TEST_ONLY parent namespace must remain unchanged\n")
			path := filepath.Join(store, "sentinel")
			if err := os.WriteFile(path, sentinel, 0600); err != nil {
				t.Fatal(err)
			}
			defer func() {
				entries, err := os.ReadDir(store)
				raw, readErr := os.ReadFile(path)
				ev.Check("parent namespace inventory and bytes unchanged", err == nil && readErr == nil && len(entries) == 1 && string(raw) == string(sentinel), "parent sentinel changed")
				ev.Record("parent-inventory", map[string]any{"files": len(entries), "sentinelSHA256": evidence.Digest(raw), "nativeLaunched": false})
			}()
			t.Setenv("GIT_DIR", parent.dir)
			t.Setenv("GIT_WORK_TREE", parent.root)
			if name == "parent-git-overrides" {
				t.Setenv("GIT_COMMON_DIR", parent.dir)
				t.Setenv("GIT_OBJECT_DIRECTORY", filepath.Join(parent.dir, "objects"))
				t.Setenv("GIT_NAMESPACE", "TEST_ONLY_parent")
			}
			// Inspect the actual native command's environment using Git's public
			// read-only process boundary. Never start native code in this RED.
			cmd := miniMaxFreshLedgerCommand(t, "TEST_ONLY_NOT_EXECUTED", "chat", "TEST_ONLY_UNUSED_GRANT", t.TempDir())
			common := miniMaxGitProbe(t, cmd, "rev-parse", "--path-format=absolute", "--git-common-dir")
			declared := miniMaxEnvValue(cmd.Env, "GIT_DIR")
			ev.Record("read-only-discovery", map[string]any{"argv": cmd.Args, "declaredPrivateGit": declared, "discoveredCommonGit": common, "parentGit": parent.dir, "nativeLaunched": false, "claims": 0, "vendorCalls": 0})
			if !ev.Check("native command owns a separate declared private namespace", declared != "" && common == declared && common != parent.dir, "declared=%s common=%s parent=%s", declared, common, parent.dir) {
				return
			}
			ev.Check("genuine source commit preserved", miniMaxGitProbe(t, cmd, "rev-parse", "HEAD") == parent.commit && miniMaxGitProbe(t, cmd, "cat-file", "-t", parent.commit) == "commit", "source HEAD differs")
			ev.Check("genuine checkout root preserved", miniMaxGitProbe(t, cmd, "rev-parse", "--show-toplevel") == parent.root, "source root differs")
			parentCmd := *cmd
			parentCmd.Env = os.Environ()
			ev.Check("parent verifier sees the same namespace", os.Getenv("GIT_DIR") == declared && os.Getenv("GIT_WORK_TREE") == parent.root && miniMaxGitProbe(t, &parentCmd, "rev-parse", "--path-format=absolute", "--git-common-dir") == common, "verifier namespace mismatch")
			ev.Check("ambient common and object overrides are absent", miniMaxEnvValue(cmd.Env, "GIT_COMMON_DIR") == "" && miniMaxEnvValue(cmd.Env, "GIT_OBJECT_DIRECTORY") == "" && miniMaxEnvValue(cmd.Env, "GIT_NAMESPACE") == "", "inherited Git overrides")
			other := miniMaxFreshLedgerCommand(t, "TEST_ONLY_NOT_EXECUTED", "chat", "TEST_ONLY_UNUSED_GRANT", t.TempDir())
			otherCommon := miniMaxGitProbe(t, other, "rev-parse", "--path-format=absolute", "--git-common-dir")
			ev.Check("independent fixture commands never share common store", otherCommon != common && otherCommon != parent.dir, "common=%s other=%s", common, otherCommon)
			entries, err := os.ReadDir(store)
			raw, readErr := os.ReadFile(path)
			ev.Check("parent namespace inventory and bytes unchanged", err == nil && readErr == nil && len(entries) == 1 && string(raw) == string(sentinel), "parent sentinel changed")
			ev.Record("namespace-inventory", map[string]any{"parentFiles": len(entries), "sentinelSHA256": evidence.Digest(raw), "childA": common, "childB": otherCommon, "sourceCommit": parent.commit})
		})
	}
}

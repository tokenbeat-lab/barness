package minimaxlive

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func sixthPinnedSelection(root *os.Root, owner string, stop sixthStop) error {
	if e := sixthNonRegularHistory(root, owner, stop); e != nil {
		return e
	}
	if e := cohortPinnedSelection(root, stop.Files, stop.NonRegular, map[string]int{strings.TrimSuffix(sixthCollection, "/"): 30195, ".evidence/provider-sites/final-acceptance-20261020/recovery6-preparation": 33}); e != nil {
		return e
	}
	dir, e := root.Open(".evidence/minimax-launches")
	if e != nil {
		return e
	}
	defer dir.Close()
	entries, e := dir.ReadDir(4097)
	if e != nil || len(entries) > 4096 {
		return fmt.Errorf("closed sixth launch selection unavailable")
	}
	found := 0
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), Recovery6RenewalID+"-") {
			if entry.Name() != Recovery6RenewalID+"-formal-zai-responses.json" && entry.Name() != Recovery6RenewalID+"-mm-global-formal-m3-chat.json" {
				return fmt.Errorf("closed sixth acquired extra claim")
			}
			found++
		}
	}
	if found != 2 {
		return fmt.Errorf("closed sixth exact2 canonical parents required")
	}
	return nil
}
func sixthNonRegularHistory(root *os.Root, owner string, stop sixthStop) error {
	if len(stop.NonRegular) != 1 {
		return fmt.Errorf("closed sixth exact one rejection fixture required")
	}
	for name, n := range stop.NonRegular {
		prefix := sixthCollection + "no-key-live/"
		suffix := "/formal-corrupt-symlink/formal-integrity.json"
		middle := strings.TrimSuffix(strings.TrimPrefix(name, prefix), suffix)
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, suffix) || middle == "" || strings.Contains(middle, "/") || n.Kind != "symlink" || n.Followed || n.Mode != 41453 || n.Target != filepath.Join(owner, strings.TrimSuffix(name, "formal-integrity.json")+"seal.json") || stop.Files[name] != "" {
			return fmt.Errorf("closed sixth rejection link metadata differs")
		}
		info, e := root.Lstat(name)
		if e != nil || info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("closed sixth negative link changed")
		}
		target, e := root.Readlink(name)
		if e != nil || target != n.Target {
			return fmt.Errorf("closed sixth negative link target changed")
		}
	}
	return nil
}

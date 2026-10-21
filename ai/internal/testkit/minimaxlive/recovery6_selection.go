package minimaxlive

import (
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
)

// The stopped fifth collection contains independently checked binaries and
// no-key bundles. Its sole intentional negative link is metadata, never data.
// Check additions as well as every pinned byte; older archive bounds stay put.
func fifthPinnedSelection(root *os.Root, owner string, stop fifthStop) error {
	if e := fifthNonRegularHistory(root, owner, stop); e != nil {
		return e
	}
	return cohortPinnedSelection(root, stop.Files, stop.NonRegular, map[string]int{strings.TrimSuffix(fifthCollection, "/"): 31103, ".evidence/provider-sites/final-acceptance-20261020/recovery5-preparation/actual-execution": 31})
}

func cohortPinnedSelection(root *os.Root, files map[string]string, nonRegular map[string]fifthNonRegular, selections map[string]int) error {
	for prefix, expected := range selections {
		want := 0
		for name := range files {
			if strings.HasPrefix(name, prefix+"/") {
				want++
			}
		}
		if want != expected {
			return fmt.Errorf("closed cohort fixed collection/preparation selection differs")
		}
		pending, found, count := []string{prefix}, 0, 0
		for len(pending) > 0 {
			name := pending[len(pending)-1]
			pending = pending[:len(pending)-1]
			before, e := root.Lstat(name)
			if e != nil || !before.IsDir() || len(strings.Split(name, "/")) > 32 {
				return fmt.Errorf("closed cohort directory missing or linked")
			}
			dir, e := root.OpenFile(name, os.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW, 0)
			if e != nil {
				return e
			}
			opened, e := dir.Stat()
			if e != nil || !os.SameFile(before, opened) {
				dir.Close()
				return fmt.Errorf("closed cohort directory changed before selection")
			}
			for {
				entries, readErr := dir.ReadDir(64)
				for _, entry := range entries {
					count++
					path := name + "/" + entry.Name()
					if count > 65536 {
						dir.Close()
						return fmt.Errorf("closed cohort entry ceiling exceeded")
					}
					if entry.Type()&os.ModeSymlink != 0 {
						if _, ok := nonRegular[path]; !ok {
							dir.Close()
							return fmt.Errorf("closed cohort acquired an unpinned link")
						}
					} else if entry.IsDir() {
						pending = append(pending, path)
					} else if !entry.Type().IsRegular() || files[path] == "" {
						dir.Close()
						return fmt.Errorf("closed cohort acquired an unpinned file")
					} else {
						found++
					}
				}
				if readErr != nil {
					dir.Close()
					if readErr != io.EOF {
						return readErr
					}
					break
				}
			}
			after, e := root.Lstat(name)
			if e != nil || !os.SameFile(before, after) || !before.ModTime().Equal(after.ModTime()) {
				return fmt.Errorf("closed cohort directory changed during selection")
			}
		}
		if found == 0 || found != want {
			return fmt.Errorf("closed cohort exact file selection differs")
		}
	}
	return nil
}

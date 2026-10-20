package minimaxlive_test

import "os/exec"

func execFixtureFIFO(path string) error { return exec.Command("mkfifo", path).Run() }

/*
Command git-hound is a Git plugin that helps prevent sensitive data from being committed
into a repository by sniffing potential commits against regular expressions
*/
package main

import (
	"flag"
	"fmt"
	"github.com/ezekg/git-hound/Godeps/_workspace/src/github.com/fatih/color"
	"github.com/ezekg/git-hound/Godeps/_workspace/src/sourcegraph.com/sourcegraph/go-diff/diff"
	"os"
)

var (
	noColor = flag.Bool("no-color", false, "Disable color output")
	config  = flag.String("config", ".githound.yml", "Hound config file")
	bin     = flag.String("bin", "git", "Executable binary to use for git command")
)

func main() {
	flag.Parse()

	if *noColor {
		color.NoColor = true
	}

	hound := &Hound{Config: *config}
	git := &Command{Bin: *bin}

	if ok := hound.New(); ok {
		out, _ := git.Exec("diff", "-U0", "--staged")
		fileDiffs, err := diff.ParseMultiFileDiff([]byte(out))
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}

		hunkCount := 0
		warnc := make(chan string)
		failc := make(chan error)
		donec := make(chan bool)

		for _, fileDiff := range fileDiffs {
			fileName := fileDiff.NewName
			hunks := fileDiff.GetHunks()

			for _, hunk := range hunks {
				go hound.Sniff(fileName, hunk, warnc, failc, donec)
				hunkCount++
			}
		}

		for c := 0; c < hunkCount; {
			select {
			case msg := <-warnc:
				fmt.Print(msg)
			case err := <-failc:
				fmt.Print(err)
				os.Exit(1)
			case <-donec:
				c++
			}
		}
	}

	out, code := git.Exec(flag.Args()...)
	fmt.Print(out)
	os.Exit(code)
}

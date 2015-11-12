package main

import (
	"flag"
	"fmt"
	"github.com/ezekg/git-hound/Godeps/_workspace/src/github.com/fatih/color"
	"github.com/ezekg/git-hound/Godeps/_workspace/src/sourcegraph.com/sourcegraph/go-diff/diff"
	"os"
)

var noColor = flag.Bool("no-color", false, "Disable color output")
var config = flag.String("config", ".githound.yml", "Hound config file")
var bin = flag.String("bin", "git", "Executable binary to use for git command")

func main() {
	flag.Parse()

	if *noColor {
		color.NoColor = true
	}

	hound := &Hound{Config: *config}
	git := &Command{Bin: *bin}

	if ok, _ := hound.New(); ok {
		out, _ := git.Exec("diff", "-U0", "--staged")
		fileDiffs, err := diff.ParseMultiFileDiff([]byte(out))
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}

		for _, fileDiff := range fileDiffs {
			fileName := fileDiff.NewName
			hunks := fileDiff.GetHunks()

			for _, hunk := range hunks {
				err := hound.Sniff(fileName, hunk)
				if err != nil {
					fmt.Print(err)
					os.Exit(1)
				}
			}
		}
	}

	out, code := git.Exec(flag.Args()...)
	fmt.Print(out)
	os.Exit(code)
}

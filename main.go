package main

import (
	"flag"
	"fmt"
	"github.com/ezekg/git-hound/Godeps/_workspace/src/github.com/fatih/color"
	"github.com/ezekg/git-hound/Godeps/_workspace/src/sourcegraph.com/sourcegraph/go-diff/diff"
	"os"
)

func main() {
	var (
		noColor = flag.Bool("no-color", false, "Disable color output")
		config  = flag.String("config", ".githound.yml", "Hound config file")
		bin     = flag.String("bin", "git", "Executable binary to use for git command")
	)

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

			errs := make(chan error)
			go func() {
				for _, hunk := range hunks {
					errs <- hound.Sniff(fileName, hunk)
				}
				close(errs)
			}()

			for err := range errs {
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

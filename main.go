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
	version     = "0.5.3"
	showVersion = flag.Bool("v", false, "Show version")
	noColor     = flag.Bool("no-color", false, "Disable color output")
	config      = flag.String("config", ".githound.yml", "Hound config file")
	bin         = flag.String("bin", "git", "Executable binary to use for git command")
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			color.Red(fmt.Sprintf("error: %s\n", r))
			os.Exit(1)
		}
	}()

	flag.Parse()

	if *showVersion {
		fmt.Printf("%s\n", version)
		os.Exit(0)
	}

	if *noColor {
		color.NoColor = true
	}

	hound := &Hound{config: *config}
	git := &Command{bin: *bin}

	if ok := hound.New(); ok {
		out, _ := git.Exec("diff", "-U0", "--staged")
		fileDiffs, err := diff.ParseMultiFileDiff([]byte(out))
		if err != nil {
			color.Red(fmt.Sprintf("%s\n", err))
			os.Exit(1)
		}

		severeSmellCount := 0
		hunkCount := 0

		smells := make(chan smell)
		done := make(chan bool)

		for _, fileDiff := range fileDiffs {
			fileName := fileDiff.NewName
			hunks := fileDiff.GetHunks()

			for _, hunk := range hunks {
				go func(hunk *diff.Hunk) {
					defer func() {
						if r := recover(); r != nil {
							color.Red(fmt.Sprintf("%s\n", r))
							os.Exit(1)
						}
					}()
					hound.Sniff(fileName, hunk, smells, done)
				}(hunk)
				hunkCount++
			}
		}

		for c := 0; c < hunkCount; {
			select {
			case s := <-smells:
				if s.severity > 1 {
					severeSmellCount++
				}

				switch s.severity {
				case 1:
					color.Yellow(fmt.Sprintf("warning: %s\n", s.String()))
				case 2:
					color.Red(fmt.Sprintf("failure: %s\n", s.String()))
				default:
					color.Red(fmt.Sprintf("error: unknown severity given - %d\n", s.severity))
				}
			case <-done:
				c++
			}
		}

		if severeSmellCount > 0 {
			fmt.Printf("%d severe smell(s) detected - please fix them before you can commit\n", severeSmellCount)
			os.Exit(1)
		}
	}

	out, code := git.Exec(flag.Args()...)
	fmt.Print(out)
	os.Exit(code)
}

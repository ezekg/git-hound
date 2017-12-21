/*
Command git-hound is a Git plugin that helps prevent sensitive data from being committed
into a repository by sniffing potential commits against regular expressions
*/
package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"os"
	"sourcegraph.com/sourcegraph/go-diff/diff"
)

var (
	version     = "0.6.2"
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

	if ok := hound.New(); !ok {
		color.Red("No config file detected")
		os.Exit(1)
	}

	var (
		runnable bool
		out      string
	)

	switch flag.Arg(0) {
	case "commit":
		out, _ = git.Exec("diff", "-U0", "--staged")
		runnable = true
	case "sniff":
		stat, _ := os.Stdin.Stat()

		// Check if anything was piped to STDIN
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			stdin, _ := ioutil.ReadAll(os.Stdin)
			out = string(stdin)
		} else {
			commit := flag.Arg(1)
			if commit == "" {
				// NOTE: This let's us get a diff containing the entire repo codebase by
				//       utilizing a magic commit hash. In reality, it's not magical,
				//       it's simply the result of sha1("tree 0\0").
				commit = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
			}
			out, _ = git.Exec("diff", commit, "--staged")
		}
	default:
		fmt.Print("Usage:\n  git-hound commit [...]\n  git-hound sniff [commit]\n")
		os.Exit(0)
	}

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
		hunks := fileDiff.Hunks

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

	if runnable {
		out, code := git.Exec(flag.Args()...)
		fmt.Print(out)
		os.Exit(code)
	}
}

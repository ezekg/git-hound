package main

import (
	"errors"
	"fmt"
	"github.com/ezekg/git-hound/Godeps/_workspace/src/github.com/fatih/color"
	"github.com/ezekg/git-hound/Godeps/_workspace/src/gopkg.in/yaml.v2"
	"github.com/ezekg/git-hound/Godeps/_workspace/src/sourcegraph.com/sourcegraph/go-diff/diff"
	"io/ioutil"
	"path/filepath"
	"regexp"
)

// A Hound contains the local configuration filename and all regexp patterns
// used for sniffing git-diffs.
type Hound struct {
	Config string
	Fails  []string `yaml:"fail"`
	Warns  []string `yaml:"warn"`
	Skips  []string `yaml:"skip"`
}

// New initializes a new Hound instance by parsing regexp patterns from a
// local configuration file and returns a status bool.
func (h *Hound) New() bool {
	config, err := h.LoadConfig()
	if err != nil {
		return false
	}

	err = h.Parse(config)
	if err != nil {
		return false
	}

	return true
}

// LoadConfig reads a local configuration file of regexp patterns and returns
// the contents of the file.
func (h *Hound) LoadConfig() ([]byte, error) {
	filename, _ := filepath.Abs(h.Config)
	return ioutil.ReadFile(filename)
}

// Parse parses a configuration byte array and returns an error.
func (h *Hound) Parse(data []byte) error {
	return yaml.Unmarshal(data, h)
}

// Sniff matches the passed git-diff hunk against all regexp patterns that
// were parsed from the local configuration.
func (h *Hound) Sniff(fileName string, hunk *diff.Hunk, warnc chan string, failc chan error, donec chan bool) {
	defer func() { donec <- true }()

	r1, _ := regexp.Compile(`^\w+\/`)
	fileName = r1.ReplaceAllString(fileName, "")
	if _, ok := h.MatchPatterns(h.Skips, []byte(fileName)); ok {
		return
	}

	r2, _ := regexp.Compile(`(?m)^\+\s*(.+)$`)
	matches := r2.FindAllSubmatch(hunk.Body, -1)

	for _, match := range matches {
		line := match[1]

		if pattern, warned := h.MatchPatterns(h.Warns, line); warned {
			msg := color.YellowString(fmt.Sprintf(
				"warning: pattern `%s` match found for `%s` starting at line %d in %s\n",
				pattern, line, hunk.NewStartLine, fileName))
			warnc <- msg
		}

		if pattern, failed := h.MatchPatterns(h.Fails, line); failed {
			msg := color.RedString(fmt.Sprintf(
				"failure: pattern `%s` match found for `%s` starting at line %d in %s\n",
				pattern, line, hunk.NewStartLine, fileName))
			failc <- errors.New(msg)
		}
	}
}

// Match matches a byte array against a regexp pattern and returns a bool.
func (h *Hound) Match(pattern string, subject []byte) bool {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	return r.Match(subject)
}

// MatchPatterns matches a byte array against an array of regexp patterns and
// returns the matched pattern and a bool.
func (h *Hound) MatchPatterns(patterns []string, subject []byte) (string, bool) {
	for _, pattern := range patterns {
		if match := h.Match(pattern, subject); match {
			return pattern, true
		}
	}

	return "", false
}

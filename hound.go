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

type Hound struct {
	Config string
	Fails  []string `yaml:"fail"`
	Warns  []string `yaml:"warn"`
	Skips  []string `yaml:"skip"`
}

// New initializes a new Hound instance by parsing regexp patterns from a
// local configuration file to prepare for diff sniffing
func (h *Hound) New() (bool, error) {
	config, err := h.LoadConfig()
	if err != nil {
		return false, err
	}

	err = h.Parse(config)
	if err != nil {
		return false, err
	}

	return true, nil
}

// LoadConfig loads a local configuration file of regexp patterns
func (h *Hound) LoadConfig() ([]byte, error) {
	filename, _ := filepath.Abs(h.Config)
	return ioutil.ReadFile(filename)
}

// Parse parses a configuration byte array
func (h *Hound) Parse(data []byte) error {
	return yaml.Unmarshal(data, h)
}

// Sniff matches the passed git-diff hunk against all regexp patterns that
// were parsed from the local configuration
func (h *Hound) Sniff(fileName string, hunk *diff.Hunk) error {
	r1, _ := regexp.Compile(`^\w+\/`)
	fileName = r1.ReplaceAllString(fileName, "")
	if _, ok := h.MatchPatterns(h.Skips, []byte(fileName)); ok {
		return nil
	}

	r2, _ := regexp.Compile(`(?m)^\+\s*(.+)$`)
	matches := r2.FindAllSubmatch(hunk.Body, -1)

	for _, match := range matches {
		line := match[1]

		if pattern, ok := h.MatchPatterns(h.Warns, line); ok {
			message := color.YellowString(fmt.Sprintf("Warning: pattern `%s` match found for `%s` starting at line %d in %s\n",
				pattern, line, hunk.NewStartLine, fileName))
			fmt.Print(message)
		}

		if pattern, ok := h.MatchPatterns(h.Fails, line); ok {
			err := color.RedString(fmt.Sprintf("Failure: pattern `%s` match found for `%s` starting at line %d in %s\n",
				pattern, line, hunk.NewStartLine, fileName))
			return errors.New(err)
		}
	}

	return nil
}

// Match matches a byte array against a regexp pattern
func (h *Hound) Match(pattern string, subject []byte) (bool, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	if r.Match(subject) {
		return true, nil
	}

	return false, nil
}

// MatchPatterns matches a byte array against an array of regexp patterns
func (h *Hound) MatchPatterns(patterns []string, subject []byte) (string, bool) {
	for _, pattern := range patterns {
		if match, _ := h.Match(pattern, subject); match {
			return pattern, true
		}
	}

	return "", false
}

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

func (h *Hound) New() (error, bool) {
	config, err := h.LoadConfig()
	if err != nil {
		return err, false
	}

	err = h.Parse(config)
	if err != nil {
		return err, false
	}

	return nil, true
}

func (h *Hound) LoadConfig() ([]byte, error) {
	filename, _ := filepath.Abs(h.Config)
	return ioutil.ReadFile(filename)
}

func (h *Hound) Parse(data []byte) error {
	return yaml.Unmarshal(data, h)
}

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

func (h *Hound) MatchPatterns(patterns []string, subject []byte) (string, bool) {
	for _, pattern := range patterns {
		if match, _ := h.Match(pattern, subject); match {
			return pattern, true
		}
	}

	return "", false
}

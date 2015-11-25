package main

import (
	"github.com/ezekg/git-hound/Godeps/_workspace/src/gopkg.in/yaml.v2"
	"github.com/ezekg/git-hound/Godeps/_workspace/src/sourcegraph.com/sourcegraph/go-diff/diff"
	"io/ioutil"
	"path/filepath"
	"regexp"
)

var (
	regexes = make(map[string]*regexp.Regexp)
)

// A Hound contains the local configuration filename and all regexp patterns
// used for sniffing git-diffs.
type Hound struct {
	Fails  []string `yaml:"fail"`
	Warns  []string `yaml:"warn"`
	Skips  []string `yaml:"skip"`
	config string
}

// New initializes a new Hound instance by parsing regexp patterns from a
// local configuration file and returns a status bool.
func (h *Hound) New() bool {
	config, err := h.loadConfig()
	if err != nil {
		return false
	}

	err = h.parse(config)
	if err != nil {
		return false
	}

	return true
}

// Sniff matches the passed git-diff hunk against all regexp patterns that
// were parsed from the local configuration.
func (h *Hound) Sniff(fileName string, hunk *diff.Hunk, smells chan<- smell, done chan<- bool) {
	defer func() { done <- true }()

	rxFileName, _ := h.regexp(`^\w+\/`)
	fileName = rxFileName.ReplaceAllString(fileName, "")
	if _, ok := h.matchPatterns(h.Skips, []byte(fileName)); ok {
		return
	}

	rxModLines, _ := h.regexp(`(?m)^\+\s*(.+)$`)
	matches := rxModLines.FindAllSubmatch(hunk.Body, -1)

	for _, match := range matches {
		line := match[1]

		if pattern, warned := h.matchPatterns(h.Warns, line); warned {
			smells <- smell{
				pattern:  pattern,
				fileName: fileName,
				line:     line,
				lineNum:  hunk.NewStartLine,
				severity: 1,
			}
		}

		if pattern, failed := h.matchPatterns(h.Fails, line); failed {
			smells <- smell{
				pattern:  pattern,
				fileName: fileName,
				line:     line,
				lineNum:  hunk.NewStartLine,
				severity: 2,
			}
		}
	}
}

// loadConfig reads a local configuration file of regexp patterns and returns
// the contents of the file.
func (h *Hound) loadConfig() ([]byte, error) {
	filename, _ := filepath.Abs(h.config)
	return ioutil.ReadFile(filename)
}

// parse parses a configuration byte array and returns an error.
func (h *Hound) parse(config []byte) error {
	return yaml.Unmarshal(config, h)
}

// getRegexp looks for the specified pattern in Hound's regexes cache, and if
// it is available, it will fetch from it. If it is not available, it
// will compile the pattern and store it in the cache. Returns a Regexp
// and an error.
func (h *Hound) regexp(pattern string) (*regexp.Regexp, error) {
	if regexes[pattern] != nil {
		return regexes[pattern], nil
	}

	r, err := regexp.Compile(pattern)
	if err == nil {
		regexes[pattern] = r
	}

	return r, err
}

// match matches a byte array against a regexp pattern and returns a bool.
func (h *Hound) match(pattern string, subject []byte) bool {
	r, err := h.regexp(pattern)
	if err != nil {
		panic(err)
	}

	return r.Match(subject)
}

// matchPatterns matches a byte array against an array of regexp patterns and
// returns the matched pattern and a bool.
func (h *Hound) matchPatterns(patterns []string, subject []byte) (string, bool) {
	for _, pattern := range patterns {
		if match := h.match(pattern, subject); match {
			return pattern, true
		}
	}

	return "", false
}

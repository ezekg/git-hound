package main

import (
	"github.com/dlclark/regexp2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"sourcegraph.com/sourcegraph/go-diff/diff"
	"sync"
)

var (
	mutex   = &sync.Mutex{}
	regexes = make(map[string]*regexp2.Regexp)
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

	err = h.parseConfig(config)
	if err != nil {
		panic(err)
	}

	return true
}

// Sniff matches the passed git-diff hunk against all regexp patterns that
// were parsed from the local configuration.
func (h *Hound) Sniff(fileName string, hunk *diff.Hunk, smells chan<- smell, done chan<- bool) {
	defer func() { done <- true }()

	rxFileName, _ := h.regexp(`^\w+\/`)
	fileName, _ = rxFileName.Replace(fileName, "", -1, -1)
	if _, ok := h.matchPatterns(h.Skips, fileName); ok {
		return
	}

	var matches []*regexp2.Match
	rxModLines, _ := h.regexp(`(?m)^\+\s*(.+)$`)
	match, _ := rxModLines.FindStringMatch(string(hunk.Body))

	if match != nil {
		matches = append(matches, match)

		m, _ := rxModLines.FindNextMatch(match)
		for m != nil {
			matches = append(matches, m)
			m, _ = rxModLines.FindNextMatch(m)
		}
	}

	for _, match := range matches {
		groups := match.Groups()
		line := groups[1].Capture.String()

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

// parseConfig parses a configuration byte array and returns an error.
func (h *Hound) parseConfig(config []byte) error {
	return yaml.Unmarshal(config, h)
}

// regexp looks for the specified pattern in Hound's regexes cache, and if
// it is available, it will fetch from it. If it is not available, it
// will compile the pattern and store it in the cache. Returns a Regexp
// and an error.
func (h *Hound) regexp(pattern string) (*regexp2.Regexp, error) {
	// Make sure that we don't encounter a race condition where multiple
	// goroutines a
	mutex.Lock()
	defer mutex.Unlock()

	if regexes[pattern] != nil {
		return regexes[pattern], nil
	}

	r, err := regexp2.Compile(pattern, 0)
	if err == nil {
		regexes[pattern] = r
	}

	return r, err
}

// match matches a byte array against a regexp pattern and returns a bool.
func (h *Hound) match(pattern string, subject string) bool {
	r, err := h.regexp(pattern)
	if err != nil {
		panic(err)
	}

	res, err := r.MatchString(subject)
	if err != nil {
		return false
	}

	return res
}

// matchPatterns matches a byte array against an array of regexp patterns and
// returns the matched pattern and a bool.
func (h *Hound) matchPatterns(patterns []string, subject string) (string, bool) {
	for _, pattern := range patterns {
		if match := h.match(pattern, subject); match {
			return pattern, true
		}
	}

	return "", false
}

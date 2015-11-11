package main

import (
	"github.com/ezekg/git-hound/Godeps/_workspace/src/sourcegraph.com/sourcegraph/go-diff/diff"
	"io/ioutil"
	"os"
	"testing"
)

func TestDiffs(t *testing.T) {
	rescueStdout := os.Stdout
	var fileName string
	var hunk *diff.Hunk

	h := &Hound{}
	c := []byte(`
warn:
  - '(?i)user(name)?\W*[:=,]\W*.+$'
fail:
  - '(?i)pass(word)?\W*[:=,]\W*.+$'
`)

	if err := h.Parse(c); err != nil {
		t.Fatalf("Should parse - %s", err)
	}

	// Should fail
	fileName, hunk = getDiff(`diff --git a/test1.go b/test1.go
index 000000..000000 000000
--- a/test1.go
+++ b/test1.go
@@ -1,2 +3,4 @@
+Password: something-secret`)
	if err := h.Sniff(fileName, hunk); err == nil {
		t.Fatalf("Should fail - %s", err)
	}

	// Should pass but output warning
	fileName, hunk = getDiff(`diff --git a/test2.go b/test2.go
index 000000..000000 000000
--- a/test2.go
+++ b/test2.go
@@ -1,2 +3,4 @@
+Username: something-secret`)
	r, w, _ := os.Pipe()
	os.Stdout = w

	if err := h.Sniff(fileName, hunk); err != nil {
		w.Close()
		os.Stdout = rescueStdout
		t.Fatalf("Should pass - %s", err)
	}

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	if len(out) <= 0 {
		t.Fatalf("Should warn - %s", out)
	}

	// Should pass
	fileName, hunk = getDiff(`diff --git a/test3.go b/test3.go
index 000000..000000 000000
--- a/test3.go
+++ b/test3.go
@@ -1,2 +3,4 @@
+Something that is okay to commit`)
	if err := h.Sniff(fileName, hunk); err != nil {
		t.Fatalf("Should pass - %s", err)
	}

	// Should only pay attention to added lines and pass
	fileName, hunk = getDiff(`diff --git a/test4.go b/test4.go
index 000000..000000 000000
--- a/test4.go
+++ b/test4.go
@@ -1,2 +3,4 @@
-Password: something-secret`)
	if err := h.Sniff(fileName, hunk); err != nil {
		t.Fatalf("Should pass - %s", err)
	}
}

func getDiff(diffContents string) (string, *diff.Hunk) {
	fileDiff, _ := diff.ParseFileDiff([]byte(diffContents))
	fileName := fileDiff.NewName

	for _, hunk := range fileDiff.GetHunks() {
		return fileName, hunk
	}

	return fileName, nil
}

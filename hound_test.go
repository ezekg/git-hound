package main

import (
	"github.com/ezekg/git-hound/Godeps/_workspace/src/sourcegraph.com/sourcegraph/go-diff/diff"
	"testing"
)

func TestDiffs(t *testing.T) {
	hound := &Hound{}
	config := []byte(`
warn:
  - '(?i)user(name)?\W*[:=,]\W*.+$'
fail:
  - '(?i)pass(word)?\W*[:=,]\W*.+$'
skip:
  - '\.test$'
`)

	if err := hound.Parse(config); err != nil {
		t.Fatalf("Should parse - %s", err)
	}

	// Should fail
	{
		fileName, hunk := getDiff(`diff --git a/test1.go b/test1.go
index 000000..000000 000000
--- a/test1.go
+++ b/test1.go
@@ -1,2 +3,4 @@
+Password: something-secret`)
		warnc := make(chan string)
		failc := make(chan error)
		donec := make(chan bool)

		go hound.Sniff(fileName, hunk, warnc, failc, donec)

		select {
		case <-failc:
			break
		case <-warnc:
			t.Fatalf("Should not warn")
		case <-donec:
			t.Fatalf("Should receive message")
		}
	}

	// Should pass but output warning
	{
		fileName, hunk := getDiff(`diff --git a/test2.go b/test2.go
index 000000..000000 000000
--- a/test2.go
+++ b/test2.go
@@ -1,2 +3,4 @@
+Username: something-secret`)
		warnc := make(chan string)
		failc := make(chan error)
		donec := make(chan bool)

		go hound.Sniff(fileName, hunk, warnc, failc, donec)

		select {
		case <-failc:
			t.Fatalf("Should not fail")
		case <-warnc:
			break
		case <-donec:
			t.Fatalf("Should receive message")
		}
	}

	// Should pass
	{
		fileName, hunk := getDiff(`diff --git a/test3.go b/test3.go
index 000000..000000 000000
--- a/test3.go
+++ b/test3.go
@@ -1,2 +3,4 @@
+Something that is okay to commit`)
		warnc := make(chan string)
		failc := make(chan error)
		donec := make(chan bool)

		go hound.Sniff(fileName, hunk, warnc, failc, donec)

		select {
		case <-failc:
			t.Fatal("Should not fail")
		case <-warnc:
			t.Fatal("Should not warn")
		case <-donec:
			break
		}
	}

	// Should only pay attention to added lines and pass
	{
		fileName, hunk := getDiff(`diff --git a/test4.go b/test4.go
index 000000..000000 000000
--- a/test4.go
+++ b/test4.go
@@ -1,2 +3,4 @@
-Password: something-secret`)
		warnc := make(chan string)
		failc := make(chan error)
		donec := make(chan bool)

		go hound.Sniff(fileName, hunk, warnc, failc, donec)

		select {
		case <-failc:
			t.Fatal("Should not fail")
		case <-warnc:
			t.Fatal("Should not warn")
		case <-donec:
			break
		}
	}

	// Should skip even with failures
	{
		fileName, hunk := getDiff(`diff --git a/test4.test b/test4.test
index 000000..000000 000000
--- a/test4.test
+++ b/test4.test
@@ -1,2 +3,4 @@
+Password: something-secret`)
		warnc := make(chan string)
		failc := make(chan error)
		donec := make(chan bool)

		go hound.Sniff(fileName, hunk, warnc, failc, donec)

		select {
		case <-failc:
			t.Fatal("Should not fail")
		case <-warnc:
			t.Fatal("Should not warn")
		case <-donec:
			break
		}
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

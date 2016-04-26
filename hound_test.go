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

	if err := hound.parseConfig(config); err != nil {
		t.Fatalf("Should parse config - %s", err)
	}

	// Should fail with a warning
	{
		fileName, hunk := getDiff(t, `diff --git a/test1.go b/test1.go
index 000000..000000 000000
--- a/test1.go
+++ b/test1.go
@@ -1,2 +3,4 @@
+Password: something-secret
+Username: something-secret`)
		warnCount := 0
		failCount := 0

		smells := make(chan smell)
		done := make(chan bool)

		go hound.Sniff(fileName, hunk, smells, done)

		for c := 0; c < 1; {
			select {
			case s := <-smells:
				switch s.severity {
				case 1:
					warnCount++
				case 2:
					failCount++
				default:
					t.Fatalf("Should receive smell")
				}
			case <-done:
				c++
			}
		}

		if warnCount == 0 {
			t.Fatalf("Should warn")
		}

		if failCount == 0 {
			t.Fatalf("Should fail")
		}
	}

	// Should fail
	{
		fileName, hunk := getDiff(t, `diff --git a/test2.go b/test2.go
index 000000..000000 000000
--- a/test2.go
+++ b/test2.go
@@ -1,2 +3,4 @@
+Password: something-secret`)
		smells := make(chan smell)
		done := make(chan bool)

		go hound.Sniff(fileName, hunk, smells, done)

		select {
		case s := <-smells:
			switch s.severity {
			case 1:
				t.Fatalf("Should not warn")
			case 2:
				t.Logf("Did fail")
			default:
				t.Fatalf("Should fail")
			}
		case <-done:
			t.Fatalf("Should receive smell")
		}
	}

	// Should pass but output warning
	{
		fileName, hunk := getDiff(t, `diff --git a/test3.go b/test3.go
index 000000..000000 000000
--- a/test3.go
+++ b/test3.go
@@ -1,2 +3,4 @@
+Username: something-secret`)
		smells := make(chan smell)
		done := make(chan bool)

		go hound.Sniff(fileName, hunk, smells, done)

		select {
		case s := <-smells:
			switch s.severity {
			case 1:
				t.Logf("Did warn")
			case 2:
				t.Fatalf("Should not fail")
			default:
				t.Fatalf("Should warn")
			}
		case <-done:
			t.Fatalf("Should receive smell")
		}
	}

	// Should pass
	{
		fileName, hunk := getDiff(t, `diff --git a/test4.go b/test4.go
index 000000..000000 000000
--- a/test4.go
+++ b/test4.go
@@ -1,2 +3,4 @@
+Something that is okay to commit`)
		smells := make(chan smell)
		done := make(chan bool)

		go hound.Sniff(fileName, hunk, smells, done)

		select {
		case s := <-smells:
			switch s.severity {
			case 1:
				t.Fatalf("Should not warn")
			case 2:
				t.Fatalf("Should not fail")
			default:
				t.Fatalf("Should not receive smell")
			}
		case <-done:
			t.Logf("Did pass")
		}
	}

	// Should only pay attention to added lines and pass
	{
		fileName, hunk := getDiff(t, `diff --git a/test5.go b/test5.go
index 000000..000000 000000
--- a/test5.go
+++ b/test5.go
@@ -1,2 +3,4 @@
-Password: something-secret`)
		smells := make(chan smell)
		done := make(chan bool)

		go hound.Sniff(fileName, hunk, smells, done)

		select {
		case s := <-smells:
			switch s.severity {
			case 1:
				t.Fatalf("Should not warn")
			case 2:
				t.Fatalf("Should not fail")
			default:
				t.Fatalf("Should not receive smell")
			}
		case <-done:
			t.Logf("Did pass")
		}
	}

	// Should skip even with failures
	{
		fileName, hunk := getDiff(t, `diff --git a/test6.test b/test6.test
index 000000..000000 000000
--- a/test6.test
+++ b/test6.test
@@ -1,2 +3,4 @@
+Password: something-secret`)
		smells := make(chan smell)
		done := make(chan bool)

		go hound.Sniff(fileName, hunk, smells, done)

		select {
		case s := <-smells:
			switch s.severity {
			case 1:
				t.Fatalf("Should not warn")
			case 2:
				t.Fatalf("Should not fail")
			default:
				t.Fatalf("Should not receive message")
			}
		case <-done:
			t.Logf("Did pass")
		}
	}
}

func getDiff(t *testing.T, diffContents string) (string, *diff.Hunk) {
	fileDiff, err := diff.ParseFileDiff([]byte(diffContents))
	if err != nil {
		t.Fatalf("Should parse fileDiff")
	}

	fileName := fileDiff.NewName

	for _, hunk := range fileDiff.GetHunks() {
		return fileName, hunk
	}

	return fileName, nil
}

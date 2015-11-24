package main

import (
	"fmt"
)

// A smell contains information about a code smell, such as line, filename,
// line number and severity.
type smell struct {
	pattern  string
	fileName string
	line     []byte
	lineNum  int32
	severity int
}

// String returns a string representation of the smell instance.
func (s *smell) String() string {
	return fmt.Sprintf("pattern `%s` match found for `%s` starting at line %d in %s",
		s.pattern, s.line, s.lineNum, s.fileName)
}

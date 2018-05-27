// Copyright 2017 The ksonnet authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package strings

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/purell"
	"github.com/pkg/errors"
	godiff "github.com/shazow/go-diff"
)

// IsASCIIIdentifier takes a string and returns true if the string does not
// contain any special characters.
func IsASCIIIdentifier(s string) bool {
	f := func(r rune) bool {
		return r < 'A' || r > 'z'
	}
	if strings.IndexFunc(s, f) != -1 {
		return false
	}
	return true
}

// QuoteNonASCII puts quotes around an identifier that contains non-ASCII
// characters.
func QuoteNonASCII(s string) string {
	if !IsASCIIIdentifier(s) {
		return fmt.Sprintf(`"%s"`, s)
	}
	return s
}

// NormalizeURL uses purell's "usually safe normalization" algorithm to
// normalize URLs. This includes removing dot segments, removing trailing
// slashes, removing unnecessary escapes, removing default ports, and setting
// the URL to lowercase.
func NormalizeURL(s string) (string, error) {
	return purell.NormalizeURLString(s, purell.FlagsUsuallySafeGreedy)
}

// InSlice returns true if the string is in the slice.
func InSlice(s string, sl []string) bool {
	for i := range sl {
		if sl[i] == s {
			return true
		}
	}

	return false
}

// Ptr returns a pointer to a string.
func Ptr(s string) *string {
	return &s
}

func Compare(expected, got string) (bool, error) {
	var err error
	expected, err = scanString(strings.NewReader(expected))
	if err != nil {
		return false, errors.Wrap(err, "scan string")
	}

	got, err = scanString(strings.NewReader(got))
	if err != nil {
		return false, errors.Wrap(err, "scan string")
	}

	rExpected := strings.NewReader(expected)
	rGot := strings.NewReader(got)

	var buf bytes.Buffer
	err = godiff.DefaultDiffer().Diff(&buf, rExpected, rGot)
	if err != nil {
		return false, err
	}

	if buf.String() != "" {
		return false, errors.New(buf.String())
	}

	return true, nil
}

func scanString(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)

	scanner.Split(scanCRLF)

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", errors.Wrap(err, "scanner error")
	}

	return strings.Join(lines, "\n"), nil
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

// scanCRLF scans for Windows style line endings.
func scanCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte{'\r', '\n'}); i >= 0 {
		// We have a full newline-terminated line.
		return i + 2, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

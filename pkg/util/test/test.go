// Copyright 2018 The ksonnet authors
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

package test

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app/mocks"
	godiff "github.com/shazow/go-diff"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ReadTestData reads a file from `testdata` and returns it as a string.
func ReadTestData(t *testing.T, name string) string {
	path := filepath.Join("testdata", name)
	data, err := ioutil.ReadFile(path)
	require.NoError(t, err)

	return string(data)
}

// StageFile stages a file on on the provided filesystem from
// testdata.
func StageFile(t *testing.T, fs afero.Fs, src, dest string) {
	in := filepath.Join("testdata", src)

	b, err := ioutil.ReadFile(in)
	require.NoError(t, err)

	dir := filepath.Dir(dest)
	err = fs.MkdirAll(dir, 0755)
	require.NoError(t, err)

	err = afero.WriteFile(fs, dest, b, 0644)
	require.NoError(t, err)
}

// StageDir stages a directory on the provided filesystem from
// testdata.
func StageDir(t *testing.T, fs afero.Fs, src, dest string) {
	root, err := filepath.Abs(filepath.Join("testdata", src))
	require.NoError(t, err)

	err = filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		cur := filepath.Join(dest, strings.TrimPrefix(path, root))
		if fi.IsDir() {
			return fs.Mkdir(cur, 0755)
		}

		copyFile(fs, path, cur)
		return nil
	})

	require.NoError(t, err)
}

func copyFile(fs afero.Fs, src, dest string) error {
	from, err := os.Open(src)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := fs.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	return err
}

// WithApp runs an enclosure with a mocked app and fs.
func WithApp(t *testing.T, root string, fn func(*mocks.App, afero.Fs)) {
	fs := afero.NewMemMapFs()

	WithAppFs(t, root, fs, fn)
}

// WithAppFs runs an enclosure with a mocked app and fs. Allow supplying the fs.
func WithAppFs(t *testing.T, root string, fs afero.Fs, fn func(*mocks.App, afero.Fs)) {
	a := &mocks.App{}
	a.On("Fs").Return(fs)
	a.On("Root").Return(root)
	a.On("LibPath", mock.AnythingOfType("string")).Return(filepath.Join(root, "lib", "v1.8.7"), nil)

	fn(a, fs)
}

// AssertOutput asserts the file in filename is equal to actual.
func AssertOutput(t *testing.T, filename, actual string) {
	path := filepath.Join("testdata", filepath.FromSlash(filename))

	b, err := ioutil.ReadFile(path)
	require.NoError(t, err, "read expected")

	CompareStrings(t, strings.TrimSpace(string(b)), strings.TrimSpace(actual))
}

func CompareStrings(t *testing.T, expected, got string) {
	expected = scanString(t, strings.NewReader(expected))
	got = scanString(t, strings.NewReader(got))

	rExpected := strings.NewReader(expected)
	rGot := strings.NewReader(got)

	var buf bytes.Buffer
	err := godiff.DefaultDiffer().Diff(&buf, rExpected, rGot)
	require.NoError(t, err)
	require.Empty(t, buf.String())
}

func scanString(t *testing.T, r io.Reader) string {
	scanner := bufio.NewScanner(r)

	scanner.Split(scanCRLF)

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	require.NoError(t, scanner.Err(), "scanner error")

	return strings.Join(lines, "\n")
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

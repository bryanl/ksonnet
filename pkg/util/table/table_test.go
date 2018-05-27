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

package table

import (
	"bytes"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/util/test"
)

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	table := New(&buf)

	table.SetHeader([]string{"name", "version", "Namespace", "SERVER"})
	table.Append([]string{"default", "v1.7.0", "default", "http://default"})
	table.AppendBulk([][]string{
		{"dev", "v1.8.0", "dev", "http://dev"},
		{"east/prod", "v1.8.0", "east/prod", "http://east-prod"},
	})

	table.Render()

	test.AssertOutput(t, "table.txt", buf.String())
}

func TestTable_no_header(t *testing.T) {
	var buf bytes.Buffer
	table := New(&buf)

	table.Append([]string{"default", "v1.7.0", "default", "http://default"})
	table.AppendBulk([][]string{
		{"dev", "v1.8.0", "dev", "http://dev"},
		{"east/prod", "v1.8.0", "east/prod", "http://east-prod"},
	})

	table.Render()

	test.AssertOutput(t, "table_no_header.txt", buf.String())
}

func TestTable_trim_space(t *testing.T) {
	var buf bytes.Buffer
	table := New(&buf)

	table.SetHeader([]string{"name", "version", "Namespace", "SERVER"})
	table.Append([]string{"default", "v1.7.0", "default", "http://default"})
	table.AppendBulk([][]string{
		{"dev", "v1.8.0", "", ""},
		{"east/prod", "v1.8.0", "east/prod", "http://east-prod"},
	})

	table.Render()

	test.AssertOutput(t, "table_trim_space.txt", buf.String())
}

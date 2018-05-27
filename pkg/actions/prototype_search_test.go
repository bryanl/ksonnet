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

package actions

import (
	"bytes"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/stretchr/testify/require"
)

func TestPrototypeSearch(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		libaries := app.LibraryRefSpecs{}

		appMock.On("Libraries").Return(libaries, nil)

		in := map[string]interface{}{
			OptionApp:   appMock,
			OptionQuery: "search",
		}

		a, err := NewPrototypeSearch(in)
		require.NoError(t, err)

		var buf bytes.Buffer
		a.out = &buf

		a.protoSearchFn = func(string, prototype.Prototypes) (prototype.Prototypes, error) {
			snippet := prototype.SnippetSchema{ShortDescription: "description"}

			return prototype.Prototypes{
				{Name: "result1", Template: snippet},
				{Name: "result2", Template: snippet},
			}, nil
		}

		err = a.Run()
		require.NoError(t, err)

		test.AssertOutput(t, "prototype/search/output.txt", buf.String())
	})
}

func TestProtoptypeSearch_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := NewPrototypeSearch(in)
	require.Error(t, err)
}

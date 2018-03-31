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
	"strconv"
	"strings"

	"github.com/ksonnet/ksonnet/component"
	"github.com/ksonnet/ksonnet/env"
	"github.com/ksonnet/ksonnet/metadata/app"
	mp "github.com/ksonnet/ksonnet/metadata/params"
	"github.com/ksonnet/ksonnet/pkg/params"
	"github.com/pkg/errors"
)

// RunParamSet runs `param set`
func RunParamSet(m map[string]interface{}) error {
	ps, err := NewParamSet(m)
	if err != nil {
		return err
	}

	return ps.Run()
}

// ParamSet sets a parameter for a component.
type ParamSet struct {
	app      app.App
	name     string
	rawPath  string
	rawValue string
	index    int
	global   bool
	envName  string

	// TODO: remove once ksonnet has more robust env param handling.
	setEnv func(ksApp app.App, envName, name, pName, value string) error

	cm component.Manager
}

// NewParamSet creates an instance of ParamSet.
func NewParamSet(m map[string]interface{}) (*ParamSet, error) {
	ol := newOptionLoader(m)

	ps := &ParamSet{
		app:      ol.loadApp(),
		name:     ol.loadString(OptionName),
		rawPath:  ol.loadString(OptionPath),
		rawValue: ol.loadString(OptionValue),
		global:   ol.loadOptionalBool(OptionGlobal),
		envName:  ol.loadOptionalString(OptionEnvName),
		index:    ol.loadOptionalInt(OptionIndex),

		cm:     component.DefaultManager,
		setEnv: setEnv,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	if ps.envName != "" && ps.global {
		return nil, errors.New("unable to set global param for environments")
	}

	return ps, nil
}

// Run runs the action.
func (ps *ParamSet) Run() error {
	value, err := params.DecodeValue(ps.rawValue)
	if err != nil {
		return errors.Wrap(err, "value is invalid")
	}

	evaluatedValue := ps.rawValue
	if _, ok := value.(string); ok {
		evaluatedValue = strconv.Quote(ps.rawValue)
	}

	if ps.envName != "" {
		return ps.setEnv(ps.app, ps.envName, ps.name, ps.rawPath, evaluatedValue)
	}

	path := strings.Split(ps.rawPath, ".")

	if ps.global {
		return ps.setGlobal(path, value)
	}

	return ps.setLocal(path, value)
}

func (ps *ParamSet) setGlobal(path []string, value interface{}) error {
	ns, err := ps.cm.Namespace(ps.app, ps.name)
	if err != nil {
		return errors.Wrap(err, "retrieve namespace")
	}

	if err := ns.SetParam(path, value); err != nil {
		return errors.Wrap(err, "set global param")
	}

	return nil
}

func (ps *ParamSet) setLocal(path []string, value interface{}) error {
	_, c, err := ps.cm.ResolvePath(ps.app, ps.name)
	if err != nil {
		return errors.Wrap(err, "could not find component")
	}

	options := component.ParamOptions{
		Index: ps.index,
	}
	if err := c.SetParam(path, value, options); err != nil {
		return errors.Wrap(err, "set param")
	}

	return nil
}

func setEnv(ksApp app.App, envName, name, pName, value string) error {
	spc := env.SetParamsConfig{
		App: ksApp,
	}

	p := mp.Params{
		pName: value,
	}

	return env.SetParams(envName, name, p, spc)
}

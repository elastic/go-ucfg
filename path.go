// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package ucfg

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type cfgPath struct {
	fields []field
	sep    string
}

type field interface {
	String() string
	SetValue(opt *options, elem value, v value) Error
	GetValue(opt *options, elem value) (value, Error)
	Remove(opt *options, elem value) (bool, Error)
}

type namedField struct {
	name string
}

type idxField struct {
	i int
}

func parsePathIdx(in string, idx int, opts *options) cfgPath {
	if in == "" {
		return cfgPath{
			sep:    opts.pathSep,
			fields: []field{idxField{idx}},
		}
	}

	p := parsePathWithOpts(in, opts)
	if idx >= 0 {
		p.fields = append(p.fields, idxField{idx})
	}

	return p
}

const (
	escapePathRegExp = `^\[.*\]$`
)

var escapePathReg = regexp.MustCompile(escapePathRegExp)

func parsePath(in, sep string, maxIdx int64, enableNumKeys, allowEscapePath bool) cfgPath {
	if sep == "" || (allowEscapePath && escapePathReg.MatchString(in)) {
		return cfgPath{
			sep:    sep,
			fields: []field{parseField(in, maxIdx, enableNumKeys)},
		}
	}

	elems := strings.Split(in, sep)
	fields := make([]field, 0, len(elems))
	// If property is the name with separators, for example "inputs.0.i"
	// fall back to original implementation
	if len(elems) > 1 {
		enableNumKeys = false
	}
	for _, elem := range elems {
		fields = append(fields, parseField(elem, maxIdx, enableNumKeys))
	}
	return cfgPath{fields: fields, sep: sep}
}

func parsePathWithOpts(in string, opts *options) cfgPath {
	return parsePath(in, opts.pathSep, opts.maxIdx, opts.enableNumKeys, opts.escapePath)
}

func parseField(in string, maxIdx int64, enableNumKeys bool) field {
	// If numeric keys are not enabled, fallback to the original implementation
	if !enableNumKeys {
		idx, err := strconv.ParseInt(in, 0, 64)
		// Limit index value to the configurable max.
		// If the idx > opts.maxIdx treat it as a regular named field.
		// This preserves the current behavour for small index fields values (<= opts.maxIdx)
		// and prevents large memory allocations or OOM if the string is large numeric value
		if err == nil && idx <= int64(maxIdx) {
			return idxField{int(idx)}
		}
	}
	return namedField{in}
}

func (p cfgPath) String() string {
	if len(p.fields) == 0 {
		return ""
	}

	if len(p.fields) == 1 {
		return p.fields[0].String()
	}

	s := make([]string, 0, len(p.fields))
	for _, f := range p.fields {
		s = append(s, f.String())
	}

	sep := p.sep
	if sep == "" {
		sep = "."
	}
	return strings.Join(s, sep)
}

func (n namedField) String() string {
	return n.name
}

func (i idxField) String() string {
	return fmt.Sprintf("%d", i.i)
}

func (p cfgPath) Has(cfg *Config, opt *options) (bool, Error) {
	fields := p.fields

	cur := value(cfgSub{cfg})
	for ; len(fields) > 0; fields = fields[1:] {
		field := fields[0]
		next, err := field.GetValue(opt, cur)
		if err != nil {
			// has checks if a value is missing -> ErrMissing is no error but a valid
			// outcome
			if err.Reason() == ErrMissing {
				err = nil
			}
			return false, err
		}

		if next == nil {
			return false, nil
		}

		cur = next
	}

	return true, nil
}

func (p cfgPath) GetValue(cfg *Config, opt *options) (value, Error) {
	fields := p.fields

	cur := value(cfgSub{cfg})
	for ; len(fields) > 1; fields = fields[1:] {
		field := fields[0]
		next, err := field.GetValue(opt, cur)
		if err != nil {
			return nil, err
		}

		if next == nil {
			return nil, raiseMissing(cfg, field.String())
		}

		cur = next
	}

	field := fields[0]
	v, err := field.GetValue(opt, cur)
	if err != nil {
		return nil, raiseMissing(cfg, field.String())
	}
	return v, nil
}

func (n namedField) GetValue(opts *options, elem value) (value, Error) {
	cfg, err := elem.toConfig(opts)
	if err != nil {
		return nil, raiseExpectedObject(opts, elem)
	}

	v, _ := cfg.fields.get(n.name)
	return v, nil
}

func (i idxField) GetValue(opts *options, elem value) (value, Error) {
	cfg, err := elem.toConfig(opts)
	if err != nil {
		if i.i == 0 {
			return elem, nil
		}

		return nil, raiseExpectedObject(opts, elem)
	}

	arr := cfg.fields.array()
	if i.i >= len(arr) {
		return nil, raiseMissing(cfg, i.String())
	}
	return arr[i.i], nil
}

func (p cfgPath) SetValue(cfg *Config, opt *options, val value) Error {
	fields := p.fields
	node := value(cfgSub{cfg})

	// 1. iterate until intermediate node not having some required child node
	for ; len(fields) > 1; fields = fields[1:] {
		field := fields[0]
		v, err := field.GetValue(opt, node)
		if err != nil {
			if err.Reason() == ErrMissing {
				break
			}
			return err
		}

		if isNil(v) {
			break
		}
		node = v
	}

	// 2. build intermediate nodes from bottom up

	for ; len(fields) > 1; fields = fields[:len(fields)-1] {
		field := fields[len(fields)-1]

		next := New()
		next.metadata = val.meta()
		v := cfgSub{next}
		if err := field.SetValue(opt, v, val); err != nil {
			return err
		}
		val = v
	}

	// 3. insert new sub-tree into config
	return fields[0].SetValue(opt, node, val)
}

func (n namedField) SetValue(opts *options, elem value, v value) Error {
	sub, ok := elem.(cfgSub)
	if !ok {
		return raiseExpectedObject(opts, elem)
	}

	sub.c.fields.set(n.name, v)
	v.SetContext(context{parent: elem, field: n.name})
	return nil
}

func (i idxField) SetValue(opts *options, elem value, v value) Error {
	sub, ok := elem.(cfgSub)
	if !ok {
		return raiseExpectedObject(opts, elem)
	}

	sub.c.fields.setAt(i.i, elem, v)
	v.SetContext(context{parent: elem, field: i.String()})
	return nil
}

func (p cfgPath) Remove(cfg *Config, opt *options) (bool, error) {
	fields := p.fields

	// Loop over intermediate objects. Returns an error if any intermediate is
	// actually no object.
	cur := value(cfgSub{cfg})
	for ; len(fields) > 1; fields = fields[1:] {
		field := fields[0]
		next, err := field.GetValue(opt, cur)
		if err != nil {
			// Ignore ErrMissing when walking down a config tree. If intermediary is
			// missing we can't remove our setting.
			if err.Reason() == ErrMissing {
				err = nil
			}

			return false, err
		}

		if next == nil {
			return false, err
		}

		cur = next
	}

	// resolve config object in case we deal with references
	tmp, err := cur.toConfig(opt)
	if err != nil {
		return false, err
	}
	cur = cfgSub{tmp}

	field := fields[0]
	return field.Remove(opt, cur)
}

func (n namedField) Remove(opts *options, elem value) (bool, Error) {
	sub, ok := elem.(cfgSub)
	if !ok {
		return false, raiseExpectedObject(opts, elem)
	}

	removed := sub.c.fields.del(n.name)
	return removed, nil
}

func (i idxField) Remove(opts *options, elem value) (bool, Error) {
	sub, ok := elem.(cfgSub)
	if !ok {
		return false, raiseExpectedObject(opts, elem)
	}

	removed := sub.c.fields.delAt(i.i)
	return removed, nil
}

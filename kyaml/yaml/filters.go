// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var Filters = map[string]func() Filter{
	"AnnotationClearer": func() Filter { return &AnnotationClearer{} },
	"AnnotationGetter":  func() Filter { return &AnnotationGetter{} },
	"AnnotationSetter":  func() Filter { return &AnnotationSetter{} },
	"ElementAppender":   func() Filter { return &ElementAppender{} },
	"ElementMatcher":    func() Filter { return &ElementMatcher{} },
	"FieldClearer":      func() Filter { return &FieldClearer{} },
	"FilterMatcher":     func() Filter { return &FilterMatcher{} },
	"FieldMatcher":      func() Filter { return &FieldMatcher{} },
	"FieldSetter":       func() Filter { return &FieldSetter{} },
	"PathGetter":        func() Filter { return &PathGetter{} },
	"PathMatcher":       func() Filter { return &PathMatcher{} },
	"Parser":            func() Filter { return &Parser{} },
	"PrefixSetter":      func() Filter { return &PrefixSetter{} },
	"ValueReplacer":     func() Filter { return &ValueReplacer{} },
	"SuffixSetter":      func() Filter { return &SuffixSetter{} },
	"TeePiper":          func() Filter { return &TeePiper{} },
}

// YFilter wraps the GrepFilter interface so it can be unmarshalled into a struct.
type YFilter struct {
	Filter
}

func (y YFilter) MarshalYAML() (interface{}, error) {
	return y.Filter, nil
}

func (y *YFilter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	meta := &ResourceMeta{}
	if err := unmarshal(meta); err != nil {
		return err
	}
	if filter, found := Filters[meta.Kind]; !found {
		var knownFilters []string
		for k := range Filters {
			knownFilters = append(knownFilters, k)
		}
		sort.Strings(knownFilters)
		return fmt.Errorf("unsupported GrepFilter Kind %s:  may be one of: [%s]",
			meta.Kind, strings.Join(knownFilters, ","))
	} else {
		y.Filter = filter()
	}
	if err := unmarshal(y.Filter); err != nil {
		return err
	}
	return nil
}

type YFilters []YFilter

func (y YFilters) Filters() []Filter {
	var f []Filter
	for i := range y {
		f = append(f, y[i].Filter)
	}
	return f
}

type FilterMatcher struct {
	Kind string `yaml:"kind"`

	// Filters are the set of Filters run by TeePiper.
	Filters YFilters `yaml:"pipeline,omitempty"`
}

func (t FilterMatcher) Filter(rn *RNode) (*RNode, error) {
	v, err := rn.Pipe(t.Filters.Filters()...)
	if v == nil || err != nil {
		return nil, err
	}
	// return the original input if the pipeline resolves to true
	return rn, err
}

type ValueReplacer struct {
	Kind string `yaml:"kind"`

	StringMatch string `yaml:"stringMatch"`
	RegexMatch  string `yaml:"regexMatch"`
	Replace     string `yaml:"replace"`
	Count       int    `yaml:"count"`
}

func (s ValueReplacer) Filter(object *RNode) (*RNode, error) {
	if s.Count == 0 {
		s.Count = -1
	}
	if s.StringMatch != "" {
		object.value.Value = strings.Replace(object.value.Value, s.StringMatch, s.Replace, s.Count)
	} else if s.RegexMatch != "" {
		r, err := regexp.Compile(s.RegexMatch)
		if err != nil {
			return nil, fmt.Errorf("ValueReplacer RegexMatch does not compile: %v", err)
		}
		object.value.Value = r.ReplaceAllString(object.value.Value, s.Replace)
	} else {
		return nil, fmt.Errorf("ValueReplacer missing StringMatch and RegexMatch")
	}
	return object, nil
}

type PrefixSetter struct {
	Kind string `yaml:"kind"`

	Value string `yaml:"value"`
}

func (s PrefixSetter) Filter(object *RNode) (*RNode, error) {
	if !strings.HasPrefix(object.value.Value, s.Value) {
		object.value.Value = s.Value + object.value.Value
	}
	return object, nil
}

type SuffixSetter struct {
	Kind string `yaml:"kind"`

	Value string `yaml:"value"`
}

func (s SuffixSetter) Filter(object *RNode) (*RNode, error) {
	if !strings.HasSuffix(object.value.Value, s.Value) {
		object.value.Value = object.value.Value + s.Value
	}
	return object, nil
}

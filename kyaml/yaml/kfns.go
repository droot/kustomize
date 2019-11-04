// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import "gopkg.in/yaml.v3"

// AnnotationClearer removes an annotation at metadata.annotations.
// Returns nil if the annotation or field does not exist.
type AnnotationClearer struct {
	Kind string `yaml:"kind,omitempty"`
	Key  string `yaml:"key,omitempty"`
}

func (c AnnotationClearer) Filter(rn *RNode) (*RNode, error) {
	return rn.Pipe(
		PathGetter{Path: []string{"metadata", "annotations"}},
		FieldClearer{Name: c.Key})
}

func ClearAnnotation(key string) AnnotationClearer {
	return AnnotationClearer{Key: key}
}

// AnnotationSetter sets an annotation at metadata.annotations.
// Creates metadata.annotations if does not exist.
type AnnotationSetter struct {
	Kind  string `yaml:"kind,omitempty"`
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

func (s AnnotationSetter) Filter(rn *RNode) (*RNode, error) {
	return rn.Pipe(
		PathGetter{Path: []string{"metadata", "annotations"}, Create: yaml.MappingNode},
		FieldSetter{Name: s.Key, Value: NewScalarRNode(s.Value)})
}

func SetAnnotation(key, value string) AnnotationSetter {
	return AnnotationSetter{Key: key, Value: value}
}

// AnnotationGetter gets an annotation at metadata.annotations.
// Returns nil if metadata.annotations does not exist.
type AnnotationGetter struct {
	Kind  string `yaml:"kind,omitempty"`
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

// AnnotationGetter returns the annotation value.
// Returns "", nil if the annotation does not exist.
func (g AnnotationGetter) Filter(rn *RNode) (*RNode, error) {
	v, err := rn.Pipe(PathGetter{Path: []string{"metadata", "annotations", g.Key}})
	if v == nil || err != nil {
		return v, err
	}
	if g.Value == "" || v.value.Value == g.Value {
		return v, err
	}
	return nil, err
}

func GetAnnotation(key string) AnnotationGetter {
	return AnnotationGetter{Key: key}
}

// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	ResourceListKind       = "ResourceList"
	ResourceListApiVersion = "kyaml.kustomize.dev/v1alpha1"
)

// ByteReadWriter reads from an input and writes to an output
type ByteReadWriter struct {
	// Reader is where ResourceNodes are decoded from.
	Reader io.Reader

	// Writer is where ResourceNodes are encoded.
	Writer io.Writer

	// OmitReaderAnnotations will configures Read to skip setting the kyaml.kustomize.dev/kio/index
	// annotation on Resources as they are Read.
	OmitReaderAnnotations bool

	// KeepReaderAnnotations if set will keep the Reader specific annotations when writing
	// the Resources, otherwise they will be cleared.
	KeepReaderAnnotations bool

	// Style is a style that is set on the Resource Node Document.
	Style yaml.Style

	FunctionConfig *yaml.RNode

	WrappingApiVersion string
	WrappingKind       string
}

func (rw *ByteReadWriter) Read() ([]*yaml.RNode, error) {
	b := &ByteReader{
		Reader:                rw.Reader,
		OmitReaderAnnotations: rw.OmitReaderAnnotations,
	}
	val, err := b.Read()
	rw.FunctionConfig = b.FunctionConfig
	rw.WrappingApiVersion = b.WrappingApiVersion
	rw.WrappingKind = b.WrappingKind
	return val, err
}

func (rw *ByteReadWriter) Write(nodes []*yaml.RNode) error {
	return ByteWriter{
		Writer:                rw.Writer,
		KeepReaderAnnotations: rw.KeepReaderAnnotations,
		Style:                 rw.Style,
		FunctionConfig:        rw.FunctionConfig,
		WrappingApiVersion:    rw.WrappingApiVersion,
		WrappingKind:          rw.WrappingKind,
	}.Write(nodes)
}

// ByteReader decodes ResourceNodes from bytes.
// By default, Read will set the kyaml.kustomize.dev/kio/index annotation on each RNode as it
// is read so they can be written back in the same order.
type ByteReader struct {
	// Reader is where ResourceNodes are decoded from.
	Reader io.Reader

	// OmitReaderAnnotations will configures Read to skip setting the kyaml.kustomize.dev/kio/index
	// annotation on Resources as they are Read.
	OmitReaderAnnotations bool

	// SetAnnotations is a map of caller specified annotations to set on resources as they are read
	// These are independent of the annotations controlled by OmitReaderAnnotations
	SetAnnotations map[string]string

	FunctionConfig *yaml.RNode

	WrappingApiVersion string
	WrappingKind       string
}

var _ Reader = &ByteReader{}

func (r *ByteReader) Read() ([]*yaml.RNode, error) {
	output := ResourceNodeSlice{}

	// by manually splitting resources -- otherwise the decoder will get the Resource
	// boundaries wrong for header comments.
	input := &bytes.Buffer{}
	_, err := io.Copy(input, r.Reader)
	if err != nil {
		return nil, err
	}
	values := strings.Split(input.String(), "\n---\n")

	index := 0
	for i := range values {
		decoder := yaml.NewDecoder(bytes.NewBufferString(values[i]))
		node, err := r.decode(index, decoder)
		if err == io.EOF {
			continue
		}
		if err != nil {
			return nil, err
		}
		if yaml.IsMissingOrNull(node) {
			// empty value
			continue
		}

		// ok if no metadata -- assume not an InputList
		meta, _ := node.GetMeta()

		// the elements are wrapped in an InputList, unwrap them
		// don't check apiVersion, we haven't standardized on the domain
		if (meta.Kind == ResourceListKind || meta.Kind == "List") &&
			node.Field("items") != nil {
			r.WrappingKind = meta.Kind
			r.WrappingApiVersion = meta.ApiVersion

			// unwrap the list
			fc := node.Field("functionConfig")
			if fc != nil {
				r.FunctionConfig = fc.Value
			}

			items := node.Field("items")
			if items != nil {
				for i := range items.Value.Content() {
					// add items
					output = append(output, yaml.NewRNode(items.Value.Content()[i]))
				}

			}
			continue
		}

		// add the node to the list
		output = append(output, node)

		// increment the index annotation value
		index++
	}
	return output, nil
}

func isEmptyDocument(node *yaml.Node) bool {
	// node is a Document with no content -- e.g. "---\n---"
	return node.Kind == yaml.DocumentNode &&
		node.Content[0].Tag == yaml.NullNodeTag
}

func (r *ByteReader) decode(index int, decoder *yaml.Decoder) (*yaml.RNode, error) {
	node := &yaml.Node{}
	err := decoder.Decode(node)
	if err == io.EOF {
		return nil, io.EOF
	}
	if err != nil {
		return nil, err
	}

	if isEmptyDocument(node) {
		return nil, nil
	}

	// set annotations on the read Resources
	// sort the annotations by key so the output Resources is consistent (otherwise the
	// annotations will be in a random order)
	n := yaml.NewRNode(node)
	if r.SetAnnotations == nil {
		r.SetAnnotations = map[string]string{}
	}
	if !r.OmitReaderAnnotations {
		r.SetAnnotations[kioutil.IndexAnnotation] = fmt.Sprintf("%d", index)
	}
	var keys []string
	for k := range r.SetAnnotations {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		_, err = n.Pipe(yaml.SetAnnotation(k, r.SetAnnotations[k]))
		if err != nil {
			return nil, err
		}
	}
	return yaml.NewRNode(node), nil
}

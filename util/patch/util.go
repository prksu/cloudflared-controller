/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package patch

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type patchType string

func (p patchType) Key() string {
	return strings.Split(string(p), ".")[0]
}

const (
	specPatch   patchType = "spec"
	statusPatch patchType = "status"
)

var preserveUnstructuredKeys = map[string]bool{
	"kind":       true,
	"apiVersion": true,
	"metadata":   true,
}

func toUnstructured(obj runtime.Object) (*unstructured.Unstructured, error) {
	// If the incoming object is already unstructured, perform a deep copy first
	// otherwise DefaultUnstructuredConverter ends up returning the inner map without
	// making a copy.
	if _, ok := obj.(runtime.Unstructured); ok {
		obj = obj.DeepCopyObject()
	}
	rawMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: rawMap}, nil
}

// unsafeUnstructuredCopy returns a shallow copy of the unstructured object given as input.
// It copies the common fields such as `kind`, `apiVersion`, `metadata` and the patchType specified.
//
// It's not safe to modify any of the keys in the returned unstructured object, the result should be treated as read-only.
func unsafeUnstructuredCopy(obj *unstructured.Unstructured, focus patchType) *unstructured.Unstructured {
	// Create the return focused-unstructured object with a preallocated map.
	res := &unstructured.Unstructured{Object: make(map[string]interface{}, len(obj.Object))}

	// Ranges over the keys of the unstructured object, think of this as the very top level of an object
	// when submitting a yaml to kubectl or a client.
	// These would be keys like `apiVersion`, `kind`, `metadata`, `spec`, `status`, etc.
	for key := range obj.Object {
		value := obj.Object[key]

		// Perform a shallow copy only for the keys we're interested in, or the ones that should be always preserved.
		if key == focus.Key() || preserveUnstructuredKeys[key] {
			res.Object[key] = value
		}
	}

	return res
}

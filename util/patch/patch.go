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
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// Patcher is a utility for ensuring the proper patching of objects.
type Patcher struct {
	client       client.Client
	gvk          schema.GroupVersionKind
	beforeObject client.Object
	before       *unstructured.Unstructured
	after        *unstructured.Unstructured
	changes      map[string]bool
}

// NewPatcher returns an initialized Patcher.
func NewPatcher(obj client.Object, crclient client.Client) (*Patcher, error) {
	// Get the GroupVersionKind of the object,
	// used to validate against later on.
	gvk, err := apiutil.GVKForObject(obj, crclient.Scheme())
	if err != nil {
		return nil, err
	}

	// Convert the object to unstructured to compare against our before copy.
	unstructuredObj, err := toUnstructured(obj)
	if err != nil {
		return nil, err
	}

	return &Patcher{
		client:       crclient,
		gvk:          gvk,
		before:       unstructuredObj,
		beforeObject: obj.DeepCopyObject().(client.Object),
	}, nil
}

// Patch will attempt to patch the given object, including its status.
func (p *Patcher) Patch(ctx context.Context, obj client.Object) error {
	// Get the GroupVersionKind of the object that we want to patch.
	gvk, err := apiutil.GVKForObject(obj, p.client.Scheme())
	if err != nil {
		return err
	}
	if gvk != p.gvk {
		return fmt.Errorf("unmatched GroupVersionKind, expected %q got %q", p.gvk, gvk)
	}

	// Convert the object to unstructured to compare against our before copy.
	p.after, err = toUnstructured(obj)
	if err != nil {
		return err
	}

	// Calculate and store the top-level field changes (e.g. "metadata", "spec", "status") we have before/after.
	p.changes, err = p.calculateChanges(obj)
	if err != nil {
		return err
	}

	// Issue patches and return errors in an aggregate.
	return kerrors.NewAggregate([]error{
		// Then proceed to patch the rest of the object.
		p.patch(ctx, obj),
		p.patchStatus(ctx, obj),
	})
}

// patch issues a patch for metadata and spec.
func (p *Patcher) patch(ctx context.Context, obj client.Object) error {
	if !p.shouldPatch("metadata") && !p.shouldPatch("spec") {
		return nil
	}
	beforeObject, afterObject, err := p.calculatePatch(obj, specPatch)
	if err != nil {
		return err
	}
	return p.client.Patch(ctx, afterObject, client.MergeFrom(beforeObject))
}

// patchStatus issues a patch if the status has changed.
func (p *Patcher) patchStatus(ctx context.Context, obj client.Object) error {
	if !p.shouldPatch("status") {
		return nil
	}
	beforeObject, afterObject, err := p.calculatePatch(obj, statusPatch)
	if err != nil {
		return err
	}
	return p.client.Status().Patch(ctx, afterObject, client.MergeFrom(beforeObject))
}

// calculatePatch returns the before/after objects to be given in a controller-runtime patch, scoped down to the absolute necessary.
func (p *Patcher) calculatePatch(afterObj client.Object, focus patchType) (client.Object, client.Object, error) {
	// Get a shallow unsafe copy of the before/after object in unstructured form.
	before := unsafeUnstructuredCopy(p.before, focus)
	after := unsafeUnstructuredCopy(p.after, focus)

	// We've now applied all modifications to local unstructured objects,
	// make copies of the original objects and convert them back.
	beforeObj := p.beforeObject.DeepCopyObject().(client.Object)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(before.Object, beforeObj); err != nil {
		return nil, nil, err
	}
	afterObj = afterObj.DeepCopyObject().(client.Object)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(after.Object, afterObj); err != nil {
		return nil, nil, err
	}
	return beforeObj, afterObj, nil
}

func (p *Patcher) shouldPatch(in string) bool {
	return p.changes[in]
}

// calculate changes tries to build a patch from the before/after objects we have
// and store in a map which top-level fields (e.g. `metadata`, `spec`, `status`, etc.) have changed.
func (p *Patcher) calculateChanges(after client.Object) (map[string]bool, error) {
	// Calculate patch data.
	patch := client.MergeFrom(p.beforeObject)
	diff, err := patch.Data(after)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate patch data - error: %v", err)
	}

	// Unmarshal patch data into a local map.
	patchDiff := map[string]interface{}{}
	if err := json.Unmarshal(diff, &patchDiff); err != nil {
		return nil, fmt.Errorf("failed to unmarshal patch data into a map - error: %v", err)
	}

	// Return the map.
	res := make(map[string]bool, len(patchDiff))
	for key := range patchDiff {
		res[key] = true
	}
	return res, nil
}

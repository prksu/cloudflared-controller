/*
Copyright 2021 Ahmad Nurus S.

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

package util

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetOriginCertSecret(ctx context.Context, crclient client.Client, namespace string, ref *corev1.TypedLocalObjectReference) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	gv, err := schema.ParseGroupVersion(corev1.SchemeGroupVersion.String())
	if err != nil {
		return nil, err
	}

	gvk := gv.WithKind(ref.Kind)
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)
	key := client.ObjectKey{Namespace: namespace, Name: ref.Name}
	if err := crclient.Get(ctx, key, u); err != nil {
		return nil, err
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, secret); err != nil {
		return nil, err
	}

	return secret, nil
}

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
	"errors"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cloudflaredv1alpha1 "github.com/prksu/cloudflared-controller/api/v1alpha1"
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

func TunnelConfigurationFromIngress(ctx context.Context, crclient client.Client, ing *networkingv1.Ingress) (*cloudflaredv1alpha1.TunnelConfiguration, error) {
	log := log.FromContext(ctx)
	tc := &cloudflaredv1alpha1.TunnelConfiguration{}
	ic := &networkingv1.IngressClass{}
	if err := crclient.Get(ctx, client.ObjectKey{Name: *ing.Spec.IngressClassName}, ic); err != nil {
		return nil, err
	}

	tcgvk, err := apiutil.GVKForObject(tc, crclient.Scheme())
	if err != nil {
		log.Error(err, "unable to parse TunnelConfiguration GVK")
		return nil, err
	}

	if tcgvk.Group != *ic.Spec.Parameters.APIGroup || tcgvk.Kind != ic.Spec.Parameters.Kind {
		return nil, errors.New("IngressClass parameters does not match with TunnelConfiguration Group Kind")
	}

	// NOTE: TunnelConfiguration is namespaced scope we use the ingress namespace to getting the resource for now
	// until we adopt Kubernetes v1.21 types where the IngressClass parameter has namespaced scope
	if err := crclient.Get(ctx, client.ObjectKey{Namespace: ing.Namespace, Name: ic.Spec.Parameters.Name}, tc); err != nil {
		log.Error(err, "unable to get TunnelConfiguration object")
		return nil, err
	}

	return tc, nil
}

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

package resources

import (
	"reflect"
	"testing"

	cloudflaredv1alpha1 "github.com/prksu/cloudflared-controller/api/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ingressResource_Tunnel(t *testing.T) {
	type fields struct {
		Ingress *networkingv1.Ingress
	}
	tests := []struct {
		name   string
		fields fields
		want   *cloudflaredv1alpha1.Tunnel
	}{
		{
			name: "default",
			fields: fields{
				Ingress: &networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "default",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
			want: &cloudflaredv1alpha1.Tunnel{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewIngressResources(tt.fields.Ingress)
			if got := r.Tunnel(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ingressResource.Tunnel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ingressResource_TunnelIngressRules(t *testing.T) {
	type fields struct {
		Ingress *networkingv1.Ingress
	}
	tests := []struct {
		name   string
		fields fields
		want   []cloudflaredv1alpha1.TunnelIngressRule
	}{
		{
			name: "default (should be turn Ingress default backend into Tunnel default route)",
			fields: fields{
				Ingress: &networkingv1.Ingress{
					Spec: networkingv1.IngressSpec{
						DefaultBackend: &networkingv1.IngressBackend{
							Service: &networkingv1.IngressServiceBackend{
								Name: "foo",
								Port: networkingv1.ServiceBackendPort{
									Number: 8000,
								},
							},
						},
					},
				},
			},
			want: []cloudflaredv1alpha1.TunnelIngressRule{
				{
					Service: "http://foo:8000",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewIngressResources(tt.fields.Ingress)
			if got := r.TunnelIngressRules(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ingressResource.TunnelIngressRules() = %v, want %v", got, tt.want)
			}
		})
	}
}

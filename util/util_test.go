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
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	cloudflaredv1alpha1 "github.com/prksu/cloudflared-controller/api/v1alpha1"
)

func TestGetOriginCertSecret(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	type args struct {
		ctx       context.Context
		crclient  client.Client
		namespace string
		ref       *corev1.TypedLocalObjectReference
	}
	tests := []struct {
		name    string
		args    args
		want    *corev1.Secret
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				ctx: context.Background(),
				crclient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-origincret-secret",
							Namespace: "default",
						},
					},
				).Build(),
				namespace: "default",
				ref: &corev1.TypedLocalObjectReference{
					Kind: "Secret",
					Name: "my-origincret-secret",
				},
			},
			want: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-origincret-secret",
					Namespace: "default",
				},
			},
		},
		{
			name: "secret not found",
			args: args{
				ctx:       context.Background(),
				crclient:  fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build(),
				namespace: "default",
				ref: &corev1.TypedLocalObjectReference{
					Kind: "Secret",
					Name: "my-origincret-secret",
				},
			},
			wantErr: true,
		},
		{
			name: "ref does not have Kind",
			args: args{
				ctx: context.Background(),
				crclient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-origincret-secret",
							Namespace: "default",
						},
					},
				).Build(),
				namespace: "default",
				ref: &corev1.TypedLocalObjectReference{
					Kind: "",
					Name: "my-origincret-secret",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOriginCertSecret(tt.args.ctx, tt.args.crclient, tt.args.namespace, tt.args.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOriginCertSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil {
				tt.want.SetResourceVersion("999")
				tt.want.SetGroupVersionKind(got.GroupVersionKind())
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOriginCertSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTunnelConfigurationFromIngress(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(cloudflaredv1alpha1.AddToScheme(scheme))

	type args struct {
		ctx      context.Context
		crclient client.Client
		ing      *networkingv1.Ingress
	}
	tests := []struct {
		name    string
		args    args
		want    *cloudflaredv1alpha1.TunnelConfiguration
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				ctx: context.Background(),
				crclient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
					&networkingv1.IngressClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "my-ingressclass",
						},
						Spec: networkingv1.IngressClassSpec{
							Parameters: &corev1.TypedLocalObjectReference{
								APIGroup: pointer.StringPtr("cloudflared.cloudflare.com"),
								Kind:     "TunnelConfiguration",
								Name:     "my-tunnel-config",
							},
						},
					},
					&cloudflaredv1alpha1.TunnelConfiguration{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-tunnel-config",
							Namespace: "default",
						},
					},
				).Build(),
				ing: &networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-ing",
						Namespace: "default",
					},
					Spec: networkingv1.IngressSpec{
						IngressClassName: pointer.StringPtr("my-ingressclass"),
					},
				},
			},
			want: &cloudflaredv1alpha1.TunnelConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-tunnel-config",
					Namespace: "default",
				},
			},
		},
		{
			name: "ingressclass parameters does not match",
			args: args{
				ctx: context.Background(),
				crclient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
					&networkingv1.IngressClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "my-ingressclass",
						},
						Spec: networkingv1.IngressClassSpec{
							Parameters: &corev1.TypedLocalObjectReference{
								APIGroup: pointer.StringPtr("cloudflared.cloudflare.com"),
								Kind:     "FooConfiguration",
								Name:     "my-tunnel-config",
							},
						},
					},
					&cloudflaredv1alpha1.TunnelConfiguration{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-tunnel-config",
							Namespace: "default",
						},
					},
				).Build(),
				ing: &networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-ing",
						Namespace: "default",
					},
					Spec: networkingv1.IngressSpec{
						IngressClassName: pointer.StringPtr("my-ingressclass"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "tunnelconfigurations not found",
			args: args{
				ctx: context.Background(),
				crclient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
					&networkingv1.IngressClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "my-ingressclass",
						},
						Spec: networkingv1.IngressClassSpec{
							Parameters: &corev1.TypedLocalObjectReference{
								APIGroup: pointer.StringPtr("cloudflared.cloudflare.com"),
								Kind:     "TunnelConfiguration",
								Name:     "my-tunnel-config",
							},
						},
					},
				).Build(),
				ing: &networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-ing",
						Namespace: "default",
					},
					Spec: networkingv1.IngressSpec{
						IngressClassName: pointer.StringPtr("my-ingressclass"),
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TunnelConfigurationFromIngress(tt.args.ctx, tt.args.crclient, tt.args.ing)
			if (err != nil) != tt.wantErr {
				t.Errorf("TunnelConfigurationFromIngress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil {
				tt.want.SetResourceVersion("999")
				tt.want.SetGroupVersionKind(got.GroupVersionKind())
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TunnelConfigurationFromIngress() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/yaml"
)

func Test_tunnelResource_Secret(t *testing.T) {
	type fields struct {
		Tunnel *cloudflaredv1alpha1.Tunnel
	}
	type args struct {
		data map[string][]byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *corev1.Secret
	}{
		{
			name: "default",
			fields: fields{
				Tunnel: &cloudflaredv1alpha1.Tunnel{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-tunnel",
						Namespace: "default",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
			args: args{data: map[string][]byte{
				"value": []byte("baz"),
			}},
			want: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tunnel-secret",
					Namespace: "default",
					Labels: map[string]string{
						"cloudflared.cloudflare.com/managed-by": "test-tunnel",
						"foo":                                   "bar",
					},
				},
				Immutable: pointer.BoolPtr(true),
				Data: map[string][]byte{
					"value": []byte("baz"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewTunnelResources(tt.fields.Tunnel)
			if got := r.Secret(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tunnelResource.Secret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tunnelResource_ConfigMap(t *testing.T) {
	type fields struct {
		Tunnel *cloudflaredv1alpha1.Tunnel
	}
	tests := []struct {
		name   string
		fields fields
		want   *corev1.ConfigMap
	}{
		{
			name: "default",
			fields: fields{
				Tunnel: &cloudflaredv1alpha1.Tunnel{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-tunnel",
						Namespace: "default",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
			want: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tunnel-config",
					Namespace: "default",
					Labels: map[string]string{
						"cloudflared.cloudflare.com/managed-by": "test-tunnel",
						"foo":                                   "bar",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewTunnelResources(tt.fields.Tunnel)
			if got := r.ConfigMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tunnelResource.ConfigMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tunnelResource_ConfigMapData(t *testing.T) {
	type fields struct {
		Tunnel *cloudflaredv1alpha1.Tunnel
	}
	tests := []struct {
		name   string
		fields fields
		want   struct {
			Tunnel          string                                  `json:"tunnel,omitempty"`
			CredentialsFile string                                  `json:"credentials-file,omitempty"`
			Ingress         []cloudflaredv1alpha1.TunnelIngressRule `json:"ingress,omitempty"`
		}
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				Tunnel: &cloudflaredv1alpha1.Tunnel{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-tunnel",
						Namespace: "default",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
					Spec: cloudflaredv1alpha1.TunnelSpec{
						IngressRules: []cloudflaredv1alpha1.TunnelIngressRule{
							{
								Service: "http://foo:8000",
							},
						},
					},
				},
			},
			want: struct {
				Tunnel          string                                  `json:"tunnel,omitempty"`
				CredentialsFile string                                  `json:"credentials-file,omitempty"`
				Ingress         []cloudflaredv1alpha1.TunnelIngressRule `json:"ingress,omitempty"`
			}{
				Tunnel:          "k8s-test-tunnel",
				CredentialsFile: "/etc/cloudflared/k8s-test-tunnel.json",
				Ingress: []cloudflaredv1alpha1.TunnelIngressRule{
					{
						Service: "http://foo:8000",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewTunnelResources(tt.fields.Tunnel)
			got, err := r.ConfigMapData()
			if (err != nil) != tt.wantErr {
				t.Errorf("tunnelResource.ConfigMapData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			b, err := yaml.Marshal(tt.want)
			if (err != nil) != tt.wantErr {
				t.Errorf("tunnelResource.ConfigMapData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := map[string]string{
				"config.yaml": string(b),
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("tunnelResource.ConfigMapData() = %v, want %v", got, want)
			}
		})
	}
}

func Test_tunnelResource_Deployment(t *testing.T) {
	type fields struct {
		Tunnel *cloudflaredv1alpha1.Tunnel
	}
	tests := []struct {
		name   string
		fields fields
		want   *appsv1.Deployment
	}{
		{
			name: "default",
			fields: fields{
				Tunnel: &cloudflaredv1alpha1.Tunnel{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-tunnel",
						Namespace: "default",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
					Spec: cloudflaredv1alpha1.TunnelSpec{
						TunnelConfigurationSpec: cloudflaredv1alpha1.TunnelConfigurationSpec{
							OriginCert: &corev1.TypedLocalObjectReference{
								Kind: "Secret",
								Name: "test-origincert",
							},
						},
					},
				},
			},
			want: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-tunnel",
					Namespace: "default",
					Labels: map[string]string{
						"cloudflared.cloudflare.com/managed-by": "test-tunnel",
						"foo":                                   "bar",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: pointer.Int32Ptr(0),
					Selector: metav1.SetAsLabelSelector(
						map[string]string{
							"cloudflared.cloudflare.com/managed-by": "test-tunnel",
						},
					),
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"cloudflared.cloudflare.com/managed-by": "test-tunnel",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "cloudflared",
									Image: "cloudflare/cloudflared:2021.5.10",
									Env: []corev1.EnvVar{
										{
											Name:  "TUNNEL_ORIGIN_CERT",
											Value: "/etc/cloudflared/cert.pem",
										},
									},
									Command: []string{"cloudflared", "tunnel"},
									Args:    []string{"--no-autoupdate", "--config", "/.cloudflared/config.yaml", "run"},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "secret",
											MountPath: "/etc/cloudflared",
										},
										{
											Name:      "config",
											MountPath: "/.cloudflared",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "secret",
									VolumeSource: corev1.VolumeSource{
										Projected: &corev1.ProjectedVolumeSource{
											Sources: []corev1.VolumeProjection{
												{
													Secret: &corev1.SecretProjection{
														LocalObjectReference: corev1.LocalObjectReference{
															Name: "test-tunnel-secret",
														},
													},
												},
												{
													Secret: &corev1.SecretProjection{
														LocalObjectReference: corev1.LocalObjectReference{
															Name: "test-origincert",
														},
													},
												},
											},
										},
									},
								},
								{
									Name: "config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "test-tunnel-config",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewTunnelResources(tt.fields.Tunnel)
			if got := r.Deployment(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tunnelResource.Deployment() = %v, want %v", got, tt.want)
			}
		})
	}
}

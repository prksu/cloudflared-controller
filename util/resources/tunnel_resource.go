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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/yaml"

	cloudflaredv1alpha1 "github.com/prksu/cloudflared-controller/api/v1alpha1"
)

type TunnelResourceGetter interface {
	TunnelName() string
	SecretName() string
	ConfigMapName() string

	Secret(data map[string][]byte) *corev1.Secret
	ConfigMap() *corev1.ConfigMap
	ConfigMapData() (map[string]string, error)

	Deployment() *appsv1.Deployment
}

type tunnelResource struct {
	*cloudflaredv1alpha1.Tunnel
}

func NewTunnelResources(tunnel *cloudflaredv1alpha1.Tunnel) TunnelResourceGetter {
	return tunnelResource{
		Tunnel: tunnel,
	}
}

func (r tunnelResource) TunnelName() string    { return "k8s-" + r.Name }
func (r tunnelResource) SecretName() string    { return r.Name + "-secret" }
func (r tunnelResource) ConfigMapName() string { return r.Name + "-config" }

func (r tunnelResource) CommonLabels() map[string]string {
	return map[string]string{
		"cloudflared.cloudflare.com/managed-by": r.Name,
	}
}

func (r tunnelResource) Secret(data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.SecretName(),
			Namespace: r.Namespace,
			Labels:    labels.Merge(r.Labels, r.CommonLabels()),
		},
		Data:      data,
		Immutable: pointer.BoolPtr(true),
	}
}

func (r tunnelResource) ConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.ConfigMapName(),
			Namespace: r.Namespace,
			Labels:    labels.Merge(r.Labels, r.CommonLabels()),
		},
	}
}

func (r tunnelResource) ConfigMapData() (map[string]string, error) {
	data := make(map[string]string)
	config := struct {
		Tunnel          string                                  `json:"tunnel,omitempty"`
		CredentialsFile string                                  `json:"credentials-file,omitempty"`
		Ingress         []cloudflaredv1alpha1.TunnelIngressRule `json:"ingress,omitempty"`
	}{
		Tunnel:          r.TunnelName(),
		CredentialsFile: "/etc/cloudflared/" + r.TunnelName() + ".json",
		Ingress:         r.Spec.IngressRules,
	}

	b, err := yaml.Marshal(config)
	data["config.yaml"] = string(b)
	return data, err
}

func (r tunnelResource) Deployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.Namespace,
			Labels:    labels.Merge(r.Labels, r.CommonLabels()),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(0),
			Selector: metav1.SetAsLabelSelector(r.CommonLabels()),
			Template: r.PodTemplate(),
		},
	}
}

func (r tunnelResource) PodTemplate() corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: r.CommonLabels(),
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
											Name: r.SecretName(),
										},
									},
								},
								{
									Secret: &corev1.SecretProjection{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: r.Spec.OriginCert.Name,
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
								Name: r.ConfigMapName(),
							},
						},
					},
				},
			},
		},
	}
}

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
	"strconv"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudflaredv1alpha1 "github.com/prksu/cloudflared-controller/api/v1alpha1"
)

type IngressResourceGetter interface {
	Tunnel() *cloudflaredv1alpha1.Tunnel
	TunnelIngressRules() []cloudflaredv1alpha1.TunnelIngressRule
}

type ingressResource struct {
	*networkingv1.Ingress
}

func NewIngressResources(ing *networkingv1.Ingress) IngressResourceGetter {
	return ingressResource{
		Ingress: ing,
	}
}

func (r ingressResource) Tunnel() *cloudflaredv1alpha1.Tunnel {
	return &cloudflaredv1alpha1.Tunnel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Ingress.Name,
			Namespace: r.Ingress.Namespace,
			Labels:    r.Ingress.Labels,
		},
		Spec: cloudflaredv1alpha1.TunnelSpec{},
	}
}

func (r ingressResource) TunnelIngressRules() []cloudflaredv1alpha1.TunnelIngressRule {
	var ir []cloudflaredv1alpha1.TunnelIngressRule
	if r.Spec.DefaultBackend != nil {
		svc := r.Spec.DefaultBackend.Service
		ir = append(ir, cloudflaredv1alpha1.TunnelIngressRule{
			Service: "http://" + svc.Name + ":" + strconv.Itoa(int(svc.Port.Number)),
		})
	}

	return ir
}

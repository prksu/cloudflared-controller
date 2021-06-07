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
	"strings"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudflaredv1alpha1 "github.com/prksu/cloudflared-controller/api/v1alpha1"
	"github.com/prksu/cloudflared-controller/util"
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
	var tirList []cloudflaredv1alpha1.TunnelIngressRule
	for _, rule := range r.Spec.Rules {
		for _, hip := range rule.HTTP.Paths {

			// Path matching according https://kubernetes.io/docs/concepts/services-networking/ingress/#examples
			//
			// - PathTypeExact:
			//	1. Add "^" at beginning and end with "$" to make it exact match
			// - PathTypePrefix:
			//	1. Add "^" at beginning and end with "$" to make it exact match
			//	2. Add "^" at beginning and end with "/" to make it matches subpath
			// - PathTypeImplementationSpecific:
			//	1. Leave as is
			switch *hip.PathType {
			case networkingv1.PathTypeExact:
				var tir cloudflaredv1alpha1.TunnelIngressRule
				tir.Path = "^" + hip.Path + "$"
				tir.Hostname = rule.Host
				tir.Service = util.FormatIngressServiceBackend(hip.Backend.Service)
				tirList = append(tirList, tir)
			case networkingv1.PathTypePrefix:
				var tir1 cloudflaredv1alpha1.TunnelIngressRule
				var tir2 cloudflaredv1alpha1.TunnelIngressRule
				path := strings.TrimSuffix(hip.Path, "/")
				tir1.Hostname = rule.Host
				tir1.Path = "^" + path + "$"
				tir1.Service = util.FormatIngressServiceBackend(hip.Backend.Service)
				tir2.Hostname = rule.Host
				tir2.Path = "^" + path + "/"
				tir2.Service = util.FormatIngressServiceBackend(hip.Backend.Service)
				tirList = append(tirList, tir1)
				tirList = append(tirList, tir2)
			default:
				var tir cloudflaredv1alpha1.TunnelIngressRule
				tir.Hostname = rule.Host
				tir.Path = hip.Path
				tir.Service = util.FormatIngressServiceBackend(hip.Backend.Service)
			}
		}
	}

	// Required default rule which match all URLs.
	// use cloudflare http_status:404 if there is no default backend specific
	switch r.Spec.DefaultBackend {
	case nil:
		tirList = append(tirList, cloudflaredv1alpha1.TunnelIngressRule{
			Service: "http_status:404",
		})
	default:
		svc := r.Spec.DefaultBackend.Service
		tirList = append(tirList, cloudflaredv1alpha1.TunnelIngressRule{
			Service: util.FormatIngressServiceBackend(svc),
		})
	}

	return tirList
}

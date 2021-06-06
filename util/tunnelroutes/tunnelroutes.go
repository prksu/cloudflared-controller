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

package tunnelroutes

import (
	"k8s.io/apimachinery/pkg/util/sets"

	cloudflaredv1alpha1 "github.com/prksu/cloudflared-controller/api/v1alpha1"
)

func FromTunnelSpec(spec cloudflaredv1alpha1.TunnelSpec) []string {
	result := sets.NewString()
	for _, rule := range spec.IngressRules {
		// TODO(prksu): Hostname with wildcard should be inserted without wildcard character (*.)
		if rule.Hostname == "" {
			continue
		}

		result.Insert(rule.Hostname)
	}

	return result.List()
}

func FromTunnelStatus(status cloudflaredv1alpha1.TunnelStatus) []string {
	return status.Routes
}

func Difference(r1, r2 []string) []string {
	return sets.NewString(r1...).Difference(sets.NewString(r2...)).List()
}

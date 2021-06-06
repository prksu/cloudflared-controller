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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TunnelConfigurationSpec defines the desired state of TunnelConfiguration
type TunnelConfigurationSpec struct {
	// OriginCert is a reference to a object that contains cloudflare tunnel origincert.
	OriginCert *corev1.TypedLocalObjectReference `json:"originCert,omitempty"`
}

// +kubebuilder:object:root=true

// TunnelConfiguration is the Schema for the tunnelconfigurations API
type TunnelConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TunnelConfigurationSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// TunnelConfigurationList contains a list of TunnelConfiguration
type TunnelConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TunnelConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TunnelConfiguration{}, &TunnelConfigurationList{})
}

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

// TunnelOriginRequest defines originRequest optional configurations.
type TunnelOriginRequest struct {
	// Timeout for establishing a new TCP connection to your origin server.
	// This excludes the time taken to establish TLS. (Default: 30s)
	// +optional
	ConnectTimeout string `json:"connectTimeout"`

	// Timeout for completing a TLS handshake to your origin server,
	// if you have chosen to connect Tunnel to an HTTPS server. (Default: 10s)
	// +optional
	TLSTimeout string `json:"tlsTimeout"`

	// Disables chunked transfer encoding. Useful if you are running
	// a WSGI server. (Default: false)
	// +optional
	DisableChunkedEncoding bool `json:"disableChunkedEncoding"`

	// Sets the HTTP Host header on requests sent to the local service.
	// +optional
	HTTPHostHeader string `json:"httpHostHeader"`

	// The timeout after which a TCP keepalive packet is sent on a connection
	// between Tunnel and the origin server. (Default: 30s)
	// +optional
	TCPKeepAlive string `json:"tcpKeepAlive"`

	// Maximum number of idle keepalive connections between Tunnel and your origin.
	// This does not restrict the total number of concurrent connections. (Default: 100)
	// +optional
	KeepAliveConnections int32 `json:"keepAliveConnections"`

	// Timeout after which an idle keepalive connection can be discarded. (Default: 1m30s)
	// +optional
	KeepAliveTimeout string `json:"keepAliveTimeout"`

	// Disables TLS verification of the certificate presented by your origin.
	// Will allow any certificate from the origin to be accepted. (Default: false)
	// +optional
	NoTLSVerify bool `json:"noTLSVerify"`

	// Hostname that cloudflared should expect from your origin server certificate.
	// +optional
	OriginServerName string `json:"originServerName"`
}

// TunnelConfigurationSpec defines the desired state of TunnelConfiguration
type TunnelConfigurationSpec struct {
	// OriginCert is a reference to a object that contains cloudflare tunnel origincert.
	OriginCert *corev1.TypedLocalObjectReference `json:"originCert,omitempty"`

	// OriginRequest is optional origin configurations. See
	// https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/configuration/ingress#origin-configurations
	// +optional
	OriginRequest *TunnelOriginRequest `json:"originRequest,omitempty"`
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

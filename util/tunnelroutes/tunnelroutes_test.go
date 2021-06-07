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
	"reflect"
	"testing"

	cloudflaredv1alpha1 "github.com/prksu/cloudflared-controller/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestFromTunnelSpec(t *testing.T) {
	type args struct {
		spec cloudflaredv1alpha1.TunnelSpec
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "default",
			args: args{
				spec: cloudflaredv1alpha1.TunnelSpec{
					IngressRules: []cloudflaredv1alpha1.TunnelIngressRule{
						{
							Hostname: "foo.example.com",
						},
						{
							Hostname: "bar.example.com",
						},
					},
				},
			},
			want: sets.NewString("foo.example.com", "bar.example.com").List(),
		},
		{
			name: "hostname with empty string (should be ignore)",
			args: args{
				spec: cloudflaredv1alpha1.TunnelSpec{
					IngressRules: []cloudflaredv1alpha1.TunnelIngressRule{
						{
							Hostname: "foo.example.com",
						},
						{
							Hostname: "bar.example.com",
						},
						{
							Hostname: "",
						},
					},
				},
			},
			want: sets.NewString("foo.example.com", "bar.example.com").List(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromTunnelSpec(tt.args.spec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromTunnelSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromTunnelStatus(t *testing.T) {
	type args struct {
		status cloudflaredv1alpha1.TunnelStatus
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "default",
			args: args{
				status: cloudflaredv1alpha1.TunnelStatus{
					Routes: []string{"foo.example.com", "bar.example.com"},
				},
			},
			want: []string{"foo.example.com", "bar.example.com"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromTunnelStatus(tt.args.status); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromTunnelStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDifference(t *testing.T) {
	type args struct {
		r1 []string
		r2 []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "default",
			args: args{
				r1: []string{"foo.example.com", "bar.example.com"},
				r2: []string{"foo.example.com"},
			},
			want: []string{"bar.example.com"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Difference(tt.args.r1, tt.args.r2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Difference() = %v, want %v", got, tt.want)
			}
		})
	}
}

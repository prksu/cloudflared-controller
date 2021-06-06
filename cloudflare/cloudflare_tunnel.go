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

package cloudflare

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type TunnelClient interface {
	Get(ctx context.Context, tunnelID uuid.UUID) (*Tunnel, error)
	GetByName(ctx context.Context, name string) (*Tunnel, error)
	List(ctx context.Context, opts *TunnelListOptions) ([]*Tunnel, error)
	Create(ctx context.Context, name string) (*Tunnel, error)
	Delete(ctx context.Context, tunnelID uuid.UUID) error
	Route(ctx context.Context, tunnelID uuid.UUID, route TunnelRoute) error
}

type TunnelCredentials struct {
	AccountTag   string
	TunnelSecret []byte
	TunnelID     uuid.UUID
	TunnelName   string
}

type Tunnel struct {
	ID              uuid.UUID          `json:"id"`
	Name            string             `json:"name"`
	CreatedAt       time.Time          `json:"created_at"`
	DeletedAt       time.Time          `json:"deleted_at"`
	CredentialsFile *TunnelCredentials `json:"credentials_file,omitempty"`
}

type TunnelListOptions struct {
	UUID      string `url:"uuid,omitempty"`
	Name      string `url:"name,omitempty"`
	IsDeleted bool   `url:"is_deleted"`
	ExistedAt string `url:"existed_at,omitempty"`
}

type TunnelRoute interface {
	json.Marshaler
	Type() string
}

type TunnelDNSRoute struct {
	Hostname          string
	OverwriteExisting bool
}

func (r *TunnelDNSRoute) Type() string {
	return "dns"
}

func (r *TunnelDNSRoute) MarshalJSON() ([]byte, error) {
	s := struct {
		Type              string `json:"type"`
		UserHostname      string `json:"user_hostname"`
		OverwriteExisting bool   `json:"overwrite_existing"`
	}{
		Type:              r.Type(),
		UserHostname:      r.Hostname,
		OverwriteExisting: r.OverwriteExisting,
	}
	return json.Marshal(&s)
}

type TunnelLBRoute struct {
	LBName string
	LBPool string
}

func (r *TunnelLBRoute) Type() string {
	return "lb"
}

func (r *TunnelLBRoute) MarshalJSON() ([]byte, error) {
	s := struct {
		Type   string `json:"type"`
		LBName string `json:"lb_name"`
		LBPool string `json:"lb_pool"`
	}{
		Type:   r.Type(),
		LBName: r.LBName,
		LBPool: r.LBPool,
	}
	return json.Marshal(&s)
}

type tunnels struct {
	client *client
}

func newTunnels(c *client) *tunnels {
	return &tunnels{
		client: c,
	}
}

// Get fetch a tunnel by name.
//
// API reference: https://api.cloudflare.com/#argo-tunnel-get-argo-tunnel
func (s *tunnels) Get(ctx context.Context, tunnelID uuid.UUID) (*Tunnel, error) {
	s.client.logger.V(1).Info("Getting tunnel details", "tunnel-id", tunnelID.String())
	tunnel := &Tunnel{}
	err := NewRequest(s.client).
		Verb(http.MethodGet).
		AccountPrefix(s.client.accountID).
		Resource("tunnels").
		ResourceID(tunnelID.String()).
		Header("Accept", "application/json;version=1").
		Do(ctx).
		Into(tunnel)
	return tunnel, err
}

// Get fetch a tunnel by name.
func (s *tunnels) GetByName(ctx context.Context, name string) (*Tunnel, error) {
	s.client.logger.V(1).Info("Getting tunnel details by name", "name", name)
	tunnelList, err := s.List(ctx, &TunnelListOptions{
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	switch len(tunnelList) {
	case 0:
		return nil, ErrNotFound
	case 1:
		return tunnelList[0], nil
	default:
		return nil, ErrUnexpectedMultipleTunnel
	}
}

// List retrieves all tunnels.
//
// API reference: https://api.cloudflare.com/#argo-tunnel-list-argo-tunnels
func (s *tunnels) List(ctx context.Context, opts *TunnelListOptions) ([]*Tunnel, error) {
	var tunnelList []*Tunnel
	if opts == nil {
		opts = &TunnelListOptions{}
	}

	s.client.logger.V(1).Info("Retriving tunnels", "options", opts)
	err := NewRequest(s.client).
		Verb(http.MethodGet).
		AccountPrefix(s.client.accountID).
		Resource("tunnels").
		Header("Accept", "application/json;version=1").
		Param(opts).
		Do(ctx).
		Into(&tunnelList)
	return tunnelList, err
}

// Create creates a new tunnel for the account.
//
// API reference: https://api.cloudflare.com/#argo-tunnel-create-argo-tunnel
func (s *tunnels) Create(ctx context.Context, name string) (*Tunnel, error) {
	s.client.logger.V(1).Info("Creating tunnel", "name", name)
	tunnel := &Tunnel{}
	secret := make([]byte, 32)
	_, _ = rand.Read(secret)
	body := struct {
		Name   string `json:"name"`
		Secret []byte `json:"tunnel_secret"`
	}{
		Name:   name,
		Secret: secret,
	}

	err := NewRequest(s.client).
		Verb(http.MethodPost).
		AccountPrefix(s.client.accountID).
		Resource("tunnels").
		Header("Accept", "application/json;version=1").
		Body(body).
		Do(ctx).
		Into(tunnel)
	return tunnel, err
}

// Delete removes a tunnel.
//
// API reference: https://api.cloudflare.com/#argo-tunnel-delete-argo-tunnel
func (s *tunnels) Delete(ctx context.Context, tunnelID uuid.UUID) error {
	s.client.logger.V(1).Info("Deleting tunnel", "tunnel-id", tunnelID.String())
	return NewRequest(s.client).
		Verb(http.MethodDelete).
		AccountPrefix(s.client.accountID).
		Resource("tunnels").
		ResourceID(tunnelID.String()).
		Header("Accept", "application/json;version=1").
		Do(ctx).
		Error()
}

func (s *tunnels) Route(ctx context.Context, tunnelID uuid.UUID, route TunnelRoute) error {
	s.client.logger.V(1).Info("Routing tunnel", "tunnel-id", tunnelID.String(), "type", route.Type())
	return NewRequest(s.client).
		Verb(http.MethodPut).
		ZonePrefix(s.client.zoneID).
		Resource("tunnels").
		ResourceID(tunnelID.String()).
		SubPath("routes").
		Header("Accept", "application/json;version=1").
		Body(route).
		Do(ctx).
		Error()
}

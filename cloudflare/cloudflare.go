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
	"errors"
	"net/http"
	"net/url"

	"github.com/cloudflare/cloudflared/certutil"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	DefaultBaseURL = "https://api.cloudflare.com/client/v4"
	APITokenEnv    = "CF_API_TOKEN"
)

var (
	ErrNotFound                 = errors.New("not found")
	ErrUnexpectedMultipleTunnel = errors.New("unexpected have multiple tunnel with same name")
)

type Client interface {
	AccountID() string
	ZoneID() string

	Tunnels() TunnelClient
	Zones() ZoneClient
}

type client struct {
	baseURL *url.URL
	client  *http.Client
	logger  logr.Logger

	accountID string
	zoneID    string
}

type ClientOption func(c *client) error

// WithAPIToken use oauth2 client with given token as static token source
//
// NOTE: When used, this option should be called before any options to avoid
// overwrite existing http.Client
func WithAPIToken(token string) ClientOption {
	return func(c *client) error {
		ctx := context.Background()
		c.client = oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
		return nil
	}
}

// WithOriginCert use origincert to populate accountID and zoneID.
// This options also wrap the http.DefaultTransport to inject
// X-Auth-User-Service-Key header
func WithOriginCert(origincert []byte) ClientOption {
	return func(c *client) error {
		cert, err := certutil.DecodeOriginCert(origincert)
		if err != nil {
			return err
		}

		client := c.client
		c.zoneID = cert.ZoneID
		c.accountID = cert.AccountID
		c.client = &http.Client{
			Transport: &ServiceKeyTransport{
				Base:       client.Transport,
				ServiceKey: cert.ServiceKey,
			},
		}

		return nil
	}
}

func WithLogger(logger logr.Logger) ClientOption {
	return func(c *client) error {
		c.logger = logger.WithName("cloudflare-client")
		return nil
	}
}

func NewClient(opts ...ClientOption) (Client, error) {
	baseURL, err := url.Parse(DefaultBaseURL)
	if err != nil {
		return nil, err
	}

	c := &client{
		client:  http.DefaultClient,
		baseURL: baseURL,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.logger == nil {
		c.logger = zapr.NewLogger(zap.NewNop())
	}

	return c, nil
}

func (c *client) AccountID() string {
	return c.accountID
}

func (c *client) ZoneID() string {
	return c.zoneID
}

func (c *client) Tunnels() TunnelClient {
	return newTunnels(c)
}

func (c *client) Zones() ZoneClient {
	return newZones(c)
}

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
	"net/http"
)

type ZoneClient interface {
	Get(ctx context.Context, zoneID string) (*Zone, error)
}

type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type zones struct {
	client *client
}

func newZones(c *client) *zones {
	return &zones{
		client: c,
	}
}

// Get fetch a zone.
//
// API reference: https://api.cloudflare.com/#zone-zone-details
func (s *zones) Get(ctx context.Context, zoneID string) (*Zone, error) {
	s.client.logger.V(1).Info("Getting zone details", "zone-id", zoneID)
	zone := &Zone{}
	err := NewRequest(s.client).
		Verb(http.MethodGet).
		Resource("zones").
		ResourceID(zoneID).
		Do(ctx).
		Into(zone)
	return zone, err
}

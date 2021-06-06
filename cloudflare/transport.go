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
	"net/http"
)

type ServiceKeyTransport struct {
	Base       http.RoundTripper
	ServiceKey string
}

func (t *ServiceKeyTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	bodyIsClosed := false
	if r.Body != nil {
		defer func() {
			if !bodyIsClosed {
				r.Body.Close()
			}
		}()
	}

	rr := r.Clone(r.Context())
	rr.Header.Add("X-Auth-User-Service-Key", t.ServiceKey)
	if t.Base == nil {
		t.Base = http.DefaultTransport
	}

	bodyIsClosed = true
	return t.Base.RoundTrip(rr)
}

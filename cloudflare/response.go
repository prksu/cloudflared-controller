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
	"encoding/json"
	"fmt"
)

type Error struct {
	Code    json.Number `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
}

func (e Error) Error() string {
	return fmt.Sprintf("code: %v, reason: %s", e.Code, e.Message)
}

type ResultInfoCursors struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

type ResultInfo struct {
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	TotalPages int               `json:"total_pages"`
	Count      int               `json:"count"`
	Total      int               `json:"total_count"`
	Cursor     string            `json:"cursor"`
	Cursors    ResultInfoCursors `json:"cursors"`
}

type Response struct {
	Success    bool            `json:"success,omitempty"`
	Errors     []Error         `json:"errors,omitempty"`
	Messages   []string        `json:"messages,omitempty"`
	Result     json.RawMessage `json:"result,omitempty"`
	ResultInfo ResultInfo      `json:"result_info,omitempty"`
}

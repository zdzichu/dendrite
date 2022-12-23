// Copyright 2022 The Matrix.org Foundation C.I.C.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package inthttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/matrix-org/dendrite/internal/caching"
	"github.com/matrix-org/dendrite/internal/httputil"
	"github.com/matrix-org/dendrite/relayapi/api"
)

// HTTP paths for the internal HTTP API
const (
	RelayAPIPerformRelayServerSyncPath = "/relayapi/performRelayServerSync"
	RelayAPIPerformStoreAsyncPath      = "/relayapi/performStoreAsync"
	RelayAPIQueryAsyncTransactionsPath = "/relayapi/queryAsyncTransactions"
)

// NewRelayAPIClient creates a RelayInternalAPI implemented by talking to a HTTP POST API.
// If httpClient is nil an error is returned
func NewRelayAPIClient(relayapiURL string, httpClient *http.Client, cache caching.ServerKeyCache) (api.RelayInternalAPI, error) {
	if httpClient == nil {
		return nil, errors.New("NewRelayInternalAPIHTTP: httpClient is <nil>")
	}
	return &httpRelayInternalAPI{
		relayAPIURL: relayapiURL,
		httpClient:  httpClient,
		cache:       cache,
	}, nil
}

type httpRelayInternalAPI struct {
	relayAPIURL string
	httpClient  *http.Client
	cache       caching.ServerKeyCache
}

func (h *httpRelayInternalAPI) PerformRelayServerSync(
	ctx context.Context,
	request *api.PerformRelayServerSyncRequest,
	response *api.PerformRelayServerSyncResponse,
) error {
	return httputil.CallInternalRPCAPI(
		"PerformRelayServerSync", h.relayAPIURL+RelayAPIPerformRelayServerSyncPath,
		h.httpClient, ctx, request, response,
	)
}

func (h *httpRelayInternalAPI) PerformStoreAsync(
	ctx context.Context,
	request *api.PerformStoreAsyncRequest,
	response *api.PerformStoreAsyncResponse,
) error {
	return httputil.CallInternalRPCAPI(
		"PerformStoreAsync", h.relayAPIURL+RelayAPIPerformStoreAsyncPath,
		h.httpClient, ctx, request, response,
	)
}

func (h *httpRelayInternalAPI) QueryAsyncTransactions(
	ctx context.Context,
	request *api.QueryAsyncTransactionsRequest,
	response *api.QueryAsyncTransactionsResponse,
) error {
	return httputil.CallInternalRPCAPI(
		"QueryAsyncTransactions", h.relayAPIURL+RelayAPIQueryAsyncTransactionsPath,
		h.httpClient, ctx, request, response,
	)
}
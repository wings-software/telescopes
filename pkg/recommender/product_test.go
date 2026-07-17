// Copyright © 2019 Banzai Cloud
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

package recommender

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/banzaicloud/telescopes/.gen/cloudinfo"
	"github.com/go-openapi/runtime"
	"github.com/goph/emperror"
	"github.com/goph/logur"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCloudInfoClient(baseURL string) CloudInfoSource {
	return NewCloudInfoClient(baseURL, logur.NewTestLogger())
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func TestAvg(t *testing.T) {
	assert.Equal(t, 0.0, avg(nil))
	assert.Equal(t, 0.0, avg([]cloudinfo.ZonePrice{}))
	assert.Equal(t, 1.5, avg([]cloudinfo.ZonePrice{
		{Price: 1.0},
		{Price: 2.0},
	}))
}

func TestDiscriminateErrCtx(t *testing.T) {
	apiErr := &runtime.APIError{Code: http.StatusNotFound, OperationName: "GetRegion"}
	tagged := discriminateErrCtx(apiErr)
	assert.Contains(t, emperror.Context(tagged), cloudInfoService)

	connErr := errors.New("connection refused")
	tagged = discriminateErrCtx(connErr)
	assert.Contains(t, emperror.Context(tagged), cloudInfoClientComponent)
}

func TestCloudInfoClient_SuccessPaths(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/providers/google/services/gke/regions/us-central1/products", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, cloudinfo.ProductDetailsResponse{
			Products: []cloudinfo.ProductDetails{{
				Category:        "general",
				Series:          "n1",
				Type:            "n1-standard-1",
				OnDemandPrice:   0.05,
				CpusPerVm:       1,
				MemPerVm:        3.75,
				GpusPerVm:       0,
				Burst:           false,
				NtwPerf:         "moderate",
				NtwPerfCategory: "low",
				CurrentGen:      true,
				Zones:           []string{"us-central1-a"},
				SpotPrice: []cloudinfo.ZonePrice{
					{Price: 0.01},
					{Price: 0.03},
				},
			}},
		})
	})
	mux.HandleFunc("/providers/google", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, cloudinfo.ProviderResponse{
			Provider: cloudinfo.Provider{Provider: "google"},
		})
	})
	mux.HandleFunc("/providers/google/services/gke", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, cloudinfo.ServiceResponse{
			Service: cloudinfo.Service{Service: "gke"},
		})
	})
	mux.HandleFunc("/providers/google/services/gke/regions/us-central1", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, cloudinfo.GetRegionResp{
			Id:    "us-central1",
			Name:  "US Central",
			Zones: []string{"us-central1-a", "us-central1-b"},
		})
	})
	mux.HandleFunc("/providers/google/services/gke/regions", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []cloudinfo.Region{{Id: "us-central1", Name: "US Central"}})
	})
	mux.HandleFunc("/providers/google/services/gke/continents", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []cloudinfo.Continent{{
			Name:    "North America",
			Regions: []cloudinfo.Region{{Id: "us-central1", Name: "US Central"}},
		}})
	})
	mux.HandleFunc("/continents", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []string{"North America", "Europe"})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cli := newTestCloudInfoClient(server.URL)

	vms, err := cli.GetProductDetails("google", "gke", "us-central1")
	require.NoError(t, err)
	require.Len(t, vms, 1)
	assert.Equal(t, "n1-standard-1", vms[0].Type)
	assert.Equal(t, 0.02, vms[0].AvgPrice)

	provider, err := cli.GetProvider("google")
	require.NoError(t, err)
	assert.Equal(t, "google", provider)

	service, err := cli.GetService("google", "gke")
	require.NoError(t, err)
	assert.Equal(t, "gke", service)

	region, err := cli.GetRegion("google", "gke", "us-central1")
	require.NoError(t, err)
	assert.Equal(t, "US Central", region)

	zones, err := cli.GetZones("google", "gke", "us-central1")
	require.NoError(t, err)
	assert.Equal(t, []string{"us-central1-a", "us-central1-b"}, zones)

	regions, err := cli.GetRegions("google", "gke")
	require.NoError(t, err)
	require.Len(t, regions, 1)
	assert.Equal(t, "us-central1", regions[0].Id)

	continents, err := cli.GetContinentsData("google", "gke")
	require.NoError(t, err)
	require.Len(t, continents, 1)
	assert.Equal(t, "North America", continents[0].Name)

	continentNames, err := cli.GetContinents()
	require.NoError(t, err)
	assert.Equal(t, []string{"North America", "Europe"}, continentNames)
}

func TestCloudInfoClient_ErrorPaths(t *testing.T) {
	mux := http.NewServeMux()
	notFound := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	}
	mux.HandleFunc("/providers/google/services/gke/regions/us-central1/products", notFound)
	mux.HandleFunc("/providers/google", notFound)
	mux.HandleFunc("/providers/google/services/gke", notFound)
	mux.HandleFunc("/providers/google/services/gke/regions/us-central1", notFound)
	mux.HandleFunc("/providers/google/services/gke/regions", notFound)
	mux.HandleFunc("/providers/google/services/gke/continents", notFound)
	mux.HandleFunc("/continents", notFound)

	server := httptest.NewServer(mux)
	defer server.Close()

	cli := newTestCloudInfoClient(server.URL)

	_, err := cli.GetProductDetails("google", "gke", "us-central1")
	require.Error(t, err)

	_, err = cli.GetProvider("google")
	require.Error(t, err)

	_, err = cli.GetService("google", "gke")
	require.Error(t, err)

	_, err = cli.GetRegion("google", "gke", "us-central1")
	require.Error(t, err)

	_, err = cli.GetZones("google", "gke", "us-central1")
	require.Error(t, err)

	_, err = cli.GetRegions("google", "gke")
	require.Error(t, err)

	_, err = cli.GetContinentsData("google", "gke")
	require.Error(t, err)

	_, err = cli.GetContinents()
	require.Error(t, err)
}

func TestCloudInfoClient_ConnectivityError(t *testing.T) {
	cli := newTestCloudInfoClient("http://127.0.0.1:1")

	_, err := cli.GetProvider("google")
	require.Error(t, err)
	assert.Contains(t, emperror.Context(err), cloudInfoClientComponent)
}

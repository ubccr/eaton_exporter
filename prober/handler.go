// Copyright 2024 Andrew E. Bruno
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prober

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ubccr/eaton_exporter/client"
)

type Prober interface {
	GetPath() string
	Register(*prometheus.Registry)
	Handler(*slog.Logger)
}

func Handler(w http.ResponseWriter, r *http.Request, logger *slog.Logger, params url.Values) {
	if params == nil {
		params = r.URL.Query()
	}

	target := params.Get("target")
	if target == "" {
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		return
	}

	modules := params.Get("module")
	if modules == "" {
		modules = "input"
	}

	probers := make([]Prober, 0)

	for _, moduleName := range strings.Split(modules, ",") {
		switch moduleName {
		case "input":
			probers = append(probers, &InputProber{})
		default:
			http.Error(w, fmt.Sprintf("Unknown module %q", moduleName), http.StatusBadRequest)
			logger.Debug("Unknown module", "module", moduleName)
			return
		}
	}

	handle, err := client.GetHandle(target)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to target %q", target), http.StatusBadRequest)
		logger.Error("Failed to connect to target", "target", target, "err", err)
		return
	}

	registry := prometheus.NewRegistry()

	for _, p := range probers {
		p.Register(registry)
		handle.AddEndpoint(p)
	}

	if err := handle.Fetch(); err != nil {
		http.Error(w, "Failed to fetch eaton endpoint", http.StatusInternalServerError)
		logger.Error("Failed to fetch eaton endpoint", "err", err)
		return
	}

	for _, p := range probers {
		p.Handler(logger)
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

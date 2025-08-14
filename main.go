// Copyright 2025 Andrew E. Bruno
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

package main

import (
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/ubccr/eaton_exporter/client"
	"github.com/ubccr/eaton_exporter/prober"
)

const (
	eatonEndpoint   = "/eaton"
	metricsEndpoint = "/metrics"
)

var (
	configFile     = kingpin.Flag("config.file", "Eaton exporter config file").Default("/etc/prometheus/eaton.conf").String()
	logLevelProber = kingpin.Flag("log.prober", "Log level for probe request logs. One of: [debug, info, warn, error]. Defaults to debug").Default("debug").String()
	listenAddress  = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9795").String()
)

func main() {
	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print("eaton_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promslog.New(promslogConfig)
	probeLogLevel := promslog.NewLevel()
	if err := probeLogLevel.Set(*logLevelProber); err != nil {
		logger.Warn("Error setting log prober level, log prober level unchanged", "err", err, "current_level", probeLogLevel.String())
	}

	logger.Info("Starting eaton_exporter", "version", version.Info())
	logger.Info("Build context", "build_context", version.BuildContext())
	logger.Info("Starting Server", "address", *listenAddress)
	logger.Info("Using config file", "path", *configFile)

	err := client.LoadConfig(*configFile)
	if err != nil {
		logger.Error("Failed to parse config", "err", err)
		os.Exit(1)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Eaton Exporter</title></head>
             <body>
             <h1>Eaton Exporter</h1>
             <p><a href='` + eatonEndpoint + `'>Eaton Metrics</a></p>
             <p><a href='` + metricsEndpoint + `'>Exporter Metrics</a></p>
             </body>
             </html>`))
	})
	http.HandleFunc(eatonEndpoint, func(w http.ResponseWriter, r *http.Request) {
		prober.Handler(w, r, logger, nil)
	})
	http.Handle(metricsEndpoint, promhttp.Handler())

	err = http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		logger.Error("err", err)
		os.Exit(1)
	}
}

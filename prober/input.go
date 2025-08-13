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
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

type phaseEndpoint struct {
	Path string `json:"@id"`
}

type inputMeasures struct {
	ActivePower float64 `json:"activePower"`
}

type inputStatus struct {
	Operating string `json:"operating"`
	Health    string `json:"health"`
}

type InputProber struct {
	Measures      *inputMeasures `json:"measures"`
	Status        *inputStatus   `json:"status"`
	PhaseEndpoint *phaseEndpoint `json:"phases"`

	powerGuage  prometheus.Gauge
	statusGuage *prometheus.GaugeVec
}

func (p *InputProber) GetPath() string {
	return "/powerDistributions/1/inputs/1"
}

func (p *InputProber) Register(registry *prometheus.Registry) {
	p.powerGuage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "eaton_pdu_active_power",
		Help: "PDU active power (W)",
	})

	registry.MustRegister(p.powerGuage)

	p.statusGuage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "eaton_pdu_input_status",
		Help: "PDU input status",
	}, []string{"operating"})

	registry.MustRegister(p.statusGuage)
}

func (p *InputProber) Handler(logger *slog.Logger) {
	p.powerGuage.Set(p.Measures.ActivePower)

	if p.Status.Health == "ok" {
		p.statusGuage.WithLabelValues(p.Status.Operating).Set(1)
	} else {
		p.statusGuage.WithLabelValues(p.Status.Operating).Set(0)
	}
}

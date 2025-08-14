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

package prober

import (
	"encoding/json"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ubccr/eaton_exporter/client"
)

const (
	InputsEndpoint = client.BasePath + "/powerDistributions/1/inputs/1"
)

type Phase struct {
	Measures struct {
		PercentLoad float64 `json:"percentLoad"`
		Current     float64 `json:"current"`
		VoltageLL   float64 `json:"voltageLL"`
	} `json:"measures"`
	Identification struct {
		PhysicalName string `json:"physicalName"`
	} `json:"Identification"`
}

type InputProber struct {
	Measures struct {
		ActivePower float64 `json:"activePower"`
	} `json:"measures"`
	Status struct {
		Operating string `json:"operating"`
		Health    string `json:"health"`
	} `json:"status"`
	PhaseEndpoint struct {
		Path string `json:"@id"`
	} `json:"phases"`

	phases         []*Phase
	powerGauge     prometheus.Gauge
	statusGauge    *prometheus.GaugeVec
	phaseVoltGauge *prometheus.GaugeVec
	phaseAmpGauge  *prometheus.GaugeVec
	phaseLoadGauge *prometheus.GaugeVec
	ec             *client.Client
}

func (p *InputProber) Register(registry *prometheus.Registry) {
	p.powerGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "eaton_pdu_active_power",
		Help: "PDU active power (W)",
	})

	registry.MustRegister(p.powerGauge)

	p.statusGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "eaton_pdu_input_status",
		Help: "PDU input status",
	}, []string{"operating"})

	registry.MustRegister(p.statusGauge)

	p.phaseLoadGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "eaton_pdu_phase_percent_load",
		Help: "PDU phase percent load (%)",
	}, []string{"phase"})

	registry.MustRegister(p.phaseLoadGauge)

	p.phaseVoltGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "eaton_pdu_phase_voltage_ll",
		Help: "PDU phase voltageLL (V)",
	}, []string{"phase"})

	registry.MustRegister(p.phaseVoltGauge)

	p.phaseAmpGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "eaton_pdu_phase_current",
		Help: "PDU phase current (A)",
	}, []string{"phase"})

	registry.MustRegister(p.phaseAmpGauge)
}

func (p *InputProber) Fetch(logger *slog.Logger) error {
	rawJson, err := p.ec.FetchEndpoint(InputsEndpoint)
	if err != nil {
		return err
	}

	err = json.Unmarshal(rawJson, p)
	if err != nil {
		return err
	}

	if p.PhaseEndpoint.Path == "" {
		return nil
	}

	p.phases = make([]*Phase, 0)

	rawJson, err = p.ec.FetchEndpoint(p.PhaseEndpoint.Path)
	if err != nil {
		return err
	}

	var root client.EndpointRoot

	err = json.Unmarshal(rawJson, &root)
	if err != nil {
		return err
	}

	for _, m := range root.Members {
		rawJson, err := p.ec.FetchEndpoint(m.ID)
		if err != nil {
			return err
		}

		var phase Phase

		err = json.Unmarshal(rawJson, &phase)
		if err != nil {
			return err
		}

		p.phases = append(p.phases, &phase)
	}

	return nil
}

func (p *InputProber) Handler(logger *slog.Logger) {
	p.powerGauge.Set(p.Measures.ActivePower)

	if p.Status.Health == "ok" {
		p.statusGauge.WithLabelValues(p.Status.Operating).Set(1)
	} else {
		p.statusGauge.WithLabelValues(p.Status.Operating).Set(0)
	}

	for _, phase := range p.phases {
		p.phaseVoltGauge.WithLabelValues(phase.Identification.PhysicalName).Set(phase.Measures.VoltageLL)
		p.phaseAmpGauge.WithLabelValues(phase.Identification.PhysicalName).Set(phase.Measures.Current)
		p.phaseLoadGauge.WithLabelValues(phase.Identification.PhysicalName).Set(phase.Measures.PercentLoad)
	}
}

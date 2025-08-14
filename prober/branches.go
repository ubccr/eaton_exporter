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
	BranchesEndpoint = client.BasePath + "/powerDistributions/1/branches"
)

type Branch struct {
	Measures struct {
		Current     float64 `json:"current"`
		PercentLoad float64 `json:"percentLoad"`
		Voltage     float64 `json:"voltageLL"`
	} `json:"measures"`
	Identification struct {
		PhysicalName string `json:"physicalName"`
	} `json:"Identification"`
	Status struct {
		Operating      string `json:"operating"`
		Health         string `json:"health"`
		BreakerTripped bool   `json:"breakerTripped"`
	} `json:"status"`
}

type BranchProber struct {
	branches     []*Branch
	statusGuage  *prometheus.GaugeVec
	breakerGuage *prometheus.GaugeVec
	voltGuage    *prometheus.GaugeVec
	ampGuage     *prometheus.GaugeVec
	loadGuage    *prometheus.GaugeVec
	ec           *client.Client
}

func (b *BranchProber) Register(registry *prometheus.Registry) {
	b.statusGuage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "eaton_pdu_branch_status",
		Help: "PDU branch status",
	}, []string{"branch", "operating"})

	registry.MustRegister(b.statusGuage)

	b.loadGuage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "eaton_pdu_branch_percent_load",
		Help: "PDU branch percent load (%)",
	}, []string{"branch"})

	registry.MustRegister(b.loadGuage)

	b.voltGuage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "eaton_pdu_branch_voltage",
		Help: "PDU branch voltage (V)",
	}, []string{"branch"})

	registry.MustRegister(b.voltGuage)

	b.ampGuage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "eaton_pdu_branch_current",
		Help: "PDU branch current (A)",
	}, []string{"branch"})

	registry.MustRegister(b.ampGuage)

	b.breakerGuage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "eaton_pdu_branch_breaker_tripped",
		Help: "PDU branch breaker tripped",
	}, []string{"branch"})

	registry.MustRegister(b.breakerGuage)
}

func (b *BranchProber) Fetch(logger *slog.Logger) error {
	rawJson, err := b.ec.FetchEndpoint(BranchesEndpoint)
	if err != nil {
		return err
	}

	var root client.EndpointRoot

	err = json.Unmarshal(rawJson, &root)
	if err != nil {
		return err
	}

	for _, m := range root.Members {
		rawJson, err := b.ec.FetchEndpoint(m.ID)
		if err != nil {
			return err
		}

		var branch Branch

		err = json.Unmarshal(rawJson, &branch)
		if err != nil {
			return err
		}

		b.branches = append(b.branches, &branch)
	}

	return nil
}

func (b *BranchProber) Handler(logger *slog.Logger) {

	for _, branch := range b.branches {
		b.loadGuage.WithLabelValues(branch.Identification.PhysicalName).Set(branch.Measures.PercentLoad)
		b.voltGuage.WithLabelValues(branch.Identification.PhysicalName).Set(branch.Measures.Voltage)
		b.ampGuage.WithLabelValues(branch.Identification.PhysicalName).Set(branch.Measures.Current)
		if branch.Status.Health == "ok" {
			b.statusGuage.WithLabelValues(branch.Identification.PhysicalName, branch.Status.Operating).Set(1)
		} else {
			b.statusGuage.WithLabelValues(branch.Identification.PhysicalName, branch.Status.Operating).Set(0)
		}
		if branch.Status.BreakerTripped {
			b.breakerGuage.WithLabelValues(branch.Identification.PhysicalName).Set(1)
		} else {
			b.breakerGuage.WithLabelValues(branch.Identification.PhysicalName).Set(0)
		}
	}
}

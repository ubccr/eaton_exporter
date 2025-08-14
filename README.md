# Eaton PDU Exporter

Prometheus exporter for Eaton PDUs. This exporter is intended to query multiple
Eaton PDUs rom an external host.

The `/eaton` metrics endpoint exposes the Eaton metrics and requires a `target`
parameter.  The `module` parameter can also be used to select which probe
commands to run, the default modules are `inputs,branches`. Available modules are:

- inputs
- branches

The `/metrics` endpoint exposes Go and process metrics for this exporter.

## Configuration

This exporter requires a eaton.conf file. Example config:

```ini
[connection:pdu1.example.com]
host=pdu1.example.com
username=admin
password=admin

[connection:pdu2.example.com]
host=pdu2.example.com
username=admin
password=admin
```

## Prometheus configs

```yaml
- job_name: eaton
  metrics_path: /eaton
  static_configs:
  - targets:
    - pdu1.example.com
    - pdu2.example.com
    labels:
      module: inputs
  - targets:
    - pdu3.example.com
    labels:
      module: input
  relabel_configs:
  - source_labels: [__address__]
    target_label: __param_target
  - source_labels: [__param_target]
    target_label: instance
  - source_labels: [module]
    target_label: __param_module
  - target_label: __address__
    replacement: 127.0.0.1:9465
```

Example systemd unit file [here](systemd/eaton_exporter.service)

## Sample Metrics

```
eaton_pdu_active_power 5520
eaton_pdu_branch_breaker_tripped{branch="A"} 0
eaton_pdu_branch_breaker_tripped{branch="B"} 0
eaton_pdu_branch_breaker_tripped{branch="C"} 0
eaton_pdu_branch_breaker_tripped{branch="D"} 0
eaton_pdu_branch_breaker_tripped{branch="E"} 0
eaton_pdu_branch_breaker_tripped{branch="F"} 0
eaton_pdu_branch_current{branch="A"} 3.22
eaton_pdu_branch_current{branch="B"} 1.623
eaton_pdu_branch_current{branch="C"} 2.037
eaton_pdu_branch_current{branch="D"} 8.482
eaton_pdu_branch_current{branch="E"} 0
eaton_pdu_branch_current{branch="F"} 11.937
eaton_pdu_branch_percent_load{branch="A"} 16.1
eaton_pdu_branch_percent_load{branch="B"} 8.115
eaton_pdu_branch_percent_load{branch="C"} 10.185
eaton_pdu_branch_percent_load{branch="D"} 42.41
eaton_pdu_branch_percent_load{branch="E"} 0
eaton_pdu_branch_percent_load{branch="F"} 59.685
eaton_pdu_branch_status{branch="A",operating="in service"} 1
eaton_pdu_branch_status{branch="B",operating="in service"} 1
eaton_pdu_branch_status{branch="C",operating="in service"} 1
eaton_pdu_branch_status{branch="D",operating="in service"} 1
eaton_pdu_branch_status{branch="E",operating="in service"} 1
eaton_pdu_branch_status{branch="F",operating="in service"} 1
eaton_pdu_branch_voltage{branch="A"} 0
eaton_pdu_branch_voltage{branch="B"} 0
eaton_pdu_branch_voltage{branch="C"} 0
eaton_pdu_branch_voltage{branch="D"} 0
eaton_pdu_branch_voltage{branch="E"} 0
eaton_pdu_branch_voltage{branch="F"} 0
eaton_pdu_input_status{operating="in service"} 1
eaton_pdu_phase_current{phase="L1"} 22.17
eaton_pdu_phase_current{phase="L2"} 12.501
eaton_pdu_phase_current{phase="L3"} 14.869
eaton_pdu_phase_percent_load{phase="L1"} 36.95
eaton_pdu_phase_percent_load{phase="L2"} 20.865
eaton_pdu_phase_percent_load{phase="L3"} 24.782
eaton_pdu_phase_voltage_ll{phase="L1"} 203.65
eaton_pdu_phase_voltage_ll{phase="L2"} 204.95
eaton_pdu_phase_voltage_ll{phase="L3"} 203.91
```

## License

eaton_exporter is released under the Apache License Version 2.0. See the LICENSE file.

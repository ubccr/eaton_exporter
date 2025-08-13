# Eaton PDU Exporter

Prometheus exporter for Eaton PDUs. This exporter
is intended to query multiple Eaton PDUs rom an external host.

The `/eaton` metrics endpoint exposes the Eaton metrics and requires a `target`
parameter.  The `module` parameter can also be used to select which probe
commands to run, the default module is `input`. Available modules are:

- input

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
      module: input
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
# HELP eaton_pdu_active_power PDU active power (W)
# TYPE eaton_pdu_active_power gauge
eaton_pdu_active_power 5520
# HELP eaton_pdu_input_status PDU input status
# TYPE eaton_pdu_input_status gauge
eaton_pdu_input_status{operating="in service"} 1
```

## License

eaton_exporter is released under the Apache License Version 2.0. See the LICENSE file.

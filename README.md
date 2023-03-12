# Prometheus Speedtest Exporter

This is a prometheus speedtest exporter written purely in [Golang][golang]. It uses the default port for the speedtest exporter `9516`.

## Usage

```bash
./prometheus-speedtest-exporter
```

### Prometheus configuration

```yaml
scrape_configs:
  - job_name: speedtest
    metrics_path: /metrics
    scrape_interval: 5m
    scrape_timeout: 30s
    static_configs:
      - targets:
          - localhost:9516
```

## Related projects

- [jeanralphaviles/prometheus_speedtest (Python)](https://github.com/jeanralphaviles/prometheus_speedtest)
- [billimek/prometheus-speedtest-exporter (Shell)](https://github.com/billimek/prometheus-speedtest-exporter)

## License

This project is licensed under the terms of the [MIT license](./LICENSE.md).

[golang]: https://go.dev/

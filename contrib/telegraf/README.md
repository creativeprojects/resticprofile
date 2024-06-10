# backup stats in InfluxDB via Telegraf

Telegraf can ingest stats produced by resticprofile in two different ways:

1. Configure [http_listener_v2](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/http_listener_v2) expecting Prometheus data, have resticprofile push to it via `prometheus-push`. You need telegraf version [1.31 or higher](https://github.com/influxdata/telegraf/issues/15453). You also need to explicitly configure urls on telegraf side, like so:

```toml
[[inputs.http_listener_v2]]
  paths = [
    "/metrics/job/profilename.backup",
    "/metrics/job/anotherprofile.backup"
  ]
  service_address = ":9999"
  http_success_code = 200
  data_format = "prometheus"
```

2. Configure telegraf to parse resticprofile's JSON status file. This works only if telegraf can access the file that resticprofile produces. It can also break if the status file format changes.

```toml
[[inputs.file]]
  files = ["/var/cache/restic/status.json"]
  data_format = "json_v2"
  [[inputs.file.json_v2]]
    measurement_name = "restic"
    [[inputs.file.json_v2.object]]
      path = "{backup_profile:profiles.@keys,stats:profiles.@values}.@group"
      disable_prepend_keys = true
      timestamp_key = "stats_backup_time"
      timestamp_format = "rfc3339nano"
      tags = ["backup_profile"]
```

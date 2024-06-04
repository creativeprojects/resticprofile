# backup stats in InfluxDB via Telegraf

Telegraf can ingest stats produced by resticprofile in three different ways:

1. Configure Telegraf to parse resticprofile's JSON status file. The parsing directive looks brittle, but as of June 2024, this seems to be the most viable method.

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

1. Configure [http_listener_v2](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/http_listener_v2) expecting Prometheus data, have resticprofile push to it via `prometheus-push`. This is currently slightly borked, due to [this bug](https://github.com/influxdata/telegraf/issues/15453). The data gets written, but resticprofile emits a warning. Once that bug is fixed, it might be a viable way to go. The downside is that you need to explicitly configure urls on telegraf side, like so:

```toml
[[inputs.http_listener_v2]]
  paths = ["/metrics/job/home.backup"]
  service_address = "127.0.0.1:9999"
  data_format = "prometheus"
```

1. Configure resticprofile to write prometheus metrics to a file via `prometheus-save-to-file`, have telegraf monitor the file. This is quite brittle, as the generated metrics file gets updated once per backup job run - which means if you run a backup for a group, some of the stats will linger for just a handful of seconds, while the last set of stats will persist untill the next backup run.

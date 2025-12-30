---
title: "Using ntfy"
weight: 10
---

This configuration executes the following:
- Makes a template profile called "default" [(learn more)]({{% relref "/configuration/variables/templates/" %}})
- Creates a `backup` on a schedule every day at 12am [(learn more)]({{% relref "/configuration/schedules/configuration/" %}})
    - After a fail or success, sends a message through the `ntfy` service using a POST request [(learn more)]({{% relref "/configuration/hooks/http_hooks/" %}})
- Creates a `forget` schedule that runs every day at 12:05am
    - Prunes all non-matching entries [(learn more)]({{% relref "/reference/profile/forget/" %}})
    - Keeps every daily backup for the past week (7 backups), every weekly backup for the past month (4 backups), and a monthly backup for 75 years.

{{< tabs groupid="example-ntfy" >}}
{{% tab title="toml" %}}
```toml
default:
  insecure-no-password: true
  backup: 
    schedule: "*-*-* 00:00:00"
    schedule-permission: user_logged_on
    send-after:
      method: POST
      url: [[YOUR_URL_HERE]]
      headers:
        - name: Title
          value: "{{.Profile.Name}} ran successfully!"
        - name: Priority
          value: "low"
    send-after-fail:
      method: POST
      url: [[YOUR_URL_HERE]]
      headers:
        - name: Title
          value: "{{.Profile.Name}} failed!"
        - name: Tags
          value: "warning"
        - name: Priority
          value: "high"
  forget:
    schedule: "*-*-* 00:05:00"
    schedule-permission: user_logged_on
    keep-within-daily: '7d'
    keep-within-weekly: '1m'
    keep-within-monthly: '75y'
    prune: true
```
{{% /tab %}}
{{< /tabs >}}

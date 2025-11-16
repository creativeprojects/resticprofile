---
title: "Use stdin in configuration"
---

Simple example sending a file via stdin

{{< tabs groupid="config-with-hcl" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[stdin]
  repository = "local:/backup/restic"
  password-file = "key"

  [stdin.backup]
    stdin = true
    stdin-filename = "stdin-test"
    tag = [ 'stdin' ]
  
[mysql]
  inherit = "stdin"

  [mysql.backup]
    stdin-command = [ 'mysqldump --all-databases --order-by-primary' ]
    stdin-filename = "dump.sql"
    tag = [ 'mysql' ]

```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

stdin:
  repository: "local:/backup/restic"
  password-file: key
  backup:
    stdin: true
    stdin-filename: stdin-test
    tag:
      - stdin

mysql:
  inherit: stdin
  backup:
    stdin-command: "mysqldump --all-databases --order-by-primary"
    stdin-filename: "dump.sql"
    tag:
      - mysql

```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
# sending stream through stdin
stdin {
    repository = "local:/backup/restic"
    password-file = "key"

    backup = {
        stdin = true
        stdin-filename = "stdin-test"
        tag = [ "stdin" ]
    }
}

mysql {
  inherit = "stdin"

  backup = {
    stdin-command = [ "mysqldump --all-databases --order-by-primary" ]
    stdin-filename = "dump.sql"
    tag = [ "mysql" ]
  }
}
```

{{% /tab %}}
{{< /tabs >}}

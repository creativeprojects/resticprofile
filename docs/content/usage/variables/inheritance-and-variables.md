---
title: "Using inheritance and variables"
---
You can use a combination of inheritance and variables in the resticprofile configuration file like so:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[generic]
  password-file = "{{ .ConfigDir }}/{{ .Profile.Name }}-key"
  repository = "/backup/{{ .Now.Weekday }}"
  lock = "$HOME/resticprofile-profile-{{ .Profile.Name }}.lock"
  initialize = true

  [generic.backup]
    check-before = true
    exclude = [ "/**/.git" ]
    exclude-caches = true
    one-file-system = false
    run-after = "echo All Done!"
    run-before = [
        "echo Hello {{ .Env.LOGNAME }}",
        "echo current dir: {{ .CurrentDir }}",
        "echo config dir: {{ .ConfigDir }}",
        "echo profile started at {{ .Now.Format "02 Jan 06 15:04 MST" }}"
    ]
    tag = [ "{{ .Profile.Name }}", "dev" ]

  [generic.retention]
    after-backup = true
    before-backup = false
    compact = false
    keep-within = "30d"
    prune = true
    tag = [ "{{ .Profile.Name }}", "dev" ]

  [generic.snapshots]
    tag = [ "{{ .Profile.Name }}", "dev" ]

[src]
  inherit = "generic"

  [src.backup]
    source = [ "{{ .Env.HOME }}/go/src" ]
  
  [src.check]
    # Weekday is an integer from 0 to 6 (starting from Sunday)
    # Nice trick to add 1 to an integer: https://stackoverflow.com/a/72465098
    read-data-subset = "{{ len (printf "a%*s" .Now.Weekday "") }}/7"

```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
---
version: "1"

generic:
    password-file: "{{ .ConfigDir }}/{{ .Profile.Name }}-key"
    repository: "/backup/{{ .Now.Weekday }}"
    lock: "$HOME/resticprofile-profile-{{ .Profile.Name }}.lock"
    initialize: true

    backup:
        check-before: true
        exclude:
        - /**/.git
        exclude-caches: true
        one-file-system: false
        run-after: echo All Done!
        run-before:
          - "echo Hello {{ .Env.LOGNAME }}"
          - "echo current dir: {{ .CurrentDir }}"
          - "echo config dir: {{ .ConfigDir }}"
          - "echo profile started at {{ .Now.Format "02 Jan 06 15:04 MST" }}"
        tag:
          - "{{ .Profile.Name }}"
          - dev

    retention:
        after-backup: true
        before-backup: false
        compact: false
        keep-within: 30d
        prune: true
        tag:
          - "{{ .Profile.Name }}"
          - dev

    snapshots:
        tag:
          - "{{ .Profile.Name }}"
          - dev

src:
    inherit: generic

    backup:
        source:
          - "{{ .Env.HOME }}/go/src"

    check:
        # Weekday is an integer from 0 to 6 (starting from Sunday)
        # Nice trick to add 1 to an integer: https://stackoverflow.com/a/72465098
        read-data-subset: "{{ len (printf "a%*s" .Now.Weekday "") }}/7"

```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"generic" = {
  "password-file" = "{{ .ConfigDir }}/{{ .Profile.Name }}-key"
  "repository" = "/backup/{{ .Now.Weekday }}"
  "lock" = "$HOME/resticprofile-profile-{{ .Profile.Name }}.lock"
  "initialize" = true

  "backup" = {
    "check-before" = true
    "exclude" = ["/**/.git"]
    "exclude-caches" = true
    "one-file-system" = false
    "run-after" = "echo All Done!"
    "run-before" = ["echo Hello {{ .Env.LOGNAME }}", "echo current dir: {{ .CurrentDir }}", "echo config dir: {{ .ConfigDir }}", "echo profile started at {{ .Now.Format "02 Jan 06 15:04 MST" }}"]
    "tag" = ["{{ .Profile.Name }}", "dev"]
  }

  "retention" = {
    "after-backup" = true
    "before-backup" = false
    "compact" = false
    "keep-within" = "30d"
    "prune" = true
    "tag" = ["{{ .Profile.Name }}", "dev"]
  }

  "snapshots" = {
    "tag" = ["{{ .Profile.Name }}", "dev"]
  }
}

"src" = {
  "inherit" = "generic"

  "backup" = {
    "source" = ["{{ .Env.HOME }}/go/src"]
  }

  "check" = {
    # Weekday is an integer from 0 to 6 (starting from Sunday)
    # Nice trick to add 1 to an integer: https://stackoverflow.com/a/72465098
    "read-data-subset" = "{{ len (printf "a%*s" .Now.Weekday "") }}/7"
  }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "generic": {
    "password-file": "{{ .ConfigDir }}/{{ .Profile.Name }}-key",
    "repository": "/backup/{{ .Now.Weekday }}",
    "lock": "$HOME/resticprofile-profile-{{ .Profile.Name }}.lock",
    "initialize": true,
    "backup": {
      "check-before": true,
      "exclude": [
        "/**/.git"
      ],
      "exclude-caches": true,
      "one-file-system": false,
      "run-after": "echo All Done!",
      "run-before": [
        "echo Hello {{ .Env.LOGNAME }}",
        "echo current dir: {{ .CurrentDir }}",
        "echo config dir: {{ .ConfigDir }}",
        "echo profile started at {{ .Now.Format "02 Jan 06 15:04 MST" }}"
      ],
      "tag": [
        "{{ .Profile.Name }}",
        "dev"
      ]
    },
    "retention": {
      "after-backup": true,
      "before-backup": false,
      "compact": false,
      "keep-within": "30d",
      "prune": true,
      "tag": [
        "{{ .Profile.Name }}",
        "dev"
      ]
    },
    "snapshots": {
      "tag": [
        "{{ .Profile.Name }}",
        "dev"
      ]
    }
  },
  "src": {
    "inherit": "generic",
    "backup": {
      "source": [
        "{{ .Env.HOME }}/go/src"
      ]
    },
    "check": {
      "read-data-subset": "{{ len (printf "a%*s" .Now.Weekday "") }}/7"
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}

This is obviously not a real world example, but it shows many of the possibilities you can do with variable expansion.

To check the generated configuration, you can use the resticprofile `show` command:

```shell
% resticprofile -c examples/template.yaml -n src show

global:
    default-command:          snapshots
    restic-lock-retry-after:  1m0s
    restic-stale-lock-age:    2h0m0s
    min-memory:               100
    send-timeout:             30s

profile src:
    repository:     /backup/Monday
    password-file:  /Users/CP/go/src/resticprofile/examples/src-key
    initialize:     true
    lock:           /Users/CP/resticprofile-profile-src.lock

    backup:
        check-before:    true
        run-before:      echo Hello CP
                         echo current dir: /Users/CP/go/src/resticprofile
                         echo config dir: /Users/CP/go/src/resticprofile/examples
                         echo profile started at 05 Sep 22 17:39 BST
        run-after:       echo All Done!
        source:          /Users/CP/go/src
        exclude:         /**/.git
        exclude-caches:  true
        tag:             src
                         dev

    retention:
        after-backup:  true
        keep-within:   30d
        path:          /Users/CP/go/src
        prune:         true
        tag:           src
                       dev

    check:
        read-data-subset:  2/7

    snapshots:
        tag:  src
              dev
```

As you can see, the `src` profile inherited from the `generic` profile. The tags `{{ .Profile.Name }}` got replaced by the name of the current profile `src`.
Now you can reuse the same generic configuration in another profile.

You might have noticed the `read-data-subset` in the `check` section which will read a seventh of the data every day, meaning the whole repository data will be checked over a week. You can find [more information about this trick](https://stackoverflow.com/a/72465098).

More examples can be found [here]({{% relref "/usage/variables/list-variables" %}}).

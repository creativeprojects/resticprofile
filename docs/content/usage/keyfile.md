---
title: "Generate a Keyfile"
weight: 10
---

## Generating random keys

Resticprofile includes a tool to generate cryptographically secure, base64-encoded random keys. Save the key to a file and use it as a strong key for Restic.

- On Linux, FreeBSD, Dragonfly, and Solaris, Reader uses `getrandom(2)`.
- On legacy Linux (< 3.17), it uses `/dev/urandom`.
- On macOS, and OpenBSD Reader, uses `arc4random_buf(3)`.
- On NetBSD, Reader uses the kern.arandom sysctl.
- On Windows, Reader uses the ProcessPrng API.

[Reference from the Go cryto library documentation](https://golang.org/pkg/crypto/rand/#pkg-variables)

```shell
resticprofile generate --random-key
```

Generates a 1024-byte random key (converted to 1368 Base64 characters) and displays it in the console.

To generate a key of a different size, specify the byte length in the command line.

```shell
resticprofile generate --random-key 2048
```

---
title: "Generate a Keyfile"
weight: 10
---

## Generating random keys

resticprofile has a handy tool to generate cryptographically secure random keys encoded in base64. You can simply put this key into a file and use it as a strong key for restic.

- On Linux and FreeBSD, the generator uses `getrandom(2)` if available, `/dev/urandom` otherwise.
- On OpenBSD and macOS, the generator uses `getentropy(2)`.
- On other Unix-like systems, the generator reads from `/dev/urandom`.
- On Windows systems, the generator uses the uses the RtlGenRandom API.
- On JS/Wasm, the generator uses the Web Crypto API.
- On WASIP1/Wasm, the generator uses `random_get` from `wasi_snapshot_preview1`. 

[Reference from the Go documentation](https://golang.org/pkg/crypto/rand/#pkg-variables)

```shell
resticprofile generate --random-key
```

generates a 1024 bytes random key (converted into 1368 base64 characters) and displays it on the console

To generate a different size of key, you can specify the bytes length on the command line:

```shell
resticprofile generate --random-key 2048
```

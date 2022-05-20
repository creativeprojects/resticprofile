---
title: "Generate a Keyfile"
date: 2022-05-16T20:21:16+01:00
weight: 10
---

## Generating random keys

resticprofile has a handy tool to generate cryptographically secure random keys encoded in base64. You can simply put this key into a file and use it as a strong key for restic.

- On Linux and FreeBSD, the generator uses `getrandom(2)` if available, `/dev/urandom` otherwise.
- On OpenBSD, the generator uses `getentropy(2)`.
- On other Unix-like systems, the generator reads from `/dev/urandom`.
- On Windows systems, the generator uses the CryptGenRandom API.
- On Wasm, the generator uses the Web Crypto API.

[Reference from the Go documentation](https://golang.org/pkg/crypto/rand/#pkg-variables)

```
$ resticprofile generate --random-key
```

generates a 1024 bytes random key (converted into 1368 base64 characters) and displays it on the console

To generate a different size of key, you can specify the bytes length on the command line:

```
$ resticprofile generate --random-key 2048
```

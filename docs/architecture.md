# capsule-cli — Architecture

How the binary is put together, what crosses the wire, and why the server can never read a
capsule.

## Runtime topology

```
capsule seal                                      Capsule server
  │                                                     │
  ├─ crypto.GenerateKey()      32 random bytes          │
  ├─ crypto.Encrypt()          AES-256-GCM              │
  │                                                     │
  ├─ api.CreatePaste() ───────────────────────────▶ POST /api/pastes
  │      { content: base64(iv‖ciphertext), … }          │
  │                                                     ▼
  │   ◀─────────────────────────────────────── { id, delete_token }
  │
  └─ prints  https://<server>/<id>#<key>
                                    └── fragment, never sent by any HTTP client


capsule reveal <url>
  │
  ├─ cmd.parseURL()            server ‖ id ‖ fragment
  ├─ api.GetContent() ────────────────────────────▶ POST /api/pastes/<id>/content
  │   ◀─────────────────────────────────────── { content }
  └─ crypto.Decrypt()          plaintext to stdout
```

The server holds ciphertext, an id, and a delete token. It never receives the key, because
the key lives after the `#` in the URL and fragments are not transmitted in HTTP requests.

## Components

| Package | Responsibility |
|---|---|
| `main` | Calls `cmd.Execute()` and nothing else |
| `cmd` | Cobra command tree, flag parsing, terminal I/O, `parseURL` |
| `internal/api` | HTTP client for the Capsule server, 30-second timeout |
| `internal/config` | Loads and saves `~/.capsule.yml` |
| `internal/crypto` | Key generation, AES-256-GCM, PBKDF2 key wrapping, base64url helpers |

`cmd` is the only package that talks to the user; `internal/*` never prints.

## Cryptography

Content encryption, in `internal/crypto/crypto.go`:

- A fresh 256-bit key per capsule, from `crypto/rand`
- AES-256-GCM with a fresh 12-byte IV, also from `crypto/rand`
- The payload sent to the server is `base64.StdEncoding(iv ‖ ciphertext)`; decryption
  splits the first 12 bytes back off as the IV

Key transport, unprotected case: the raw key is `base64.RawURLEncoding`-encoded and becomes
the URL fragment.

Key transport, password case (`seal --password`):

1. 32 random salt bytes, 12 random IV bytes
2. `PBKDF2(password, salt, 600000, 32, SHA-256)` produces the wrapping key
3. The content key is sealed with AES-GCM under the wrapping key
4. The fragment becomes `base64url(encryptedKey).base64url(salt).base64url(iv)`

`reveal` decides which case it is with `crypto.IsPasswordProtected`, which simply tests the
fragment for a `.`. That is the entire protocol signal — there is no flag and no server
round trip to find out. A wrong password fails at `gcm.Open` and is reported as
`wrong password or corrupted key`.

## URL parsing

`cmd.parseURL` accepts any absolute URL and derives three things:

- `ServerURL` — `scheme://host`, so `reveal` and `revoke` follow the link you were given
  rather than the configured server
- `ID` — the last non-empty path segment
- `Fragment` — everything after `#`

A URL without a scheme or host is rejected, as is one whose path ends up empty. `reveal`
additionally rejects a URL with no fragment, since there would be no key to decrypt with.

## Server API consumed

| Method | Path | Used by | Notes |
|---|---|---|---|
| `POST` | `/api/pastes` | `seal` | JSON body, accepts `200` or `201` |
| `POST` | `/api/pastes/{id}/content` | `reveal` | No request body, returns `{ content }` |
| `DELETE` | `/api/pastes/{id}` | `revoke` | Requires the `X-Delete-Token` header |
| `GET` | `/api/pastes/{id}` | — | Implemented as `api.Client.GetMetadata`, not wired to any command |

The create request carries `content`, `burn_after_read`, `expires_in`, `has_password`, and
an optional `syntax`. `has_password` is advisory metadata for the web client; the CLI does
not rely on the server to enforce it. Any non-success status is surfaced with the server's
own response body appended.

## Authentication

There is none. A capsule is protected by capability, not identity: whoever holds the URL
fragment can decrypt, whoever holds the delete token can revoke. Nothing in the repo reads
a credential, a token file, or an OIDC issuer, so this CLI does not participate in the
suite's Authentik SSO. It also emits no `pool` events.

## State on disk

The single state file is `~/.capsule.yml`, written with mode `0644`, holding one key. See
[configuration.md](configuration.md). Capsules themselves are never cached locally.

## Versioning

`cmd.version` defaults to `dev` and is replaced at build time by GoReleaser through
`-X github.com/FacileStudio/capsule-cli/cmd.version={{.Version}}`. A binary built with a
plain `go build` therefore reports `dev` from `capsule --version`.

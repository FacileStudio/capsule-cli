# capsule-cli

Terminal client for Capsule, the zero-knowledge encrypted paste service. All encryption
happens on your machine; the server only ever stores ciphertext.

The binary is named `capsule`. It talks to any Capsule-compatible server and defaults to
`https://capsule.facile.studio`.

## What it does

- Encrypts content client-side with AES-256-GCM and uploads only the ciphertext
- Puts the decryption key in the URL fragment, which HTTP clients never send to the server
- Reads content from an argument or from stdin, so it composes with pipes
- Optionally wraps the key with a password using PBKDF2-SHA256 at 600,000 iterations
- Fetches and decrypts a capsule back to stdout
- Revokes a capsule early with the delete token returned at creation time
- Points at a self-hosted server through a single YAML config file

## Stack

| Layer | Tech |
|---|---|
| CLI | Go 1.26, cobra 1.10, `golang.org/x/crypto` (PBKDF2), `golang.org/x/term`, fatih/color |
| Storage | `~/.capsule.yml`, YAML via `gopkg.in/yaml.v3` |
| Deploy | GoReleaser, GitHub Actions on tag push, Homebrew tap `FacileStudio/tap` |

## Install

```sh
brew install FacileStudio/tap/capsule
```

With a Go toolchain:

```sh
go install github.com/FacileStudio/capsule-cli@latest
```

The binary lands in `$(go env GOPATH)/bin/capsule`. The `install.sh` script in this repo
does the same thing through a shallow clone of `main`.

## Usage

```sh
capsule seal "api_key=sk-1234" --expires 1h
echo "secret" | capsule seal --no-burn
capsule reveal https://capsule.facile.studio/cap_abc123#3xK9pQ
capsule revoke https://capsule.facile.studio/cap_abc123 --token 7f2ae4
```

`seal` prints the shareable URL on stdout and the delete token on stderr, so redirecting
stdout captures only the link. `reveal` prints plaintext on stdout and nothing else.

Full command reference: [docs/usage.md](docs/usage.md).

## Configuration

The CLI reads no environment variables. Its only state is `~/.capsule.yml`, created with
defaults the first time any command runs.

| Key | What it does |
|---|---|
| `server_url` | Base URL used by `seal`. Defaults to `https://capsule.facile.studio` |

`reveal` and `revoke` ignore this key and derive the server from the URL you pass them.

Full reference: [docs/configuration.md](docs/configuration.md).

## Structure

```
main.go      Entrypoint, delegates to cmd.Execute()
cmd/         Cobra commands — seal, reveal, revoke, config, plus the URL parser
internal/    api/ (HTTP client), config/ (YAML state), crypto/ (AES-GCM, PBKDF2)
docs/        Architecture, configuration, development, usage
```

## Documentation

| Doc | What's in it |
|---|---|
| [Architecture](docs/architecture.md) | Crypto flow, URL fragment contract, server API calls |
| [Configuration](docs/configuration.md) | The config file, its keys, and defaults |
| [Development](docs/development.md) | Build, run, and release the binary |
| [Usage](docs/usage.md) | Every command and flag, with examples |

---

Part of the [Facile Suite](https://facile.studio) — self-hosted tools for creative studios
and freelancers. One login, zero cloud dependency.

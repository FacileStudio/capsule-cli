# capsule-cli — Development

Local setup, the build, and how a release reaches Homebrew.

## Prerequisites

- Go 1.26.1 or later, as pinned by the `go` directive in `go.mod`
- A reachable Capsule server if you want to exercise `seal` and `reveal` end to end

There is no `mise.toml`, no `Makefile`, no linter config, and no `scripts/check.sh` in this
repo. The toolchain is plain `go`.

## Build and run

```sh
go build -o capsule .
./capsule --help
```

Run without building:

```sh
go run . seal "hello"
```

Install into your Go bin directory:

```sh
go install .
```

A locally built binary reports `dev` for `capsule --version`, because the real version is
injected only by the release build. To reproduce the release string:

```sh
go build -ldflags "-s -w -X github.com/FacileStudio/capsule-cli/cmd.version=0.0.0-local" -o capsule .
```

## Tests

```sh
go test ./...
```

There are currently no test files in the repo, so this passes trivially. If you touch
`internal/crypto` or `internal/api`, add one — the crypto package is pure and easy to test
round-trip: `Encrypt` then `Decrypt` with the same key, and `WrapKey` then `UnwrapKey` with
the same password.

## Adding a command

Commands live in `cmd/` and register themselves in an `init` function:

```go
func init() {
	myCmd.Flags().StringVar(&myFlag, "flag", "", "What it does")
	rootCmd.AddCommand(myCmd)
}
```

Keep user-facing output in `cmd`; `internal/*` returns errors and values, never prints.
Success output goes to stdout, everything advisory to stderr, so the tool stays pipeable.

## Layout

```
main.go                        Entrypoint
cmd/root.go                    Root command and version wiring
cmd/seal.go                    Encrypt and upload
cmd/reveal.go                  Fetch and decrypt
cmd/revoke.go                  Delete with a token
cmd/config.go                  config and config set server
cmd/url.go                     parseURL helper
internal/api/client.go         HTTP client and request/response types
internal/config/config.go      Load, Save, Path
internal/crypto/crypto.go      AES-256-GCM and PBKDF2 wrapping
```

## Release

Releases are tag-driven. Pushing a tag matching `v*` runs
`.github/workflows/release.yml`, which checks out with full history, sets up the stable Go
toolchain, and runs `goreleaser release --clean`.

```sh
git tag v0.2.0
git push --tags
```

GoReleaser (`.goreleaser.yml`) builds `CGO_ENABLED=0` binaries for `linux` and `darwin` on
`amd64` and `arm64`, names the binary `capsule`, packs `tar.gz` archives as
`capsule-cli_<version>_<os>_<arch>`, writes `checksums.txt`, and pushes a formula named
`capsule` to the `FacileStudio/homebrew-tap` repository. The only secret it needs is the
workflow's own `GITHUB_TOKEN`.

To check the config without publishing:

```sh
goreleaser release --snapshot --clean
```

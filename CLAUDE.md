# Capsule CLI

Zero-knowledge encrypted paste sharing from the terminal. All encryption is client-side (AES-256-GCM); the server never sees plaintext or keys.

## Tech Stack

- **Language**: Go 1.26
- **CLI framework**: cobra (spf13/cobra)
- **Crypto**: AES-256-GCM encryption, PBKDF2 key wrapping (600k iterations, SHA-256) for password protection
- **Config**: YAML (`~/.capsule.yml`) via gopkg.in/yaml.v3
- **Release**: GoReleaser + GitHub Actions (tag-triggered), Homebrew tap via FacileStudio/homebrew-tap
- **Dependencies**: fatih/color (terminal colors), golang.org/x/term (password input), golang.org/x/crypto (pbkdf2)

## Key Commands

```bash
# Build
go build -o capsule .

# Run (dev)
go run .

# Install locally
go install .

# Run tests (none exist yet)
go test ./...

# Release (tag-triggered via CI)
git tag v0.x.x && git push --tags
```

There is no Makefile, no linter config, and no test files in the project currently.

## Project Structure

```
.
├── main.go                       # Entrypoint — delegates to cmd.Execute()
├── cmd/
│   ├── root.go                   # Root cobra command, version flag
│   ├── seal.go                   # Encrypt content and upload (supports stdin, password, burn, expiry, syntax)
│   ├── reveal.go                 # Fetch ciphertext and decrypt locally
│   ├── revoke.go                 # Delete a capsule using a delete token
│   ├── config.go                 # Show/set config (server URL)
│   └── url.go                    # URL parser — extracts server, paste ID, and key fragment
├── internal/
│   ├── api/
│   │   └── client.go             # HTTP client for the Capsule server API
│   ├── config/
│   │   └── config.go             # YAML config loader/saver (~/.capsule.yml)
│   └── crypto/
│       └── crypto.go             # AES-256-GCM encrypt/decrypt, PBKDF2 key wrapping
├── .goreleaser.yml               # GoReleaser config (linux/darwin, amd64/arm64)
├── .github/workflows/release.yml # CI: release on tag push
└── install.sh                    # Clone-and-install script
```

## Architecture Notes

- The encryption key is placed in the URL fragment (`#`), which browsers and HTTP clients never send to the server. The server only stores ciphertext.
- Password protection wraps the AES key with PBKDF2-derived key + AES-GCM. The fragment becomes `encryptedKey.salt.iv` (dot-separated base64url). The presence of a `.` in the fragment signals password protection (`IsPasswordProtected`).
- Version is injected at build time via ldflags: `-X github.com/FacileStudio/capsule-cli/cmd.version={{.Version}}`
- Default server: `https://capsule.facile.studio`
- API endpoints: `POST /api/pastes` (create), `GET /api/pastes/:id` (metadata), `POST /api/pastes/:id/content` (fetch ciphertext), `DELETE /api/pastes/:id` (revoke, requires `X-Delete-Token` header)
- Config file path: `~/.capsule.yml` (auto-created with defaults on first run)

## Conventions

- Commit messages use gitmoji prefix style (e.g. `🚀 init:`, `🔧 fix:`)
- No inline comments in code
- No tests yet — add them if you touch crypto or API logic
- Binary artifacts (`capsule`, `capsule-cli`, `dist/`) are gitignored

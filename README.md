# Capsule

Zero-knowledge encrypted paste sharing from the terminal. The server never sees your plaintext — all encryption happens client-side using AES-256-GCM.

## Install

### Go

```bash
go install github.com/FacileStudio/capsule-cli@latest
```

### Homebrew

```bash
brew install FacileStudio/tap/capsule
```

### From source

```bash
curl -fsSL https://raw.githubusercontent.com/FacileStudio/capsule-cli/main/install.sh | bash
```

## Usage

### Seal (encrypt & share)

```bash
# From argument
capsule seal "my secret message"

# From stdin (pipe-friendly)
echo "secret" | capsule seal
cat secret.txt | capsule seal

# With options
capsule seal "api_key=sk-1234" --burn --expires 1h
capsule seal "fn main() {}" --syntax rust --no-burn --expires 7d

# Password-protected
capsule seal "top secret" --password
```

### Reveal (fetch & decrypt)

```bash
# Prints plaintext to stdout
capsule reveal https://capsule.facile.dev/cap_abc123#keyFragment

# Pipe to file
capsule reveal https://capsule.facile.dev/cap_abc123#keyFragment > secret.txt
```

### Revoke (burn a capsule)

```bash
capsule revoke https://capsule.facile.dev/cap_abc123 --token abc123...def456
```

### Config

```bash
# Show current config
capsule config

# Point to a different server
capsule config set server https://my-capsule.example.com
```

## How it works

1. **Seal**: generates a random 256-bit AES key, encrypts your content with AES-256-GCM, uploads the ciphertext, and puts the key in the URL fragment (never sent to the server).
2. **Reveal**: extracts the key from the URL fragment, fetches the ciphertext, decrypts locally.
3. **Password protection** (optional): wraps the AES key with PBKDF2 (600,000 iterations, SHA-256) + AES-GCM. The wrapped key, salt, and IV are encoded in the URL fragment. Decryption requires both the URL and the password.

The server only ever stores ciphertext. No plaintext, no keys.

## Self-hosting

Capsule works with any Capsule-compatible server. Set your server URL:

```bash
capsule config set server https://your-server.example.com
```

## License

MIT

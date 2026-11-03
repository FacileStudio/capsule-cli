# capsule-cli — Usage

Every command, every flag, and what each one actually does. Taken from the cobra tree in
`cmd/`.

## Synopsis

```
capsule [command]
```

| Command | What it does |
|---|---|
| `seal` | Encrypt content and upload it |
| `reveal` | Fetch a capsule and decrypt it |
| `revoke` | Delete a capsule with its delete token |
| `login [url]` | Sign in by approving a link in any browser |
| `logout` | End the session and forget the token |
| `whoami` | Show the signed-in user |
| `keys list` | List registered API keys |
| `keys create` | Create a new API key |
| `keys revoke` | Revoke an API key by ID |
| `config` | Show configuration |
| `config set server` | Change the server URL |
| `help` | Cobra's built-in help |

Global flags: `-h, --help` on every command, `--json` for structured JSON output, `--no-color` to disable colors, and `-v, --version` on the root command.

```sh
capsule --version
```

Prints `capsule version dev` for a locally built binary and the tagged version for a
released one.

## `capsule seal [content]`

Encrypts content client-side and uploads the ciphertext. Prints the shareable URL on
**stdout** and the delete token on **stderr**.

Content comes from the arguments, or from stdin when stdin is not a terminal. Multiple
arguments are joined with a single space. With neither — an interactive terminal and no
argument — the command fails with `provide content as an argument or pipe via stdin`. Empty
content is rejected too.

| Flag | Type | Default | What it does |
|---|---|---|---|
| `--burn` | bool | `true` | Destroy the capsule after the first read |
| `--no-burn` | bool | `false` | Keep the capsule readable until it expires; wins over `--burn` |
| `-e, --expires` | string | `24h` | Expiration window. The flag help lists `1h`, `24h`, `7d`, `30d` |
| `-p, --password` | bool | `false` | Prompt for a password and wrap the key with it |
| `-s, --syntax` | string | — | Syntax highlighting hint stored with the capsule, for the web viewer |

Burn-after-read is on by default, so the plain form is the paranoid form:

```sh
capsule seal "api_key=sk-1234"
```

From a pipe, kept for a week, readable more than once:

```sh
cat deploy-notes.txt | capsule seal --no-burn --expires 7d
```

With a syntax hint:

```sh
capsule seal "fn main() {}" --syntax rust
```

Password-protected. The prompt is written to stderr and the password is read from the
terminal without echo; an empty password is rejected:

```sh
capsule seal "top secret" --password
Password:
https://capsule.facile.studio/cap_abc123#Zm9v.YmFy.YmF6

Delete token: 7f2ae4…
```

Capture only the URL, keeping the token visible in the terminal:

```sh
url=$(capsule seal "deploy key" --expires 1h)
```

## `capsule reveal <url>`

Fetches a capsule and writes the plaintext to stdout with no trailing newline added and no
decoration. Takes exactly one argument and has no flags.

The server is taken from the URL you pass, not from your config, so you can read a capsule
hosted anywhere. The URL must include the `#` fragment; without it the command fails with
`no key fragment found in URL (missing #)`.

```sh
capsule reveal https://capsule.facile.studio/cap_abc123#3xK9pQ
```

To a file:

```sh
capsule reveal https://capsule.facile.studio/cap_abc123#3xK9pQ > secret.txt
```

If the fragment contains a `.`, the capsule is password-protected and you are prompted
before the fetch:

```sh
capsule reveal https://capsule.facile.studio/cap_abc123#Zm9v.YmFy.YmF6
Password:
```

A wrong password reports `wrong password or corrupted key`. A burned or expired capsule
fails at the fetch step with the server's status and message.

## `capsule revoke <url>`

Deletes a capsule before it expires. Takes exactly one argument.

| Flag | Type | Default | What it does |
|---|---|---|---|
| `--token` | string | — | The delete token printed by `seal`. Required |

The fragment is not needed here — revoking does not decrypt anything — so the bare link is
enough.

```sh
capsule revoke https://capsule.facile.studio/cap_abc123 --token 7f2ae4
Capsule burned.
```

Omitting `--token` fails at flag parsing, since cobra marks it required.

## `capsule config`

Prints the current configuration and the file it was read from. No arguments, no flags.

```sh
capsule config
server_url: https://capsule.facile.studio

Config file: /Users/you/.capsule.yml
```

Running it on a machine with no config file creates `~/.capsule.yml` with the default
server URL, then prints it.

## `capsule config set server <url>`

Sets the server URL used by `seal` and writes it back to `~/.capsule.yml`.

```sh
capsule config set server https://capsule.example.com
Server URL set to https://capsule.example.com
```

`capsule config set` on its own has no action and prints help for its subcommands.
`server_url` is the only key the CLI knows how to set.

## `capsule login [url]`

Signs in by approving a link in any browser. The server defaults to
`https://capsule.facile.studio`; an argument, `CAPSULE_SERVER_URL`, `CAPSULE_URL` or a stored
value take precedence in that order. The CLI prints an approval link through the OAuth device
grant and waits until you approve it, on this machine or another one. Nothing listens on a port
and nothing is typed, so it works over SSH.

```sh
capsule login
capsule login https://capsule.example.com --listener
echo "$TOKEN" | capsule login --token-stdin
```

A server without device login falls back to the loopback flow with a one-line notice.
`--listener` forces that flow: a browser on this machine redirects to a local port.
`--no-browser` prints the URL instead of opening it. `--no-listener` shows a code on the page
for the command to read from stdin. `--token` and `--token-stdin` store an existing token after
checking it. If the server says the account is not linked, sign in once at the server in a
browser, then run `capsule login` again.

## `capsule logout` and `capsule whoami`

`logout` ends the session on the server and clears the stored token. `whoami` prints the
signed-in email, or exits non-zero when not logged in.

## `capsule keys list`

Lists registered API keys, optionally filtered by application name.

```sh
capsule keys list
capsule keys list --app webapp
capsule keys list --json
```

| Flag | Type | Default | What it does |
|---|---|---|---|
| `--app` | string | — | Filter keys by application name |

## `capsule keys create --app <name>`

Creates a new API key. Prints the raw token on stdout (or as JSON if `--json` is specified).

```sh
capsule keys create --app webapp
capsule keys create --app webapp --public --origins "https://example.com,https://app.example.com" --quota 1000
```

| Flag | Type | Default | What it does |
|---|---|---|---|
| `--app` | string | — | Application name (required) |
| `--public` | bool | `false` | Create a public browser key instead of a secret key |
| `--origins` | string | — | Comma-separated allowed origins for public keys |
| `--quota` | int | `0` | Daily event quota limit for public keys (0 means unlimited) |

## `capsule keys revoke <id>`

Revokes an API key by its numeric ID. Prompts for confirmation when run interactively unless `--yes` is provided.

```sh
capsule keys revoke 42
capsule keys revoke 42 --yes
capsule keys revoke 42 --json
```

| Flag | Type | Default | What it does |
|---|---|---|---|
| `-y, --yes` | bool | `false` | Confirm revocation without prompting |

## Authentication

Paste operations (`seal`, `reveal`, `revoke`) require no user login. Holding the URL fragment grants read access; holding the delete token grants deletion rights.

API key management commands (`keys list`, `keys create`, `keys revoke`) authenticate using the token stored in `~/.capsule.yml` or via the `CAPSULE_TOKEN` environment variable.

## Config location

`~/.capsule.yml`. Details and keys in [configuration.md](configuration.md).

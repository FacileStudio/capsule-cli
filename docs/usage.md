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
| `config` | Show configuration |
| `config set server` | Change the server URL |
| `help` | Cobra's built-in help |

Global flags: `-h, --help` on every command, and `-v, --version` on the root command.

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

## Authentication

There is no login, no token file, and no SSO. Access to a capsule is proved by holding its
URL fragment; the right to delete it early is proved by holding its delete token. Neither
is stored on disk by this CLI, so treat the `seal` output as the only copy.

## Config location

`~/.capsule.yml`. Details and keys in [configuration.md](configuration.md).

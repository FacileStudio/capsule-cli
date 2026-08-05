# capsule-cli — Configuration

Everything the CLI reads from outside its own arguments, which is one YAML file and no
environment variables at all.

## Environment variables

None. The source contains no `os.Getenv` call. Anything you want to change is either a
command flag or the config file below.

## Config file

| Path | `~/.capsule.yml` |
|---|---|
| Format | YAML |
| Permissions | `0644`, set by the CLI when it writes the file |
| Created | Automatically, on the first command that loads config |

Resolution is `os.UserHomeDir()` joined with `.capsule.yml`. There is no `--config` flag, no
`XDG_CONFIG_HOME` support, and no per-project override.

## Keys

| Key | Required | Default | What it does |
|---|---|---|---|
| `server_url` | no | `https://capsule.facile.studio` | Base URL that `seal` uploads to |

A complete config file:

```yaml
server_url: https://capsule.facile.studio
```

Unknown keys are ignored by the YAML decoder. An empty or missing `server_url` falls back
to the default, so a truncated file degrades to stock behavior rather than an error. A file
that is present but malformed YAML is a hard failure: the command exits with
`loading config: parsing config: …`.

## Which commands read it

| Command | Uses `server_url` |
|---|---|
| `seal` | yes — the upload target, with trailing slashes trimmed |
| `config` | yes — prints it, and the resolved config path |
| `config set server` | yes — rewrites it |
| `reveal` | no — the server is taken from the URL argument |
| `revoke` | no — same |

This split is deliberate. Sealing needs a default target; revealing already knows where the
capsule lives because the link says so. It also means you can read a capsule from someone
else's server without touching your config.

## Pointing at a self-hosted server

```sh
capsule config set server https://capsule.example.com
capsule config
```

The second command prints the stored value and the file path it came from. To go back to
the hosted instance, set it again to `https://capsule.facile.studio`, or delete
`~/.capsule.yml` and let the next run recreate it with defaults.

## What is not stored

No tokens, no history, no cached capsules, no key material. Delete tokens are printed to
stderr once at `seal` time and never persisted — if you lose one, the capsule can only be
removed by its expiry or by a burn-after-read.

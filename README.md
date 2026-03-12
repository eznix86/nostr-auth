# Nostr Auth

![Demo GIF](./art/demo.gif)

Nostr Auth lets you keep your identity while providing a simple way to manage access to your platforms. It makes your identity portable across web applications, so you can use Nostr-based authentication with existing services — without introducing unnecessary complexity.

## Why Nostr Auth?

Existing authentication solutions are powerful and widely adopted in the web2 world, but they are not always a natural fit for Nostr-based users. Nostr Auth sits in front of your existing web applications and adds a Nostr-based authentication layer, so you can protect resources while preserving your users' portable identity.

## Use Cases

- Adding Nostr authentication to existing web applications
- Protecting internal or private resources
- Enabling identity-based access control for Nostr users

> [!NOTE]
> This version only supports cookie-based authentication, meaning it works within a single primary domain and its subdomains (e.g. `example.com` and `app.example.com`). Cross-domain authentication is not currently supported (`anotherdomain.com`).

## Requirements

- [Taskfile](https://taskfile.dev/)
- Go 1.26+
- Overmind (`go install github.com/air-verse/air@latest`)
- Air (`go install github.com/air-verse/air@latest`)
- Golangci-lint (`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`)
- [bun](https://bun.sh/)
- Docker and Docker Compose (to try demo)

## Running Locally

- Fork
- `bun install`
- `go install`
- `cp config.json.example config.json`
- Edit the `config.json`
- `cp .env.example .env`
- Edit the `.env`
- `task dev`

Should be available on http://localhost:3000

## Config

`config.json` controls both access rules and branding.

```json
{
  "auth": {
    "enabled": true,
    "groups": {
      "admins": [
        "alice@example.com",
        "npub1..."
      ]
    },
    "apps": {
      "default": {
        "config": {
          "domains": [
            "app.example.com"
          ]
        },
        "users": [
          "group:admins"
        ]
      }
    }
  },
  "branding": {
    "background": {
      "source": {
        "type": "preset",
        "variant": "canyon-falls"
      }
    }
  }
}
```

- `auth.enabled` turns authorization on or off
- `auth.groups` defines reusable user groups with NIP-05 identifiers, `npub`, or nested `group:<name>` references
- `auth.apps` defines which domains are protected and which users or groups can access them
- `branding.background.source.type` currently supports `preset`
- `branding.background.source.variant` can be `canyon-falls` (default), `fields-road`, `mountain-valley`, or `storm-valley`
- if `branding` is omitted, the app falls back to `canyon-falls`

> [!IMPORTANT]
> Prefer `npub` values for real access control. An `npub` identifies a specific public key, so it stays stable over time. A NIP-05 identifier is easier to read and useful in testing or development, but its ownership can change, which may unintentionally change who matches a rule.

## Testing Demo

- `bun install`
- `go install`
- `task demo` — check out the examples [README.md](./examples/proxy/README.md)

## Supports

- Kubernetes Gateway API
- Kubernetes Ingress
- Nginx
- Traefik
- Envoy
- Caddy

## License

[MIT](./LICENSE)

## Credits

- Photo by <a href="https://unsplash.com/@wilstewart3">Wil Stewart</a> on <a href="https://unsplash.com/photos/landscape-photography-of-brown-mountain-pHANr-CpbYM">Unsplash</a>

- Photo by <a href="https://sandeep.ramgolam.com/">Sandeep Ramgolam</a> on <a href="https://sandeep.ramgolam.com/wallpapers">Wallpapers</a>

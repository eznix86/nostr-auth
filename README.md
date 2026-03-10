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
- `cp auth.json.example auth.json`
- Edit the `auth.json`
- `cp .env.example .env`
- Edit the `.env`
- `task dev`

Should be available on http://localhost:3000

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

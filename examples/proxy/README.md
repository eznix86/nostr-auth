# Proxy Demo

This folder keeps examples isolated from the main app.

## What runs where

- `nostr-auth` app: `http://127.0.0.1:3000`
- upstream protected demo: `http://127.0.0.1:8080`
- nginx demo: `http://127.0.0.1:8081`
- Traefik demo: `http://127.0.0.1:8082`
- Caddy demo: `http://127.0.0.1:8083`
- Envoy demo: `http://127.0.0.1:8084`

## Start the demo

```sh
task demo
```

## How it works

- each proxy protects the upstream app on port `8080`
- each proxy calls `http://host.docker.internal:3000/auth/check`
- if you are not logged in, the proxy sends you to `nostr-auth`
- after signing in, `nostr-auth` redirects you back to the original URL
- if you are logged in, the proxy forwards user and group headers to the upstream demo

## Forwarded headers

- user: `Remote-User`, `X-Forwarded-User`, `X-Auth-Request-User`
- email: `X-Auth-Request-Email`
- groups: `X-Forwarded-Groups`, `X-Auth-Request-Groups`
- profile: `X-Auth-Request-Preferred-Username`, `X-Auth-Request-Name`, `X-Auth-Request-Picture`

## Notes

- these examples are intentionally minimal and local-first
- `host.docker.internal` works on macOS and is also mapped through `host-gateway` in the compose file for Linux Docker setups
- nginx uses `auth_request`, so it handles the login redirect itself after a `401`
- Traefik and Caddy pass the auth service response back to the browser directly
- Envoy uses `ext_authz` and can call `/auth/check` with the original request method
- Envoy may call `/auth/check` with the original path appended, so the app accepts both `/auth/check` and `/auth/check/*`
- nginx talks to the auth service over `host.docker.internal`, but browser redirects go to `http://localhost:3000`

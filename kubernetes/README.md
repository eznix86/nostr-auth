# Kubernetes

- namespace: `nostr-auth`
- ingress host: `nostr-auth.brunobernard.dev`
- workload: `StatefulSet`

## Before deploy

- copy `kubernetes/06-secret.yaml.template` to `kubernetes/06-secret.yaml`
- replace `APP_KEY` with a real secret
- adjust `kubernetes/02-auth-configmap.yaml` for your real auth policy

## Deploy

```sh
task deploy
```

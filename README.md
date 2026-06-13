# rollops-plugin-launchdarkly

A [Rollops](https://github.com/klarlabs-studio/rollops) feature-flag provider
plugin backed by [LaunchDarkly](https://launchdarkly.com/). It drives a flag's
on/off state and its environment-level percentage rollout (fallthrough) to track
a Rollops canary in lockstep — as a rollout steps 10% → 50% → 100%, the flag's
rollout follows.

## How it works

Rollops calls the plugin per progressive step (and/or on promote) with the flag
key, target environment, and current traffic percentage. The plugin PATCHes the
flag with a JSON Patch that sets the environment's `on` state and replaces its
fallthrough with a percentage rollout. LaunchDarkly weights are per-mille, so a
percentage is scaled by 1000 (25% → 25000). The flag's first two variations are
assumed to be the boolean true/false pair (variation 0 = on, 1 = off).

## Configuration

Credentials come from the plugin's own environment, never from the Rollops
target spec:

| Env var                | Required | Default                        | Description            |
|------------------------|----------|--------------------------------|------------------------|
| `LAUNCHDARKLY_API_URL` | no       | `https://app.launchdarkly.com` | Base URL               |
| `LAUNCHDARKLY_TOKEN`   | yes      | —                              | API access token       |
| `LAUNCHDARKLY_PROJECT` | no       | `default`                      | Project key            |

## Install

```sh
rollops plugin install launchdarkly
```

Or build and pin manually with `make build` / `make checksum`, then wire into a
rollout spec:

```yaml
featureFlags:
  plugin: ~/.rollops/plugins/launchdarkly
  sha256: <pin>
  flag: checkout
  environment: production
  when: both
```

## License

MIT

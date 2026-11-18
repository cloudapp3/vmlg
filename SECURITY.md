# Security Policy

## Supported versions

Security fixes are applied to the latest tagged release and the `main` branch.

| Version | Supported |
| --- | --- |
| Latest tagged release (`0.1.x`) | Yes |
| `main` | Yes, for the next release |
| Older releases | No |

## Reporting a vulnerability

Do **not** open a public issue for a suspected security vulnerability (such as command execution, SSRF bypass, amplification bypass, or authentication flaws).

Use [GitHub Private Vulnerability Reporting](https://github.com/cloudapp3/vmlg/security/advisories/new) to report:

- affected version, deployment mode (`standalone`, `hub`, or `agent`), and OS
- clear description of the security impact
- minimal reproduction steps
- any suggested mitigation

Please omit private tokens, real production keys, or unrelated host details.

## Operational deployment notes

- Place public-facing `standalone` or `hub` deployments behind an HTTPS reverse proxy (e.g., Caddy, Nginx, Traefik, Cloudflare) and configure `server.trusted_proxies` ONLY with the actual CIDRs of your reverse proxy so `X-Forwarded-For` cannot be spoofed.
- Never commit `hub.agent_token`, `signing_secret`, or `metrics_token` to version control; prefer environment variables or `*_FILE` Docker secrets.
- Leave `guard.allow_private_targets = false` on public internet deployments so RFC1918, loopback, link-local, and cloud metadata addresses remain blocked.

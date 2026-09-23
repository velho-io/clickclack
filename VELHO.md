# Velho ClickClack deployment

Team chat at **https://chat.velho.io**. This fork tracks upstream
`openclaw/clickclack`. Branch `velho-release` is a release tag plus one Velho
commit (`Dockerfile.railway`, `railway.toml`, this file), and Railway
deploys that branch.

## Where things are

| What | Value |
|---|---|
| Railway | workspace **Velho** → project **velho-chat** → service **clickclack** (env `production`) |
| Data | Railway volume at `/app/data` (SQLite `clickclack.db` + `uploads/`) |
| Backups | Railway volume backups: daily (6-day retention) + weekly (27-day retention) |
| DNS | Cloudflare `chat.velho.io` CNAME → `yd2jun57.up.railway.app` (DNS-only) |
| Health | `GET /readyz` (Railway healthcheck) |
| Workspace | "Velho", slug `clickclack` (keep the slug: GitHub sign-ins join the workspace with this slug) |

Service variables: `CLICKCLACK_PUBLIC_URL=https://chat.velho.io`,
`CLICKCLACK_PASSWORD_AUTH_ENABLED=true`, `CLICKCLACK_DEV_BOOTSTRAP=false`,
`CLICKCLACK_ENVIRONMENT=production`, `CLICKCLACK_WEB_VERSION=<tag>`,
`RAILWAY_RUN_UID=0` (the image runs as a non-root user; a Railway volume
mounts as root), `PORT=8080`.

## Upgrading

```sh
git fetch upstream --tags
git checkout velho-release && git rebase --onto vX.Y.Z <old-tag> velho-release
git push --force-with-lease origin velho-release   # Railway redeploys
railway variables --set CLICKCLACK_WEB_VERSION=vX.Y.Z
```

Migrations run on boot. Read the upstream CHANGELOG before upgrading.

## Admin commands

`railway ssh` splits quoted arguments, so a name like "First Last" breaks
the flags. Upload a script instead:

```sh
cat > s.sh <<'EOF'
U=$(clickclack admin user create --data /app/data --name "First Last" --email first@velho.io --workspace wsp_01m36p6p1dhp5vra1jzhzdsk5w)
echo "$U"
printf '%s' 'TEMP-PASSWORD' | clickclack admin user set-password --data /app/data --user "$U"
EOF
railway ssh --service clickclack "echo $(base64 -i s.sh | tr -d '\n') | base64 -d > /tmp/s.sh && sh /tmp/s.sh; rm -f /tmp/s.sh"
```

The person signs in with email + temporary password and changes it under
Account settings → Password. Do not use `admin magic-link create` for an
existing password user: it creates a separate account.

Hot backup: `clickclack backup --data /app/data --out /app/data/backup-$(date +%F).db`.

## Optional follow-ups

- **GitHub sign-in** (self-service for velho-io org members). Create an
  OAuth app under github.com/organizations/velho-io/settings/applications
  with callback `https://chat.velho.io/api/auth/github/callback`, then set
  `CLICKCLACK_GITHUB_CLIENT_ID`, `CLICKCLACK_GITHUB_CLIENT_SECRET`, and
  `CLICKCLACK_GITHUB_ALLOWED_ORG=velho-io`. GitHub identities are **not**
  linked to existing password accounts by email. Existing users should keep
  password login, or move ownership to their GitHub account first.
- **Phone push**: register an application at pushover.net and set
  `CLICKCLACK_PUSHOVER_API_TOKEN`. Each person then pastes their own
  Pushover user key under Account settings → Notifications.
- **Bots / alerts to #ops-alerts**: `clickclack admin bot create ... --scopes bot:write`
  (see docs/bot-installs.md). The incoming webhook only reads the `text` field.

Known limit: behind Railway's proxy, the server sees the proxy address, not
the real client IP. Per-address login rate limits therefore apply to all
users together.

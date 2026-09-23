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
| Mobile app | https://mobile.velho.io: Android APK download page (service `mobile-download`, published from `~/Projects/velho/clickclack-mobile` with `scripts/publish-download.sh`) |
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

## Sign-in

- **GitHub** (velho-io org only): org-owned OAuth app "Velho Chat", callback
  `https://chat.velho.io/api/auth/github/callback`. Vars:
  `CLICKCLACK_GITHUB_CLIENT_ID`, `CLICKCLACK_GITHUB_CLIENT_SECRET`,
  `CLICKCLACK_GITHUB_ALLOWED_ORG=velho-io`. ClickClack never links GitHub to an
  existing account by email, so the four founding accounts were linked by
  hand: an `identities` row with `provider='github'` and
  `provider_subject=<GitHub numeric user id>` (`gh api users/<login> --jq .id`).
  Do the same *before* a new password user first signs in with GitHub, or
  they end up with two accounts. The pre-change backup is
  `/app/data/backup-pre-github-link.db`.
- **Password**: for people without GitHub (see Admin commands).

## Push notifications

We don't use Pushover because it charges each user per platform. ClickClack
sends Pushover-style requests to our own relay instead: service `velho-push`
in the same Railway project, at https://push.velho.io (repo
`~/Projects/velho/velho-push`, deployed with `railway up`). The relay has its
own volume at `/data` and holds the only app token in `PUSH_APP_TOKENS`, which
matches `CLICKCLACK_PUSHOVER_API_TOKEN` here.
`CLICKCLACK_PUSHOVER_API_URL=https://push.velho.io/1/messages.json` comes from
the merged `velho/pushover-url` branch, which is also an upstream PR candidate.

Phone setup today (free): the free ntfy app, subscribed to a private
random topic on ntfy.sh. Each person's topic and key are in
`~/.config/velho/chat-push-keys.txt`. The person pastes the key in Account
settings → Notifications → Mobile push. Create more with
`curl -s https://push.velho.io/v1/ntfy -H 'Content-Type: application/json' -d '{"topic":"velho-<name>-<random>"}'`.
ntfy.sh topics are public to anyone who knows the name, so keep topics random.
`CLICKCLACK_PUSH_RELAY_URL=https://push.velho.io` publishes the relay at the
public `GET /api/push-relay` endpoint (branch `velho/push-relay-discovery`,
merged). After sign-in, the mobile app reads it and offers push once, so
nobody types a relay address.
The ClickClack mobile app (`~/Projects/velho/clickclack-mobile`) will register
Expo push tokens with the relay on its own.

## Bots / alerts to #ops-alerts

`clickclack admin bot create ... --scopes bot:write` (see docs/bot-installs.md).
The incoming webhook only reads the `text` field.

Known limit: behind Railway's proxy, the server sees the proxy address, not
the real client IP. Per-address login rate limits therefore apply to all
users together.

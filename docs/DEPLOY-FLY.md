# Deploy QuietCord Backend on Fly.io

Minimal cost setup: **one Fly Machine** (API) + **Fly Redis** (Upstash-backed). No VPS.

## Prerequisites

1. [Fly CLI](https://fly.io/docs/hands-on/install-flyctl/) and `fly auth login`
2. GitHub repo cloned: `QuietCord/Backend`
3. A **Discord application** (OAuth2) for cloud login — can reuse the same app as Rich Presence or create a dedicated one

## 1. Create the Fly app

```powershell
cd C:\Projects\QuietCord-Backend
fly apps create quietcord-api
```

If the name `quietcord-api` is taken, edit `app = "..."` in `fly.toml` and use that name everywhere below.

Optional: change `primary_region` in `fly.toml` (e.g. `lax`, `sjc`, `mia`) before deploy.

## 2. Redis (required)

The API stores OAuth secrets and settings backups in Redis. On Fly, use managed Redis:

```powershell
fly redis create
```

Follow prompts (name e.g. `quietcord-redis`, same region as the app). When finished:

```powershell
fly redis status
```

Copy the **Private URL** (often `rediss://...` with TLS). You will set it as `REDIS_URI` (this repo’s env name, not `REDIS_URL`).

## 3. Secrets and config

Generate two peppers (keep them safe; rotating them invalidates existing cloud accounts):

```powershell
# PowerShell — run twice for two different values
-join ((1..64) | ForEach-Object { "{0:x2}" -f (Get-Random -Maximum 256) })
```

Set secrets (replace placeholders):

```powershell
$APP = "quietcord-api"
$BASE = "https://$APP.fly.dev"

fly secrets set `
  REDIS_URI="rediss://default:YOUR_PASSWORD@YOUR_HOST.upstash.io:6379" `
  DISCORD_CLIENT_ID="..." `
  DISCORD_CLIENT_SECRET="..." `
  DISCORD_REDIRECT_URI="$BASE/v1/oauth/callback" `
  ROOT_REDIRECT="https://github.com/QuietCord/Backend" `
  PEPPER_SECRETS="..." `
  PEPPER_SETTINGS="..." `
  -a $APP
```

Optional while testing (only your Discord user ID):

```powershell
fly secrets set ALLOWED_USERS="123456789012345678" -a quietcord-api
```

Optional size cap (default 32MB):

```powershell
fly secrets set SIZE_LIMIT="32000000" -a quietcord-api
```

## 4. Discord Developer Portal

In your app → **OAuth2** → **Redirects**, add:

```text
https://quietcord-api.fly.dev/v1/oauth/callback
```

(Use your real Fly app hostname if different.)

## 5. Deploy

```powershell
fly deploy -a quietcord-api
```

Check health:

```powershell
curl https://quietcord-api.fly.dev/v1/
```

Root `/` should redirect to `ROOT_REDIRECT`.

## 6. Wire the Quiet client

In `Quiet/src/shared/brand.ts`:

```ts
export const CLOUD_API_URL = "https://quietcord-api.fly.dev/";
```

Rebuild/inject PTB. In **Settings → Cloud**, enable integrations; desktop may prompt for a CSP override for `connect-src` the first time.

## Cost notes

- **Machine**: `auto_stop_machines` in `fly.toml` stops the VM when idle (cold start on first request).
- **Redis**: Fly Redis bills via Upstash; free tier is enough for personal / small team use.
- **HTTPS**: `*.fly.dev` is included.

## Troubleshooting

| Symptom | Check |
|--------|--------|
| 401 on settings | OAuth not completed; peppers changed; wrong `REDIS_URI` |
| OAuth redirect error | `DISCORD_REDIRECT_URI` must match portal exactly |
| App won’t start | `fly logs -a quietcord-api` |
| Redis TLS errors | Use full `rediss://` URL from `fly redis status` |

## Local dev (unchanged)

```powershell
copy .env.example .env
# fill .env, then:
docker compose up -d
```

Use `REDIS_URI=redis:6379` in `.env` for Compose, not the Upstash URL.

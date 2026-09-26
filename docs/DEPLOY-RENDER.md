# QuietCord API on Render (+ Upstash Redis)

Stack: **Render Web Service** (Dockerfile) + **Upstash Redis** (free). No Fly, no VPS.

Public URL (default service name): `https://quietcord-api.onrender.com`

## 1. Redis (Upstash)

1. [console.upstash.com](https://console.upstash.com) → sign up (free).
2. **Create database** → region close to you (e.g. US-West).
3. Copy the **Redis URL** (TLS), looks like `rediss://default:xxxx@xxxx.upstash.io:6379`.
4. In Render you will set env var **`REDIS_URI`** to that exact string.

## 2. Peppers (once)

PowerShell:

```powershell
function New-Pepper { -join ((1..64) | ForEach-Object { "{0:x2}" -f (Get-Random -Maximum 256) }) }
New-Pepper  # → PEPPER_SECRETS
New-Pepper  # → PEPPER_SETTINGS
```

Save both; rotating them wipes existing cloud accounts.

## 3. Render

1. [dashboard.render.com](https://dashboard.render.com) → sign up (GitHub login is easiest).
2. **New → Blueprint**.
3. Connect repo **`QuietCord/Backend`**, branch `main` (uses `render.yaml` in root).
4. After the service is created, open **quietcord-api → Environment** and add:

| Key | Value |
|-----|--------|
| `REDIS_URI` | Upstash `rediss://...` |
| `DISCORD_CLIENT_ID` | e.g. `1553407390182019282` |
| `DISCORD_CLIENT_SECRET` | Discord Developer Portal → OAuth2 |
| `DISCORD_REDIRECT_URI` | `https://quietcord-api.onrender.com/v1/oauth/callback` |
| `PEPPER_SECRETS` | hex from step 2 |
| `PEPPER_SETTINGS` | hex from step 2 |
| `ROOT_REDIRECT` | `https://github.com/QuietCord/Backend` (AGPL — not Vencord/Vencloud) |

Optional: `ALLOWED_USERS=your_discord_user_id` (comma-separated) to lock the instance.

5. **Manual Deploy** (or push to `main`) and wait for build.

Health: `https://quietcord-api.onrender.com/v1/` should return 200.

> **Free tier:** the service sleeps after ~15 min idle; first request may take ~30–60s (cold start).

### Without Blueprint

**New → Web Service** → repo `QuietCord/Backend`, **Language: Docker**, same env vars, **Health Check Path:** `/v1/`.

## 4. Discord Developer Portal

OAuth2 → **Redirects** → add:

```text
https://quietcord-api.onrender.com/v1/oauth/callback
```

(Change hostname if you picked another service name on Render.)

## 5. Quiet client

In `Quiet/src/shared/brand.ts`:

```ts
export const CLOUD_API_URL = "https://quietcord-api.onrender.com/";
```

Rebuild / `sync:ptb`. Settings → Cloud → enable integrations.

## Local dev

Unchanged: `.env` + `docker compose up` with `REDIS_URI=redis:6379`.

## Troubleshooting

| Issue | Fix |
|-------|-----|
| Build fails | Render logs; ensure Dockerfile builds locally |
| 401 on API | Complete OAuth in client; check peppers / Redis |
| OAuth redirect mismatch | `DISCORD_REDIRECT_URI` must match portal exactly |
| Slow first request | Free tier cold start; normal |

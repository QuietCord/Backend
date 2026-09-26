# QuietCord API on SnapDeploy (+ Upstash Redis)

Deploy: **[snapdeploy.dev/containers](https://snapdeploy.dev/containers)** → GitHub → **`QuietCord/Backend`** (Dockerfile).

Public URL: `https://<your-app>.snapdeploy.app`

## Redis (Upstash) — not SnapDeploy’s Redis form

SnapDeploy’s **managed Redis / DB pack** is optional. For **Upstash**:

1. Do **not** paste Upstash into a “SnapDeploy Redis” slot if it rejects external URLs.
2. Open your container → **Environment variables** (edit on running container is supported).
3. Add a **custom** variable (name exactly):

| Key | Value |
|-----|--------|
| `REDIS_URI` | Full TLS URL from Upstash |

Upstash console → your DB → copy **Redis URL**, then use **`rediss://`** (double s):

```text
rediss://default:YOUR_TOKEN@noted-weevil-302224.upstash.io:6379
```

Alternative name (same value): `REDIS_URL` — the backend accepts either.

## Required env vars (SnapDeploy)

| Key | Example |
|-----|---------|
| `HOST` | `0.0.0.0` |
| `PORT` | `8080` (or whatever SnapDeploy maps; they often inject `PORT`) |
| `REDIS_URI` | Upstash `rediss://...` |
| `DISCORD_CLIENT_ID` | `1553407390182019282` |
| `DISCORD_CLIENT_SECRET` | Discord portal OAuth2 |
| `DISCORD_REDIRECT_URI` | `https://YOUR-APP.snapdeploy.app/v1/oauth/callback` |
| `ROOT_REDIRECT` | `https://github.com/QuietCord/Backend` |
| `PEPPER_SECRETS` | 128 hex chars (`scripts/new-peppers.ps1`) |
| `PEPPER_SETTINGS` | 128 hex chars |

Optional: `ALLOWED_USERS=your_discord_user_id`

## Discord

OAuth2 redirect must match **`DISCORD_REDIRECT_URI`** exactly.

## Quiet client

`CLOUD_API_URL` in `Quiet/src/shared/brand.ts`:

```ts
export const CLOUD_API_URL = "https://YOUR-APP.snapdeploy.app/";
```

Rebuild + `sync:ptb`.

## Health check

`https://YOUR-APP.snapdeploy.app/v1/` → 200 when Redis + env are OK.

Free tier sleeps ~15 min idle; first request can be slow (wake).

# QuietCord Backend

Fork of [Vencord/Backend](https://github.com/Vencord/Backend) (AGPL) for the **Quiet** client cloud sync.

- **Org:** https://github.com/QuietCord  
- **Client repo:** https://github.com/QuietCord/Quiet  
- **Upstream merges:** `git fetch upstream && git merge upstream/main`

## Deploy (recommended)

**Render + Upstash** — [docs/DEPLOY-RENDER.md](docs/DEPLOY-RENDER.md)

**SnapDeploy + Upstash** — [docs/DEPLOY-SNAPDEPLOY.md](docs/DEPLOY-SNAPDEPLOY.md) (no card on free tier)

Optional: [docs/DEPLOY-FLY.md](docs/DEPLOY-FLY.md) if you use Fly.io instead.

Configure `.env` from `.env.example` for local Docker (Discord OAuth, Redis, peppers). Set `ROOT_REDIRECT` to this repo or your privacy docs.

Client setting: `CLOUD_API_URL` in `Quiet/src/shared/brand.ts` must point to your deployed HTTPS base URL (trailing slash).

## Tier 3 API (Quiet)

All routes use the same base64 `Authorization` token as settings sync unless noted.

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| POST | `/v1/telemetry/perf` | User | Anonymous perf snapshot (schema v1). Rate limit: 1/hour/user. |
| GET | `/v1/telemetry/perf/recent` | `X-Quiet-Admin-Secret` | Last 100 samples (debug). |
| POST | `/v1/settings/merge` | User | 3-way merge with `localEtag`, `remoteEtag`, base64 zlib local blob. |
| POST | `/v1/hooks/register` | User | Discord webhook (+ optional email field) for release pings. |
| DELETE | `/v1/hooks/register` | User | Remove hook registration. |
| POST | `/v1/hooks/release` | `X-Quiet-Release-Secret` | CI broadcast to all registered webhooks. |

Environment:

- `QUIET_RELEASE_NOTIFY_SECRET` — shared secret for release broadcast and admin telemetry read.

### CI release ping (example)

```bash
curl -X POST "$CLOUD_API_URL/v1/hooks/release" \
  -H "Content-Type: application/json" \
  -H "X-Quiet-Release-Secret: $QUIET_RELEASE_NOTIFY_SECRET" \
  -d '{"tag":"v1.0.0","url":"https://github.com/QuietCord/Quiet/releases/tag/v1.0.0"}'
```

Settings history: each `PUT /v1/settings` keeps the last 16 compressed blobs for merge base resolution.

`POST /v1/settings/merge` always returns a `merged` partial blob (non-conflicting namespaces). When `complete` is false, the client resolves `conflicts` locally and uploads the composed backup.

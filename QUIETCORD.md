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

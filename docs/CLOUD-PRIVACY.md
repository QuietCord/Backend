# QuietCord Cloud — privacy

This document describes **QuietCord/Backend** (Vencloud-compatible API) when you enable **Cloud** in the Quiet client.

## What we store

- **Discord user ID** (from OAuth `identify`) to bind your account.
- **API secret** (hashed with server-side peppers; used to authenticate settings sync).
- **Your Quiet/Vencord settings backup** (encrypted-at-rest only in the sense of access control — stored as JSON in Redis, size-limited by `SIZE_LIMIT`).

We do **not** need your Discord password. OAuth is handled by Discord.

## Where data lives

- **Redis** (e.g. Upstash) configured by the instance operator.
- **API** runs on infrastructure chosen by the operator (e.g. SnapDeploy).

## Your choices

- You can **delete settings** or **delete your cloud account** from Quiet → Settings → Cloud.
- Instance operators may set `ALLOWED_USERS` to restrict who can register.

## Source & license

- Backend source: https://github.com/QuietCord/Backend (AGPL-3.0).
- Client: https://github.com/QuietCord/Quiet (GPL-3.0).

For the public instance at `backend-285bb.containers.snapdeploy.app`, contact the operator listed in the QuietCord org.

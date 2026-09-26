package globals

import (
	"os"

	"github.com/redis/go-redis/v9"
)

// environment variables
var (
	HOST = os.Getenv("HOST")
	PORT = os.Getenv("PORT")

	// REDIS_URI or REDIS_URL (Railway/SnapDeploy often use REDIS_URL).
	// Full URL: rediss://default:pass@host:6379 (Upstash). Docker Compose: redis:6379
	REDIS_URI = func() string {
		if u := os.Getenv("REDIS_URI"); u != "" {
			return u
		}
		return os.Getenv("REDIS_URL")
	}()

	ROOT_REDIRECT = os.Getenv("ROOT_REDIRECT")

	DISCORD_CLIENT_ID     = os.Getenv("DISCORD_CLIENT_ID")
	DISCORD_CLIENT_SECRET = os.Getenv("DISCORD_CLIENT_SECRET")
	DISCORD_REDIRECT_URI  = os.Getenv("DISCORD_REDIRECT_URI")

	PEPPER_SETTINGS = os.Getenv("PEPPER_SETTINGS")
	PEPPER_SECRETS  = os.Getenv("PEPPER_SECRETS")

	// Shared secret for CI release broadcasts and optional admin telemetry reads.
	RELEASE_NOTIFY_SECRET = os.Getenv("QUIET_RELEASE_NOTIFY_SECRET")

	SIZE_LIMIT int // initialised in main

	ALLOWED_USERS map[string]bool // initialised in main
)

// other app globals, initialised in main
var (
	// redis client
	RDB *redis.Client
)

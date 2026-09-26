package routes

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	g "github.com/vencord/backend/globals"
	"github.com/vencord/backend/util"
)

// POST /v1/telemetry/perf — anonymous perf snapshot (opt-in from QuietPerformance).

type perfTelemetryBody struct {
	Schema        int    `json:"schema"`
	DiscordBuild  int    `json:"discordBuild"`
	QuietGitHash  string `json:"quietGitHash"`
	BuildChannel  string `json:"buildChannel"`
	GuildClass    string `json:"guildClass"`
	PatchHealth   struct {
		Applied     int      `json:"applied"`
		Active      int      `json:"active"`
		BrokenCount int      `json:"brokenCount"`
		BrokenIds   []string `json:"brokenIds"`
		Skipped     int      `json:"skipped"`
	} `json:"patchHealth"`
	Metrics struct {
		RamMb        float64 `json:"ramMb"`
		Fps          float64 `json:"fps"`
		FrameP95Ms   float64 `json:"frameP95Ms"`
		FluxPerSec   float64 `json:"fluxPerSec"`
		LongTasksMin float64 `json:"longTasksPerMin"`
	} `json:"metrics"`
	Preset        string `json:"preset"`
	SafeMode      bool   `json:"safeMode"`
	BenchmarkMode bool   `json:"benchmarkMode"`
}

func POSTTelemetryPerf(c *fiber.Ctx) error {
	userId := c.Context().UserValue("userId").(string)

	var body perfTelemetryBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(&fiber.Map{"error": "Invalid JSON body"})
	}

	if body.Schema != 1 {
		return c.Status(400).JSON(&fiber.Map{"error": "Unsupported schema"})
	}

	if body.DiscordBuild <= 0 || len(body.QuietGitHash) == 0 || len(body.QuietGitHash) > 64 {
		return c.Status(400).JSON(&fiber.Map{"error": "Invalid snapshot metadata"})
	}

	if len(body.PatchHealth.BrokenIds) > 32 {
		return c.Status(400).JSON(&fiber.Map{"error": "Too many patch ids"})
	}

	rateKey := "telemetry:rate:" + util.Hash(g.PEPPER_SETTINGS+userId)
	ok, err := g.RDB.SetNX(c.Context(), rateKey, "1", time.Hour).Result()
	if err != nil {
		panic(err)
	}
	if !ok {
		return c.JSON(&fiber.Map{"accepted": false, "reason": "rate_limited"})
	}

	payload, _ := json.Marshal(body)
	aggKey := "telemetry:perf:recent"
	g.RDB.LPush(c.Context(), aggKey, string(payload))
	g.RDB.LTrim(c.Context(), aggKey, 0, 4999)

	// Rolling counters for dashboards (best-effort).
	if body.PatchHealth.BrokenCount > 0 {
		g.RDB.Incr(c.Context(), "telemetry:counter:broken_patches")
	}
	g.RDB.Incr(c.Context(), "telemetry:counter:samples")
	g.RDB.HIncrBy(c.Context(), "telemetry:builds", strconv.Itoa(body.DiscordBuild), 1)

	return c.JSON(&fiber.Map{"accepted": true})
}

func GETTelemetryPerfRecent(c *fiber.Ctx) error {
	if g.RELEASE_NOTIFY_SECRET == "" || c.Get("X-Quiet-Admin-Secret") != g.RELEASE_NOTIFY_SECRET {
		return c.SendStatus(404)
	}

	items, err := g.RDB.LRange(c.Context(), "telemetry:perf:recent", 0, 99).Result()
	if err == redis.Nil {
		return c.JSON(&fiber.Map{"samples": []string{}})
	} else if err != nil {
		panic(err)
	}

	return c.JSON(&fiber.Map{"samples": items})
}

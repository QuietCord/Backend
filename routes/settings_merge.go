package routes

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	g "github.com/vencord/backend/globals"
	"github.com/vencord/backend/util"
)

type mergeRequest struct {
	LocalEtag  string `json:"localEtag"`
	RemoteEtag string `json:"remoteEtag"`
	Local      string `json:"local"` // base64 raw deflate of local export; required when local != remote
}

func settingsKey(userId string) string {
	return "settings:" + util.Hash(g.PEPPER_SETTINGS+userId)
}

func settingsHistoryKey(userId string) string {
	return "settings:hist:" + util.Hash(g.PEPPER_SETTINGS+userId)
}

func appendSettingsHistory(ctx context.Context, userId string, written int64, value []byte) {
	entry, _ := json.Marshal(map[string]any{
		"written": written,
		"value":   base64.StdEncoding.EncodeToString(value),
	})
	g.RDB.LPush(ctx, settingsHistoryKey(userId), string(entry))
	g.RDB.LTrim(ctx, settingsHistoryKey(userId), 0, 15)
}

func loadHistoryVersion(ctx context.Context, userId, etag string) ([]byte, bool) {
	items, err := g.RDB.LRange(ctx, settingsHistoryKey(userId), 0, 15).Result()
	if err != nil {
		return nil, false
	}
	for _, item := range items {
		var row map[string]any
		if json.Unmarshal([]byte(item), &row) != nil {
			continue
		}
		written := fmtWritten(row["written"])
		if written != etag {
			continue
		}
		raw, ok := row["value"].(string)
		if !ok {
			continue
		}
		decoded, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			continue
		}
		return decoded, true
	}
	return nil, false
}

func fmtWritten(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case json.Number:
		return t.String()
	default:
		return ""
	}
}

func POSTSettingsMerge(c *fiber.Ctx) error {
	userId := c.Context().UserValue("userId").(string)

	var req mergeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(&fiber.Map{"error": "Invalid JSON body"})
	}

	if req.LocalEtag == "" || req.RemoteEtag == "" {
		return c.Status(400).JSON(&fiber.Map{"error": "localEtag and remoteEtag required"})
	}

	if req.LocalEtag == req.RemoteEtag {
		return c.JSON(&fiber.Map{"complete": true, "unchanged": true})
	}

	remoteWritten, err := g.RDB.HGet(c.Context(), settingsKey(userId), "written").Result()
	if err == redis.Nil {
		return c.Status(404).JSON(&fiber.Map{"error": "No cloud settings"})
	} else if err != nil {
		panic(err)
	}

	if remoteWritten != req.RemoteEtag {
		return c.Status(409).JSON(&fiber.Map{
			"error":          "remoteEtag stale",
			"currentRemote":  remoteWritten,
		})
	}

	remoteCompressed, err := g.RDB.HGet(c.Context(), settingsKey(userId), "value").Result()
	if err != nil {
		panic(err)
	}

	localCompressed, err := base64.StdEncoding.DecodeString(req.Local)
	if err != nil || len(localCompressed) == 0 {
		return c.Status(400).JSON(&fiber.Map{"error": "local settings blob required (base64 deflate)"})
	}

	remoteRaw, err := util.InflateSettings([]byte(remoteCompressed))
	if err != nil {
		panic(err)
	}
	localRaw, err := util.InflateSettings(localCompressed)
	if err != nil {
		return c.Status(400).JSON(&fiber.Map{"error": "Invalid local deflate payload"})
	}

	localBundle, err := util.ParseSettingsBundle(localRaw)
	if err != nil {
		return c.Status(400).JSON(&fiber.Map{"error": "Invalid local settings JSON"})
	}
	remoteBundle, err := util.ParseSettingsBundle(remoteRaw)
	if err != nil {
		panic(err)
	}

	var baseBundle *util.SettingsBundle
	baseEtag := minEtag(req.LocalEtag, req.RemoteEtag)
	if blob, ok := loadHistoryVersion(c.Context(), userId, baseEtag); ok {
		baseRaw, err := util.InflateSettings(blob)
		if err == nil {
			baseBundle, _ = util.ParseSettingsBundle(baseRaw)
		}
	}

	mergeResult, err := util.MergeSettingsBundles(baseBundle, localBundle, remoteBundle)
	if err != nil {
		return c.Status(500).JSON(&fiber.Map{"error": err.Error()})
	}

	resp := fiber.Map{
		"complete":  mergeResult.Complete,
		"conflicts": mergeResult.Conflicts,
	}

	if mergeResult.Merged != nil {
		mergedJSON, err := json.Marshal(mergeResult.Merged)
		if err != nil {
			panic(err)
		}
		deflated, err := util.DeflateSettings(mergedJSON)
		if err != nil {
			panic(err)
		}
		resp["merged"] = base64.StdEncoding.EncodeToString(deflated)
	}

	return c.JSON(resp)
}

func minEtag(a, b string) string {
	ai, errA := strconv.ParseInt(a, 10, 64)
	bi, errB := strconv.ParseInt(b, 10, 64)
	if errA != nil || errB != nil {
		if a < b {
			return a
		}
		return b
	}
	if ai <= bi {
		return a
	}
	return b
}

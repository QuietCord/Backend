package routes

import (
	"encoding/json"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/imroc/req/v3"

	g "github.com/vencord/backend/globals"
	"github.com/vencord/backend/util"
)

type hookRegistration struct {
	WebhookUrl      string `json:"webhookUrl"`
	Email           string `json:"email"`
	NotifyReleases  bool   `json:"notifyReleases"`
}

func hooksKey(userId string) string {
	return "hooks:" + util.Hash(g.PEPPER_SETTINGS+userId)
}

func POSTHooksRegister(c *fiber.Ctx) error {
	userId := c.Context().UserValue("userId").(string)

	var body hookRegistration
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(&fiber.Map{"error": "Invalid JSON body"})
	}

	body.WebhookUrl = strings.TrimSpace(body.WebhookUrl)
	body.Email = strings.TrimSpace(body.Email)

	if body.WebhookUrl == "" && body.Email == "" {
		return c.Status(400).JSON(&fiber.Map{"error": "webhookUrl or email required"})
	}

	if body.WebhookUrl != "" {
		u, err := url.Parse(body.WebhookUrl)
		if err != nil || u.Scheme != "https" || !strings.Contains(u.Host, "discord") {
			return c.Status(400).JSON(&fiber.Map{"error": "webhookUrl must be a https Discord webhook"})
		}
	}

	if body.Email != "" {
		if _, err := mail.ParseAddress(body.Email); err != nil {
			return c.Status(400).JSON(&fiber.Map{"error": "Invalid email"})
		}
	}

	payload, _ := json.Marshal(body)
	g.RDB.Set(c.Context(), hooksKey(userId), string(payload), 0)
	g.RDB.SAdd(c.Context(), "hooks:subs", util.Hash(g.PEPPER_SETTINGS+userId))

	return c.JSON(&fiber.Map{"registered": true})
}

func DELETEHooksRegister(c *fiber.Ctx) error {
	userId := c.Context().UserValue("userId").(string)
	hash := util.Hash(g.PEPPER_SETTINGS + userId)
	g.RDB.Del(c.Context(), hooksKey(userId))
	g.RDB.SRem(c.Context(), "hooks:subs", hash)
	return c.SendStatus(204)
}

type releaseBroadcast struct {
	Tag     string `json:"tag"`
	Url     string `json:"url"`
	Message string `json:"message"`
}

// POST /v1/hooks/release — call from CI when Quiet main ships a release.
func POSTHooksRelease(c *fiber.Ctx) error {
	if g.RELEASE_NOTIFY_SECRET == "" || c.Get("X-Quiet-Release-Secret") != g.RELEASE_NOTIFY_SECRET {
		return c.SendStatus(404)
	}

	var body releaseBroadcast
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(&fiber.Map{"error": "Invalid JSON body"})
	}

	if body.Tag == "" {
		return c.Status(400).JSON(&fiber.Map{"error": "tag required"})
	}

	if body.Message == "" {
		body.Message = "New Quiet release on main: " + body.Tag
		if body.Url != "" {
			body.Message += " — " + body.Url
		}
	}

	members, err := g.RDB.SMembers(c.Context(), "hooks:subs").Result()
	if err != nil {
		panic(err)
	}

	client := req.C().SetTimeout(10 * time.Second)
	sent := 0

	for _, userHash := range members {
		raw, err := g.RDB.Get(c.Context(), "hooks:"+userHash).Result()
		if err != nil {
			continue
		}
		var reg hookRegistration
		if json.Unmarshal([]byte(raw), &reg) != nil || !reg.NotifyReleases {
			continue
		}

		if reg.WebhookUrl != "" {
			_, _ = client.R().
				SetBodyJsonMarshal(map[string]any{
					"content": body.Message,
				}).
				Post(reg.WebhookUrl)
			sent++
		}
		// Email delivery intentionally omitted — wire your provider in a fork if needed.
	}

	return c.JSON(&fiber.Map{"notified": sent})
}

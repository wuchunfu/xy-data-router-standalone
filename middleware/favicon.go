package middleware

import (
	"embed"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

//go:embed assets/favicon.ico
var fav embed.FS

func Favicon() fiber.Handler {
	icon, err := fav.ReadFile("assets/favicon.ico")
	if err != nil {
		log.Fatalln("Failed to read favicon.ico:", err, "\nbye.")
	}
	iconLen := strconv.Itoa(len(icon))

	return func(c *fiber.Ctx) error {
		if len(c.Path()) != 12 || c.Path() != "/favicon.ico" {
			return c.Next()
		}
		if c.Method() != fiber.MethodGet && c.Method() != fiber.MethodHead {
			if c.Method() != fiber.MethodOptions {
				c.Status(fiber.StatusMethodNotAllowed)
			} else {
				c.Status(fiber.StatusOK)
			}
			c.Set(fiber.HeaderAllow, "GET, HEAD, OPTIONS")
			c.Set(fiber.HeaderContentLength, "0")
			return nil
		}
		c.Set(fiber.HeaderContentLength, iconLen)
		c.Set(fiber.HeaderContentType, "image/x-icon")
		c.Set(fiber.HeaderCacheControl, "public, max-age=31536000")
		return c.Status(fiber.StatusOK).Send(icon)
	}
}

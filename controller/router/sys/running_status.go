package sys

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/internal/json"
	"github.com/fufuok/xy-data-router/service"
)

func runningStatusHandler(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	return c.Send(json.MustJSONIndent(service.RunningStatus()))
}

func runningQueueStatusHandler(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	return c.Send(json.MustJSONIndent(service.RunningQueueStatus()))
}

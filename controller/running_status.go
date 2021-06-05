package controller

import (
	"github.com/fufuok/utils"
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/service"
)

func runningStatusHandler(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return c.Send(utils.MustJSONIndent(service.RunningStatus()))
}

func runningQueueStatusHandler(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return c.Send(utils.MustJSONIndent(service.RunningQueueStatus()))
}

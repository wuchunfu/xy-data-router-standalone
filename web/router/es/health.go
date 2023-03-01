package es

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/service/es"
)

func healthHandler(c *fiber.Ctx) error {
	params := getParams()
	defer putParams(params)

	resp := es.GetResponse()
	defer es.PutResponse(resp)
	resp.Response, resp.Err = es.Client.Cluster.Health(
		es.Client.Cluster.Health.WithTimeout(defaultESAPITimeout),
	)

	return sendResult(c, resp, params)
}

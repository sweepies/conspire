package middleware

import (
	"math/rand"

	"github.com/gofiber/fiber/v2"
)

// RandomIndex returns the random index middleware
func RandomIndex(hostnames map[string]bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if !hostnames[ctx.Hostname()] {
			// use a random hostname's index and favicon
			hostKeys := make([]string, len(hostnames))

			i := 0
			for k := range hostnames {
				hostKeys[i] = k
				i++
			}

			ctx.Request().SetHost(hostKeys[rand.Intn(len(hostKeys))])
		}

		return ctx.Next()
	}
}

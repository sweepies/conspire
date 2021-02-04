package middleware

import "github.com/gofiber/fiber/v2"

var attributions = map[string]string{
	"lime.fan": "Icon made by Freepik from www.flaticon.com",
}

// Attribution returns the icon attribution middleware
func Attribution() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		val, ok := attributions[ctx.Hostname()]

		if ok {
			ctx.Set("X-Attribution", val)
		}

		return ctx.Next()
	}
}

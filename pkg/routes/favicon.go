package routes

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// Favicon returns the favicon handler ("/favicon.ico")
func Favicon() fiber.Handler {

	return func(ctx *fiber.Ctx) error {
		ctx.Set(fiber.HeaderContentType, "image/x-icon")

		host := ctx.Hostname()
		file := filepath.Join("static", "favicon", host+".ico")

		// ask for forgiveness rather than permission
		err := ctx.SendFile(file)

		if err != nil {
			go log.Debug().Err(err).Str("key", ctx.Path()).Msg("Request for key failed")

			err = ctx.SendFile(filepath.Join("static", "favicon", "default.ico"))

			if err == nil {
				return ctx.SendStatus(fiber.StatusOK)
			}
		}

		return nil
	}
}

package routes

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// Favicon returns the favicon handler
func Favicon() fiber.Handler {

	return func(ctx *fiber.Ctx) error {
		ctx.Set(fiber.HeaderContentType, "image/x-icon")

		host := ctx.Hostname()
		file := filepath.Join("static", "favicon", host+".ico")

		// ask for forgiveness rather than permission
		err := ctx.SendFile(file)

		if err != nil {
			err2 := ctx.SendFile(filepath.Join("static", "favicon", "default.ico"))

			if err2 != nil {
				go log.Err(err).Msg("Error sending default favicon")
				return fiber.ErrInternalServerError
			}
		}

		return nil
	}
}

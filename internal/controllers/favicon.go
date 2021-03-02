package controllers

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// Favicon returns the favicon handler
func Favicon(staticPath string) fiber.Handler {

	return func(ctx *fiber.Ctx) error {
		ctx.Set(fiber.HeaderContentType, "image/x-icon")

		file := filepath.Join(staticPath, "favicon", ctx.Hostname()+".ico")

		err := ctx.SendFile(file)

		if err != nil {
			err2 := ctx.SendFile(filepath.Join(staticPath, "favicon", "default.ico"))

			if err2 != nil {
				go log.Err(err).Msg("Error sending default favicon")
				return fiber.ErrInternalServerError
			}
		}

		return ctx.SendStatus(200)
	}
}

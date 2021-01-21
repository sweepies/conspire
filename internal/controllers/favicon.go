package controllers

import (
	"io"
	"io/fs"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// Favicon returns the favicon handler
func Favicon(staticFs fs.FS, hostnames map[string]bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		ctx.Set(fiber.HeaderContentType, "image/x-icon")

		if !hostnames[ctx.Hostname()] {
			f, err := staticFs.Open("favicon/default.ico")

			if err != nil {
				go log.Err(err).Msg("Error sending default favicon")
				return fiber.ErrNotFound
			}

			bytes, _ := io.ReadAll(f)
			return ctx.Send(bytes)
		}

		f, err := staticFs.Open("favicon/" + ctx.Hostname() + ".ico")

		if err != nil {
			return fiber.ErrNotFound
		}

		bytes, _ := io.ReadAll(f)
		return ctx.Send(bytes)
	}
}

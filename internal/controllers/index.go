package controllers

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// Index returns the index page handler
func Index(staticPath string) fiber.Handler {

	return func(ctx *fiber.Ctx) error {
		globPath := filepath.Join(staticPath, "index", ctx.Hostname()+".*")

		files, _ := filepath.Glob(globPath)

		if len(files) != 0 {
			return ctx.SendFile(files[0])
		}

		err := ctx.SendFile(filepath.Join(staticPath, "index", "default.jpg"))

		if err != nil {
			go log.Err(err).Msg("Error sending default index")
			return fiber.ErrInternalServerError
		}

		return nil
	}
}

package routes

import (
	"fmt"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// Index returns the index page handler
func Index() fiber.Handler {

	return func(ctx *fiber.Ctx) error {

		files, _ := filepath.Glob(fmt.Sprintf("static/index/%s.*", ctx.Hostname()))

		if len(files) != 0 {
			return ctx.SendFile(files[0])
		}

		err := ctx.SendFile(filepath.Join("static", "index", "default.jpg"))

		if err != nil {
			go log.Err(err).Msg("Error sending default index")
			return fiber.ErrInternalServerError
		}

		return nil
	}
}

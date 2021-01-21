package controllers

import (
	"io"
	"io/fs"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// Index returns the index page handler
func Index(staticFs fs.FS, hostnames map[string]bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {

		if !hostnames[ctx.Hostname()] {
			f, err := staticFs.Open("index/default.jpg")

			if err != nil {
				go log.Err(err).Msg("Error sending default index")
				return fiber.ErrNotFound
			}

			ctx.Set(fiber.HeaderContentType, "image/jpeg")

			bytes, _ := io.ReadAll(f)
			return ctx.Send(bytes)
		}

		matches, err := fs.Glob(staticFs, "index/"+ctx.Hostname()+".*")

		// Glob only errors on a malformed pattern (so, a malformed hostname)
		if err != nil {
			return fiber.ErrBadRequest
		}

		if len(matches) != 0 {
			name := matches[0]

			f, err := staticFs.Open(name)

			// this will be a result of an invalid path (so, again a malformed hostname)
			if err != nil {
				return fiber.ErrBadRequest
			}

			bytes, _ := io.ReadAll(f)
			return ctx.Send(bytes)
		}

		return fiber.ErrNotFound
	}
}

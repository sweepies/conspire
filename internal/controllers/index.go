package controllers

import (
	"io"
	"io/fs"
	"mime"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// Index returns the index page handler
func Index(staticFs fs.FS, hostnames map[string]bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		host := ctx.Hostname()

		if !hostnames[host] {
			f, err := staticFs.Open("index/default.html")

			if err != nil {
				go log.Err(err).Msg("Error sending default index")
				return fiber.ErrNotFound
			}

			ctx.Set(fiber.HeaderContentType, "text/html")

			bytes, _ := io.ReadAll(f)
			return ctx.Send(bytes)
		}

		matches, err := fs.Glob(staticFs, "index/"+host+".*")

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

			contentType := mime.TypeByExtension(filepath.Ext(name))
			ctx.Set(fiber.HeaderContentType, contentType)

			bytes, _ := io.ReadAll(f)
			return ctx.Send(bytes)
		}

		return fiber.ErrNotFound
	}
}

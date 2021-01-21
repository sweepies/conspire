package routes

import (
	"fmt"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

// Index returns the index handler ("/")
func Index() fiber.Handler {

	return func(ctx *fiber.Ctx) error {

		files, _ := filepath.Glob(fmt.Sprintf("static/index/%s.*", ctx.Hostname()))

		if len(files) != 0 {
			return ctx.SendFile(files[0])
		}

		err := ctx.SendFile(filepath.Join("static", "index", "default.jpg"))

		if err == nil {
			return ctx.SendStatus(fiber.StatusOK)
		}

		return nil
	}
}

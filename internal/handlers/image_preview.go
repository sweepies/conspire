package handlers

import (
	"mime"
	"path"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ImagePreview returns the file preview handler
func ImagePreview(ctx *fiber.Ctx) error {
	file := ctx.Params("file")
	typ := mime.TypeByExtension(filepath.Ext(file))

	fields := fiber.Map{
		"fileName": file,
		"URL":      path.Join("/", file),
	}

	if strings.HasPrefix(typ, "image") {
		return ctx.Render("templates/image_preview", fields)
	}

	return ctx.Render("templates/file_preview", fields)
}

package controllers

import (
	"mime"
	"net/url"
	"path"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/sweepyoface/conspire/internal/configuration"
	"github.com/sweepyoface/conspire/pkg/s3util"
)

// Upload returns the file upload handler
func Upload(config *configuration.Config, s3 *s3util.Helper) fiber.Handler {
	bucket := config.S3.Bucket

	return func(ctx *fiber.Ctx) error {

		fileHead, err := ctx.FormFile("file")

		if err != nil {
			if ctx.Get("Content-Type") != "multipart/form-data" {
				ctx.Status(fiber.StatusBadRequest)
				return ctx.JSON(fiber.Map{"status": "error", "message": "invalid content type"})
			}
			ctx.Status(fiber.StatusBadRequest)
			return ctx.JSON(fiber.Map{"status": "error", "message": "file not found"})
		}

		file, err2 := fileHead.Open()

		if err2 != nil {
			go log.Err(err2).Msg("Unexpected error opening uploaded file")
			return fiber.ErrInternalServerError
		}

		bytes := []byte{}

		_, err3 := file.Read(bytes)

		if err3 != nil {
			go log.Err(err3).Msg("Unexpected error reading uploaded file")
			return fiber.ErrInternalServerError
		}

		defer file.Close()

		mimeType := mime.TypeByExtension(filepath.Ext(fileHead.Filename))

		if mimeType == "" {
			mimeType = fiber.MIMEOctetStream
		}

		// TODO: Add a configuration option to randomly generate file names, instead of accepting user input
		_, err4 := s3.UploadObject(bucket, fileHead.Filename, file, mimeType, config.DefaultCacheControl, config.SetPublicACL)

		if err4 != nil {
			go log.Err(err4).Msg("Unexpected error uploading file")
			return fiber.ErrInternalServerError
		}

		url, _ := url.Parse(ctx.BaseURL())

		url.Path = path.Join(url.Path, fileHead.Filename)

		ctx.Status(fiber.StatusCreated)
		// TODO: Add an option to choose randomly from a list of hostnames
		return ctx.JSON(fiber.Map{"status": "success", "url": url.String()})
	}
}

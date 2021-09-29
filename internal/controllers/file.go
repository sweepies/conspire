package controllers

import (
	"mime"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/sweepyoface/conspire/pkg/util"
)

// File returns the file serving handler
func File(s3 *util.S3) fiber.Handler {
	bucket := viper.GetString("s3-bucket")

	return func(ctx *fiber.Ctx) error {
		file := ctx.Params("file")

		// fetch object metadata
		_, err := s3.HeadObject(bucket, file)

		if err != nil {
			if strings.Contains(err.Error(), "Not Found") {
				return fiber.ErrNotFound
			}
			if strings.Contains(err.Error(), "Moved Permanently") {
				go log.Err(err).Msg("Unexpected S3 error. A 301 redirect potentially indicates that the configured bucket is not in the configured region")
				return fiber.ErrInternalServerError
			}
			go log.Err(err).Msg("Unexpected S3 error")
			return fiber.ErrInternalServerError
		}

		contentType := mime.TypeByExtension(filepath.Ext(file))

		if contentType == "" {
			contentType = fiber.MIMEOctetStream
		}

		writer := ctx.Response().BodyWriter()

		err = s3.DownloadObject(bucket, file, writer)

		if err != nil {
			go log.Err(err).Msg("Unexpected S3 error")
			return fiber.ErrInternalServerError
		}

		ctx.Set(fiber.HeaderContentType, contentType)
		ctx.Set(fiber.HeaderCacheControl, viper.GetString("cache-control"))

		return nil
	}
}

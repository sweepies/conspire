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
		metadata, err := s3.HeadObject(bucket, file)

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

		var contentType string

		if metadata.ContentType == nil {
			contentType = mime.TypeByExtension(filepath.Ext(file))

			if contentType == "" {
				contentType = fiber.MIMEOctetStream
			}
		} else {
			contentType = *metadata.ContentType
		}

		bytes, err := s3.DownloadObject(bucket, file)

		if err != nil {
			go log.Err(err).Msg("Unexpected S3 error")
			return fiber.ErrInternalServerError
		}

		ctx.Set(fiber.HeaderContentType, contentType)

		if metadata.CacheControl == nil {
			ctx.Set(fiber.HeaderCacheControl, viper.GetString("cache"))
		} else {
			ctx.Set(fiber.HeaderCacheControl, *metadata.CacheControl)
		}

		return ctx.Send(bytes)
	}
}

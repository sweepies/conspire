package controllers

import (
	"mime"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/sweepyoface/conspire/internal/configuration"
	"github.com/sweepyoface/conspire/pkg/s3util"
)

// File returns the file serving handler
func File(config *configuration.Config, s3 *s3util.Helper, forbiddenIs404 bool) fiber.Handler {
	bucket := config.S3.Bucket

	return func(ctx *fiber.Ctx) error {
		file := ctx.Params("file")

		// fetch object metadata
		metadata, err := s3.HeadObject(bucket, file)

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				if aerr.Code() == "NotFound" {
					return fiber.ErrNotFound
				}

				go log.Err(err).Msg("Unexpected S3 error")
				return fiber.ErrInternalServerError
			}
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
			ctx.Set(fiber.HeaderCacheControl, config.DefaultCacheControl)
		} else {
			ctx.Set(fiber.HeaderCacheControl, *metadata.CacheControl)
		}

		return ctx.Send(bytes)
	}
}

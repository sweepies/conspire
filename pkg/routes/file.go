package routes

import (
	"mime"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/sweepyoface/conspire/pkg/s3util"
)

// File returns the file serving handler
func File(s3 *s3util.Helper, param string) fiber.Handler {
	bucket := viper.GetString("s3_bucket")

	return func(ctx *fiber.Ctx) error {
		file := ctx.Params(param)

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

		// TODO: Add a configuration option to enable or disable Discord embeds

		// send HTML to the discord crawler
		// this shows a preview embed and preserves the url in chat
		if strings.Contains(ctx.Get("User-Agent"), "Discord") && strings.HasPrefix(contentType, "image") {
			ctx.Set(fiber.HeaderCacheControl, "private")

			return ctx.Render("image", fiber.Map{
				"Url": ctx.Path(),
			})
		}

		bytes, err := s3.DownloadObject(bucket, file)

		if err != nil {
			go log.Err(err).Msg("Unexpected S3 error")
			return fiber.ErrInternalServerError
		}

		ctx.Set(fiber.HeaderContentType, contentType)

		if metadata.CacheControl == nil {
			ctx.Set(fiber.HeaderCacheControl, "no-store")
		} else {
			ctx.Set(fiber.HeaderCacheControl, *metadata.CacheControl)
		}

		return ctx.Send(bytes)
	}
}

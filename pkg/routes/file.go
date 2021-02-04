package routes

import (
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/sweepyoface/conspire/pkg/s3util"
)

func newPublicFetchURLHandler(fetchURL url.URL, forbiddenIs404 bool, param string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		file := ctx.Params(param)

		fetchURL.Path = path.Join(file)

		resp, err := http.Get(fetchURL.String())

		if err != nil {
			go log.Err(err).Msg("Error while fetching public URL")
			return fiber.ErrInternalServerError
		}

		if resp.StatusCode == fiber.StatusNotFound {
			return fiber.ErrNotFound
		}

		if resp.StatusCode == fiber.StatusForbidden && forbiddenIs404 {
			return fiber.ErrNotFound
		}

		body, bodyReadErr := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		if bodyReadErr != nil {
			go log.Err(err).Msg("Error reading response body")
			return fiber.ErrInternalServerError
		}

		if resp.StatusCode >= 400 {
			go log.Error().Str("status", resp.Status).Bytes("body", body).Msg("Error while fetching public URL")
			return fiber.ErrInternalServerError
		}

		contentType := resp.Header.Get("Content-Type")
		cacheControl := resp.Header.Get("Cache-Control")

		if contentType == "" {
			contentType = mime.TypeByExtension(filepath.Ext(file))
		}

		if cacheControl == "" {
			cacheControl = viper.GetString("default_cache_control")
		}

		ctx.Set(fiber.HeaderContentType, contentType)
		ctx.Set(fiber.HeaderCacheControl, cacheControl)

		return ctx.Send(body)
	}
}

// File returns the file serving handler
func File(s3 *s3util.Helper, forbiddenIs404 bool, param string) fiber.Handler {
	bucket := viper.GetString("s3_bucket")

	pubURL, err := url.Parse(viper.GetString("public_fetch_url"))

	if err == nil && pubURL.String() != "" {
		return newPublicFetchURLHandler(*pubURL, forbiddenIs404, param)
	}

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
			ctx.Set(fiber.HeaderCacheControl, viper.GetString("default_cache_control"))
		} else {
			ctx.Set(fiber.HeaderCacheControl, *metadata.CacheControl)
		}

		return ctx.Send(bytes)
	}
}

package controllers

import (
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"path"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/sweepyoface/conspire/internal/configuration"
)

// FilePubURL returns the file serving handler for a publically accessible URL
func FilePubURL(config *configuration.Config, forbiddenIs404 bool) fiber.Handler {
	fetchURL, err := url.Parse(config.PublicFetchURL)

	if err != nil {
		log.Fatal().Err(err).Msg("Error while parsing public fetch URL")
	}

	return func(ctx *fiber.Ctx) error {
		file := ctx.Params("file")

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
			cacheControl = config.DefaultCacheControl
		}

		ctx.Set(fiber.HeaderContentType, contentType)
		ctx.Set(fiber.HeaderCacheControl, cacheControl)

		return ctx.Send(body)
	}
}

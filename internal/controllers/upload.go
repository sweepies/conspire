package controllers

import (
	"mime"
	"net/http"
	"net/url"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/sweepyoface/conspire/pkg/util"
)

// Upload returns the file upload handler
func Upload(s3 *util.S3) fiber.Handler {
	bucket := viper.GetString("s3-bucket")

	return func(ctx *fiber.Ctx) error {
		fileName := ctx.Params("file")

		fileHead, err := ctx.FormFile("file")

		if err != nil {
			ctx.Status(fiber.StatusBadRequest)
			return ctx.JSON(fiber.Map{"status": "error", "message": "file not found"})
		}

		file, errOpen := fileHead.Open()

		if errOpen != nil {
			go log.Err(errOpen).Msg("Error opening uploaded file")
			return fiber.ErrInternalServerError
		}

		// read 512 bytes for sniffing
		buffer := make([]byte, 512)

		_, errRead := file.Read(buffer)
		if err != nil {
			go log.Err(errRead).Msg("Error reading uploaded file")
		}

		// don't trust the client, sniff true file extension
		contentType := http.DetectContentType(buffer)
		exts, err := mime.ExtensionsByType(contentType)

		var ext string

		if err != nil || len(exts) == 0 {
			ext = ".bin"
		} else {
			ext = exts[0]
		}

		// on linux, the first listed association with 'image/jpeg' is '.jpeg' when most of the world uses '.jpg'
		if ext == ".jpeg" {
			ext = ".jpg"
		}

		// change the file extension to the one sniffed
		fileName = fileName[0:len(fileName)-len(path.Ext(fileName))] + ext

		// check for existing file
		exists, err := s3.ObjectExists(bucket, fileName)

		if err != nil {
			go log.Err(err).Msg("Error checking if object exists")
			return fiber.ErrInternalServerError
		}

		if exists {
			ctx.Status(fiber.StatusConflict)
			return ctx.JSON(fiber.Map{"status": "error", "message": "file already exists"})
		}

		// TODO: Add an option to randomly generate file names (or validate input), instead of accepting user input
		_, errUpload := s3.UploadObject(bucket, fileName, file, contentType, viper.GetString("cache"), viper.GetBool("acl"))

		if errUpload != nil {
			go log.Err(errUpload).Msg("Error uploading file")
			return fiber.ErrInternalServerError
		}

		url, _ := url.Parse(ctx.BaseURL())

		url.Path = path.Join(url.Path, fileName)

		ctx.Status(fiber.StatusCreated)
		return ctx.JSON(fiber.Map{"status": "success", "url": url.String()})
	}
}

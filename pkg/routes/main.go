package routes

import (
	"bytes"
	"html/template"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func connect() *session.Session {
	var resolver endpoints.ResolverFunc = func(service, region string, opts ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		return endpoints.ResolvedEndpoint{
			URL: endpoints.AddScheme(viper.GetString("s3_endpoint"), false),
		}, nil
	}

	awsSession, err := session.NewSession(&aws.Config{
		Region:           aws.String(viper.GetString("s3_region")),
		EndpointResolver: resolver,
	})

	if err != nil {
		log.Fatal().Err(err).Send()
	}

	return awsSession
}

func parseTemplate(url string) (string, error) {
	templ, err := template.ParseFiles(path.Join("static", "template.html"))

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	type TemplateData struct {
		URL string
	}

	err = templ.Execute(&buf, TemplateData{URL: url})

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Main returns the main handler ("/:file")
func Main() fiber.Handler {

	awsSession := connect()
	s3Service := s3.New(awsSession)
	s3Downloader := s3manager.NewDownloader(awsSession)

	return func(ctx *fiber.Ctx) error {
		// fetch object metadata
		headObjectInput := s3.HeadObjectInput{
			Bucket: aws.String(viper.GetString("s3_bucket")),
			Key:    aws.String(ctx.Params("file")),
		}
		metadata, err := s3Service.HeadObject(&headObjectInput)

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				go log.Debug().Err(aerr).Str("key", ctx.Params("file")).Msg("Request for key failed")

				if aerr.Code() == "NotFound" {
					return fiber.ErrNotFound
				}

				go log.Error().Msg("Unexpected S3 error")
				return fiber.ErrInternalServerError
			}
		}

		// determine whether to deliver the object or html
		if strings.Contains(ctx.Get("User-Agent"), "Discord") && strings.HasPrefix(*metadata.ContentType, "image") {
			html, err := parseTemplate(ctx.OriginalURL())

			if err != nil {
				go log.Err(err).Msg("Unexpected error parsing template")
				return fiber.ErrInternalServerError
			}

			ctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
			return ctx.SendString(html)
		}

		getObjectInput := s3.GetObjectInput{
			Bucket: aws.String(viper.GetString("s3_bucket")),
			Key:    aws.String(ctx.Params("file")),
		}

		buf := aws.NewWriteAtBuffer([]byte{})
		size, err := s3Downloader.Download(buf, &getObjectInput)

		if err != nil {
			go log.Err(err).Msg("Unexpected S3 error")
			return fiber.ErrInternalServerError
		}

		ctx.Set(fiber.HeaderContentType, *metadata.ContentType)
		ctx.Set(fiber.HeaderCacheControl, "public, max-age=31536000")
		return ctx.SendStream(bytes.NewReader(buf.Bytes()), int(size))
	}
}

package main

import (
	"os"
	"path"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/sweepyoface/conspire/pkg/middleware"
	"github.com/sweepyoface/conspire/pkg/routes"
	"github.com/sweepyoface/conspire/pkg/s3util"
)

// VERSION is the current version of this package
const VERSION = "0.0.7"

var shouldHave []string = []string{
	"S3_REGION",
	"S3_BUCKET",
}

var (
	app *fiber.App
	s3  *s3util.Helper
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Str("version", VERSION).Msg("Starting conspire")

	configure()

	chanS3 := make(chan *s3util.Helper)
	go initS3(chanS3)

	chanAuth := make(chan fiber.Handler)
	go initAuth(chanAuth)

	app = fiber.New(fiber.Config{
		Views:        html.New(path.Join("static", "templates"), ".html"),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  0,                 // uses ReadTimeout
		BodyLimit:    100 * 1024 * 1024, // 100MiB
	})

	// optimized for fast startup

	app.Get("/", routes.Index())
	app.Get("/favicon.ico", routes.Favicon())

	app.Use(recover.New())
	app.Use(middleware.Attribution())
	app.Use("/upload", <-chanAuth)

	s3 = <-chanS3

	app.Get("/:file", routes.File(s3, "file"))
	app.Post("/upload", routes.Upload(s3))

	log.Fatal().Err(app.Listen(":8080")).Send()
}

func configure() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warn().Msg("Config file not found, proceeding with environment variables")
		} else {
			log.Fatal().Msg(err.Error())
		}
	}

	viper.AutomaticEnv()
	viper.SetDefault("s3_endpoint", "s3.amazonaws.com")
	viper.SetDefault("s3_region", "us-east-1")

	var dontHave []string

	for _, key := range shouldHave {
		viper.BindEnv(key)

		if len(viper.GetString(key)) == 0 {
			dontHave = append(dontHave, key)
		}
	}

	if len(dontHave) != 0 {
		log.Fatal().Strs("values", dontHave).Msg("Missing required configuration values")
	}
}

func initS3(c chan *s3util.Helper) {
	c <- s3util.New()
}

func initAuth(c chan fiber.Handler) {
	c <- middleware.Auth()
}

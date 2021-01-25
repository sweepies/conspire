package main

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/sweepyoface/conspire/pkg/middleware"
	"github.com/sweepyoface/conspire/pkg/routes"
)

// VERSION is the current version of this package
const VERSION = "0.0.6"

var shouldHave []string = []string{
	"S3_REGION",
	"S3_BUCKET",
}

var (
	app *fiber.App
)

func main() {
	configure()

	app = fiber.New()

	app.Use(recover.New())

	if viper.GetBool("cache_enabled") {
		app.Use(cache.New(cache.Config{
			Expiration:   30 * time.Minute,
			CacheControl: true,
			Next: func(ctx *fiber.Ctx) bool {
				return strings.Contains(ctx.Get("User-Agent"), "Discord")
			},
			KeyGenerator: func(ctx *fiber.Ctx) string {
				path := ctx.Path()

				// seperate keys for each hostname for the favicon and index routes
				if path == "/favicon.ico" || path == "/" {
					path = path + " " + ctx.Hostname()
				}

				return path
			},
		}))
	}

	app.Use("/upload", middleware.Auth())
	app.Use(middleware.Attribution())

	// most specific first
	app.Get("/", routes.Index())
	app.Get("/favicon.ico", routes.Favicon())
	app.Get("/:file", routes.Main())

	app.Listen(":8080")
}

func configure() {
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Str("version", VERSION).Msg("Starting conspire")

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
	viper.SetDefault("cache_enabled", true)

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

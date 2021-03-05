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
	"github.com/sweepyoface/conspire/internal/configuration"
	"github.com/sweepyoface/conspire/internal/controllers"
	"github.com/sweepyoface/conspire/internal/handlers"
	"github.com/sweepyoface/conspire/internal/middleware"
	"github.com/sweepyoface/conspire/pkg/s3util"
)

var (
	// CommitHash is the git commit hash of the current version
	CommitHash string
	// CommitDate is the git commit date of the current version
	CommitDate string
)

var (
	config configuration.Config
	app    *fiber.App
	s3     *s3util.Helper
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	config = configuration.Configure(CommitHash, CommitDate)

	chanS3 := make(chan *s3util.Helper)
	go initS3(chanS3)

	chanAuth := make(chan fiber.Handler)
	go initAuth(chanAuth)

	log.Info().Str("version", CommitHash).Msg("Starting conspire")
	app = fiber.New(fiber.Config{
		Views:        html.New(path.Join("static", "templates"), ".html"),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  0,                 // uses ReadTimeout
		BodyLimit:    100 * 1024 * 1024, // 100MiB
	})

	// optimized for fast startup

	app.Get("/", controllers.Index("static"))
	app.Get("/favicon.ico", controllers.Favicon("static"))

	app.Use(recover.New())
	app.Use(middleware.Attribution())
	app.Use("/upload", <-chanAuth)

	s3 = <-chanS3

	app.Get("/:file/preview", handlers.ImagePreview)
	app.Get("/:file", controllers.File(&config, s3, true))
	app.Post("/upload", controllers.Upload(&config, s3))

	log.Fatal().Err(app.Listen(":8080")).Send()
}

func initS3(c chan *s3util.Helper) {
	c <- s3util.New(&config)
}

func initAuth(c chan fiber.Handler) {
	c <- middleware.Auth()
}

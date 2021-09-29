package cmd

import (
	"io/fs"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/sweepyoface/conspire/internal/controllers"
	"github.com/sweepyoface/conspire/internal/handlers"
	"github.com/sweepyoface/conspire/internal/middleware"
	"github.com/sweepyoface/conspire/pkg/util"
	"gopkg.in/go-playground/validator.v9"
)

var (
	EmbedFs fs.FS
	app     *fiber.App
	s3      *util.S3
)

var Config struct {
	S3Endpoint   string `validate:"omitempty" viper:"s3-endpoint"`
	S3Bucket     string `validate:"v_required" viper:"s3-bucket"`
	SetPublicACL bool   `validate:"omitempty" viper:"acl"`
	CacheControl string `validate:"omitempty" viper:"cache-control"`
	RandomIndex  bool   `validate:"omitempty" viper:"random-index"`
}

var rootCmd = &cobra.Command{
	Use:              "conspire",
	Short:            "An S3-based file sharing server",
	PersistentPreRun: initConfig,
	Run:              run,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	rootCmd.Flags().StringVar(&Config.S3Endpoint, "s3-endpoint", "https://s3.amazonaws.com", "the S3 endpoint to use")
	rootCmd.Flags().StringVar(&Config.S3Bucket, "s3-bucket", "", "the S3 bucket to use")
	rootCmd.Flags().BoolVar(&Config.SetPublicACL, "acl", false, "pass this flag to set a public read ACL on objects")
	rootCmd.Flags().StringVar(&Config.CacheControl, "cache-control", "public, max-age=31536000", "the Cache-Control string to use for uploaded objects")
	rootCmd.Flags().BoolVar(&Config.RandomIndex, "random-index", false, "pass this flag to use a random hostname's index and favicon instead of the default when one cannot be found")
}

// Runs before startup to check flags
func initConfig(cmd *cobra.Command, args []string) {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.BindPFlags(cmd.Flags())
	viper.AutomaticEnv()

	validate := validator.New()

	// use the value of 'viper' tag as the name of the StructField for validation
	validate.RegisterTagNameFunc(func(fl reflect.StructField) string {
		return fl.Tag.Get("viper")
	})

	// values aren't loaded into the Config struct unless they are Cobra flags
	// override the 'required' validation to instead check if the value is stored with viper
	validate.RegisterValidation("v_required", func(fl validator.FieldLevel) bool {
		return viper.GetString(fl.FieldName()) != ""
	}, true)

	// validate
	if err := validate.Struct(&Config); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			validation := err.Tag()
			if validation == "vrequired" {
				validation = "required"
			}
			log.Error().Str("flag", err.Namespace()).Str("validation", validation).Msg("Config flag failed validation")
		}
		log.Fatal().Msg("One or more config flags failed validation")
	}
}

func run(cmd *cobra.Command, args []string) {
	// init S3 connection
	chanS3Conn := make(chan *util.S3)
	go util.NewS3(chanS3Conn, viper.GetString("s3-endpoint"))

	log.Info().Str("version", "todo").Msg("Starting conspire")

	staticFs, err := fs.Sub(EmbedFs, "static")

	if err != nil {
		log.Fatal().Err(err).Msg("Error loading static fs")
	}

	// find all unique hostnames in static directories
	hostnames := make(map[string]bool)

	walkDirFunc := func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		nameParts := strings.Split(d.Name(), ".")
		nameParts = nameParts[:len(nameParts)-1]

		basename := strings.Join(nameParts, ".")

		if basename == "default" || hostnames[basename] {
			return nil
		}

		hostnames[basename] = true
		log.Debug().Str("hostname", basename).Msg("Found hostname")

		return nil
	}

	fs.WalkDir(staticFs, "favicon", walkDirFunc)
	fs.WalkDir(staticFs, "index", walkDirFunc)

	app = fiber.New(fiber.Config{
		Views:        html.NewFileSystem(http.FS(staticFs), ".html"),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  0,                 // uses ReadTimeout
		BodyLimit:    100 * 1024 * 1024, // 100MiB
	})

	app.Use(recover.New())

	if viper.GetBool("random-index") {
		app.Use(middleware.RandomIndex(hostnames))
	}

	app.Get("/", controllers.Index(staticFs, hostnames))
	app.Get("/favicon.ico", controllers.Favicon(staticFs, hostnames))

	s3 = <-chanS3Conn

	app.Get("/:file/preview", handlers.ImagePreview)
	app.Get("/:file", controllers.File(s3))
	// app.Post("/:file", controllers.Upload(s3)) TODO authenticate this

	port := os.Getenv("PORT")

	if len(port) == 0 {
		port = "8080"
	}

	log.Fatal().Err(app.Listen(":" + port)).Send()
}

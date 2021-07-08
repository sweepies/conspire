package cmd

import (
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html"
	homedir "github.com/mitchellh/go-homedir"
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
	cfgFile string
	app     *fiber.App
	s3      *util.S3
)

var Config struct {
	S3Endpoint   string `validate:"vrequired" viper:"s3-endpoint"`
	S3Bucket     string `validate:"vrequired" viper:"s3-bucket"`
	SetPublicACL bool   `validate:"omitempty" viper:"acl"`
	CacheControl string `validate:"omitempty" viper:"cache"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:    "conspire",
	Short:  "An S3-based file sharing server",
	PreRun: preRun,
	Run:    run,
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.conspire.yaml)")
	rootCmd.Flags().StringVar(&Config.S3Endpoint, getConfigViperTagFromField("S3Endpoint"), "https://s3.amazonaws.com", "the S3 endpoint to use")
	rootCmd.Flags().StringVar(&Config.S3Bucket, getConfigViperTagFromField("S3Bucket"), "", "the S3 bucket to use (e.g. my-files)")
	rootCmd.Flags().BoolVar(&Config.SetPublicACL, getConfigViperTagFromField("SetPublicACL"), false, "set to true to set a public read ACL on objects")
	rootCmd.Flags().StringVar(&Config.CacheControl, getConfigViperTagFromField("CacheControl"), "public, max-age=31536000", "the Cache-Control string to use for uploaded objects")

	// bind flags to viper
	viper.BindPFlags(rootCmd.Flags())

	// replace use '_' as '-' for environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// look for $HOME/conspire.(json|toml|yaml|yml|hcl|ini|env|properties)
		viper.AddConfigPath(home)
		viper.SetConfigName("conspire")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Debug().Str("file", viper.ConfigFileUsed()).Msg("Using config file")
	}
}

// Runs before startup to check flags
func preRun(cmd *cobra.Command, args []string) {
	// we need to check if viper is getting correct values, not the config struct itself
	validate := validator.New()

	// use the value of 'viper' tag as the name of the StructField for validation
	validate.RegisterTagNameFunc(func(fl reflect.StructField) string {
		name := strings.SplitN(fl.Tag.Get("viper"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// override the 'required' validation to instead check if the value is stored with viper
	validate.RegisterValidation("vrequired", func(fl validator.FieldLevel) bool {
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

	// init auth middleware
	chanAuthMiddleware := make(chan fiber.Handler)
	go middleware.Auth(chanAuthMiddleware)

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

	app.Get("/", controllers.Index(staticFs, hostnames))
	app.Get("/favicon.ico", controllers.Favicon(staticFs, hostnames))

	app.Use(recover.New())
	app.Use(middleware.Attribution())
	app.Use("/upload", <-chanAuthMiddleware)

	s3 = <-chanS3Conn

	app.Get("/:file/preview", handlers.ImagePreview)
	app.Get("/:file", controllers.File(s3))
	app.Post("/:file", controllers.Upload(s3))

	// TODO: move to controller
	app.Get("/:file/preview", func(ctx *fiber.Ctx) error {
		file := ctx.Params("file")
		typ := mime.TypeByExtension(filepath.Ext(file))

		fields := fiber.Map{
			"fileName": file,
			"URL":      path.Join("/", file),
		}

		if strings.HasPrefix(typ, "image") {
			return ctx.Render("static/templates/image_preview", fields)
		}

		return ctx.Render("static/templates/file_preview", fields)
	})

	port := os.Getenv("PORT")

	if len(port) == 0 {
		port = "8080"
	}

	log.Fatal().Err(app.Listen(":" + port)).Send()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

// really want to avoid setting the viper key name more than once. this is a horrible hack
func getConfigViperTagFromField(fl string) string {
	field, ok := reflect.TypeOf(&Config).Elem().FieldByName(fl)

	viperTag, ok2 := field.Tag.Lookup("viper")

	if !ok || !ok2 {
		panic("Field name specified in flags or viper tag on field not found")
	}

	return viperTag

}

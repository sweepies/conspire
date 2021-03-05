package configuration

import (
	"fmt"
	"os"

	goConfig "github.com/pcelvng/go-config"
	"github.com/rs/zerolog/log"
	"gopkg.in/go-playground/validator.v9"
)

// Config is the main config type
type Config struct {
	S3                  S3Configuration `env:"S3" flag:"s3" validate:"required"`
	PublicFetchURL      string          `env:"PUBLIC_FETCH_URL" flag:"url" help:"The URL to fetch from instead of using the S3 API" validate:"omitempty,url"`
	SetPublicACL        bool            `env:"SET_PUBLIC_ACL" flag:"acl" help:"Whether or not to set a public read ACL on objects" validate:"omitempty"`
	DefaultCacheControl string          `env:"CACHE_CONTROL" flag:"cache" help:"Cache-Control string to use" validate:"required"`

	ShowVersion bool `flag:"version,v" help:"Show the program version" validate:"omitempty"`
}

// S3Configuration is the S3 config type
type S3Configuration struct {
	Endpoint string `env:"ENDPOINT" flag:"endpoint" help:"The S3 endpoint e.g. s3.amazonaws.com" validate:"required,fqdn"`
	Region   string `env:"REGION" flag:"region" help:"The S3 region e.g. us-east-1" validate:"required"`
	Bucket   string `env:"BUCKET" flag:"bucket" help:"The S3 bucket" validate:"required"`
}

var validate *validator.Validate

// Configure loads the application configuration
func Configure(hash, date string) Config {
	// defaults
	config := Config{
		S3: S3Configuration{
			Endpoint: "s3.amazonaws.com",
			Region:   "us-east-1",
		},
		DefaultCacheControl: "public, max-age=31536000",
	}

	err := goConfig.Load(&config)

	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config")
	}

	if config.ShowVersion {
		fmt.Println("conspire version", hash+"-unstable", "("+date+")")
		os.Exit(0)
	}

	validate = validator.New()

	errValidate := validate.Struct(config)

	if errValidate != nil {
		for _, err := range errValidate.(validator.ValidationErrors) {
			log.Error().Str("option", err.Namespace()).Str("validation", err.Tag()).Msg("Config option failed validation")
		}

		log.Fatal().Msg("One or more config options couldn't be loaded")
	}

	return config
}

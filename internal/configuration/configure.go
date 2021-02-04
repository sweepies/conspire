package configuration

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var shouldHave []string = []string{
	"S3_BUCKET",
}

// Configure bootstraps the application configuration
func Configure() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warn().Msg("Config file not found, proceeding with environment variables")
		} else {
			log.Fatal().Err(err).Send()
		}
	}

	viper.AutomaticEnv()
	viper.SetDefault("s3_endpoint", "s3.amazonaws.com")
	viper.SetDefault("s3_region", "us-east-1")
	viper.SetDefault("default_cache_control", "public, max-age=31536000")

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

// ConfigureTest bootstraps the application configuration for testing
func ConfigureTest() {
	viper.SetConfigName("config_test")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal().Msg("Config file not found")
		} else {
			log.Fatal().Err(err).Send()
		}
	}

	viper.SetDefault("s3_endpoint", "s3.amazonaws.com")
	viper.SetDefault("s3_region", "us-east-1")
	viper.SetDefault("default_cache_control", "public, max-age=31536000")

	var dontHave []string

	for _, key := range shouldHave {
		if len(viper.GetString(key)) == 0 {
			dontHave = append(dontHave, key)
		}
	}

	if len(dontHave) != 0 {
		log.Fatal().Strs("values", dontHave).Msg("Missing required configuration values")
	}
}

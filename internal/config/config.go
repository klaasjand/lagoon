package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/klaasjand/lagoon-dev/internal/repository"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	RepoConfigs []repository.RepoConfig
)

func LoadConfig() error {
	log.Info().Msg("Loading config")

	viper.SetConfigName("lagoon") // Name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // Look for config in the working directory

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			return errors.New("no config file found")
		} else {
			// Config file was found but another error was produced
			return err
		}
	}

	if err := viper.UnmarshalKey("repositories", &RepoConfigs); err != nil {
		return errors.New("unable to decode repo configs")
	}

	if len(RepoConfigs) == 0 {
		return errors.New("no repo configs found")
	}

	validate := validator.New()
	validate.RegisterValidation("repo_id", repository.ValidateId)
	validate.RegisterValidation("repo_path", repository.ValidatePathAbs)
	validate.RegisterValidation("repo_cron", repository.ValidateCron)

	if err := validate.Var(&RepoConfigs, "dive"); err != nil {
		return errors.Errorf("missing required repo config attributes %v", err)
	}

	return nil
}

package internal

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/spf13/viper"
)

var (
	cfgSingleton sync.Once
	Instance     *Config
)

type Config struct {
	WebURL  string `mapstructure:"WEB_URL"`
	BaseURL string `mapstructure:"API_URL"`
	CertKey string `mapstructure:"CERT_KEY"`
	CertVal string `mapstructure:"CERT_VAL"`
}

func LoadEnv() {
	log.Println("Load configuration file . . . .")
	cfgSingleton.Do(func() {
		viper.AutomaticEnv()
		if err := viper.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if errors.As(err, &configFileNotFoundError) {
				panic(".env file not found!, please copy .env.example and paste as .env")
			}
			panic(fmt.Sprintf("ENV_ERROR: %s", err.Error()))
		}
		log.Println("configuration file: ready")
		if err := viper.Unmarshal(&Instance); err != nil {
			panic(fmt.Sprintf("ENV_ERROR: %s", err.Error()))
		}
	})
}

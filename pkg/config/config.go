package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config is global object that holds all application level variables.
var Config appConfig

type appConfig struct {
	LogLevel              string `mapstructure:"LOG_LEVEL"`
	CityCode              string `mapstructure:"CITY_CODE"`
	PostgresHostname      string `mapstructure:"POSTGRES_HOSTNAME"`
	PostgresPort          int    `mapstructure:"POSTGRES_PORT"`
	PostgresUser          string `mapstructure:"POSTGRES_USER"`
	PostgresPassword      string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDB            string `mapstructure:"POSTGRES_DB"`
	GinMode               string `mapstructure:"GIN_MODE"`
	JWTSecret             string `mapstructure:"JWT_SECRET"`
	JWTRefreshTokenSecret string `mapstructure:"JWT_REFRESH_TOKEN_SECRET"`
	SessionSecret         string `mapstructure:"SESSION_SECRET"`
	ReferralCode          string `mapstructure:"REFERRAL_CODE"`
}

// LoadConfig loads config from files
func LoadConfig(configPaths ...string) error {
	v := viper.New()
	v.SetConfigName("app")
	v.SetConfigType("env")
	v.AutomaticEnv()
	for _, path := range configPaths {
		v.AddConfigPath(path)
	}
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read the configuration file: %s", err)
	}
	return v.Unmarshal(&Config)
}

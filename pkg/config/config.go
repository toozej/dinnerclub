package config

import (
	"os"

	"github.com/caarlos0/env/v7"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// Config is global object that holds all application level variables.
var Config appConfig

type appConfig struct {
	LogLevel              string `env:"LOG_LEVEL,required"`
	CityCode              string `env:"CITY_CODE,required"`
	PostgresHostname      string `env:"POSTGRES_HOSTNAME"`
	PostgresPort          int    `env:"POSTGRES_PORT"`
	PostgresUser          string `env:"POSTGRES_USER"`
	PostgresPassword      string `env:"POSTGRES_PASSWORD"`
	PostgresDB            string `env:"POSTGRES_DB"`
	DatabaseURL           string `env:"DATABASE_URL"`
	GinMode               string `env:"GIN_MODE,required"`
	JWTSecret             string `env:"JWT_SECRET,required"`
	JWTRefreshTokenSecret string `env:"JWT_REFRESH_TOKEN_SECRET,required"`
	SessionSecret         string `env:"SESSION_SECRET,required"`
	ReferralCode          string `env:"REFERRAL_CODE,required"`
}

func LoadConfig(configPaths ...string) error {
	// try to load config from *.env files
	err := LoadConfigFromFiles(configPaths)
	if err != nil {
		log.Debugf("unable to parse environment variables from files: %e", err)
	}

	// parse environment (both from *.env and general OS environment) into config var
	err = env.Parse(&Config)
	if err != nil {
		log.Fatalf("unable to parse environment variables: %e", err)
		return err
	}
	return nil
}

// LoadConfigFromFiles loads config from *.env files
func LoadConfigFromFiles(configPaths []string) error {
	for _, path := range configPaths {
		_, cerr := os.Stat(path)
		if cerr == nil {
			// if the path exists, load it
			eerr := godotenv.Load(path)
			if eerr != nil {
				log.Fatalf("Error loading env file %s", path)
			}
		} else if os.IsNotExist(cerr) {
			// File or directory does not exist
			log.Debugf("Config path %s does not exist: %v", configPaths, cerr)
			return cerr
		}
	}
	return nil
}

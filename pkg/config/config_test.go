package config

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Set up test data
	expectedConfig := appConfig{
		LogLevel:              "test_log_level",
		CityCode:              "test_city_code",
		PostgresHostname:      "test_postgres_hostname",
		PostgresPort:          12345,
		PostgresUser:          "test_postgres_user",
		PostgresPassword:      "test_postgres_password",
		PostgresDB:            "test_postgres_db",
		GinMode:               "test_gin_mode",
		JWTSecret:             "test_jwt_secret",
		JWTRefreshTokenSecret: "test_jwt_refresh_token_secret",
		SessionSecret:         "test_session_secret",
	}

	// Call LoadConfig and check the result
	err := LoadConfig("./testdata/config")
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}
	if Config != expectedConfig {
		t.Errorf("Loaded config does not match expected config. Got %v, expected %v", Config, expectedConfig)
	}
}

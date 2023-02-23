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
		DatabaseURL:           "test_database_url",
		GinMode:               "test_gin_mode",
		JWTSecret:             "test_jwt_secret",
		JWTRefreshTokenSecret: "test_jwt_refresh_token_secret",
		SessionSecret:         "test_session_secret",
		ReferralCode:          "test_referral_code",
	}

	// Call LoadConfig and check the result
	err := LoadConfig("./testdata/config/app.env")
	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}
	if Config != expectedConfig {
		t.Errorf("Loaded config does not match expected config. Got %v, expected %v", Config, expectedConfig)
	}
}

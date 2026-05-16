package controllers

import (
	"testing"
)

func TestDefaultErrorConfig(t *testing.T) {
	config := DefaultErrorConfig()

	if config == nil {
		t.Error("Expected config to be initialized")
		return
	}
	if config.LogSensitiveData {
		t.Error("Expected LogSensitiveData to be false in default config")
	}
}

func TestDevelopmentErrorConfig(t *testing.T) {
	config := DevelopmentErrorConfig()

	if config == nil {
		t.Error("Expected config to be initialized")
		return
	}
	if !config.LogSensitiveData {
		t.Error("Expected LogSensitiveData to be true in development config")
	}
}

func TestGetErrorConfig(t *testing.T) {
	// Test default behavior
	config := GetErrorConfig()
	if config == nil {
		t.Error("Expected config to be returned")
		return
	}

	// Just test that we can get a config
	if config.LogSensitiveData {
		t.Error("Expected default config to have LogSensitiveData=false")
	}
}

func TestNewErrorLogger(t *testing.T) {
	config := DefaultErrorConfig()
	logger := NewErrorLogger(config)

	if logger == nil {
		t.Error("Expected logger to be initialized")
	}
}

func TestUpdateErrorLogger(t *testing.T) {
	config := DevelopmentErrorConfig()

	// This should not panic
	UpdateErrorLogger(config)

	// Verify the global logger was updated
	if GlobalErrorLogger == nil {
		t.Error("Expected global error logger to be set")
	}
}

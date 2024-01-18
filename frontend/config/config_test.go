package config_test

import (
	"testing"

	"github.com/benjamonnguyen/opendoorchat/frontend/config"
)

func TestLoadConfig(t *testing.T) {
	// want
	const wantBackendApiKey = "backendApiKey"
	const wantBackendBaseUrl = "http://localhost:8080"

	cfg, _ := config.LoadConfig("test.env")
	if cfg.BackendBaseUrl != wantBackendBaseUrl {
		t.Errorf("cfg.BackendBaseUrl: got %s, want %s\n", cfg.BackendBaseUrl, wantBackendBaseUrl)
	}
}

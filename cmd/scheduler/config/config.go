package config

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/kelseyhightower/envconfig"
)

var (
	Name      = "indexer-scheduler"
	Version   string
	GitSHA    string
	Timestamp string
)

const (
	modeDevelopment = "development"
	modeProduction  = "production"
)

// Config holds the configuration data
type Config struct {
	AppEnv      string `json:"app_env" envconfig:"APP_ENV" default:"development"`
	DatabaseURL string `json:"database_url" envconfig:"DATABASE_URL" required:"true"`
	Address     string `json:"address" envconfig:"ADDRESS" default:"127.0.0.1"`
	Port        string `json:"port" envconfig:"PORT" default:"8075"`

	// Rollbar
	RollbarAccessToken string `json:"rollbar_access_token" envconfig:"ROLLBAR_ACCESS_TOKEN"`
	RollbarServerRoot  string `json:"rollbar_server_root" envconfig:"ROLLBAR_SERVER_ROOT" default:"github.com/figment-networks/indexer-scheduler"`

	SchedulesConfig    string `json:"schedules_config" envconfig:"SCHEDULES_CONFIG"`
	DestinationsConfig string `json:"destinations_config" envconfig:"DESTINATIONS_CONFIG"`

	HealthCheckInterval time.Duration `json:"health_check_interval" envconfig:"HEALTH_CHECK_INTERVAL" default:"10s"`
}

// FromFile reads the config from a file
func FromFile(path string, config *Config) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, config)
}

// FromEnv reads the config from environment variables
func FromEnv(config *Config) error {
	return envconfig.Process("", config)
}

package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config defines the overall configuration for the application.
type Config struct {
	HTTP  HTTP  `yaml:"http"`
	NEONE NEONE `yaml:"neone"`
}

// HTTP defines the configuration for the HTTP server.
type HTTP struct {
	Addr           string `yaml:"addr"`
	StaticFilesDir string `yaml:"staticFilesDir"`
}

// NEONE defines the configuration for connecting to the NE-ONE Server.
type NEONE struct {
	RequestTimeout    time.Duration     `yaml:"requestTimeout"`
	RateLimiterPolicy RateLimiterPolicy `yaml:"rateLimiterPolicy"`
	RetryPolicy       RetryPolicy       `yaml:"retryPolicy"`
}

// RateLimiterPolicy defines the configuration for rate limiting requests to the
// NE-ONE Server.
type RateLimiterPolicy struct {
	MaxExecutionsPerMinute uint          `yaml:"maxExecutionsPerMinute"`
	MaxWaitTime            time.Duration `yaml:"maxWaitTime"`
}

// RetryPolicy defines the configuration for retrying failed requests to the
// NE-ONE Server.
type RetryPolicy struct {
	MaxAttempts int           `yaml:"maxAttempts"`
	Delay       time.Duration `yaml:"delay"`
	MaxDelay    time.Duration `yaml:"maxDelay"`
}

// Load reads the configuration from a YAML file and returns a Config struct.
func Load() (Config, error) {
	const configPath = "config.yaml"

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// Default returns a Config struct with default values for all fields. This can
// be used as a fallback if loading from file fails.
func Default() Config {
	const (
		requestTimeout                  = 3 * time.Minute
		rateLimitMaxExecutionsPerMinute = 100
		rateLimitMaxWaitTime            = 30 * time.Second
		retryMaxAttempts                = 10
		retryDelay                      = 1 * time.Second
		retryMaxDelay                   = 30 * time.Second
	)

	return Config{
		HTTP: HTTP{
			Addr:           ":8181",
			StaticFilesDir: "cmd/frontend/dist",
		},
		NEONE: NEONE{
			RequestTimeout: requestTimeout,
			RateLimiterPolicy: RateLimiterPolicy{
				MaxExecutionsPerMinute: rateLimitMaxExecutionsPerMinute,
				MaxWaitTime:            rateLimitMaxWaitTime,
			},
			RetryPolicy: RetryPolicy{
				MaxAttempts: retryMaxAttempts,
				Delay:       retryDelay,
				MaxDelay:    retryMaxDelay,
			},
		},
	}
}

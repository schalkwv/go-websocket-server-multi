// Package config provides configuration for the application.
package config

// Config holds the configuration for the application.
type Config struct {
	Env              string `envconfig:"ENV"`
	AppVersion       string `envconfig:"APP_VERSION"`
	Port             string `envconfig:"PORT" default:"4041"`
	TimeoutHTTP      int    `envconfig:"TIMEOUT_HTTP" default:"30"`
	RetryMaxAttempts int    `envconfig:"RETRY_MAX_ATTEMPTS" default:"2"`
	RetryWaitMin     int    `envconfig:"RETRY_WAIT_MIN" default:"2"`
	RetryWaitMax     int    `envconfig:"RETRY_WAIT_MAX" default:"3"`
	// DSN                string `envconfig:"DB_DSN" required:"true"`
}

package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"golang.org/x/exp/slices"
)

const (
	// Environments
	ENV_ENVIRONMENT string = "ENVIRONMENT"
	ENV_APP_NAME    string = "APP_NAME"
	ENV_APP_TYPE    string = "APP_TYPE"
	ENV_CLOUD       string = "CLOUD"

	ENV_OTEL_EXPORTER_OTLP_ENDPOINT string = "OTEL_EXPORTER_OTLP_ENDPOINT"
	ENV_OTEL_EXPORTER_OTLP_HEADERS  string = "OTEL_EXPORTER_OTLP_HEADERS"

	ENV_PORT                  string = "PORT"
	ENV_SQL_DB_MIGRATION      string = "SQL_DB_MIGRATION"
	ENV_CLOUD_HOST            string = "CLOUD_HOST"
	ENV_CLOUD_REGION          string = "CLOUD_REGION"
	ENV_CLOUD_SECRET          string = "CLOUD_SECRET"
	ENV_CLOUD_TOKEN           string = "CLOUD_TOKEN"
	ENV_CLOUD_DISABLE_SSL     string = "CLOUD_DISABLE_SSL"
	ENV_CLOUD_AWS_ROLE_ARN    string = "CLOUD_AWS_ROLE_ARN"
	ENV_CACHE_URI             string = "CACHE_URI"
	ENV_CACHE_PASSWORD        string = "CACHE_PASSWORD"
	ENV_SQL_DB_NAME           string = "SQL_DB_NAME"
	ENV_SQL_DB_HOST           string = "SQL_DB_HOST"
	ENV_SQL_DB_PORT           string = "SQL_DB_PORT"
	ENV_SQL_DB_USER           string = "SQL_DB_USER"
	ENV_SQL_DB_PASSWORD       string = "SQL_DB_PASSWORD"
	ENV_SQL_DB_SSL_MODE       string = "SQL_DB_SSL_MODE"
	ENV_SQL_DB_MAX_OPEN_CONNS string = "SQL_DB_MAX_OPEN_CONNS"
	ENV_SQL_DB_MAX_IDLE_CONNS string = "SQL_DB_MAX_IDLE_CONNS"
	ENV_LOG_LEVEL             string = "LOG_LEVEL"
	ENV_COLIBRI_MESSAGING     string = "COLIBRI_MESSAGING"

	// Environment values
	ENVIRONMENT_PRODUCTION        string = "production"
	ENVIRONMENT_SANDBOX           string = "sandbox"
	ENVIRONMENT_DEVELOPMENT       string = "development"
	ENVIRONMENT_TEST              string = "test"
	APP_TYPE_SERVICE              string = "service"
	APP_TYPE_SERVERLESS           string = "serverless"
	APP_TYPE_CLI                  string = "cli"
	CLOUD_AWS                     string = "aws"
	CLOUD_GCP                     string = "gcp"
	CLOUD_FIREBASE                string = "firebase"
	CLOUD_NONE                    string = "none"
	MESSAGING_CLOUD_DEFAULT       string = "CLOUD_DEFAULT"
	MESSAGING_RABBITMQ            string = "RABBITMQ"
	SQL_DB_CONNECTION_URI_DEFAULT string = "host=%s port=%s user=%s password=%s dbname=%s application_name='%s' sslmode=%s"
	VERSION                              = "v0.1.9"

	// Errors messages
	errorEnvironmentNotConfiguredMsg string = "environment is not configured. Set production, sandbox, development or test"
	errorAppNameNotConfiguredMsg     string = "app name is not configured"
	errorAppTypeNotConfiguredMsg     string = "app type is not configured. Set service, serverless or cli"
	errorCloudNotConfiguredMsg       string = "cloud is not configured. Set aws, azure, gcp, firebase or none"
	errorParsingIntegerMsg           string = "could not parse %s, permitted int value, got %v: %w"
	errorParsingBooleanMsg           string = "could not parse %s, permitted 'true' or 'false', got %v: %w"
)

var (
	ENVIRONMENT                = ""
	APP_NAME                   = ""
	APP_TYPE                   = ""
	APP_VERSION                = ""
	WAIT_GROUP_TIMEOUT_SECONDS = 90 // 1.5 minutes

	OTEL_EXPORTER_OTLP_ENDPOINT = ""
	OTEL_EXPORTER_OTLP_HEADERS  = ""

	PORT = 8080

	CLOUD              = ""
	CLOUD_HOST         = ""
	CLOUD_REGION       = ""
	CLOUD_SECRET       = ""
	CLOUD_TOKEN        = ""
	CLOUD_DISABLE_SSL  = true
	CLOUD_AWS_ROLE_ARN = ""

	SQL_DB_NAME           = ""
	SQL_DB_CONNECTION_URI = ""
	SQL_DB_MIGRATION      = false
	SQL_DB_MAX_OPEN_CONNS = 10
	SQL_DB_MAX_IDLE_CONNS = 3

	COLIBRI_MESSAGING = MESSAGING_CLOUD_DEFAULT

	CACHE_URI      = ""
	CACHE_PASSWORD = ""
)

// Load loads and validates all environment variables. It's used in app initialization.
func Load() error {
	_ = godotenv.Load()

	ENVIRONMENT = os.Getenv(ENV_ENVIRONMENT)
	if !slices.Contains([]string{ENVIRONMENT_PRODUCTION, ENVIRONMENT_SANDBOX, ENVIRONMENT_DEVELOPMENT, ENVIRONMENT_TEST}, ENVIRONMENT) {
		return errors.New(errorEnvironmentNotConfiguredMsg)
	}

	APP_NAME = os.Getenv(ENV_APP_NAME)
	if APP_NAME == "" {
		return errors.New(errorAppNameNotConfiguredMsg)
	}

	APP_TYPE = os.Getenv(ENV_APP_TYPE)
	if !slices.Contains([]string{APP_TYPE_SERVICE, APP_TYPE_SERVERLESS, APP_TYPE_CLI}, APP_TYPE) {
		return errors.New(errorAppTypeNotConfiguredMsg)
	}

	CLOUD = os.Getenv(ENV_CLOUD)
	if !slices.Contains([]string{CLOUD_AWS, CLOUD_GCP, CLOUD_FIREBASE, CLOUD_NONE}, CLOUD) {
		return errors.New(errorCloudNotConfiguredMsg)
	}

	OTEL_EXPORTER_OTLP_ENDPOINT = os.Getenv(ENV_OTEL_EXPORTER_OTLP_ENDPOINT)
	OTEL_EXPORTER_OTLP_HEADERS = os.Getenv(ENV_OTEL_EXPORTER_OTLP_HEADERS)

	if err := convertIntEnv(&PORT, ENV_PORT); err != nil {
		return err
	}

	if err := convertIntEnvWithDefault(&WAIT_GROUP_TIMEOUT_SECONDS, "WAIT_GROUP_TIMEOUT_SECONDS", WAIT_GROUP_TIMEOUT_SECONDS); err != nil {
		return err
	}

	if err := convertIntEnv(&SQL_DB_MAX_OPEN_CONNS, ENV_SQL_DB_MAX_OPEN_CONNS); err != nil {
		return err
	}

	if err := convertIntEnv(&SQL_DB_MAX_IDLE_CONNS, ENV_SQL_DB_MAX_IDLE_CONNS); err != nil {
		return err
	}

	if err := convertBoolEnv(&SQL_DB_MIGRATION, ENV_SQL_DB_MIGRATION); err != nil {
		return err
	}

	if err := convertBoolEnv(&CLOUD_DISABLE_SSL, ENV_CLOUD_DISABLE_SSL); err != nil {
		return err
	}

	if messagingEnv := os.Getenv(ENV_COLIBRI_MESSAGING); messagingEnv != "" {
		if messagingEnv != MESSAGING_CLOUD_DEFAULT && messagingEnv != MESSAGING_RABBITMQ {
			return fmt.Errorf("invalid COLIBRI_MESSAGING value: %s. Allowed values: %s, %s", messagingEnv, MESSAGING_CLOUD_DEFAULT, MESSAGING_RABBITMQ)
		}
		COLIBRI_MESSAGING = messagingEnv
	}

	CLOUD_HOST = os.Getenv(ENV_CLOUD_HOST)
	CLOUD_REGION = os.Getenv(ENV_CLOUD_REGION)
	CLOUD_SECRET = os.Getenv(ENV_CLOUD_SECRET)
	CLOUD_TOKEN = os.Getenv(ENV_CLOUD_TOKEN)
	CLOUD_AWS_ROLE_ARN = os.Getenv(ENV_CLOUD_AWS_ROLE_ARN)

	CACHE_URI = os.Getenv(ENV_CACHE_URI)
	CACHE_PASSWORD = os.Getenv(ENV_CACHE_PASSWORD)

	SQL_DB_NAME = os.Getenv(ENV_SQL_DB_NAME)
	SQL_DB_CONNECTION_URI = fmt.Sprintf(SQL_DB_CONNECTION_URI_DEFAULT,
		os.Getenv(ENV_SQL_DB_HOST),
		os.Getenv(ENV_SQL_DB_PORT),
		os.Getenv(ENV_SQL_DB_USER),
		os.Getenv(ENV_SQL_DB_PASSWORD),
		SQL_DB_NAME,
		APP_NAME,
		os.Getenv(ENV_SQL_DB_SSL_MODE))

	return nil
}

// convertBoolEnv loads the value of an environment variable, converts it to boolean and insert the result into a pointer.
func convertBoolEnv(env *bool, envName string) error {
	if envString := os.Getenv(envName); envString != "" {
		var err error
		if *env, err = strconv.ParseBool(envString); err != nil {
			return fmt.Errorf(errorParsingBooleanMsg, envName, envString, err)
		}
	}
	return nil
}

// convertIntEnv loads the value of an environment variable, converts it to interger and insert the result into a pointer.
func convertIntEnv(env *int, envName string) error {
	if envString := os.Getenv(envName); envString != "" {
		var err error
		if *env, err = strconv.Atoi(envString); err != nil {
			return fmt.Errorf(errorParsingIntegerMsg, envName, envString, err)
		}
	}
	return nil
}

// convertIntEnvWithDefault loads the value of an environment variable, converts it to interger and insert the result into a pointer.
func convertIntEnvWithDefault(env *int, envName string, fallback int) error {
	envString := getEnvWithDefault(envName, fallback)
	var err error
	if *env, err = strconv.Atoi(envString); err != nil {
		return fmt.Errorf(errorParsingIntegerMsg, envName, envString, err)
	}
	return nil
}

// getEnvWithDefault loads the value of an environment variable.
func getEnvWithDefault(key string, defaultValue int) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return fmt.Sprintf("%v", defaultValue)
	}
	return value
}

// IsProductionEnvironment returns a boolean if is production environment.
func IsProductionEnvironment() bool {
	return ENVIRONMENT == ENVIRONMENT_PRODUCTION
}

// IsSandboxEnvironment returns a boolean if is sandbox environment.
func IsSandboxEnvironment() bool {
	return ENVIRONMENT == ENVIRONMENT_SANDBOX
}

// IsDevelopmentEnvironment returns a boolean if is development environment.
func IsDevelopmentEnvironment() bool {
	return ENVIRONMENT == ENVIRONMENT_DEVELOPMENT
}

// IsTestEnvironment returns a boolean if is test environment.
func IsTestEnvironment() bool {
	return ENVIRONMENT == ENVIRONMENT_TEST
}

// IsCloudEnvironment returns a boolean if is production or sandbox environment.
func IsCloudEnvironment() bool {
	return IsProductionEnvironment() || IsSandboxEnvironment()
}

// IsLocalEnvironment returns a boolean if is development or test environment.
func IsLocalEnvironment() bool {
	return IsDevelopmentEnvironment() || IsTestEnvironment()
}

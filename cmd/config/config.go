package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/zs-dima/auth-service/pkg/tool"
)

type Config struct {
	GrpcAddress       string
	GrpcApiKey        string
	GrpcApiReflection bool
	HttpAddress       string
	HttpApiKey        string
	OpenTelemetry     bool
	JwtSecretKey      string
	DB                *DbConfig
	Log               *LogConfig
}

type LogConfig struct {
	Level string
	File  string
	Human bool
}

type DbConfig struct {
	Uri      string
	Password string
}

func NewConfig() (*Config, error) {
	grpcAddress := os.Getenv("grpc_address")
	if grpcAddress == "" {
		return nil, fmt.Errorf("missing environment variable: grpc_address")
	}
	grpcApiKey := os.Getenv("grpc_api_key")
	_, grpcApiReflection := os.LookupEnv("grpc_api_reflection")

	httpAddress := os.Getenv("http_address")
	if httpAddress == "" {
		return nil, fmt.Errorf("missing environment variable: http_address")
	}
	httpApiKey := os.Getenv("http_api_key")

	jwtSecretKey := os.Getenv("jwt_secret_key")
	if jwtSecretKey == "" {
		return nil, fmt.Errorf("missing environment variable: jwt_secret_key")
	}

	_, opentelemetry := os.LookupEnv("opentelemetry")

	dbUri := tool.GetFileValue("db_uri")
	if dbUri == "" {
		return nil, fmt.Errorf("missing environment variable: db_uri")
	}

	dbPassword := tool.GetFileValue("db_password")
	if dbPassword == "" {
		return nil, fmt.Errorf("missing environment variable: db_password")
	}

	logLevel := strings.ToUpper(os.Getenv("log_level"))
	_, human := os.LookupEnv("human_logging")
	logFile := os.Getenv("log_file")

	return &Config{
		GrpcAddress:       grpcAddress,
		GrpcApiKey:        grpcApiKey,
		GrpcApiReflection: grpcApiReflection,
		HttpAddress:       httpAddress,
		HttpApiKey:        httpApiKey,
		JwtSecretKey:      jwtSecretKey,
		OpenTelemetry:     opentelemetry,
		Log: &LogConfig{
			Level: logLevel,
			File:  logFile,
			Human: human,
		},
		DB: &DbConfig{
			Uri:      dbUri,
			Password: dbPassword,
		},
	}, nil
}

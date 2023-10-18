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

	grpcAddress := os.Getenv("GRPC_ADDRESS")
	if grpcAddress == "" {
		return nil, fmt.Errorf("missing environment variable: GRPC_ADDRESS")
	}

	grpcApiKey := os.Getenv("GRPC_API_KEY")
	_, grpcApiReflection := os.LookupEnv("GRPC_API_REFLECTION")

	httpAddress := os.Getenv("HTTP_ADDRESS")
	if httpAddress == "" {
		return nil, fmt.Errorf("missing environment variable: HTTP_ADDRESS")
	}
	httpApiKey := os.Getenv("HTTP_API_KEY")

	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	if jwtSecretKey == "" {
		return nil, fmt.Errorf("missing environment variable: JWT_SECRET_KEY")
	}

	_, opentelemetry := os.LookupEnv("opentelemetry")

	dbUri := tool.GetFileValue("DB_URI")
	if dbUri == "" {
		return nil, fmt.Errorf("missing environment variable: DB_URI")
	}

	dbPassword := tool.GetFileValue("DB_PASSWORD")
	if dbPassword == "" {
		return nil, fmt.Errorf("missing environment variable: DB_PASSWORD")
	}

	logLevel := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	_, human := os.LookupEnv("HUMAN_LOGGING")
	logFile := os.Getenv("LOG_FILE")

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

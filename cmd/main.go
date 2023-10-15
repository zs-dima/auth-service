package main

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	zerolog "github.com/rs/zerolog/log"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	config "github.com/zs-dima/auth-service/cmd/config"
	logger "github.com/zs-dima/auth-service/cmd/log"
	jwt_interceptor "github.com/zs-dima/auth-service/internal/api/interceptor"
	api "github.com/zs-dima/auth-service/internal/api/service"
	build "github.com/zs-dima/auth-service/internal/build"

	pb "github.com/zs-dima/auth-service/internal/gen/proto"
)

// go mod tidy
// go mod vendor
// go get && go build
// go get github.com/spf13/viper@none

func main() {
	config, err := config.NewConfig()
	if err != nil {
		zerolog.Fatal().Msgf("failed to load configuration: %v", err)
		os.Exit(1)
	}

	log, file := logger.SetupLogging(config.Log)
	if file != nil {
		defer func() { _ = file.Close() }()
	}

	log.Info().
		Str("version", build.Version).
		Str("runtime", runtime.Version()).
		Int("pid", os.Getpid()).
		Str("grpc", config.GrpcAddress).
		Str("http", config.HttpAddress).
		Int("gomaxprocs", runtime.GOMAXPROCS(0)).Msg("starting service")

	ctx := context.Background()
	dbUri := strings.Replace(config.DB.Uri, ":@", ":"+url.QueryEscape(config.DB.Password)+"@", 1)
	dbPool, err := pgxpool.New(ctx, dbUri)
	if err != nil {
		log.Fatal().Msgf("unable to connect to database: %v", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	jwtOptions := &jwt_interceptor.JwtInterceptorOptions{
		SecretKey: config.JwtSecretKey,
		AllowedMethods: []string{
			"/auth.AuthService/SignIn",
		},
	}

	// Set up gRPC server options
	var grpcOpts = []grpc.ServerOption{
		grpc.MaxRecvMsgSize(32 * 1024 * 1024),
		grpc.MaxSendMsgSize(32 * 1024 * 1024),
		grpc.ChainStreamInterceptor(
			grpc_logging.StreamServerInterceptor(logger.InterceptorLogger(*log)),
			jwt_interceptor.StreamServerInterceptor(jwtOptions),
			grpc_recovery.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
		),
		grpc.ChainUnaryInterceptor(
			grpc_logging.UnaryServerInterceptor(logger.InterceptorLogger(*log)),
			jwt_interceptor.UnaryServerInterceptor(jwtOptions),
			grpc_recovery.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
		),
	}

	// Listen on the specified [host]:port
	grpcApiConn, err := net.Listen("tcp", config.GrpcAddress)
	if err != nil {
		log.Fatal().Msgf("cannot listen to address %s", config.GrpcAddress)
	}

	if config.GrpcApiKey != "" {
		grpcOpts = append(grpcOpts, api.GRPCKeyAuth(config.GrpcApiKey))
	}
	if config.OpenTelemetry {
		grpcOpts = append(grpcOpts, grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))
	}

	// Create new server, add middleware
	grpcServer := grpc.NewServer(grpcOpts...)

	// Register gRPC service servers
	api.RegisterAuthServiceServer(
		grpcServer,
		&api.AuthServiceServerConfig{JwtSecretKey: config.JwtSecretKey},
		dbPool,
		log,
		false,
	)

	// Set up health checks
	hs := health.NewServer()
	hs.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, hs)

	if config.GrpcApiReflection {
		reflection.Register(grpcServer)
	}

	go func() {
		// Start gRPC server
		if err := grpcServer.Serve(grpcApiConn); err != nil {
			log.Fatal().Msgf("serve GRPC API: %v", err)
		}
	}()

	log.Info().
		Str("address", config.GrpcAddress).
		Msg("serving GRPC API")

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	gwmux := grpc_runtime.NewServeMux()

	// Register Web API services to the gateway
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	pb.RegisterAuthServiceHandlerFromEndpoint(ctx, gwmux, config.GrpcAddress, opts)
	if err != nil {
		log.Fatal().Msgf("failed to register Web API service: %v", err)
	}
	gwServer := &http.Server{
		Addr:    config.HttpAddress,
		Handler: gwmux,
	}

	log.Info().
		Str("address", config.HttpAddress).
		Msg("serving Web API")

	go func() {
		if err := gwServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("failed to serve HTTP: %v", err)
		}
	}()

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Info().Msg("shutting down gRPC server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	if err := gwServer.Shutdown(ctx); err != nil {
		log.Fatal().Msgf("failed to stop HTTP server: %v", err)
	}

	log.Info().Msg("gRPC server stopped")
}

package api

import (
	"context"
	"crypto/subtle"

	"github.com/rs/zerolog"

	pb "github.com/zs-dima/auth-service/internal/gen/proto"

	model "github.com/zs-dima/auth-service/internal/gen/db"

	"github.com/jackc/pgx/v5/pgxpool"

	tool "github.com/zs-dima/auth-service/pkg/tool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func authorize(ctx context.Context, key []byte) error {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if len(md["authorization"]) > 0 && subtle.ConstantTimeCompare([]byte(md["authorization"][0]), key) == 1 {
			return nil
		}
	}
	return status.Error(codes.Unauthenticated, "unauthenticated")
}

// GRPCKeyAuth allows to set simple authentication based on string key from configuration.
// Client should provide per RPC credentials: set authorization key to metadata with value `apikey <KEY>`.
func GRPCKeyAuth(key string) grpc.ServerOption {
	authKey := []byte("apikey " + key)
	return grpc.UnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if err := authorize(ctx, authKey); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	})
}

// AuthServiceServerConfig for GRPC API Service.
type AuthServiceServerConfig struct {
	JwtSecretKey string
}

// RegisterGRPCServerAPI registers GRPC API service in provided GRPC server.
func RegisterAuthServiceServer(
	server *grpc.Server,
	config *AuthServiceServerConfig,
	dbPool *pgxpool.Pool,
	log *zerolog.Logger,
	useOpenTelemetry bool,

) error {
	service := newAuthServiceServer(config, dbPool, log, useOpenTelemetry)
	pb.RegisterAuthServiceServer(server, service)
	return nil
}

type AuthServiceServer struct {
	pb.UnimplementedAuthServiceServer

	config           *AuthServiceServerConfig
	DB               *model.Queries
	DbPool           *pgxpool.Pool
	Log              *zerolog.Logger
	Err              *tool.GrpcStatusTool
	useOpenTelemetry bool
}

func newAuthServiceServer(
	c *AuthServiceServerConfig,
	dbPool *pgxpool.Pool,
	log *zerolog.Logger,
	useOpenTelemetry bool,
) *AuthServiceServer {
	return &AuthServiceServer{
		config:           c,
		DB:               model.New(dbPool),
		DbPool:           dbPool,
		Err:              tool.NewGrpcStatusTool(log),
		Log:              log,
		useOpenTelemetry: useOpenTelemetry,
	}
}

// func (s *AuthServiceServer) mustEmbedUnimplementedAuthServiceServer() {}

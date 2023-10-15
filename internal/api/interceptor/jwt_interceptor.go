package jwt_interceptor

import (
	"context"
	"fmt"
	"strings"
	"time"

	tool "github.com/zs-dima/auth-service/pkg/tool/jwt"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type JwtInterceptorOptions struct {
	SecretKey      string
	AllowedMethods []string
}

func validate(ctx context.Context, secretKey string) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, status.Errorf(codes.Unauthenticated, "missing context metadata")
	}
	tokenData, ok := md["authorization"]
	if !ok || len(tokenData) < 1 {
		return ctx, status.Errorf(codes.Unauthenticated, "missing JWT token")
	}

	parts := strings.Split(tokenData[0], " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ctx, status.Errorf(codes.Unauthenticated, "malformatted JWT token, %v", tokenData)
	}
	tokenStr := parts[1]

	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		// Signing algorithm is "HS256"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Errorf(codes.Unauthenticated, "unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secretKey), nil
	})

	iss, _ := (*claims)["iss"].(string)
	aud, _ := (*claims)["aud"].(string)
	exp, _ := (*claims)["exp"].(float64)

	if err != nil ||
		token == nil ||
		!token.Valid ||
		iss != tool.Issuer ||
		aud != tool.Audience ||
		int64(exp) < time.Now().Unix() {
		return ctx, status.Errorf(codes.Unauthenticated, "unauthenticated")
	}

	authInfo, err := extractAuthInfo(claims)
	if err != nil {
		return ctx, status.Errorf(codes.Unauthenticated, "unauthenticated")
	}

	return context.WithValue(ctx, tool.UserClaimsKey, authInfo), nil
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

func StreamServerInterceptor(options *JwtInterceptorOptions) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		if contains(options.AllowedMethods, info.FullMethod) {
			return handler(srv, stream)
		}

		newCtx, err := validate(stream.Context(), options.SecretKey)
		if err != nil {
			return err
		}

		wrapped := &wrappedStream{stream, newCtx}
		return handler(srv, wrapped)
	}
}

func UnaryServerInterceptor(options *JwtInterceptorOptions) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if contains(options.AllowedMethods, info.FullMethod) {
			return handler(ctx, req)
		}

		newCtx, err := validate(ctx, options.SecretKey)

		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func extractAuthInfo(claims *jwt.MapClaims) (*tool.JwtAuthInfo, error) {
	userInfo, err := extractUserInfo(claims)
	if err != nil {
		return nil, err
	}

	deviceIdStr, ok := (*claims)["device"].(string)
	if !ok {
		return nil, fmt.Errorf("no device claim")
	}
	deviceId, err := uuid.Parse(deviceIdStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse device claim")
	}

	installationIdStr, ok := (*claims)["installation"].(string)
	if !ok {
		return nil, fmt.Errorf("no installation claim")
	}
	installationId, err := uuid.Parse(installationIdStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse installation claim")
	}

	return &tool.JwtAuthInfo{
			UserInfo:       userInfo,
			DeviceId:       &deviceId,
			InstallationId: &installationId,
		},
		nil
}

func extractUserInfo(claims *jwt.MapClaims) (*tool.JwtUserInfo, error) {
	userEmail, ok := (*claims)["userEmail"].(string)
	if !ok {
		return nil, fmt.Errorf("no userEmail claim")
	}
	userIdStr, ok := (*claims)["userId"].(string)
	if !ok {
		return nil, fmt.Errorf("no userId claim")
	}
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse userId claim")
	}

	role, ok := (*claims)["role"].(string)
	if !ok {
		return nil, fmt.Errorf("no roleId claim")
	}

	return &tool.JwtUserInfo{
		Id:    &userId,
		Role:  tool.JwtUserRole(role),
		Email: userEmail,
	}, nil
}

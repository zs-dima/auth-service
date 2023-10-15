package tool

import (
	"fmt"

	"github.com/rs/zerolog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

type GrpcStatusTool struct {
	Log *zerolog.Logger
}

func NewGrpcStatusTool(
	log *zerolog.Logger,
) *GrpcStatusTool {
	return &GrpcStatusTool{
		Log: log,
	}
}

func (s *GrpcStatusTool) Unauthenticated(
	userEmail string,
	errs ...error,
) error {
	for _, err := range errs {
		s.Log.Error().Msg(err.Error())
	}
	return s.status(
		codes.Unauthenticated,
		"Incorrect login",
		fmt.Sprintf("Unauthenticated user %s, user had not been found", userEmail),
	)
}

func (s *GrpcStatusTool) PermissionDenied(
	userEmail string,
	errs ...error,
) error {
	for _, err := range errs {
		s.Log.Error().Msg(err.Error())
	}
	return s.status(
		codes.PermissionDenied,
		"Permission denied",
		fmt.Sprintf("Permission denied: %s", userEmail),
	)
}

func (s *GrpcStatusTool) Internal(
	title string,
	details string,
	errs ...error,
) error {
	for _, err := range errs {
		s.Log.Error().Msg(err.Error())
	}
	return s.status(codes.Internal, title, details)
}

func (s *GrpcStatusTool) InvalidArgument(
	title string,
	details string,
) error {
	return s.status(codes.InvalidArgument, title, details)
}

func (s *GrpcStatusTool) status(
	code codes.Code,
	title string,
	details string,
) error {
	s.Log.Error().Msgf(details)

	st, err := status.
		New(code, title).
		WithDetails(&errdetails.DebugInfo{
			Detail: details,
		})
	if err != nil {
		return status.New(codes.Internal, "error creating detailed error").Err()
	}
	return st.Err()
}

// func ExtractUserEmail(ctx context.Context, jwtSecretKey string) (string, error) {
// 	claims, err := extractClaims(ctx, jwtSecretKey)
// 	if err != nil {
// 		return "", err
// 	}

// 	userEmail, ok := claims["userEmail"].(string)
// 	if !ok {
// 		return "", fmt.Errorf("no userEmail claim")
// 	}

// 	return userEmail, nil
// }

// type ContextUserInfo struct {
// 	Id     *uuid.UUID
// 	RoleId *uuid.UUID
// 	Email  string
// }

// func ExtractUserInfo(ctx context.Context, jwtSecretKey string) (*ContextUserInfo, error) {
// 	claims, err := extractClaims(ctx, jwtSecretKey)
// 	if err != nil {
// 		return nil, err
// 	}

// 	userInfo, err := extractUserInfo(claims)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return userInfo, nil
// }

// type ContextAuthInfo struct {
// 	UserInfo       *ContextUserInfo
// 	DeviceId       *uuid.UUID
// 	InstallationId *uuid.UUID
// }

// func ExtractAuthInfo(ctx context.Context, jwtSecretKey string) (*ContextAuthInfo, error) {
// 	claims, err := extractClaims(ctx, jwtSecretKey)
// 	if err != nil {
// 		return nil, err
// 	}

// 	userInfo, err := extractUserInfo(claims)
// 	if err != nil {
// 		return nil, err
// 	}

// 	deviceIdStr, ok := claims["device"].(string)
// 	if !ok {
// 		return nil, fmt.Errorf("no device claim")
// 	}
// 	deviceId, err := uuid.Parse(deviceIdStr)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse device claim")
// 	}

// 	installationIdStr, ok := claims["installation"].(string)
// 	if !ok {
// 		return nil, fmt.Errorf("no installation claim")
// 	}
// 	installationId, err := uuid.Parse(installationIdStr)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse installation claim")
// 	}

// 	return &ContextAuthInfo{
// 			UserInfo:       userInfo,
// 			DeviceId:       &deviceId,
// 			InstallationId: &installationId,
// 		},
// 		nil
// }

// func extractUserInfo(claims jwt.MapClaims) (*ContextUserInfo, error) {
// 	userEmail, ok := claims["userEmail"].(string)
// 	if !ok {
// 		return nil, fmt.Errorf("no userEmail claim")
// 	}
// 	userIdStr, ok := claims["userId"].(string)
// 	if !ok {
// 		return nil, fmt.Errorf("no userId claim")
// 	}
// 	userId, err := uuid.Parse(userIdStr)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse userId claim")
// 	}

// 	roleIdStr, ok := claims["roleId"].(string)
// 	if !ok {
// 		return nil, fmt.Errorf("no roleId claim")
// 	}
// 	roleId, err := uuid.Parse(roleIdStr)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse roleId claim")
// 	}

// 	return &ContextUserInfo{
// 		Id:     &userId,
// 		RoleId: &roleId,
// 		Email:  userEmail,
// 	}, nil
// }

// func extractClaims(ctx context.Context, jwtSecretKey string) (jwt.MapClaims, error) {
// 	md, ok := metadata.FromIncomingContext(ctx)
// 	if !ok {
// 		return nil, fmt.Errorf("failed to extract metadata from context")
// 	}

// 	authHeader, ok := md["authorization"]
// 	if !ok || len(authHeader) < 1 {
// 		return nil, fmt.Errorf("no authorization header")
// 	}

// 	tokenString := strings.TrimPrefix(authHeader[0], "Bearer ")
// 	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 		}
// 		return []byte(jwtSecretKey), nil
// 	})

// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse token:%d", err)
// 	}

// 	claims, ok := token.Claims.(jwt.MapClaims)
// 	if !ok || !token.Valid {
// 		return nil, fmt.Errorf("invalid token")
// 	}

// 	return claims, nil
// }

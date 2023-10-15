package api

import (
	"context"
	"errors"
	"fmt"
	_ "image/jpeg"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgtype"
	model "github.com/zs-dima/auth-service/internal/gen/db"
	pb "github.com/zs-dima/auth-service/internal/gen/proto"
	tool "github.com/zs-dima/auth-service/pkg/tool"
	image_tool "github.com/zs-dima/auth-service/pkg/tool/image_tool"
	jwt "github.com/zs-dima/auth-service/pkg/tool/jwt"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func (s *AuthServiceServer) Auth(ctx context.Context, request *emptypb.Empty) (*pb.ResultReply, error) {
	authInfo := jwt.ExtractAuthInfo(ctx)
	userEmail := authInfo.UserInfo.Email

	s.Log.Info().Msgf("Authenticating %s...", userEmail)

	_, err := s.DB.GetActiveUser(ctx, userEmail)
	if err != nil {
		return nil, s.Err.Unauthenticated(userEmail, err)
	}

	s.Log.Info().Msgf("%s authenticated successfully", userEmail)

	return &pb.ResultReply{
		Result: true,
	}, nil
}

func (s *AuthServiceServer) RefreshToken(ctx context.Context, request *pb.RefreshTokenRequest) (*pb.RefreshTokenReply, error) {
	authInfo := jwt.ExtractAuthInfo(ctx)
	userEmail := authInfo.UserInfo.Email
	userId := authInfo.UserInfo.Id

	s.Log.Info().Msgf("Refreshing %s token...", userEmail)

	encryptor := tool.Encryptor{}
	refreshTokenHash, err := s.DB.LoadRefreshToken(ctx, *userId)
	if err != nil || !encryptor.Validate(request.RefreshToken, refreshTokenHash) {
		return nil, s.Err.Unauthenticated(userEmail, err)
	}

	deviceId := authInfo.DeviceId
	installationId := authInfo.InstallationId

	user, err := s.DB.GetActiveUser(ctx, userEmail)
	if err != nil {
		return nil, s.Err.Unauthenticated(userEmail, err)
	}

	generator := jwt.TokenGenerator{}
	accessToken, err := generator.GenerateAccessToken(&user, deviceId, installationId, s.config.JwtSecretKey)
	if err != nil {
		return nil, s.Err.PermissionDenied(userEmail, err)
	}

	refreshToken, expiresAt, err := generator.GenerateRefreshToken()
	if err != nil {
		return nil, s.Err.Internal(
			"failed to sign in",
			fmt.Sprintf("failed to hash %s refresh token", userEmail),
			err,
		)
	}

	refreshTokenHash, err = encryptor.Hash(refreshToken)
	if err != nil {
		return nil, s.Err.Internal(
			"failed to sign in",
			fmt.Sprintf("failed to hash %s refresh token", userEmail),
			err,
		)
	}

	err = s.DB.SaveUserSession(
		ctx,
		model.SaveUserSessionParams{
			UserID:       user.ID,
			RefreshToken: refreshTokenHash,
			ExpiresAt:    pgtype.Timestamp{Time: expiresAt, Valid: true},
		})
	if err != nil {
		return nil, s.Err.PermissionDenied(userEmail, err)
	}

	s.Log.Info().Msgf("%s token refreshed successfully", userEmail)

	res := &pb.RefreshTokenReply{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}

	return res, nil
}

func (s *AuthServiceServer) SignIn(ctx context.Context, request *pb.SignInRequest) (*pb.AuthInfo, error) {
	userEmail := request.Email

	s.Log.Info().Msgf("Signing in %s ...", userEmail)

	if strings.TrimSpace(userEmail) == "" {
		return nil, s.Err.Unauthenticated(userEmail)
	}

	deviceId := tool.RpcIdToId(request.DeviceInfo.Id)
	installationId := tool.RpcIdToId(request.InstallationId)

	user, err := s.DB.GetActiveUser(ctx, userEmail)
	if err != nil {
		return nil, s.Err.Unauthenticated(userEmail, err)
	}

	encryptor := tool.Encryptor{}
	// pwd, _ := encryptor.Hash("admin")
	// s.Log.Warn().Msgf(pwd)
	if !encryptor.Validate(request.Password, user.Password) {
		return nil, s.Err.Unauthenticated(userEmail)
	}

	generator := jwt.TokenGenerator{}
	accessToken, err := generator.GenerateAccessToken(&user, deviceId, installationId, s.config.JwtSecretKey)
	if err != nil {
		return nil, s.Err.Internal(
			"failed to sign in",
			fmt.Sprintf("failed to generate %s access token", userEmail),
			err,
		)
	}

	refreshToken, expiresAt, err := generator.GenerateRefreshToken()
	if err != nil {
		return nil, s.Err.Internal(
			"failed to sign in",
			fmt.Sprintf("failed to generate %s refresh token", userEmail),
			err,
		)
	}

	refreshTokenHash, err := encryptor.Hash(refreshToken)
	if err != nil {
		return nil, s.Err.Internal(
			"failed to sign in",
			fmt.Sprintf("failed to hash %s refresh token", userEmail),
			err,
		)
	}

	err = s.DB.SaveUserSession(ctx, model.SaveUserSessionParams{
		UserID:       user.ID,
		RefreshToken: refreshTokenHash,
		ExpiresAt:    pgtype.Timestamp{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, s.Err.PermissionDenied(userEmail, err)
	}

	s.Log.Info().Msgf("Signed in successfully")

	res := &pb.AuthInfo{
		UserId:       tool.IdToRpcId(&user.ID),
		UserName:     user.Name,
		UserRole:     pb.UserRole(pb.UserRole_value[string(user.Role)]),
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}

	if user.Blurhash.Valid {
		res.Blurhash = &user.Blurhash.String
	}

	return res, nil
}

func (s *AuthServiceServer) SignOut(ctx context.Context, in *emptypb.Empty) (*pb.ResultReply, error) {
	authInfo := jwt.ExtractAuthInfo(ctx)
	userEmail := authInfo.UserInfo.Email

	s.Log.Info().Msgf("Signing out %s...", userEmail)

	user, err := s.DB.GetActiveUser(ctx, userEmail)
	if err != nil {
		return nil, s.Err.Unauthenticated(userEmail, err)
	}

	err = s.DB.EndUserSession(ctx, user.ID)
	if err != nil {
		return nil, s.Err.PermissionDenied(userEmail, err)
	}

	s.Log.Info().Msgf("Signed out successfully")

	res := &pb.ResultReply{
		Result: true,
	}

	return res, nil
}

func (s *AuthServiceServer) ResetPassword(ctx context.Context, request *pb.ResetPasswordRequest) (*pb.ResultReply, error) {
	userEmail := request.Email

	s.Log.Info().Msgf("Resetting %s password ...", userEmail)

	_, err := s.DB.GetActiveUser(ctx, userEmail)
	if err != nil {
		return nil, s.Err.Unauthenticated(userEmail, err)
	}

	// TODO Email reset password link

	s.Log.Info().Msgf("Reset password link sent to %s successfully", userEmail)

	res := &pb.ResultReply{
		Result: true,
	}

	return res, nil
}

func (s *AuthServiceServer) SetPassword(ctx context.Context, request *pb.SetPasswordRequest) (*pb.ResultReply, error) {
	userEmail := request.Email

	s.Log.Info().Msgf("Updating %s password ...", userEmail)

	encryptor := tool.Encryptor{}
	passwordHash, err := encryptor.Hash(request.Password)
	if err != nil {
		return nil, s.Err.Internal(
			"Failed to update password",
			fmt.Sprintf("Failed to hash %s password", userEmail),
			err,
		)
	}

	err = s.DB.UpdateUserPassword(
		ctx,
		model.UpdateUserPasswordParams{
			Email:    userEmail,
			Password: passwordHash,
		})
	if err != nil {
		return nil, s.Err.Unauthenticated(userEmail, err)
	}

	// TODO Email notification

	s.Log.Info().Msgf("%s password updated successfully", userEmail)

	res := &pb.ResultReply{
		Result: true,
	}

	return res, nil
}

func (s *AuthServiceServer) LoadUsersInfo(request *emptypb.Empty, stream pb.AuthService_LoadUsersInfoServer) error {
	ctx := stream.Context()

	authInfo := jwt.ExtractAuthInfo(ctx)
	userEmail := authInfo.UserInfo.Email

	s.Log.Info().Msgf("Loading users info %s", userEmail)

	users, err := s.DB.LoadUsers(ctx)
	if err != nil {
		return s.Err.Unauthenticated(userEmail, err)
	}

	for _, user := range users {
		userInfo := &pb.UserInfo{
			Id:      tool.IdToRpcId(&user.ID),
			Name:    user.Name,
			Email:   user.Email,
			Role:    pb.UserRole(pb.UserRole_value[string(user.Role)]),
			Deleted: user.Deleted.Bool,
		}

		if user.Blurhash.Valid {
			userInfo.Blurhash = &user.Blurhash.String
		}

		if err := stream.Send(userInfo); err != nil {
			return err
		}

		s.Log.Info().Msgf("Loaded users info %s successfully", userEmail)
	}

	return nil
}

func (s *AuthServiceServer) LoadUserAvatar(request *pb.LoadUserAvatarRequest, stream pb.AuthService_LoadUserAvatarServer) error {
	ctx := stream.Context()

	authInfo := jwt.ExtractAuthInfo(ctx)
	userEmail := authInfo.UserInfo.Email

	var userIds []uuid.UUID
	for _, id := range request.UserId {
		userIds = append(userIds, *tool.RpcIdToId(id))
	}

	s.Log.Info().Msgf("Loading %s users avatars", userEmail)

	avatars, err := s.DB.LoadUserAvatar(ctx, userIds)

	for _, avatar := range avatars {
		avatarInfo := &pb.UserAvatar{
			UserId: tool.IdToRpcId(&avatar.UserID),
		}

		if avatar.Avatar != nil {
			avatarInfo.Avatar = avatar.Avatar
		}

		if err := stream.Send(avatarInfo); err != nil {
			return err
		}

		s.Log.Info().Msgf("Loaded %s users avatars successfully", userEmail)
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return s.Err.Unauthenticated(userEmail, err)
	}

	s.Log.Info().Msgf("Loaded %s users avatars successfully", userEmail)

	return nil
}

func (s *AuthServiceServer) LoadUsers(request *pb.UserId, stream pb.AuthService_LoadUsersServer) error {
	ctx := stream.Context()

	authInfo := jwt.ExtractAuthInfo(ctx)
	userEmail := authInfo.UserInfo.Email

	s.Log.Info().Msgf("Loading users %s", userEmail)

	users, err := s.DB.LoadUsers(ctx)
	if err != nil {
		return s.Err.Unauthenticated(userEmail, err)
	}

	for _, user := range users {
		userInfo := &pb.User{
			Id:      tool.IdToRpcId(&user.ID),
			Name:    user.Name,
			Email:   user.Email,
			Role:    pb.UserRole(pb.UserRole_value[string(user.Role)]),
			Deleted: user.Deleted.Bool,
		}

		if user.Blurhash.Valid {
			userInfo.Blurhash = &user.Blurhash.String
		}

		if err := stream.Send(userInfo); err != nil {
			return err
		}

		s.Log.Info().Msgf("Loaded users %s successfully", userEmail)
	}

	return nil
}

func (s *AuthServiceServer) CreateUser(ctx context.Context, request *pb.CreateUserRequest) (*pb.ResultReply, error) {
	authInfo := jwt.ExtractAuthInfo(ctx)
	userEmail := authInfo.UserInfo.Email
	userId := tool.RpcIdToId(request.Id)

	s.Log.Info().Msgf("Saving %s user ...", userEmail)

	encryptor := tool.Encryptor{}
	passwordHash, err := encryptor.Hash(request.Password)
	if err != nil {
		return nil, s.Err.Internal(
			"Failed to save user",
			fmt.Sprintf("Failed to hash %s password:", userEmail),
			err,
		)
	}

	_, err = s.DB.CreateUser(
		ctx,
		model.CreateUserParams{
			ID:       *userId,
			Name:     request.Name,
			Email:    request.Email,
			Password: passwordHash,
			Role:     model.UserRole(pb.UserRole_name[int32(request.Role)]),
			Deleted:  request.Deleted,
		})
	if err != nil {
		return nil, s.Err.Internal(
			fmt.Sprintf("Failed to save %s", userEmail),
			fmt.Sprintf("Failed to save user: %s", userEmail),
			err,
		)
	}

	s.Log.Info().Msgf("%s user saved successfully", userEmail)

	res := &pb.ResultReply{
		Result: true,
	}

	return res, nil
}

func (s *AuthServiceServer) UpdateUser(ctx context.Context, request *pb.UpdateUserRequest) (*pb.ResultReply, error) {
	authInfo := jwt.ExtractAuthInfo(ctx)
	userEmail := authInfo.UserInfo.Email
	userId := tool.RpcIdToId(request.Id)

	s.Log.Info().Msgf("Saving %s user ...", userEmail)

	err := s.DB.UpdateUser(
		ctx,
		model.UpdateUserParams{
			ID:      *userId,
			Name:    request.Name,
			Email:   request.Email,
			Role:    model.UserRole(pb.UserRole_name[int32(request.Role)]),
			Deleted: request.Deleted,
		})
	if err != nil {
		return nil, s.Err.Internal(
			fmt.Sprintf("Failed to save %s", userEmail),
			fmt.Sprintf("Failed to save user: %s", userEmail),
			err,
		)
	}

	s.Log.Info().Msgf("%s user saved successfully", userEmail)

	res := &pb.ResultReply{
		Result: true,
	}

	return res, nil
}

func (s *AuthServiceServer) SaveUserPhoto(ctx context.Context, request *pb.UserPhoto) (*pb.ResultReply, error) {
	authInfo := jwt.ExtractAuthInfo(ctx)
	userEmail := authInfo.UserInfo.Email
	userId := tool.RpcIdToId(request.UserId)

	s.Log.Info().Msgf("Saving %s user photo ...", userEmail)

	avatar, blurhash, err := image_tool.ToAvatar(request.Photo)
	if err != nil {
		return nil, s.Err.Internal(
			"Failed to generate photo",
			fmt.Sprintf("Failed to generate %s photo", userEmail),
			err,
		)
	}

	tx, err := s.DbPool.Begin(ctx)
	if err != nil {
		return nil, s.Err.Internal(
			"Failed to save photo",
			fmt.Sprintf("Failed to save %s photo", userEmail),
			err,
		)
	}
	defer tx.Rollback(ctx)
	qtx := s.DB.WithTx(tx)

	err = qtx.UpdateUserBlurhash(
		ctx,
		model.UpdateUserBlurhashParams{
			ID:       *userId,
			Blurhash: pgtype.Text{String: blurhash, Valid: blurhash != ""},
		})
	if err != nil {
		return nil, s.Err.Internal(
			"Failed to save blurhash",
			fmt.Sprintf("Failed to save %s blurhash", userEmail),
			err,
		)
	}

	err = qtx.SaveUserPhoto(
		ctx,
		model.SaveUserPhotoParams{
			UserID: *userId,
			Avatar: avatar,
			Photo:  request.Photo,
		})
	if err != nil {
		return nil, s.Err.Internal(
			"Failed to save avatar",
			fmt.Sprintf("Failed to save %s avatar", userEmail),
			err,
		)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, s.Err.Internal(
			"Failed to save photo",
			fmt.Sprintf("Failed to save %s photo", userEmail),
			err,
		)
	}

	s.Log.Info().Msgf("%s avatar saved successfully", userEmail)

	res := &pb.ResultReply{
		Result: true,
	}

	return res, nil
}

/*
// TODO https://github.com/sqlc-dev/sqlc/issues/712
func Run(ctx context.Context, db *sql.DB, readOnly ReadOnly, uid string, auth AuthProvider, role DBRole, fn func(context.Context, *Queries) error) (finalErr error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		finalErr = multierror.Append(finalErr, conn.Close()).ErrorOrNil()
	}()

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{ReadOnly: bool(readOnly)})
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`SELECT set_config('role', $1, true), set_config('request.uid', $2, true), set_config('request.auth', $3, true)`,
		string(role), uid, string(auth))
	if err != nil {
		return multierror.Append(errors.Wrap(err, "set security context"), tx.Rollback())
	}

	if err = fn(ctx, New(tx)); err != nil {
		return multierror.Append(err, tx.Rollback())
	}

	return tx.Commit()
}
*/

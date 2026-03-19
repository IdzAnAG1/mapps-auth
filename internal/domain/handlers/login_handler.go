package handlers

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "mapps_auth/generated/mobileapps/proto/auth/v1"
	db_gen "mapps_auth/internal/db/gen"
	"mapps_auth/internal/domain/jwt"
	"mapps_auth/internal/domain/utils"
	"mapps_auth/internal/domain/validation"
)

func LoginHandler(
	ctx context.Context,
	req *v1.LoginRequest,
	logger *slog.Logger,
	db db_gen.Queries,
	jwtManager *jwt.Manager,
) (*v1.LoginResponse, error) {
	logger.Debug("login request received", "email", req.Email)

	if err := validation.ValidateLoginRequest(req.Email, req.Password); err != nil {
		logger.Debug("login validation failed", "error", err)
		return nil, err
	}

	user, err := db.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Debug("login rejected: user not found", "email", req.Email)
			return nil, status.Error(codes.NotFound, "user not found")
		}
		logger.Debug("failed to fetch user by email", "email", req.Email, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}
	logger.Debug("user fetched", "user_id", user.ID)

	if err := utils.CheckPasswordHash(user.PasswordHash, req.Password); err != nil {
		logger.Debug("login rejected: invalid password", "user_id", user.ID)
		return nil, status.Error(codes.Unauthenticated, "invalid password")
	}

	token, err := jwtManager.Generate(user.ID, user.Email)
	if err != nil {
		logger.Debug("failed to generate jwt", "user_id", user.ID, "error", err)
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	logger.Debug("login successful", "user_id", user.ID)
	return &v1.LoginResponse{
		UserId:      user.ID,
		AccessToken: token,
	}, nil
}

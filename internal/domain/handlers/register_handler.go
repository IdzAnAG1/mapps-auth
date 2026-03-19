package handlers

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "mapps_auth/generated/mobileapps/proto/auth/v1"
	db_gen "mapps_auth/internal/db/gen"
	"mapps_auth/internal/domain/jwt"
	"mapps_auth/internal/domain/models"
	"mapps_auth/internal/domain/validation"
)

func RegisterHandler(
	ctx context.Context,
	req *v1.RegisterRequest,
	logger *slog.Logger,
	conn *pgx.Conn,
	jwtManager *jwt.Manager,
) (*v1.RegisterResponse, error) {
	logger.Debug("register request received", "email", req.Email, "username", req.Username)

	if err := validation.ValidateRegisterRequest(req.Email, req.Password, req.Username); err != nil {
		logger.Debug("register validation failed", "error", err)
		return nil, err
	}

	q := db_gen.New(conn)

	isExists, err := q.FindUserByEmail(ctx, req.Email)
	if err != nil {
		logger.Debug("failed to check email existence", "email", req.Email, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}
	if isExists {
		logger.Debug("registration rejected: email already taken", "email", req.Email)
		return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
	}

	isExists, err = q.FindUserByUsername(ctx, req.Username)
	if err != nil {
		logger.Debug("failed to check username existence", "username", req.Username, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}
	if isExists {
		logger.Debug("registration rejected: username already taken", "username", req.Username)
		return nil, status.Error(codes.AlreadyExists, "user with this username already exists")
	}

	user, err := models.NewUser(req.Email, req.Password, req.Username)
	if err != nil {
		logger.Debug("failed to build user model", "email", req.Email, "error", err)
		return nil, status.Error(codes.Internal, "failed to create user")
	}
	logger.Debug("user model created", "user_id", user.ID)

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Debug("failed to begin transaction", "user_id", user.ID, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := q.WithTx(tx)

	if err := qtx.CreateUser(ctx, db_gen.CreateUserParams{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PassHash,
		Username:     user.Username,
	}); err != nil {
		logger.Debug("failed to insert user", "user_id", user.ID, "error", err)
		return nil, status.Error(codes.Internal, "failed to create user")
	}
	logger.Debug("user inserted", "user_id", user.ID)

	if err := qtx.AssignDefaultRole(ctx, user.ID); err != nil {
		logger.Debug("failed to assign default role", "user_id", user.ID, "error", err)
		return nil, status.Error(codes.Internal, "failed to assign role")
	}
	logger.Debug("default role assigned", "user_id", user.ID)

	if err := tx.Commit(ctx); err != nil {
		logger.Debug("failed to commit transaction", "user_id", user.ID, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}
	logger.Debug("registration transaction committed", "user_id", user.ID)

	token, err := jwtManager.Generate(user.ID, user.Email)
	if err != nil {
		logger.Debug("failed to generate jwt", "user_id", user.ID, "error", err)
		return nil, status.Error(codes.Internal, "failed to generate token")
	}
	logger.Debug("jwt generated", "user_id", user.ID)

	return &v1.RegisterResponse{
		UserId:      user.ID,
		AccessToken: token,
	}, nil
}

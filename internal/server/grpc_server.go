package server

import (
	"context"
	"log/slog"

	v1 "mapps_auth/generated/mobileapps/proto/auth/v1"
	"mapps_auth/internal/db"
	"mapps_auth/internal/domain/handlers"
	"mapps_auth/internal/domain/jwt"
)

type GrpcAuthServer struct {
	v1.UnimplementedAuthServer
	Logger     *slog.Logger
	DB         *db.DB
	JWTManager *jwt.Manager
}

func (gs *GrpcAuthServer) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	return handlers.RegisterHandler(ctx, req, gs.Logger, gs.DB.Conn, gs.JWTManager)
}

func (gs *GrpcAuthServer) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	return handlers.LoginHandler(ctx, req, gs.Logger, *gs.DB.Queries, gs.JWTManager)
}

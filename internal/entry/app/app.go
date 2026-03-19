package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	authv1 "mapps_auth/generated/mobileapps/proto/auth/v1"
	"mapps_auth/internal/config"
	"mapps_auth/internal/db"
	"mapps_auth/internal/domain/jwt"
	"mapps_auth/internal/interruptor"
	logger "mapps_auth/internal/logger"
	"mapps_auth/internal/server"

	"google.golang.org/grpc"
)

type App struct {
	tcpPort    string
	logger     *slog.Logger
	db         db.DB
	jwtManager *jwt.Manager
}

func NewApp() (*App, error) {
	cfg, err := config.LoadAndGetConfig()
	if err != nil {
		return nil, err
	}
	log := logger.New(cfg.Logger.Level)
	log.Info("logger initialized", "level", cfg.Logger.Level)

	log.Debug("config loaded",
		"server_port", cfg.AuthServer.Port,
		"db_host", cfg.Database.Host,
		"db_port", cfg.Database.Port,
		"db_name", cfg.Database.Name,
		"db_user", cfg.Database.User,
		"db_connection_attempts", cfg.Database.NoCA,
		"db_timeout_sec", cfg.Database.Timeout,
	)

	database := db.NewDB(
		cfg.GetPostgresLink(),
		cfg.Database.NoCA,
		cfg.Database.Timeout,
		log,
	)
	log.Debug("database client created")

	if err = database.ConnectWithDB(); err != nil {
		if database.Conn != nil {
			if !database.Conn.IsClosed() {
				log.Debug("closing leftover db connection after failed connect")
				if closeErr := database.Conn.Close(context.Background()); closeErr != nil {
					log.Error("failed to close leftover db connection", "error", closeErr)
					return nil, closeErr
				}
			}
		}
		return nil, err
	}

	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.TTL)
	log.Debug("jwt manager initialized", "ttl", cfg.JWT.TTL)

	log.Debug("app initialized", "port", cfg.AuthServer.Port)
	return &App{
		tcpPort:    cfg.AuthServer.Port,
		logger:     log,
		db:         *database,
		jwtManager: jwtManager,
	}, nil
}

func (app *App) Run() error {
	app.logger.Info("starting auth server", "port", app.tcpPort)

	tcp, err := net.Listen("tcp", fmt.Sprintf(":%s", app.tcpPort))
	if err != nil {
		app.logger.Error("failed to start tcp listener", "port", app.tcpPort, "error", err)
		return err
	}
	app.logger.Debug("tcp listener started", "addr", tcp.Addr().String())

	defer func() {
		err = tcp.Close()
		if err == nil || errors.Is(err, net.ErrClosed) {
			app.logger.Info("tcp listener closed")
			return
		}
		app.logger.Error("failed to close tcp listener", "error", err)
	}()

	srv := grpc.NewServer()
	app.logger.Debug("grpc server created")

	iter := interruptor.NewInterruptor(srv, app.logger, app.db)
	iter.Run()

	authServ := server.GrpcAuthServer{
		Logger:     app.logger,
		DB:         &app.db,
		JWTManager: app.jwtManager,
	}
	authv1.RegisterAuthServer(srv, &authServ)
	app.logger.Info("auth grpc service registered")

	app.logger.Info("serving grpc", "addr", tcp.Addr().String())
	if err = srv.Serve(tcp); err != nil {
		app.logger.Error("grpc server stopped with error", "error", err)
		return err
	}
	return nil
}

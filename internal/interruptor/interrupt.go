package interruptor

import (
	"log/slog"
	"mapps_auth/internal/db"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
)

type Interruptor struct {
	gRPCInterruptor *grpc.Server
	signal          chan os.Signal
	logger          *slog.Logger
	database        db.DB
}

func NewInterruptor(srv *grpc.Server, logger *slog.Logger, db db.DB) *Interruptor {
	return &Interruptor{
		srv,
		make(chan os.Signal, 1),
		logger,
		db,
	}
}

func (i *Interruptor) Run() {
	i.startCatchingSignal()
	go func() {
		i.shutdown()
	}()
}

func (i *Interruptor) startCatchingSignal() {
	i.logger.Debug("registering signal handlers", "signals", []string{"SIGTERM", "SIGINT"})
	signal.Notify(i.signal, syscall.SIGTERM, syscall.SIGINT)
	i.logger.Info("starting signal catching")
}

func (i *Interruptor) shutdown() {
	sig := <-i.signal
	i.logger.Info("shutdown signal received", "signal", sig.String())

	i.logger.Debug("stopping grpc server gracefully")
	i.gRPCInterruptor.GracefulStop()
	i.logger.Info("grpc server stopped")

	i.logger.Debug("closing database connection")
	if err := i.database.Close(); err != nil {
		i.logger.Error("error closing database", "error", err)
		return
	}
	i.logger.Info("database connection closed")
}

package db

import (
	"context"
	"fmt"
	"log/slog"
	db_gen "mapps_auth/internal/db/gen"
	"time"

	"github.com/jackc/pgx/v5"
)

type DB struct {
	pgUrl   string
	Conn    *pgx.Conn
	Queries *db_gen.Queries
	NoCA    int
	Timeout int
	logger  *slog.Logger
}

func NewDB(pgUrl string, noca, tim int, log *slog.Logger) *DB {
	return &DB{
		pgUrl:   pgUrl,
		NoCA:    noca,
		Timeout: tim,
		logger:  log,
	}
}

func (db *DB) ConnectWithDB() (err error) {
	db.logger.Debug("connecting to database", "attempts", db.NoCA, "timeout_sec", db.Timeout)

	for i := 0; i < db.NoCA; i++ {
		db.logger.Debug("connection attempt", "attempt", fmt.Sprintf("%d/%d", i+1, db.NoCA))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(db.Timeout))
		db.Conn, err = pgx.Connect(ctx, db.pgUrl)
		cancel()

		if err == nil {
			break
		}

		db.logger.Debug("connection attempt failed", "attempt", i+1, "error", err)
		if i < db.NoCA-1 {
			db.logger.Debug("retrying in 5s")
			time.Sleep(5 * time.Second)
		}
	}

	if err != nil {
		db.logger.Error("failed to connect to db", "error", err)
		return err
	}

	db.Queries = db_gen.New(db.Conn)
	db.logger.Info("connected to db")
	return nil
}

func (db *DB) Close() error {
	db.logger.Debug("closing database connection")
	err := db.Conn.Close(context.Background())
	if err != nil {
		db.logger.Error("failed to close database connection", "error", err)
		return err
	}
	db.logger.Debug("database connection closed")
	return nil
}

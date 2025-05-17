package postgres

import (
	"context"
	"fmt"
	"github.com/Muaz717/gym_app/app/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, cfg config.DB) (*Storage, error) {
	const op = "storage.postgres.New"

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.Username,
		cfg.DBPassword,
		cfg.Host,
		cfg.DBPort,
		cfg.DBName,
	)

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

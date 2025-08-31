package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	MaxConnLifetime = time.Hour
	MaxConnIdleTime = 30 * time.Minute
)

type Postgres interface {
	Pool() *pgxpool.Pool
	Close()
}

type Config struct {
	Host      string
	Port      uint16
	User      string
	Password  string
	Name      string
	SSLMode   string
	MaxConns  int32
	MinConns  int32
	Migration Migration
}

type Migration struct {
	Path      string
	AutoApply bool
}

type postgres struct {
	db *pgxpool.Pool
}

func New(cfg *Config) (postgresDB Postgres, err error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnLifetime = MaxConnLifetime
	config.MaxConnIdleTime = MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err = pool.Ping(context.Background()); err != nil {
		pool.Close()

		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if cfg.Migration.AutoApply {
		m, err := migrate.New(
			fmt.Sprintf("file://%s", cfg.Migration.Path),
			connString,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create migration: %w", err)
		}

		defer func() {
			srcErr, dbErr := m.Close()
			if srcErr != nil || dbErr != nil {
				err = fmt.Errorf("failed to close migration instance: source error: %w, database error: %w", srcErr, dbErr)
			}
		}()

		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return nil, fmt.Errorf("failed to migrate to database: %w", err)
		}

	}

	return &postgres{db: pool}, nil
}

func (p *postgres) Pool() *pgxpool.Pool {
	return p.db
}

func (p *postgres) Close() {
	p.db.Close()
}

package postgres

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/shrtyk/subscriptions-service/internal/config"
)

func MustCreateConnectionPool(cfg *config.PostgresCfg) *sql.DB {
	dsn := buildDSN(cfg)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		msg := fmt.Sprintf("failed create connections pool for DSN: '%s': %s", dsn, err)
		panic(msg)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdletime)

	if err = db.Ping(); err != nil {
		panic(fmt.Sprintf("failed ping DB: %s", err))
	}

	return db
}

func buildDSN(cfg *config.PostgresCfg) string {
	url := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Path:   cfg.DBName,
	}

	q := url.Query()
	q.Set("sslmode", cfg.SSLMode)
	url.RawQuery = q.Encode()

	return url.String()
}

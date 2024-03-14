package db

import (
	"EXAM3/user_service/config"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
)

type Postgres struct {
	Builder squirrel.StatementBuilderType
	DB      *sql.DB
}

func New(cfg config.Config) (*Postgres, error) {
	postgres := &Postgres{}
	pgxUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDatabase,
	)

	postgres.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	var err error
	postgres.DB, err = sql.Open("postgres", pgxUrl)
	if err != nil {
		return nil, err
	}

	return postgres, nil
}

func (p *Postgres) Close() {
	p.DB.Close()
}

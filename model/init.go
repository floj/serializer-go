package model

import (
	"context"
	"database/sql"
	_ "embed"

	"github.com/floj/serializer-go/config"
	_ "github.com/lib/pq"
)

//go:embed schema.sql
var ddl string

func InitDB(conf config.Config) (*sql.DB, error) {
	ctx := context.Background()

	db, err := sql.Open("postgres", conf.DBURI)
	if err != nil {
		return nil, err
	}

	if _, err := db.ExecContext(ctx, ddl); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

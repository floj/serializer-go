package model

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	"github.com/floj/serializer-go/config"
	"github.com/tursodatabase/go-libsql"
)

//go:embed schema.sql
var ddl string

type finalizer struct {
	fns []func() error
}

func (f *finalizer) Finalize() error {
	errs := []error{}
	for _, fn := range f.fns {
		errs = append(errs, fn())
	}
	return errors.Join(errs...)
}

func InitDB(conf config.DbConfig) (*sql.DB, func() error, error) {
	if conf.Type != "turso" {
		return nil, nil, fmt.Errorf("only turso DB is supported at the moment")
	}

	ctx := context.Background()

	fin := finalizer{}

	connector, err := libsql.NewEmbeddedReplicaConnector(conf.LocalPath, conf.RemoteURL,
		libsql.WithAuthToken(conf.AuthToken),
		libsql.WithSyncInterval(conf.SyncInterval),
	)

	if err != nil {
		fin.Finalize()
		return nil, nil, fmt.Errorf("error creating connector: %w", err)
	}
	fin.fns = append(fin.fns, func() error { _, err := connector.Sync(); return err }, connector.Close)

	db := sql.OpenDB(connector)
	fin.fns = append(fin.fns, db.Close)

	if _, err := db.ExecContext(ctx, ddl); err != nil {
		fin.Finalize()
		return nil, nil, err
	}
	return db, fin.Finalize, nil
}

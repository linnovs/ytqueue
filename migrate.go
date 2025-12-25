package main

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/adrg/xdg"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func openAndMigrateDB() (*sql.DB, error) {
	dbFile, err := xdg.DataFile("ytqueue/videos.db")
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", fmt.Sprintf("file://%s?_pragma=foreign_keys(1)", dbFile))
	if err != nil {
		return nil, err
	}

	srcD, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return nil, err
	}

	dbD, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance("iofs", srcD, "sqlite", dbD)
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	return db, nil
}

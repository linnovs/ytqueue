package main

import (
	"embed"
	"errors"
	"fmt"

	"github.com/adrg/xdg"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func migrateDB() error {
	d, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return err
	}

	dbFile, err := xdg.DataFile("ytqueue/videos.db")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, fmt.Sprintf("sqlite://%s", dbFile))
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

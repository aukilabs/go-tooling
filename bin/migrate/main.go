package main

import (
	"github.com/aukilabs/go-tooling/pkg/cli"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := struct {
		PostgresURL string `env:"MIGGRATE_POSTGRES_URL" help:"The url to connect to Postgres."`
		Sources     string `env:"MIGGRATE_SOURCES"      help:"The directory where the migration files are located."`
		Steps       int    `env:"MIGRATE_STEPS"         help:"The steps forward or backward to migrate to. 0 migrates the DB to its latest state."`
		Help        bool   `env:"-"                     help:"Show help."`
	}{
		PostgresURL: "postgres://test:test@localhost:5432/hds?sslmode=disable",
		Sources:     "file://data/migrations",
	}

	cli.Register().
		Help("Launches SQL migration tool to update a Postgres database schema.").
		Options(&cfg)
	cli.Load()

	m, err := migrate.New(cfg.Sources, cfg.PostgresURL)
	if err != nil {
		logrus.WithError(err).Error("creation migration failed")
		return
	}

	if cfg.Steps != 0 {
		err = m.Steps(cfg.Steps)
	} else {
		err = m.Up()
	}
	if err != nil && err != migrate.ErrNoChange {
		logrus.WithError(err).Error("migration failed")
	}
}

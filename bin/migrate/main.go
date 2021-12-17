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
		URL     string `env:"MIGGRATE_URL"     help:"The url to connect to Postgres."`
		Sources string `env:"MIGGRATE_SOURCES" help:"The directory where the migration files are located."`
		Steps   int    `env:"MIGRATE_STEPS"    help:"The steps forward or backward to migrate to. 0 migrates the DB to its latest state."`
		Help    bool   `env:"-"                help:"Show help."`
	}{
		Sources: "file://data/migrations",
	}

	cli.Register().
		Help("Launches SQL migration tool to update a Postgres database schema.").
		Options(&cfg)
	cli.Load()

	if cfg.URL == "" {
		logrus.Fatal("database url is not set. \033[2m=> migrate -url [postgres_url]\033[00m")
	}

	m, err := migrate.New(cfg.Sources, cfg.URL)
	if err != nil {
		logrus.WithError(err).Error("creating migration failed")
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

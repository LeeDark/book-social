package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/LeeDark/book-social/internal/config"
	"github.com/LeeDark/book-social/internal/coverops"
	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/storage/covers"
	"github.com/LeeDark/book-social/internal/storage/postgresql"
	"github.com/LeeDark/book-social/internal/storage/sqlite"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, output io.Writer) error {
	if len(args) == 0 || (args[0] != "install" && args[0] != "verify") {
		return errors.New("usage: coverctl install|verify -writers-stopped [install flags: -bundle DIR -manifest-sha256 HASH -actor-id ID]")
	}

	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(output)
	isWriterStopConfirmed := flags.Bool("writers-stopped", false, "confirm all cover writers are stopped")
	var bundleDir, manifestChecksum string
	var actorID int
	if args[0] == "install" {
		flags.StringVar(&bundleDir, "bundle", "", "directory with manifest.json and prepared images")
		flags.StringVar(&manifestChecksum, "manifest-sha256", "", "independently pinned manifest checksum")
		flags.IntVar(&actorID, "actor-id", 0, "existing administrator ID for the audit")
	}

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if !*isWriterStopConfirmed || flags.NArg() != 0 {
		return errors.New("stop every cover writer and pass -writers-stopped without positional arguments")
	}
	if strings.TrimSpace(os.Getenv("APP_DB_DSN")) == "" || strings.TrimSpace(os.Getenv("APP_COVERS_DIR")) == "" {
		return errors.New("explicit APP_DB_DSN and APP_COVERS_DIR are required for maintenance")
	}

	var bundle *coverops.Bundle
	if args[0] == "install" {
		if actorID <= 0 || bundleDir == "" || manifestChecksum == "" {
			return errors.New("install requires -bundle, -manifest-sha256 and a positive -actor-id")
		}

		var err error
		bundle, err = coverops.LoadBundle(bundleDir, manifestChecksum)
		if err != nil {
			return err
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, catalog, operator, err := openRepositories(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if args[0] == "verify" {
		if info, err := os.Stat(cfg.Covers.Dir); err != nil || !info.IsDir() {
			return errors.New("verify requires an existing cover directory")
		}
	}

	store, err := covers.New(cfg.Covers.Dir)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	if args[0] == "verify" {
		count, err := coverops.Verify(ctx, catalog, store)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(output, "Verified %d local cover references (including optional thumbnails).\n", count)
		return err
	}

	changed, err := coverops.NewInstaller(catalog, operator, store).Install(ctx, bundle, actorID)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(output, "Installed bundle %q: %d front covers changed.\n", bundle.ID, changed)
	return err
}

func openRepositories(ctx context.Context, cfg config.Config) (*sql.DB, books.BookRepository, books.OperatorRepository, error) {
	if cfg.Env == config.EnvDev {
		// Do not create or seed an accidental empty SQLite target during maintenance.
		path := strings.SplitN(strings.TrimPrefix(cfg.DB.DSN, "file:"), "?", 2)[0]
		if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
			return nil, nil, nil, errors.New("cover maintenance requires an existing migrated SQLite database")
		}

		db, err := sqlite.Open(ctx, cfg.DB.DSN)
		if err != nil {
			return nil, nil, nil, err
		}

		return db, sqlite.NewBookRepository(db), sqlite.NewCatalogOperatorRepository(db), nil
	}

	db, err := postgresql.Open(ctx, cfg.DB.DSN)
	if err != nil {
		return nil, nil, nil, err
	}

	return db, postgresql.NewBookRepository(db), postgresql.NewCatalogOperatorRepository(db), nil
}

package postgres

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/lib/pq"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const MIGRATION_TABLE = "meters_migrations"
const SEED_MIGRATION_TABLE = "meters_seed"

//go:embed scripts/migrations
var postgresMigrationFiles embed.FS

//go:embed scripts/seed
var postgresSeedFiles embed.FS

func MigrateDB_UsingConnection_Postgres(sqlDB *sql.DB, files embed.FS, directoryInFS string, migrationsTable string) error {
	_files, err := iofs.New(files, directoryInFS)
	if err != nil {
		log.Fatal(err)
	}

	dbInstance, err := postgres.WithInstance(sqlDB, &postgres.Config{MigrationsTable: migrationsTable})
	if err != nil {
		slog.Error("error", "err", err)
		return err
	}

	m, err := migrate.NewWithInstance("iofs", _files, "DUMMY", dbInstance)

	if err != nil {
		slog.Error("error", "err", err)
		return fmt.Errorf("failed creating new migration: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("error", "err", err)
		return fmt.Errorf("failed while migrating: %w", err)
	}

	return nil
}

func MigratePostgres(dbURL string) error {
	slog.Info("ChatService: Connecting to PostgreSQL database")
	sqlDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("error", "err", err)
		return err
	}
	defer sqlDB.Close()

	return MigrateDB_UsingConnection_Postgres(sqlDB, postgresMigrationFiles, "db/postgres/scripts/migrations", "chatservice_postgres_migrations")
}

func SeedPostgres(dbURL string) error {
	slog.Info("ChatService: Seeding PostgreSQL database", "dbURL", dbURL)
	sqlDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("error", "err", err)
		return err
	}
	defer sqlDB.Close()

	return MigrateDB_UsingConnection_Postgres(sqlDB, postgresSeedFiles, "db/postgres/scripts/seed", "chatservice_postgres_seed")
}

// GetMigrationFiles returns the embedded migration files
func GetMigrationFiles() embed.FS {
	return postgresMigrationFiles
}

// MigrateDatabase runs the database migrations using the embedded files
func MigrateDatabase(dbURL string) error {
	slog.Info("Running database migrations")
	sqlDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer sqlDB.Close()

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return MigrateDB_UsingConnection_Postgres(
		sqlDB,
		postgresMigrationFiles,
		"scripts/migrations",
		MIGRATION_TABLE,
	)
}

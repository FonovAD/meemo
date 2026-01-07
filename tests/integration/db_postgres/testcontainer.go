package db_postgres

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testDBName   = "testdb"
	testUser     = "testuser"
	testPassword = "testpassword"
)

type TestPostgresContainer struct {
	Container *tcpostgres.PostgresContainer
	DB        *sqlx.DB
}

func SetupPostgresContainer(t *testing.T) (*TestPostgresContainer, func()) {
	ctx := context.Background()

	postgresContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase(testDBName),
		tcpostgres.WithUsername(testUser),
		tcpostgres.WithPassword(testPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to postgres: %v", err)
	}

	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	tc := &TestPostgresContainer{
		Container: postgresContainer,
		DB:        db,
	}

	cleanup := func() {
		if err := cleanupTables(db); err != nil {
			t.Logf("Warning: failed to cleanup tables: %v", err)
		}
		if err := db.Close(); err != nil {
			t.Logf("Warning: failed to close db: %v", err)
		}
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("Warning: failed to terminate container: %v", err)
		}
	}

	return tc, cleanup
}

func runMigrations(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrationsPath := getMigrationsPath()

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func getMigrationsPath() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "..", "..", "..", "migrations")
}

func cleanupTables(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// TODO: Добавить обработку ошибки
	defer func() { _ = tx.Rollback() }()

	tables := []string{"files", "users"}
	for _, table := range tables {
		if _, err := tx.Exec("TRUNCATE TABLE " + table + " CASCADE"); err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	return tx.Commit()
}

func initTestDB(t *testing.T) (*sqlx.DB, func()) {
	tc, cleanup := SetupPostgresContainer(t)
	return tc.DB, cleanup
}

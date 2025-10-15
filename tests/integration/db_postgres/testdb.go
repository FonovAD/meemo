package db_postgres

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	"log"
	pg "meemo/internal/infrastructure/storage/pg"
	"os"
	"path/filepath"
	"testing"
)

func SetupTestDB(t *testing.T, config pg.PGConfig) (*sqlx.DB, func()) {
	db, err := sqlx.Open("postgres",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode))
	if err != nil {
		t.Fatal(err)
	}

	if err := runMigration(db); err != nil {
		t.Fatal(err)
	}

	return db, func() {
		if err := cleanupDB(db); err != nil {
			t.Fatal(err)
		}
		db.Close()
	}
}

func runMigration(db *sqlx.DB) error {
	log.Println("Start migration")

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return err
	}
	log.Println("driver")

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "../../../migrations"
	}

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+absPath,
		"postgres", driver,
	)
	log.Println("migrate.NewWithInstance")
	if err != nil {
		return err
	}
	err = m.Up()
	log.Println("m.Up")
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	log.Println("Миграции успешно применены")
	return nil
}

func cleanupDB(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Commit()
	query := `
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = 'public' 
    `
	rows, err := tx.Query(query)
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating table names: %w", err)
	}
	if len(tables) == 0 {
		return tx.Commit()
	}
	for _, table := range tables {
		_, err := tx.Exec("DROP TABLE IF EXISTS " + table + " CASCADE")
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

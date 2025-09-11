package utils

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// RunMigrations runs database migrations
func RunMigrations(databaseURL string) error {
	// Open database connection
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create migration driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migration instance
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		// Attempt to auto-recover if database is dirty
		if v, isDirty, verErr := m.Version(); verErr == nil && isDirty {
			if forceErr := m.Force(int(v)); forceErr != nil {
				return fmt.Errorf("failed to force migration version %d clean: %w", v, forceErr)
			}
			if retryErr := m.Up(); retryErr != nil && retryErr != migrate.ErrNoChange {
				return fmt.Errorf("failed to run migrations after force-clean (version %d): %w", v, retryErr)
			}
			log.Println("Database migrations completed successfully after clearing dirty state")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// RollbackMigrations rolls back the last migration
func RollbackMigrations(databaseURL string) error {
	// Open database connection
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create migration driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migration instance
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()

	// Rollback last migration
	if err := m.Steps(-1); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	log.Println("Database migration rollback completed successfully")
	return nil
}

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load()

	up := flag.Bool("up", false, "Run all pending migrations")
	down := flag.Bool("down", false, "Rollback all migrations")
	steps := flag.Int("steps", 0, "Number of migration steps (positive=up, negative=down)")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Println("DATABASE_URL is required")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		fmt.Printf("Failed to create migration driver: %v\n", err)
		os.Exit(1)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://database/migrations",
		"postgres", driver,
	)
	if err != nil {
		fmt.Printf("Failed to create migrator: %v\n", err)
		os.Exit(1)
	}

	if *up {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			fmt.Printf("Migration up failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations applied successfully")
	} else if *down {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			fmt.Printf("Migration down failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("All migrations rolled back")
	} else if *steps != 0 {
		if err := m.Steps(*steps); err != nil && err != migrate.ErrNoChange {
			fmt.Printf("Migration steps failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Applied %d migration steps\n", *steps)
	} else {
		version, dirty, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				fmt.Println("No migrations applied yet")
			} else {
				fmt.Printf("Failed to get version: %v\n", err)
			}
		} else {
			fmt.Printf("Current version: %d, dirty: %v\n", version, dirty)
		}
		fmt.Println("\nUsage:")
		fmt.Println("  -up      Run all pending migrations")
		fmt.Println("  -down    Rollback all migrations")
		fmt.Println("  -steps N Apply N migration steps (positive=up, negative=down)")
	}
}

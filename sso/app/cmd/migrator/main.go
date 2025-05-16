package main

import (
	"errors"
	"flag"
	"fmt"

	// Библиотека для миграций
	"github.com/golang-migrate/migrate/v4"
	// Драйвер для выполнения миграций в postgres
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Драйвер для получения миграций из файлов
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var storagePath, migrationsPath, migrationsTable, command string

	flag.StringVar(&storagePath, "storage-path", "", "path to the storage")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to the migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of the migrations table")
	flag.StringVar(&command, "command", "", "migration command (up, down)")
	flag.Parse()

	if storagePath == "" {
		panic("storage-path is required")
	}

	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	if command == "" {
		panic("command is required")
	}

	sourceUrl := fmt.Sprintf("file://%s", migrationsPath)
	databaseUrl := fmt.Sprintf("postgres://%s?sslmode=disable&x-migrations-table=%s", storagePath, migrationsTable)

	m, err := migrate.New(sourceUrl, databaseUrl)
	if err != nil {
		panic(err)
	}

	if command == "up" {
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")

				return
			}
			panic(err)
		}
	} else if command == "down" {
		if err := m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to rollback")
				return
			}
			panic(err)
		}
	} else {
		panic("Unknown command. Command must be up or down")
	}

	fmt.Println("migrations applied successfully")
}

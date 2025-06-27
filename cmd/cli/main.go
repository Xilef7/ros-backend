package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"restaurant-ordering-system/internal/pkg/config"

	"github.com/jackc/pgx/v5"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cli [migrate|seed]")
		os.Exit(1)
	}

	cmd := os.Args[1]

	// Load configuration
	cfg, err := config.LoadConfig("configs/config.json")
	if err != nil {
		os.Exit(1)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())
	if err := conn.Ping(context.Background()); err != nil {
		fmt.Printf("failed to ping database: %v\n", err)
		os.Exit(1)
	}

	switch cmd {
	case "migrate":
		doMigrate(conn)
	case "seed":
		doSeed(conn)
	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}
}

func doMigrate(conn *pgx.Conn) {
	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		fmt.Println("Failed to list migration files:", err)
		os.Exit(1)
	}
	for _, file := range files {
		if strings.Contains(filepath.Base(file), "seed") {
			continue
		}
		fmt.Printf("Running migration: %s\n", file)
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Failed to read %s: %v\n", file, err)
			os.Exit(1)
		}
		if _, err := conn.Exec(context.Background(), string(content)); err != nil {
			fmt.Printf("Migration failed for %s: %v\n", file, err)
			os.Exit(1)
		}
	}
	fmt.Println("Migration complete.")
}

func doSeed(db *pgx.Conn) {
	files, err := filepath.Glob("migrations/*seed*.sql")
	if err != nil {
		fmt.Println("Failed to list seed files:", err)
		os.Exit(1)
	}
	for _, file := range files {
		fmt.Printf("Running seed: %s\n", file)
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Failed to read %s: %v\n", file, err)
			os.Exit(1)
		}
		if _, err := db.Exec(context.Background(), string(content)); err != nil {
			fmt.Printf("Seeding failed for %s: %v\n", file, err)
			os.Exit(1)
		}
	}
	fmt.Println("Seeding complete.")
}

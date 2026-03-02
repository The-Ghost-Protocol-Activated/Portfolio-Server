package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func InitDB() (*sql.DB, error) {
	var dsn string

	// Render injects DATABASE_URL for managed Postgres — use it directly when present.
	if url := os.Getenv("DATABASE_URL"); url != "" {
		dsn = url
	} else {
		// Fallback: build DSN from individual variables (local dev).
		host := os.Getenv("DB_HOST")
		if host == "" {
			host = "localhost"
		}

		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "5432"
		}

		user := os.Getenv("DB_USER")
		if user == "" {
			user = "postgres"
		}

		password := os.Getenv("DB_PASSWORD")
		if password == "" {
			password = "postgres"
		}

		dbname := os.Getenv("DB_NAME")
		if dbname == "" {
			dbname = "blogdb"
		}

		// sslmode=disable is fine for local; Render's DATABASE_URL already
		// carries the correct sslmode=require parameter.
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname,
		)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS blog_posts (
		id SERIAL PRIMARY KEY,
		slug VARCHAR(255) UNIQUE NOT NULL,
		title VARCHAR(255) NOT NULL,
		content TEXT NOT NULL,
		excerpt TEXT,
		published_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_published BOOLEAN DEFAULT false,
		tags TEXT[]
	);

	CREATE INDEX IF NOT EXISTS idx_blog_posts_slug ON blog_posts(slug);
	CREATE INDEX IF NOT EXISTS idx_blog_posts_published ON blog_posts(is_published, published_at DESC);
	`

	_, err := db.Exec(query)
	return err
}

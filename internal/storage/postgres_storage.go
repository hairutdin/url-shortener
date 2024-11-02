package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

const createTableQuery = `
CREATE TABLE IF NOT EXISTS shortened_urls (
    uuid UUID PRIMARY KEY,
    short_url VARCHAR(255) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

type PostgresStorage struct {
	db *pgx.Conn
}

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	db, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	if _, err := db.Exec(context.Background(), createTableQuery); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	return &PostgresStorage{db: db}, nil
}

func (p *PostgresStorage) CreateShortURL(uuid, shortURL, originalURL string) error {
	_, err := p.db.Exec(context.Background(),
		"INSERT INTO shortened_urls (uuid, short_url, original_url) VALUES ($1, $2, $3)",
		uuid, shortURL, originalURL)
	return err
}

func (p *PostgresStorage) GetOriginalURL(shortURL string) (string, error) {
	var originalURL string
	err := p.db.QueryRow(context.Background(),
		"SELECT original_url FROM shortened_urls WHERE short_url=$1", shortURL).
		Scan(&originalURL)
	return originalURL, err
}

func (p *PostgresStorage) Ping() error {
	return p.db.Ping(context.Background())
}

func (p *PostgresStorage) Close() error {
	return p.db.Close(context.Background())
}

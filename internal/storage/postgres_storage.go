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

func (p *PostgresStorage) CreateBatchURLs(urls []BatchURLRequest) ([]BatchURLOutput, error) {
	tx, err := p.db.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	outputs := make([]BatchURLOutput, 0, len(urls))

	for _, url := range urls {
		_, err := tx.Exec(context.Background(),
			`INSERT INTO shortened_urls (uuid, short_url, original_url) VALUES ($1, $2, $3)`,
			url.UUID, url.ShortURL, url.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("failed to insert batch URL: %w", err)
		}

		output := BatchURLOutput{
			CorrelationID: url.UUID,
			ShortURL:      url.ShortURL,
		}
		outputs = append(outputs, output)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return outputs, nil
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

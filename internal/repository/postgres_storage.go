package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgx/v5"
)

const createTableQuery = `
CREATE TABLE IF NOT EXISTS shortened_urls (
    uuid UUID PRIMARY KEY,
    short_url VARCHAR(255) UNIQUE NOT NULL,
    original_url TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

type PostgresStorage struct {
	DB *pgx.Conn
}

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	DB, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	if _, err := DB.Exec(context.Background(), createTableQuery); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &PostgresStorage{DB: DB}, nil
}

func (p *PostgresStorage) CreateShortURL(uuid, shortURL, originalURL string) (string, error) {
	const query = `
		INSERT INTO shortened_urls (uuid, short_url, original_url)
		VALUES ($1, $2, $3)
		ON CONFLICT (original_url) DO NOTHING
		RETURNING short_url;
	`

	var existingShortURL string
	err := p.DB.QueryRow(context.Background(), query, uuid, shortURL, originalURL).Scan(&existingShortURL)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			existingShortURL, err := p.GetShortURLByOriginal(originalURL)
			if err != nil {
				return "", fmt.Errorf("failed to fetch existing short URL: %w", err)
			}
			return existingShortURL, fmt.Errorf("URL already exists with short URL: %s", existingShortURL)
		}
		return "", err
	}
	return existingShortURL, nil
}

func (p *PostgresStorage) GetShortURLByOriginal(originalURL string) (string, error) {
	const query = `SELECT short_url FROM shortened_urls WHERE original_url = $1`
	var shortURL string
	err := p.DB.QueryRow(context.Background(), query, originalURL).Scan(&shortURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to get short URL: %w", err)
	}
	return shortURL, nil
}

func (p *PostgresStorage) CreateBatchURLs(urls []BatchURLRequest) ([]BatchURLOutput, error) {
	tx, err := p.DB.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(context.Background()); err != nil && err != pgx.ErrTxClosed {
			fmt.Printf("rollback failed: %v\n", err)
		}
	}()

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
	err := p.DB.QueryRow(context.Background(),
		"SELECT original_url FROM shortened_urls WHERE short_url=$1", shortURL).
		Scan(&originalURL)
	return originalURL, err
}

func (p *PostgresStorage) Ping() error {
	return p.DB.Ping(context.Background())
}

func (p *PostgresStorage) Close() error {
	return p.DB.Close(context.Background())
}

package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
)

var DB *pgx.Conn

func ConnectDB(databaseDSN string) error {
	var err error
	DB, err = pgx.Connect(context.Background(), databaseDSN)
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		return err
	}

	err = DB.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("failed to ping the database: %v", err)
	}

	log.Println("Successfully connected to the database")
	return nil
}

func PingDB() error {
	return DB.Ping(context.Background())
}

func CloseDB() {
	if DB != nil {
		DB.Close(context.Background())
	}
}

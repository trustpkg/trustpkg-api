package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	goDotEnv "github.com/joho/godotenv"
)

type postgresConfig struct {
	host     string
	port     string
	user     string
	password string
	dbName   string
}

var Pool *pgxpool.Pool

func ConnectDb() {
	if err := goDotEnv.Load(".env"); err != nil {
		log.Println(".env file not found, using system environment variables")
	}

	dsn := strings.TrimSpace(os.Getenv("DB_URL"))
	if dsn == "" {
		var dbConfig = postgresConfig{
			host:     os.Getenv("POSTGRES_HOST"),
			port:     os.Getenv("POSTGRES_PORT"),
			user:     os.Getenv("POSTGRES_USER"),
			password: os.Getenv("POSTGRES_PASSWORD"),
			dbName:   os.Getenv("POSTGRES_DB"),
		}

		if dbConfig.host == "" || dbConfig.port == "" || dbConfig.user == "" || dbConfig.dbName == "" {
			fmt.Fprintln(os.Stderr, "Missing DB config. Set DB_URL or POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB (copy .env.example to .env).")
			os.Exit(1)
		}

		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbConfig.user, dbConfig.password, dbConfig.host, dbConfig.port, dbConfig.dbName)
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = pool.Ping(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to ping database: %v\n", err)
		os.Exit(1)
	}

	Pool = pool
	
	log.Println("Connected to postgres database successfully")
}
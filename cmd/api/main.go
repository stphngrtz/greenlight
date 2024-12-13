package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stphngrtz/greenlight/internal/data"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

type application struct {
	config config
	logger *slog.Logger
	models data.Models
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	err := godotenv.Load()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	var config config

	flag.IntVar(&config.port, "port", 4000, "API server port")
	flag.StringVar(&config.env, "env", "development", "Environment (development, staging, production)")
	flag.StringVar(&config.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&config.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&config.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&config.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")
	flag.Float64Var(&config.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&config.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&config.limiter.enabled, "limiter-enabled", true, "Rate limiter enabled")
	flag.Parse()

	db, err := openDB(config)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	app := &application{
		config: config,
		logger: logger,
		models: data.NewModels(db),
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(config config) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(config.db.maxOpenConns)
	db.SetMaxIdleConns(config.db.maxIdleConns)
	db.SetConnMaxIdleTime(config.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

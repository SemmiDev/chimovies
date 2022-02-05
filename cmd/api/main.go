package main

import (
	"context"
	"database/sql"
	"expvar"
	"os"
	"runtime"
	"time"

	"github.com/SemmiDev/chimovies/config"
	"github.com/SemmiDev/chimovies/internal/data"
	"github.com/SemmiDev/chimovies/internal/jsonlog"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type app struct {
	config config.Config
	logger *jsonlog.Logger
	models data.Models
	// mailer mailer.Mailer
}

func main() {
	config := config.LoadConfig(".")
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(config)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	logger.PrintInfo("database connection pool established", nil)
	expvar.NewString("version").Set(config.Version)
	expvar.NewString("buidltime").Set(config.BuildTime)
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	app := &app{
		config: config,
		logger: logger,
		models: data.NewModels(db),
		// mailer: mailer.New(config.SMTPHost, config.SMTPPort, config.Username, config.Password, config.Sender),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

func openDB(config config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", config.PostgreDSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxIdleTime(config.MaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

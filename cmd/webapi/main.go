package main

import (
	"context"
	"database/sql"
	"errors"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ardanlabs/conf"
	"github.com/femito1/WASA/service/api"
	"github.com/femito1/WASA/service/database"
	"github.com/femito1/WASA/service/globaltime"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := run(); err != nil {
		logrus.WithError(err).Error("application terminated with error")
		os.Exit(1)
	}
}

func run() error {
	// Seed the random number generator.
	rand.Seed(globaltime.Now().UnixNano())

	// Load configuration.
	cfg, err := loadConfiguration()
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			return nil
		}
		return err
	}

	// Initialize logging.
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	if cfg.Debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}
	logger.Info("application initializing")

	// Start the database.
	logger.Info("initializing database support")
	dbconn, err := sql.Open("sqlite3", cfg.DB.Filename)
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return err
	}
	defer func() {
		logger.Debug("database stopping")
		_ = dbconn.Close()
	}()
	db, err := database.New(dbconn)
	if err != nil {
		logger.WithError(err).Error("error creating AppDatabase")
		return err
	}

	// Start the API server.
	logger.Info("initializing API server")
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	serverErrors := make(chan error, 1)

	apirouter, err := api.New(api.Config{
		Logger:   logger,
		Database: db,
	})
	if err != nil {
		logger.WithError(err).Error("error creating the API server instance")
		return err
	}
	router := apirouter.Handler()

	router, err = registerWebUI(router)
	if err != nil {
		logger.WithError(err).Error("error registering web UI handler")
		return err
	}

	router = applyCORSHandler(router)

	apiserver := &http.Server{
		Addr:              cfg.Web.APIHost,
		Handler:           router,
		ReadTimeout:       cfg.Web.ReadTimeout,
		ReadHeaderTimeout: cfg.Web.ReadTimeout,
		WriteTimeout:      cfg.Web.WriteTimeout,
	}

	// Start the API server in a goroutine.
	go func() {
		logger.Infof("API listening on %s", apiserver.Addr)
		err := apiserver.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Error("ListenAndServe error")
			serverErrors <- err
		} else {
			serverErrors <- nil
		}
		logger.Info("stopping API server")
	}()

	// Wait for a shutdown signal or server error.
	select {
	case err := <-serverErrors:
		if err != nil {
			return err
		}
	case sig := <-shutdown:
		logger.Infof("signal %v received, start shutdown", sig)

		if err := apirouter.Close(); err != nil {
			logger.WithError(err).Warn("graceful shutdown of API router error")
		}

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := apiserver.Shutdown(ctx); err != nil {
			logger.WithError(err).Warn("error during graceful shutdown of HTTP server")
			_ = apiserver.Close()
		}

		if sig == syscall.SIGSTOP {
			return errors.New("integrity issue caused shutdown")
		}
	}

	return nil
}

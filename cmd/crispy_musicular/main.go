package main

import (
	"context"
	"io"
	"os"
	"path"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	"github.com/hoffs/crispy-musicular/pkg/backup"
	"github.com/hoffs/crispy-musicular/pkg/backup/actions"
	"github.com/hoffs/crispy-musicular/pkg/config"
	"github.com/hoffs/crispy-musicular/pkg/http"
	"github.com/hoffs/crispy-musicular/pkg/storage"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	err := os.MkdirAll("log", 0777)
	if err != nil {
		log.Error().Err(err).Msg("failed to create log directory")
		return
	}

	logFile, err := os.OpenFile(path.Join(getEnv("LOG_DIR", "log"), "crispy.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Error().Err(err).Msg("failed to open log file")
		return
	}

	log.Logger = log.Output(io.MultiWriter(os.Stdout, logFile))
	// load .env before anything else
	_ = godotenv.Load()
	_ = godotenv.Load(".env.local")

	conf, err := config.Load(getEnv("CONFIG_PATH", "conf.yaml"))
	if err != nil {
		log.Error().Err(err).Msg("failed to load config")
		return
	}

	r, err := storage.NewRepository(conf.DbPath)
	if err != nil {
		log.Error().Err(err).Msg("failed to load database")
		return
	}

	auth, err := auth.NewService(r)
	if err != nil {
		log.Error().Err(err).Msg("failed to create auth service")
		return
	}

	jsonBackup, err := actions.NewJsonBackupAction(conf)
	if err != nil {
		log.Error().Err(err).Msg("failed to create backuper json action")
		return
	}

	driveBackup, err := actions.NewGoogleDriveBackupAction(conf, auth)
	if err != nil {
		log.Error().Err(err).Msg("failed to create backuper google drive action")
		return
	}

	backuper, err := backup.NewBackuper(conf, auth, r, jsonBackup, driveBackup)
	if err != nil {
		log.Error().Err(err).Msg("failed to create backuper")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go backuper.RunPeriodically(ctx)

	// this is blocking
	err = http.RegisterHandlers(conf, auth, backuper)
	if err != nil {
		log.Error().Err(err).Msg("failed to register handlers")
		return
	}
}

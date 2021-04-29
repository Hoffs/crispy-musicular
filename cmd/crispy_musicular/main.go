package main

import (
	"github.com/hoffs/crispy-musicular/pkg/auth"
	"github.com/hoffs/crispy-musicular/pkg/backup"
	"github.com/hoffs/crispy-musicular/pkg/config"
	"github.com/hoffs/crispy-musicular/pkg/http"
	"github.com/hoffs/crispy-musicular/pkg/storage"
	"github.com/rs/zerolog/log"
)

func main() {
	r, err := storage.NewRepository("data.db")
	if err != nil {
		log.Error().Err(err).Msg("failed to load database")
		return
	}

	conf, err := config.Load("conf.yaml")
	if err != nil {
		log.Error().Err(err).Msg("failed to load config")
		return
	}

	auth, err := auth.NewService(r)
	if err != nil {
		log.Error().Err(err).Msg("failed to create auth service")
		return
	}

	backuper, err := backup.NewBackuper(conf, auth, r)
	if err != nil {
		log.Error().Err(err).Msg("failed to create backuper")
		return
	}

	err = http.RegisterHandlers(conf, auth, backuper)
	if err != nil {
		log.Error().Err(err).Msg("failed to register handlers")
		return
	}
}

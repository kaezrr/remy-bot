package main

import (
	"os"

	"github.com/kaezrr/remy-bot/internal/bot"
	"github.com/kaezrr/remy-bot/internal/config"
	"github.com/kaezrr/remy-bot/internal/store"
	"github.com/kaezrr/remy-bot/internal/wa"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("Remy starting up...")

	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load file")
	}

	s, err := store.NewDBStore(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start database")
	}

	// In memory store for testing
	// s := store.NewMemStore()

	if err := wa.Run(cfg, s, bot.Handle); err != nil {
		log.Fatal().Err(err).Msg("whatsapp runtime error")
	}
}

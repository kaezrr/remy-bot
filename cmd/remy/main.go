package main

import (
	"os"
	"time"

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

	// Create session directory and data directory
	if err := os.MkdirAll("data/session", 0755); err != nil {
		log.Fatal().Err(err).Msg("failed to create session directory")
	}

	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config file")
	}

	timezone, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid timezone")
	}

	log.Info().Str("timezone", cfg.Timezone).Msg("Using configured timezone")
	name, offset := time.Now().UTC().In(timezone).Zone()
	log.Info().
		Str("zone", name).
		Int("offset_seconds", offset).
		Msg("Timezone info")

	s, err := store.NewDBStore(cfg.Database, timezone)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start database")
	}

	if err := wa.Run(cfg, s, bot.Handle); err != nil {
		log.Fatal().Err(err).Msg("whatsapp runtime error")
	}
}

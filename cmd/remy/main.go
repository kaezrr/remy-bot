package main

import (
	"os"

	"github.com/kaezrr/remy-bot/internal/bot"
	"github.com/kaezrr/remy-bot/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("Remy starting up...")

	cfg, err := config.Load("config/config.json")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load file")
	}

	log.Info().
		Str("database", cfg.Database).
		Str("session_dir", cfg.SessionDir).
		Str("prefix", cfg.Prefix).
		Msg("config loaded")

	resp := bot.Handle(".d", cfg.Prefix)
	log.Info().Msg(resp.Text)
}

package main

import (
	"os"

	"github.com/kaezrr/remy-bot/internal/bot"
	"github.com/kaezrr/remy-bot/internal/config"
	"github.com/kaezrr/remy-bot/internal/store"
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

	s := store.New()

	tests := []string{
		// -----------------
		// HELP COMMAND
		// -----------------
		".h",
		".x", // invalid top-level command
		".",  // prefix only

		// -----------------
		// BASKET COMMANDS
		// -----------------
		".b",          // no args
		".b list",     // empty list
		".b add",      // missing name
		".b add dbms", // good add
		".b add DBMS", // should error (already exists but case-insensitive)
		".b add os",   // another basket
		".b list",     // list 2 baskets
		".b del",      // missing name
		".b del math", // non-existent basket
		".b del dbms", // delete existing
		".b list",     // confirm deletion

		// -----------------
		// DEADLINE COMMANDS
		// -----------------
		".d",                                    // no args
		".d list",                               // empty list
		".d add",                                // missing args
		".d add 2025-02-01",                     // missing time + title
		".d add 2025-02-01 23:59",               // missing title
		".d add bad-date 10:00 test",            // invalid date
		".d add 2025-02-01 bad test",            // invalid time
		".d add 2025-02-01 23:59 OS Assignment", // success
		".d add 2025-03-05 09:00 DBMS Quiz",     // second
		".d list",                               // list both
		".d del",                                // missing id
		".d del abc",                            // invalid id
		".d del 99",                             // nonexistent id
		".d del 1",                              // delete valid
		".d list",                               // confirm deletion

		// -----------------
		// PIN COMMANDS
		// -----------------
		".p",                      // no args
		".p list",                 // missing basket
		".p list os",              // basket exists but empty
		".p add",                  // missing args
		".p add os",               // missing content
		".p add os link to notes", // good add (multi-word)
		".p add os lab sheet",     // second pin
		".p list os",              // should show 2 pins
		".p del",                  // missing basket + id
		".p del os",               // missing id
		".p del os abcd",          // non-integer id
		".p del math 1",           // basket doesn't exist
		".p del os 99",            // pin doesn't exist
		".p del os 1",             // delete valid pin
		".p list os",              // confirm deletion
	}

	for _, cmd := range tests {
		resp := bot.Handle(cmd, cfg.Prefix, s)
		log.Info().Msgf("cmd: %s\n%s", cmd, resp.Text)
	}
}

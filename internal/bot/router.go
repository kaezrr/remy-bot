package bot

import (
	"errors"
	"strings"

	"github.com/kaezrr/remy-bot/internal/store"
	"github.com/rs/zerolog/log"
)

const HELP = `Usage:
.d	deadlines
.b	Manage baskets
.p	Manage pins
.h	Print this message

Use <cmd> to list further commands`

const DEADLINE_HELP = `Usage:
.d list						List all pending deadlines
.d remove <id>				Remove a deadline
.d add <date> <title>		Add a new deadline`

const BASKET_HELP = `Usage:
.b list						List all baskets
.b remove <name>			Remove a basket
.b add <name>				Add a new basket`

const PIN_HELP = `Usage:
.p list <basket>			List all pins of a basket
.p remove <basket> <id>		Remove a pin from a basket
.p add <basket> <title>		Add a new pin`

type Response struct {
	Text string
}

func Handle(input string, prefix string, s *store.Store) Response {
	after, found := strings.CutPrefix(input, prefix)

	if !found {
		return Response{Text: ""}
	}

	parts := strings.Fields(after)

	if len(parts) == 0 {
		return Response{Text: HELP}
	}

	switch parts[0] {
	case "d":
		return Response{Text: "deadline commands coming soon"}

	case "b":
		result, err := basketHandler(parts[1:], s)
		if err != nil {
			log.Error().Err(err).Msg("basket handler error")
			return Response{Text: err.Error()}
		}
		return Response{Text: result}

	case "p":
		return Response{Text: "pin commands coming soon"}

	case "h":
		return Response{Text: HELP}
	}

	return Response{Text: HELP}
}

func basketHandler(parts []string, s *store.Store) (string, error) {
	if len(parts) == 0 {
		return BASKET_HELP, nil
	}

	// Which action to take
	switch parts[0] {
	case "add":
		if len(parts) < 2 {
			return "", errors.New("basket name required")
		}
		name := parts[1]

		err := s.AddBasket(name)
		if err != nil {
			return "", err
		}

		return "basket created successfully", nil

	case "list":
		baskets := s.ListBaskets()
		if len(baskets) == 0 {
			return "there are no baskets", nil
		}

		out := "List of baskets:\n"
		for _, b := range baskets {
			out += "- " + b + "\n"
		}

		return out, nil

	case "remove":
		if len(parts) < 2 {
			return "", errors.New("basket name required")
		}
		name := parts[1]

		err := s.DeleteBasket(name)
		if err != nil {
			return "", err
		}

		return "basket deleted successfully", nil
	}

	return BASKET_HELP, nil
}

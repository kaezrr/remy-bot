package bot

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kaezrr/remy-bot/internal/store"
	"github.com/rs/zerolog/log"
)

const HELP = `Usage:
.d  deadlines
.b  Manage baskets
.p  Manage pins
.h  Print this message

Use <cmd> to list further commands`

const DEADLINE_HELP = `Usage:
.d list                       list all deadlines
.d del <id>                   remove a deadline
.d add <date> <time> <title>  add a new deadline
`

const BASKET_HELP = `Usage:
.b list            list all baskets
.b add <name>      add a new basket
.b del <name>      remove a basket`

const PIN_HELP = `Usage:
.p list <basket>              list all pins in a basket
.p add <basket> <content>     add a new pin
.p del <basket> <id>          remove a pin from a basket`

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
		result, err := deadlineHandler(parts[1:], s)
		if err != nil {
			log.Error().Err(err).Msg("deadline handler error")
			return Response{Text: err.Error()}
		}
		return Response{Text: result}

	case "b":
		result, err := basketHandler(parts[1:], s)
		if err != nil {
			log.Error().Err(err).Msg("basket handler error")
			return Response{Text: err.Error()}
		}
		return Response{Text: result}

	case "p":
		result, err := pinHandler(parts[1:], s)
		if err != nil {
			log.Error().Err(err).Msg("pin handler error")
			return Response{Text: err.Error()}
		}
		return Response{Text: result}

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

		out := "list of baskets:\n"
		for _, b := range baskets {
			out += "- " + b + "\n"
		}

		return out, nil

	case "del":
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

func deadlineHandler(parts []string, s *store.Store) (string, error) {
	if len(parts) == 0 {
		return DEADLINE_HELP, nil
	}

	switch parts[0] {
	case "list":
		deadlines := s.ListDeadlines()
		if len(deadlines) == 0 {
			return "there are no pending deadlines", nil
		}

		out := "upcoming deadlines:\n"
		for _, d := range deadlines {
			out += fmt.Sprintf("%d. %s (%s)\n", d.ID, d.Title, d.DateTime)
		}

		return out, nil

	case "add":
		if len(parts) < 4 {
			return "", errors.New("usage: .d add <date> <time> <title>")
		}

		date := parts[1]
		time := parts[2]
		title := strings.Join(parts[3:], " ")

		d := s.AddDeadline(title, date+" "+time)

		return fmt.Sprintf("deadline #%d added: %s (%s)", d.ID, d.Title, d.DateTime), nil

	case "del":
		if len(parts) < 2 {
			return "", errors.New("deadline id required")
		}
		idStr := parts[1]
		id, err := strconv.Atoi(idStr)

		if err != nil {
			return "", errors.New("id must be an integer")
		}

		err = s.DeleteDeadline(id)
		if err != nil {
			return "", err
		}

		return "deadline deleted successfully", nil
	}

	return DEADLINE_HELP, nil
}

func pinHandler(_ []string, _ *store.Store) (string, error) {
	return "pins coming soon", nil
}

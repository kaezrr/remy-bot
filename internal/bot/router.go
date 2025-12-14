package bot

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/kaezrr/remy-bot/internal/store"
	"github.com/rs/zerolog/log"
)

type Response struct {
	Text string
}

const HELP = `Available commands:
.d  deadlines
.b  Manage baskets
.p  Manage pins
.t  Random coin toss
.h  Print this message

Type any command to see its usage`

func Handle(ctx context.Context, input string, prefix string, s store.Store) Response {
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
		result, err := deadlineHandler(ctx, parts[1:], s)
		if err != nil {
			log.Error().Err(err).Msg("deadline handler error")
			return Response{Text: err.Error()}
		}
		return Response{Text: result}

	case "b":
		result, err := basketHandler(ctx, parts[1:], s)
		if err != nil {
			log.Error().Err(err).Msg("basket handler error")
			return Response{Text: err.Error()}
		}
		return Response{Text: result}

	case "p":
		result, err := pinHandler(ctx, parts[1:], s)
		if err != nil {
			log.Error().Err(err).Msg("pin handler error")
			return Response{Text: err.Error()}
		}
		return Response{Text: result}

	case "t":
		randomBool := rand.IntN(2)
		var result string
		if randomBool == 1 {
			result = "heads"
		} else {
			result = "tails"
		}
		return Response{Text: result}

	case "h":
		return Response{Text: HELP}
	}

	return Response{Text: HELP}
}

const BASKET_HELP = `Usage:
.b get             list all baskets
.b add [name]      add a new basket
.b del [name]      remove a basket`

func basketHandler(ctx context.Context, parts []string, s store.Store) (string, error) {
	if len(parts) == 0 {
		return BASKET_HELP, nil
	}

	// Which action to take
	switch parts[0] {
	case "add":
		if len(parts) < 2 {
			return "", errors.New("missing basket name")
		}
		name := parts[1]

		if err := s.AddBasket(ctx, name); err != nil {
			return "", err
		}

		return "basket created successfully", nil

	case "get":
		baskets, err := s.ListBaskets(ctx)

		if err != nil {
			return "", err
		}

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
			return "", errors.New("missing basket name")
		}
		name := parts[1]

		if err := s.DeleteBasket(ctx, name); err != nil {
			return "", err
		}

		return "basket deleted successfully", nil
	}

	return BASKET_HELP, nil
}

const DEADLINE_HELP = `Usage:
.d get                        list all deadlines
.d del [id]                   remove a deadline
.d add [date] [time] [title]  add a new deadline`

func deadlineHandler(ctx context.Context, parts []string, s store.Store) (string, error) {
	if len(parts) == 0 {
		return DEADLINE_HELP, nil
	}

	switch parts[0] {
	case "get":
		deadlines, err := s.ListDeadlines(ctx)

		if err != nil {
			return "", err
		}

		if len(deadlines) == 0 {
			return "no upcoming deadlines", nil
		}

		tz := s.Timezone()
		out := "upcoming deadlines:\n"
		for _, d := range deadlines {
			localTime := d.DueAt.In(tz).Format(store.DisplayFormat)
			out += fmt.Sprintf("%d. %s (%s)\n", d.ID, d.Title, localTime)
		}

		return out, nil

	case "add":
		if len(parts) < 2 {
			return "missing date, time, and title", nil
		}

		if len(parts) < 3 {
			return "missing time and title", nil
		}

		if len(parts) < 4 {
			return "missing title", nil
		}

		date := parts[1]
		timeStr := parts[2]

		inputTimeStr := date + " " + timeStr

		const InputFormat = "2006-01-02 15:04"

		tz := s.Timezone()

		localDeadlineTime, err := time.ParseInLocation(
			InputFormat,
			inputTimeStr,
			tz,
		)
		if err != nil {
			return "", errors.New("invalid date/time format. Use YYYY-MM-DD HH:MM")
		}

		dueAt := localDeadlineTime.UTC()
		title := strings.Join(parts[3:], " ")

		d, err := s.AddDeadline(ctx, title, dueAt)
		if err != nil {
			return "", err
		}

		displayTime := d.DueAt.In(tz).Format(store.DisplayFormat)

		return fmt.Sprintf(
			"deadline #%d added: %s (%s)",
			d.ID,
			d.Title,
			displayTime,
		), nil

	case "del":
		if len(parts) < 2 {
			return "", errors.New("missing deadline id")
		}
		idStr := parts[1]
		id, err := strconv.Atoi(idStr)

		if err != nil {
			return "", errors.New("id must be an integer")
		}

		if err = s.DeleteDeadline(ctx, id); err != nil {
			return "", err
		}

		return fmt.Sprintf("deadline #%d deleted successfully", id), nil
	}

	return DEADLINE_HELP, nil
}

const PIN_HELP = `Usage:
.p get [basket]            list all pins in a basket
.p add [basket] [content]  add a new pin
.p del [id]                remove a pin from a basket`

func pinHandler(ctx context.Context, parts []string, s store.Store) (string, error) {
	if len(parts) == 0 {
		return PIN_HELP, nil
	}

	switch parts[0] {
	case "get":
		if len(parts) < 2 {
			return "", errors.New("missing basket name")
		}

		name := parts[1]
		pins, err := s.ListPins(ctx, name)

		if err != nil {
			return "", err
		}

		if len(pins) == 0 {
			return "no pins in basket " + name, nil
		}

		out := name + " pins:\n"
		for _, p := range pins {
			out += fmt.Sprintf("%d. %s\n", p.ID, p.Content)
		}

		return out, nil

	case "add":
		if len(parts) < 2 {
			return "", errors.New("missing basket name")
		}

		if len(parts) < 3 {
			return "", errors.New("missing pin content")
		}

		name := parts[1]
		content := strings.Join(parts[2:], " ")
		pin, err := s.AddPin(ctx, name, content)

		if err != nil {
			return "", err
		}

		return fmt.Sprintf("pin #%d added to %s", pin.ID, name), nil

	case "del":
		if len(parts) < 2 {
			return "", errors.New("missing pin id")
		}

		idStr := parts[1]
		id, err := strconv.Atoi(idStr)

		if err != nil {
			return "", errors.New("id must be an integer")
		}

		if err = s.DeletePin(ctx, id); err != nil {
			return "", err
		}

		return fmt.Sprintf("pin #%d successfully deleted", id), nil
	}

	return PIN_HELP, nil
}

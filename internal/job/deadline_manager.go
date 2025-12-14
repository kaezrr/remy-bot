package job

import (
	"context"
	"fmt"
	"time"

	"github.com/kaezrr/remy-bot/internal/store"
	"github.com/rs/zerolog/log"

	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	waTypes "go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type DeadlineManager struct {
	Client    *whatsmeow.Client
	Store     store.Store
	TargetJID waTypes.JID
}

func sendGroupMessage(client *whatsmeow.Client, jid waTypes.JID, text string) {
	waMsg := &waE2E.Message{
		Conversation: proto.String(text),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.SendMessage(ctx, jid, waMsg); err != nil {
		log.Error().Err(err).Str("jid", jid.String()).Msg("Job: failed to send message")
	}
}

func (dm *DeadlineManager) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	log.Info().Msg("Job: Deadline manager started, checking every 10 minutes.")

	dm.runChecks(ctx) // Run checks immediately upon startup

	for {
		select {
		case <-ticker.C:
			log.Info().Msg("Job: Deadline manager checking now.")
			dm.runChecks(ctx)
		case <-ctx.Done():
			log.Info().Msg("Job: Deadline manager shutting down.")
			return
		}
	}
}

func (dm *DeadlineManager) runChecks(ctx context.Context) {
	now := time.Now().UTC()

	deadlines, err := dm.Store.ListDueDeadlines(ctx, now)
	if err != nil {
		log.Error().Err(err).Msg("Job: failed to fetch due deadlines")
		return
	}

	for _, d := range deadlines {
		if d.NextRemindIndex == -1 {
			log.Info().
				Int("id", d.ID).
				Str("title", d.Title).
				Msg("Job: deadline expired")

			msg := fmt.Sprintf(
				"*DEADLINE EXPIRED*\n*%s*\nIt was due at: %s",
				d.Title,
				d.DueAt.In(dm.Store.Timezone()).Format(store.DisplayFormat),
			)
			sendGroupMessage(dm.Client, dm.TargetJID, msg)

			if err := dm.Store.DeleteDeadline(ctx, d.ID); err != nil {
				log.Error().
					Err(err).
					Int("id", d.ID).
					Msg("Job: failed to delete expired deadline")
			}
			continue
		}

		remaining := max(d.DueAt.Sub(now), 0)

		log.Info().
			Int("id", d.ID).
			Str("title", d.Title).
			Int("index", d.NextRemindIndex).
			Dur("remaining", remaining).
			Msg("Job: sending reminder")

		msg := fmt.Sprintf(
			"*REMINDER*\n*%s*\nIt is due in *%s*",
			d.Title,
			formatDuration(remaining),
		)
		sendGroupMessage(dm.Client, dm.TargetJID, msg)

		// Schedule next event
		nextIndex := d.NextRemindIndex - 1

		if nextIndex < 0 {
			if err := dm.Store.UpdateNextReminder(ctx, d.ID, d.DueAt, -1); err != nil {
				log.Error().
					Err(err).
					Int("id", d.ID).
					Msg("Job: failed to schedule expiration")
			}
		} else {
			nextTime := d.DueAt.Add(-store.ReminderSchedule[nextIndex])
			if err := dm.Store.UpdateNextReminder(ctx, d.ID, nextTime, nextIndex); err != nil {
				log.Error().
					Err(err).
					Int("id", d.ID).
					Msg("Job: failed to schedule next reminder")
			}
		}
	}

	log.Info().Msg("Job: reminder cycle finished")
}

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "now"
	}

	d = d.Round(time.Minute)

	days := d / (24 * time.Hour)
	if days >= 1 {
		daysInt := int(days)
		return fmt.Sprintf("%d day%s", daysInt, pluralize(daysInt))
	}

	hours := d / time.Hour
	if hours >= 1 {
		hoursInt := int(hours)
		return fmt.Sprintf("%d hour%s", hoursInt, pluralize(hoursInt))
	}

	minutes := d / time.Minute
	if minutes >= 1 {
		minutesInt := int(minutes)
		return fmt.Sprintf("%d minute%s", minutesInt, pluralize(minutesInt))
	}

	return "less than a minute"
}

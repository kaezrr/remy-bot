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

// ReminderIntervals defines the schedule (time BEFORE the deadline) for reminders.
var ReminderIntervals = []time.Duration{
	48 * time.Hour, // Index 0: 2 days before
	24 * time.Hour, // Index 1: 1 day before
	12 * time.Hour, // Index 2: 12 hours before
	6 * time.Hour,  // Index 3: 6 hours before
	3 * time.Hour,  // Index 4: 3 hours before
}

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
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	log.Info().Msg("Job: Deadline manager started, checking every 30 minutes.")

	dm.runChecks(ctx) // Run checks immediately upon startup

	for {
		select {
		case <-ticker.C:
			dm.runChecks(ctx)
		case <-ctx.Done():
			log.Info().Msg("Job: Deadline manager shutting down.")
			return
		}
	}
}

func (dm *DeadlineManager) runChecks(ctx context.Context) {
	deletedDeadlines, err := dm.Store.DeleteExpiredDeadlines(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to clean up expired deadlines")
	}

	for _, d := range deletedDeadlines {
		message := fmt.Sprintf(
			"*Deadline Completed/Expired!*\nTask: *%s*\nWas due on: %s\n",
			d.Title,
			d.DueAt.Format(store.DisplayFormat),
		)
		sendGroupMessage(dm.Client, dm.TargetJID, message)
	}

	deadlines, err := dm.Store.GetAllActiveDeadlines(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to find active deadlines")
		return
	}

	now := time.Now()
	for _, d := range deadlines {
		deadlineTime := d.DueAt

		for i, interval := range ReminderIntervals {
			reminderTime := deadlineTime.Add(-interval)

			if reminderTime.Before(now) && i >= d.ReminderCount {

				timeUntil := deadlineTime.Sub(now)
				formattedTimeUntil := formatDuration(timeUntil)

				message := fmt.Sprintf(
					"*Deadline Alert!*\nTask: *%s*\nThis is due in *%s*!",
					d.Title,
					formattedTimeUntil,
				)

				sendGroupMessage(dm.Client, dm.TargetJID, message)

				nextReminderCount := i + 1

				if err := dm.Store.UpdateReminderState(ctx, d.ID, nextReminderCount); err != nil {
					log.Error().Err(err).Int("id", d.ID).Msg("failed to update reminder state")
				}

				break
			}
		}
	}
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

package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kaezrr/remy-bot/internal/store"
	"github.com/rs/zerolog/log"

	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	waTypes "go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

const (
	Reminder48h = 48 * time.Hour
	Reminder24h = 24 * time.Hour
	Reminder12h = 12 * time.Hour
	Reminder6h  = 6 * time.Hour
	Reminder3h  = 3 * time.Hour
)

const (
	Flag48h = 1 << iota // 1 (Binary 00001)
	Flag24h             // 2 (Binary 00010)
	Flag12h             // 4 (Binary 00100)
	Flag6h              // 8 (Binary 01000)
	Flag3h              // 16 (Binary 10000)
)

var reminderMap = []struct {
	Duration time.Duration
	Flag     int
}{
	{Duration: Reminder48h, Flag: Flag48h},
	{Duration: Reminder24h, Flag: Flag24h},
	{Duration: Reminder12h, Flag: Flag12h},
	{Duration: Reminder6h, Flag: Flag6h},
	{Duration: Reminder3h, Flag: Flag3h},
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
	deadlines, err := dm.Store.ListDeadlines(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Job: Failed to list deadlines for checks")
		return
	}

	now := time.Now().UTC()
	var wg sync.WaitGroup

	for _, d := range deadlines {
		timeRemaining := d.DueAt.Sub(now)

		if timeRemaining <= 0 {
			wg.Add(1)
			go func(deadline store.Deadline) {
				defer wg.Done()

				if timeRemaining > -(10 * time.Minute) {
					log.Info().Int("id", deadline.ID).Str("title", deadline.Title).Msg("Job: Triggering deadline expiry notification")

					msg := fmt.Sprintf(
						"*DEADLINE EXPIRED*\nThe deadline for *%s* has officially passed!\nIt was due at %s",
						deadline.Title,
						deadline.DueAt.In(dm.Store.Timezone()).Format(store.DisplayFormat),
					)
					sendGroupMessage(dm.Client, dm.TargetJID, msg)
				} else {
					log.Warn().Int("id", deadline.ID).Str("title", deadline.Title).Dur("passed", timeRemaining).Msg("Job: Deadline expired long ago, skipping notification.")
				}

				log.Info().Int("id", deadline.ID).Str("title", deadline.Title).Msg("Job: Deleting expired deadline")
				if err := dm.Store.DeleteDeadline(ctx, deadline.ID); err != nil {
					log.Error().Err(err).Int("id", deadline.ID).Msg("Job: Failed to delete expired deadline")
				}
			}(d)
			continue
		}

		for _, rm := range reminderMap {
			buffer := 10 * time.Minute

			if timeRemaining <= rm.Duration+buffer && timeRemaining > rm.Duration-buffer {

				if (d.ReminderCount & rm.Flag) == 0 {

					log.Info().Int("id", d.ID).Str("title", d.Title).Dur("remaining", timeRemaining).Msg("Job: Triggering deadline reminder")

					wg.Add(1)
					go func(deadline store.Deadline, flag int) {
						defer wg.Done()

						// 1. Send Message
						msg := fmt.Sprintf(
							"*REMINDER* \nYour deadline for *%s* is approaching!\n\nTime remaining: *%s*",
							deadline.Title,
							formatDuration(rm.Duration),
						)
						sendGroupMessage(dm.Client, dm.TargetJID, msg)

						newCount := deadline.ReminderCount | flag

						if err := dm.Store.UpdateReminderState(ctx, deadline.ID, newCount); err != nil {
							log.Error().Err(err).Int("id", deadline.ID).Msg("Job: Failed to update ReminderState after sending reminder")
						}
					}(d, rm.Flag)
				}
			}
		}
	}

	wg.Wait()
	log.Info().Msg("Job: Deadline check cycle finished.")
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

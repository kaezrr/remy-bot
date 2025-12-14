package store

import (
	"context"
	"errors"
	"time"
)

var ReminderSchedule = []time.Duration{
	3 * time.Hour,
	6 * time.Hour,
	12 * time.Hour,
	24 * time.Hour,
	48 * time.Hour,
}

func computeInitialReminder(dueAt, now time.Time) (time.Time, int) {
	for i := len(ReminderSchedule) - 1; i >= 0; i-- {
		t := dueAt.Add(-ReminderSchedule[i])
		if t.After(now) {
			return t, i
		}
	}
	// No reminders left → next event is expiration
	return dueAt, -1
}

func (dbs *DBStore) AddDeadline(ctx context.Context, title string, dueAt time.Time) (Deadline, error) {
	now := time.Now().UTC()
	dueAt = dueAt.UTC()
	nextReminder, nextIndex := computeInitialReminder(dueAt, now)

	const query = `
		INSERT INTO deadlines (
			title,
			due_at,
			next_reminder,
			next_remind_index
		)
		VALUES (?, ?, ?, ?);
	`

	res, err := dbs.db.ExecContext(
		ctx,
		query,
		title,
		dueAt.Format(time.RFC3339),
		nextReminder.Format(time.RFC3339),
		nextIndex,
	)
	if err != nil {
		return Deadline{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Deadline{}, err
	}

	d := Deadline{
		ID:              int(id),
		Title:           title,
		DueAt:           dueAt,
		NextReminder:    nextReminder,
		NextRemindIndex: nextIndex,
	}

	return d, nil
}

func (dbs *DBStore) ListDeadlines(ctx context.Context) ([]Deadline, error) {
	const query = `
		SELECT id, title, due_at, next_reminder, next_remind_index FROM deadlines
		ORDER BY due_at ASC;`

	rows, err := dbs.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deadlines := []Deadline{}

	for rows.Next() {
		var (
			d               Deadline
			dueAtStr        string
			nextReminderStr string
		)

		if err := rows.Scan(
			&d.ID,
			&d.Title,
			&dueAtStr,
			&nextReminderStr,
			&d.NextRemindIndex,
		); err != nil {
			return nil, err
		}

		d.DueAt, err = time.Parse(time.RFC3339, dueAtStr)
		if err != nil {
			return nil, err
		}

		d.NextReminder, err = time.Parse(time.RFC3339, nextReminderStr)
		if err != nil {
			return nil, err
		}

		deadlines = append(deadlines, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return deadlines, nil
}

func (dbs *DBStore) DeleteDeadline(ctx context.Context, id int) error {
	const query = `
		DELETE FROM deadlines
		WHERE id = ?;`

	res, err := dbs.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("deadline does not exist")
	}

	return nil
}

func (dbs *DBStore) ListDueDeadlines(ctx context.Context, now time.Time) ([]Deadline, error) {
	const query = `
		SELECT id, title, due_at, next_reminder, next_remind_index FROM deadlines
		WHERE next_reminder <= ?
		ORDER BY next_reminder ASC;`

	rows, err := dbs.db.QueryContext(ctx, query, now.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deadlines := []Deadline{}

	for rows.Next() {
		var (
			d               Deadline
			dueAtStr        string
			nextReminderStr string
		)

		if err := rows.Scan(&d.ID, &d.Title, &dueAtStr, &nextReminderStr, &d.NextRemindIndex); err != nil {
			return nil, err
		}

		d.DueAt, err = time.Parse(time.RFC3339, dueAtStr)
		if err != nil {
			return nil, err
		}

		d.NextReminder, err = time.Parse(time.RFC3339, nextReminderStr)
		if err != nil {
			return nil, err
		}

		deadlines = append(deadlines, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return deadlines, nil
}

func (dbs *DBStore) UpdateNextReminder(ctx context.Context, id int, nextTime time.Time, nextIndex int) error {
	const query = `
		UPDATE deadlines
		SET
			next_reminder = ?,
			next_remind_index = ?
		WHERE id = ?; `

	res, err := dbs.db.ExecContext(ctx, query, nextTime.UTC().Format(time.RFC3339), nextIndex, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("deadline not found")
	}

	return nil
}

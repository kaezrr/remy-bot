package store

import (
	"context"
	"errors"
	"time"
)

func (dbs *DBStore) AddDeadline(ctx context.Context, title string, dueAt time.Time) (Deadline, error) {
	dueAt = dueAt.UTC()
	const query = `
		INSERT INTO deadlines (title, due_at, reminder_count)
		VALUES (?, ?, 0);`

	res, err := dbs.db.ExecContext(ctx, query, title, dueAt.Format(time.RFC3339))
	if err != nil {
		return Deadline{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Deadline{}, err
	}

	d := Deadline{
		ID:            int(id),
		Title:         title,
		DueAt:         dueAt,
		ReminderCount: 0, // Set in the struct
	}

	return d, nil
}

func (dbs *DBStore) ListDeadlines(ctx context.Context) ([]Deadline, error) {
	const query = `
		SELECT id, title, due_at, reminder_count
		FROM deadlines
		ORDER BY due_at ASC;`

	rows, err := dbs.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deadlines := []Deadline{}

	for rows.Next() {
		var d Deadline
		var dueAtString string

		if err := rows.Scan(&d.ID, &d.Title, &dueAtString, &d.ReminderCount); err != nil {
			return nil, err
		}

		d.DueAt, err = time.Parse(time.RFC3339, dueAtString)
		if err != nil {
			return nil, err // Handle parsing error
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

func (dbs *DBStore) UpdateReminderState(ctx context.Context, id int, newCount int) error {
	const query = `
		UPDATE deadlines
		SET reminder_count = ?
		WHERE id = ?;`

	res, err := dbs.db.ExecContext(ctx, query, newCount, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("deadline not found for update")
	}

	return nil
}

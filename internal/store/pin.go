package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

func (dbs *DBStore) AddPin(ctx context.Context, basketName string, content string) (Pin, error) {
	const query1 = `SELECT id FROM baskets WHERE name = ?`
	const query2 = `
		INSERT INTO pins (content, basket_id)
		VALUES (?, ?);
	`

	var basketID int
	err := dbs.db.QueryRow(query1, strings.ToLower(basketName)).Scan(&basketID)
	if err == sql.ErrNoRows {
		return Pin{}, errors.New("basket does not exist")
	}
	if err != nil {
		return Pin{}, err
	}

	res, err := dbs.db.ExecContext(ctx, query2, content, basketID)
	if err != nil {
		return Pin{}, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return Pin{}, err
	}

	p := Pin{
		ID:      int(lastID),
		Content: content,
	}

	return p, nil
}
func (dbs *DBStore) ListPins(ctx context.Context, basketName string) ([]Pin, error) {
	const query1 = "SELECT id FROM baskets WHERE name = ?"
	const query2 = "SELECT id, content FROM pins WHERE basket_id = ? ORDER BY id ASC"

	var basketID int
	err := dbs.db.QueryRow(query1, basketName).Scan(&basketID)
	if err == sql.ErrNoRows {
		return nil, errors.New("basket does not exist")
	}
	if err != nil {
		return nil, err
	}

	rows, err := dbs.db.QueryContext(ctx, query2, basketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pins := []Pin{}

	for rows.Next() {
		var p Pin

		if err := rows.Scan(&p.ID, &p.Content); err != nil {
			return nil, err
		}

		pins = append(pins, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pins, nil
}

func (dbs *DBStore) DeletePin(ctx context.Context, id int) error {
	const query = `
		DELETE FROM pins 
		WHERE id = ?;
	`
	res, err := dbs.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("pin does not exist")
	}

	return nil
}

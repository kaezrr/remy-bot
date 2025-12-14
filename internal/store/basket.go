package store

import (
	"context"
	"errors"
	"strings"
)

func (dbs *DBStore) AddBasket(ctx context.Context, name string) error {
	const query = `
		INSERT INTO baskets (name)
		VALUES (?);`

	_, err := dbs.db.ExecContext(ctx, query, strings.ToLower(name))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errors.New("basket already exists")
		}
		return err
	}

	return nil
}

func (dbs *DBStore) ListBaskets(ctx context.Context) ([]string, error) {
	const query = `
		SELECT name FROM baskets
		ORDER BY name ASC;`

	rows, err := dbs.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	baskets := []string{}

	for rows.Next() {
		var s string

		if err := rows.Scan(&s); err != nil {
			return nil, err
		}

		baskets = append(baskets, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return baskets, nil
}

func (dbs *DBStore) DeleteBasket(ctx context.Context, name string) error {
	const query = `
		DELETE FROM baskets
		WHERE name = ?;`

	res, err := dbs.db.ExecContext(ctx, query, strings.ToLower(name))
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("basket does not exist")
	}

	return nil
}

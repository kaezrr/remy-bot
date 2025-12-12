package store

import (
	"database/sql"
	"errors"
	"strings"

	_ "modernc.org/sqlite"
)

type DBStore struct {
	db *sql.DB
}

const CREATE_DEADLINES_TABLE = `
CREATE TABLE IF NOT EXISTS deadlines(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	datetime TEXT NOT NULL
);`

const CREATE_BASKETS_TABLE = `
CREATE TABLE IF NOT EXISTS baskets(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT UNIQUE NOT NULL
);`

const CREATE_PINS_TABLE = `
CREATE TABLE IF NOT EXISTS pins(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	content TEXT NOT NULL,
	basket_id INTEGER NOT NULL,
	FOREIGN KEY(basket_id) REFERENCES baskets(id) 
		ON DELETE CASCADE
);`

func mustExec(db *sql.DB, query string) {
	if _, err := db.Exec(query); err != nil {
		panic(err)
	}
}

func NewDBStore(path string) (*DBStore, error) {
	db, err := sql.Open("sqlite", path)

	if err != nil {
		return nil, err
	}

	mustExec(db, "PRAGMA journal_mode = WAL;")
	mustExec(db, "PRAGMA foreign_keys = ON;")
	mustExec(db, CREATE_DEADLINES_TABLE)
	mustExec(db, CREATE_BASKETS_TABLE)
	mustExec(db, CREATE_PINS_TABLE)

	return &DBStore{db: db}, nil
}

func (dbs *DBStore) AddDeadline(title string, datetime string) (Deadline, error) {
	const query = `
		INSERT INTO deadlines (title, datetime)
		VALUES (?, ?);`

	res, err := dbs.db.Exec(query, title, datetime)
	if err != nil {
		return Deadline{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Deadline{}, err
	}

	d := Deadline{
		ID:       int(id),
		Title:    title,
		DateTime: datetime,
	}

	return d, nil
}

func (dbs *DBStore) ListDeadlines() ([]Deadline, error) {
	const query = `
		SELECT id, title, datetime
		FROM deadlines
		ORDER BY datetime ASC;`

	rows, err := dbs.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deadlines := []Deadline{}

	for rows.Next() {
		var d Deadline

		if err := rows.Scan(&d.ID, &d.Title, &d.DateTime); err != nil {
			return nil, err
		}

		deadlines = append(deadlines, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return deadlines, nil
}

func (dbs *DBStore) DeleteDeadline(id int) error {
	const query = `
		DELETE FROM deadlines
		WHERE id = ?;`

	res, err := dbs.db.Exec(query, id)
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

func (dbs *DBStore) AddBasket(name string) error {
	const query = `
		INSERT INTO baskets (name)
		VALUES (?);`

	_, err := dbs.db.Exec(query, strings.ToLower(name))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errors.New("basket already exists")
		}
		return err
	}

	return nil
}

func (dbs *DBStore) ListBaskets() ([]string, error) {
	const query = `
		SELECT name FROM baskets
		ORDER BY name ASC;`

	rows, err := dbs.db.Query(query)
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

func (dbs *DBStore) DeleteBasket(name string) error {
	const query = `
		DELETE FROM baskets
		WHERE name = ?;`

	res, err := dbs.db.Exec(query, strings.ToLower(name))
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

func (dbs *DBStore) AddPin(basketName string, content string) (Pin, error) {
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

	res, err := dbs.db.Exec(query2, content, basketID)
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
func (dbs *DBStore) ListPins(basketName string) ([]Pin, error) {
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

	rows, err := dbs.db.Query(query2, basketID)
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

func (dbs *DBStore) DeletePin(basketName string, id int) error {
	const query1 = `SELECT id FROM baskets WHERE name = ?`
	const query2 = `
		DELETE FROM pins 
		WHERE basket_id = ? AND id = ?;
	`
	var basketID int
	err := dbs.db.QueryRow(query1, strings.ToLower(basketName)).Scan(&basketID)
	if err == sql.ErrNoRows {
		return errors.New("basket does not exist")
	}
	if err != nil {
		return err
	}

	res, err := dbs.db.Exec(query2, basketID, id)
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

var _ Store = (*DBStore)(nil)

package store

import (
	"database/sql"
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

func (d *DBStore) AddDeadline(title string, datetime string) (Deadline, error)

// func (d *DBStore) ListDeadlines() ([]Deadline, error)
// func (d *DBStore) DeleteDeadline(id int) error
//
// func (d *DBStore) AddBasket(name string) error
// func (d *DBStore) ListBaskets() ([]string, error)
// func (d *DBStore) DeleteBasket(name string) error
//
// func (d *DBStore) AddPin(basketName string, content string) (Pin, error)
// func (d *DBStore) ListPins(basketName string) ([]Pin, error)
// func (d *DBStore) DeletePin(basketName string, id int) error

var _ Store = (*DBStore)(nil)

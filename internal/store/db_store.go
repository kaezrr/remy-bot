package store

import (
	"database/sql"
	"time"

	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

type DBStore struct {
	db       *sql.DB
	timezone *time.Location
}

const CREATE_DEADLINES_TABLE = `
CREATE TABLE IF NOT EXISTS deadlines(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	due_at TEXT NOT NULL,
	reminder_count INTEGER DEFAULT 0 NOT NULL
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

func NewDBStore(path string, timezone *time.Location) (*DBStore, error) {
	db, err := sql.Open("sqlite", path)

	if err != nil {
		return nil, err
	}

	mustExec(db, "PRAGMA journal_mode = WAL;")
	mustExec(db, "PRAGMA foreign_keys = ON;")
	mustExec(db, CREATE_DEADLINES_TABLE)
	mustExec(db, CREATE_BASKETS_TABLE)
	mustExec(db, CREATE_PINS_TABLE)

	log.Info().Msg("connected to SQLite database!")

	return &DBStore{db: db, timezone: timezone}, nil
}

func (dbs *DBStore) Timezone() *time.Location {
	return dbs.timezone
}

var _ Store = (*DBStore)(nil)

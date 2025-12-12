package store

import (
	"context"
	"time"
)

type Deadline struct {
	ID            int
	Title         string
	DateTime      string
	ReminderCount int
}

const TimeStorageFormat = "2006-01-02T15:04:05Z"
const DisplayFormat = "Mon, Jan 2 at 3:04 PM"

func (d *Deadline) Time() time.Time {
	t, _ := time.Parse(TimeStorageFormat, d.DateTime)
	return t.Local()
}

type Pin struct {
	ID      int
	Content string
}

type Basket struct {
	Name      string
	Pins      []Pin
	NextPinID int
}

type Store interface {
	AddDeadline(ctx context.Context, title string, datetime string) (Deadline, error)
	ListDeadlines(ctx context.Context) ([]Deadline, error)
	DeleteDeadline(ctx context.Context, id int) error

	DeleteExpiredDeadlines(ctx context.Context) ([]*Deadline, error)
	GetAllActiveDeadlines(ctx context.Context) ([]*Deadline, error)
	UpdateReminderState(ctx context.Context, id int, newCount int) error

	AddBasket(ctx context.Context, name string) error
	ListBaskets(ctx context.Context) ([]string, error)
	DeleteBasket(ctx context.Context, name string) error

	AddPin(ctx context.Context, basketName string, content string) (Pin, error)
	ListPins(ctx context.Context, basketName string) ([]Pin, error)
	DeletePin(ctx context.Context, basketName string, id int) error
}

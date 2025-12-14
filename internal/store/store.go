package store

import (
	"context"
	"time"
)

const DisplayFormat = "Mon, Jan 2 at 3:04 PM"

type Deadline struct {
	ID            int
	Title         string
	DueAt         time.Time
	ReminderCount int
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
	AddDeadline(ctx context.Context, title string, duaAt time.Time) (Deadline, error)
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
	DeletePin(ctx context.Context, id int) error

	Timezone() *time.Location
}

package store

import (
	"errors"
	"sort"
	"strings"
	"time"
)

type Deadline struct {
	ID       int
	Title    string
	DateTime string
}

func (d Deadline) Time() time.Time {
	dt, _ := time.Parse("2006-01-02 15:04", d.DateTime)
	return dt
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

type MemStore struct {
	deadlines      []Deadline
	nextDeadlineID int
	baskets        map[string]*Basket
}

func NewMemStore() *MemStore {
	return &MemStore{
		deadlines:      []Deadline{},
		nextDeadlineID: 1,
		baskets:        map[string]*Basket{},
	}
}

func (s *MemStore) AddDeadline(title string, datetime string) (Deadline, error) {
	deadline := Deadline{
		ID:       s.nextDeadlineID,
		Title:    title,
		DateTime: datetime,
	}

	s.nextDeadlineID += 1
	s.deadlines = append(s.deadlines, deadline)

	sort.Slice(s.deadlines, func(i, j int) bool {
		return s.deadlines[i].Time().Before(s.deadlines[j].Time())
	})

	return deadline, nil
}

func (s *MemStore) ListDeadlines() ([]Deadline, error) {
	return s.deadlines, nil
}

func (s *MemStore) DeleteDeadline(id int) error {
	for i, d := range s.deadlines {
		if d.ID != id {
			continue
		}
		s.deadlines = append(s.deadlines[:i], s.deadlines[i+1:]...)
		return nil
	}

	return errors.New("deadline does not exist")
}

func (s *MemStore) AddBasket(name string) error {
	name = strings.ToLower(name)

	if _, ok := s.baskets[name]; ok {
		return errors.New("basket already exists")
	}

	s.baskets[name] = &Basket{
		Name:      name,
		NextPinID: 1,
		Pins:      []Pin{},
	}

	return nil
}

func (s *MemStore) ListBaskets() ([]string, error) {
	baskets := make([]string, 0)

	for _, b := range s.baskets {
		baskets = append(baskets, b.Name)
	}

	return baskets, nil
}

func (s *MemStore) DeleteBasket(name string) error {
	name = strings.ToLower(name)

	if _, ok := s.baskets[name]; !ok {
		return errors.New("basket does not exist")
	}

	delete(s.baskets, name)
	return nil
}

func (s *MemStore) AddPin(basketName string, content string) (Pin, error) {
	basketName = strings.ToLower(basketName)
	basket, ok := s.baskets[basketName]

	if !ok {
		return Pin{}, errors.New("basket does not exist")
	}

	pin := Pin{
		ID:      basket.NextPinID,
		Content: content,
	}

	basket.NextPinID += 1
	basket.Pins = append(basket.Pins, pin)
	return pin, nil
}

func (s *MemStore) ListPins(basketName string) ([]Pin, error) {
	basketName = strings.ToLower(basketName)
	basket, ok := s.baskets[basketName]

	if !ok {
		return nil, errors.New("basket does not exist")
	}

	return basket.Pins, nil
}

func (s *MemStore) DeletePin(basketName string, id int) error {
	basketName = strings.ToLower(basketName)
	basket, ok := s.baskets[basketName]

	if !ok {
		return errors.New("basket does not exist")
	}

	for i, p := range basket.Pins {
		if p.ID != id {
			continue
		}
		basket.Pins = append(basket.Pins[:i], basket.Pins[i+1:]...)
		return nil
	}

	return errors.New("pin does not exist")
}

var _ Store = (*MemStore)(nil)

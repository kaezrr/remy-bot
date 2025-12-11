package store

import (
	"errors"
	"strings"
)

type Deadline struct {
	ID       int
	Title    string
	DateTime string
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

type Store struct {
	deadlines      []Deadline
	nextDeadlineID int
	baskets        map[string]*Basket
}

func New() *Store {
	return &Store{
		deadlines:      []Deadline{},
		nextDeadlineID: 1,
		baskets:        map[string]*Basket{},
	}
}

// TODO: sort deadlines by datetime
func (s *Store) AddDeadline(title string, datetime string) Deadline {
	deadline := Deadline{
		ID:       s.nextDeadlineID,
		Title:    title,
		DateTime: datetime,
	}

	s.nextDeadlineID += 1
	s.deadlines = append(s.deadlines, deadline)
	return deadline
}

func (s *Store) ListDeadlines() []Deadline {
	return s.deadlines
}

func (s *Store) DeleteDeadline(id int) error {
	for i, d := range s.deadlines {
		if d.ID != id {
			continue
		}
		s.deadlines = append(s.deadlines[:i], s.deadlines[i+1:]...)
		return nil
	}

	return errors.New("deadline does not exist")
}

func (s *Store) AddBasket(name string) error {
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

func (s *Store) ListBaskets() []string {
	baskets := make([]string, 0)

	for _, b := range s.baskets {
		baskets = append(baskets, b.Name)
	}

	return baskets
}

func (s *Store) DeleteBasket(name string) error {
	name = strings.ToLower(name)

	if _, ok := s.baskets[name]; !ok {
		return errors.New("basket does not exist")
	}

	delete(s.baskets, name)
	return nil
}

func (s *Store) AddPin(basketName string, content string) (Pin, error) {
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

func (s *Store) ListPins(basketName string) ([]Pin, error) {
	basketName = strings.ToLower(basketName)
	basket, ok := s.baskets[basketName]

	if !ok {
		return nil, errors.New("basket does not exist")
	}

	return basket.Pins, nil
}

func (s *Store) DeletePin(basketName string, id int) error {
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
